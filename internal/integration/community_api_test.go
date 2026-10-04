//go:build integration

// Package integration exercises the real route graph against a live PostgreSQL
// database. It is excluded from the default build by the `integration` tag so
// unit test runs stay hermetic.
//
//	go test -tags integration ./internal/integration/...
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/internal/server"
	"github.com/gin-gonic/gin"
)

const testPassword = "IntegrationTest!2345"

type apiClient struct {
	t      *testing.T
	base   string
	client *http.Client
	token  string
}

type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

func newAPI(t *testing.T) *apiClient {
	t.Helper()

	if os.Getenv("INTEGRATION_DATABASE_URL") == "" && os.Getenv("DB_NAME") == "" {
		t.Skip("set DB_* or INTEGRATION_DATABASE_URL to run integration tests")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	if pool == nil {
		t.Fatal("postgres pool is nil")
	}

	gin.SetMode(gin.TestMode)
	srv := server.New(cfg, pool)
	ts := httptest.NewServer(srv.Handler())

	t.Cleanup(func() {
		ts.Close()
		pool.Close()
	})

	c := &apiClient{t: t, base: ts.URL + "/api/v1", client: ts.Client()}
	c.expectStatus(http.MethodGet, "/health/ready", nil, "", http.StatusOK)
	return c
}

func (c *apiClient) do(method, path string, body any) (int, envelope) {
	c.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		c.t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		c.t.Fatalf("read response: %v", err)
	}

	var out envelope
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			c.t.Fatalf("%s %s returned non-JSON body %q", method, path, string(raw))
		}
	}
	return res.StatusCode, out
}

// decode unmarshals the envelope payload into out and fails the test when the
// call did not succeed.
func (c *apiClient) decode(method, path string, body any, wantStatus int, out any) {
	c.t.Helper()

	status, env := c.do(method, path, body)
	if status != wantStatus {
		c.t.Fatalf("%s %s: status = %d, want %d (body %q)", method, path, status, wantStatus, env.Error)
	}
	if out != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			c.t.Fatalf("%s %s: decode data: %v (body %q)", method, path, err, string(env.Data))
		}
	}
}

func (c *apiClient) expectStatus(method, path string, body any, token string, wantStatus int) {
	c.t.Helper()

	previous := c.token
	if token != "" {
		c.token = token
	}
	defer func() { c.token = previous }()

	status, env := c.do(method, path, body)
	if status != wantStatus {
		c.t.Fatalf("%s %s: status = %d, want %d (body %q)", method, path, status, wantStatus, env.Error)
	}
}

func (c *apiClient) decodeData(method, path string, body any, out any) {
	c.t.Helper()
	c.decode(method, path, body, http.StatusOK, out)
}

func registerTestUser(t *testing.T, c *apiClient, roleHint string) (email, access, refresh string) {
	t.Helper()

	email = fmt.Sprintf("itest-%d@campus.test", time.Now().UnixNano())
	var created struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}
	c.decode(http.MethodPost, "/auth/register", map[string]string{
		"email":    email,
		"password": testPassword,
	}, http.StatusCreated, &created)

	if created.UserID == "" || created.Role != "MEMBER" {
		t.Fatalf("register returned unexpected payload: %+v", created)
	}

	// Registration deliberately issues no tokens, so the client logs in next.
	var login struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"user"`
	}
	c.decode(http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": testPassword,
	}, http.StatusOK, &login)

	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("login did not return a usable session: %+v", login)
	}
	if login.User.ID != created.UserID {
		t.Fatalf("login user id = %q, register user id = %q", login.User.ID, created.UserID)
	}
	_ = roleHint
	return email, login.AccessToken, login.RefreshToken
}

func TestCommunityPlatformEndToEnd(t *testing.T) {
	c := newAPI(t)

	email, access, refresh := registerTestUser(t, c, "member")
	c.token = access

	t.Run("authentication", func(t *testing.T) {
		var me struct {
			Email       string `json:"email"`
			DisplayName string `json:"display_name"`
			Role        string `json:"role"`
		}
		c.decodeData(http.MethodGet, "/auth/me", nil, &me)
		if me.Email != email {
			t.Fatalf("GET /auth/me email = %q, want %q", me.Email, email)
		}
		if me.Role != "MEMBER" {
			t.Fatalf("new user role = %q, want MEMBER", me.Role)
		}

		var profile struct {
			DisplayName string `json:"display_name"`
		}
		c.decode(http.MethodPatch, "/auth/me", map[string]string{"display_name": "Integration Tester"}, http.StatusOK, &profile)
		if profile.DisplayName != "Integration Tester" {
			t.Fatalf("PATCH /auth/me display_name = %q", profile.DisplayName)
		}

		c.decodeData(http.MethodGet, "/auth/me", nil, &profile)
		if profile.DisplayName != "Integration Tester" {
			t.Fatal("display_name did not persist")
		}

		c.expectStatus(http.MethodPatch, "/auth/me", map[string]string{"display_name": "x"}, "", http.StatusBadRequest)

		loginEmail, loginAccess, loginRefresh := email, "", ""
		var login struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
		c.decode(http.MethodPost, "/auth/login", map[string]string{"email": loginEmail, "password": testPassword}, http.StatusOK, &login)
		loginAccess, loginRefresh = login.AccessToken, login.RefreshToken
		if loginAccess == "" || loginRefresh == "" {
			t.Fatal("login did not return both tokens")
		}

		var refreshed struct {
			AccessToken string `json:"access_token"`
		}
		c.decode(http.MethodPost, "/auth/refresh", map[string]string{"refresh_token": loginRefresh}, http.StatusOK, &refreshed)
		if refreshed.AccessToken == "" {
			t.Fatal("refresh did not return an access token")
		}

		// The refreshed access token must work.
		c.expectStatus(http.MethodGet, "/auth/me", nil, refreshed.AccessToken, http.StatusOK)

		c.expectStatus(http.MethodPost, "/auth/refresh", map[string]string{"refresh_token": "garbage"}, "", http.StatusUnauthorized)
		c.expectStatus(http.MethodGet, "/auth/me", nil, "not-a-token", http.StatusUnauthorized)
		c.expectStatus(http.MethodGet, "/clubs", nil, "", http.StatusUnauthorized)
	})

	t.Run("clubs", func(t *testing.T) {
		var clubs struct {
			Clubs []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"clubs"`
		}
		c.decodeData(http.MethodGet, "/clubs", nil, &clubs)
		if len(clubs.Clubs) == 0 {
			t.Skip("no clubs seeded in this database")
		}

		clubID := clubs.Clubs[0].ID
		c.decodeData(http.MethodGet, "/clubs/"+clubID, nil, nil)

		c.decodeData(http.MethodPost, "/clubs/"+clubID+"/join", nil, nil)

		var mine struct {
			Clubs []struct {
				ID string `json:"id"`
			} `json:"clubs"`
		}
		c.decodeData(http.MethodGet, "/me/clubs", nil, &mine)
		if !containsID(mine.Clubs, clubID) {
			t.Fatalf("joined club %s missing from /me/clubs", clubID)
		}

		// Joining twice is idempotent rather than an error.
		c.decodeData(http.MethodPost, "/clubs/"+clubID+"/join", nil, nil)
		c.decodeData(http.MethodPost, "/clubs/"+clubID+"/leave", nil, nil)

		c.decodeData(http.MethodGet, "/me/clubs", nil, &mine)
		if containsID(mine.Clubs, clubID) {
			t.Fatal("club still present in /me/clubs after leaving")
		}
	})

	t.Run("discussions", func(t *testing.T) {
		var created struct {
			ID string `json:"id"`
		}
		c.decode(http.MethodPost, "/community/discussions", map[string]string{
			"title":    "Integration discussion",
			"body":     "Posted by the integration suite.",
			"category": "GENERAL",
		}, http.StatusCreated, &created)
		if created.ID == "" {
			t.Fatal("discussion id missing from create response")
		}

		c.decode(http.MethodPost, "/community/discussions/"+created.ID+"/replies", map[string]string{
			"body": "First reply from the integration suite.",
		}, http.StatusCreated, nil)

		var thread struct {
			ID      string `json:"id"`
			Replies []struct {
				Body string `json:"body"`
			} `json:"replies"`
		}
		c.decodeData(http.MethodGet, "/community/discussions/"+created.ID, nil, &thread)
		if len(thread.Replies) == 0 {
			t.Fatal("thread has no replies after posting one")
		}

		var list struct {
			Discussions []struct {
				ID string `json:"id"`
			} `json:"discussions"`
		}
		c.decodeData(http.MethodGet, "/community/discussions", nil, &list)
		if len(list.Discussions) == 0 {
			t.Fatal("discussion list is empty")
		}

		c.expectStatus(http.MethodPost, "/community/discussions", map[string]string{"title": "no", "body": ""}, "", http.StatusBadRequest)
		c.expectStatus(http.MethodGet, "/community/discussions/"+created.ID, nil, "", http.StatusOK)
	})

	t.Run("events", func(t *testing.T) {
		var events struct {
			Events []struct {
				ID                   string `json:"id"`
				Status               string `json:"status"`
				RegistrationDeadline string `json:"registration_deadline"`
			} `json:"events"`
		}
		c.decodeData(http.MethodGet, "/events", nil, &events)

		var published string
		for _, e := range events.Events {
			if e.Status == "PUBLISHED" {
				published = e.ID
				break
			}
		}
		if published == "" {
			t.Skip("no published events available to register for")
		}

		var waitlist struct {
			Entries []struct {
				Position int `json:"position"`
			} `json:"entries"`
		}
		c.decodeData(http.MethodGet, "/events/"+published+"/waitlist", nil, &waitlist)

		c.decodeData(http.MethodPost, "/events/"+published+"/cancel", nil, nil)
		c.expectStatus(http.MethodGet, "/events/"+published, nil, "", http.StatusOK)
	})

	t.Run("opportunities", func(t *testing.T) {
		var opps struct {
			Opportunities []struct {
				ID string `json:"id"`
			} `json:"opportunities"`
		}
		c.decodeData(http.MethodGet, "/opportunities", nil, &opps)
		if len(opps.Opportunities) == 0 {
			t.Skip("no opportunities seeded in this database")
		}
		oppID := opps.Opportunities[0].ID

		c.decodeData(http.MethodGet, "/opportunities/"+oppID, nil, nil)
		c.decodeData(http.MethodPost, "/opportunities/"+oppID+"/save", nil, nil)
		c.decodeData(http.MethodDelete, "/opportunities/"+oppID+"/save", nil, nil)

		c.decode(http.MethodPost, "/community/opportunities/"+oppID+"/apply", map[string]string{
			"cover_note": "Applied by the integration suite.",
		}, http.StatusCreated, nil)

		// A second application for the same opportunity must be rejected.
		c.expectStatus(http.MethodPost, "/community/opportunities/"+oppID+"/apply", map[string]string{}, "", http.StatusConflict)

		var mine struct {
			Applications []struct {
				OpportunityID string `json:"opportunity_id"`
			} `json:"applications"`
		}
		c.decodeData(http.MethodGet, "/me/opportunities", nil, &mine)
		if len(mine.Applications) == 0 {
			t.Fatal("application not visible in /me/opportunities")
		}

		c.decodeData(http.MethodDelete, "/community/opportunities/"+oppID+"/application", nil, nil)
	})

	t.Run("services_and_bookings", func(t *testing.T) {
		var services struct {
			Resources []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				IsBookable bool   `json:"is_bookable"`
			} `json:"resources"`
		}
		c.decodeData(http.MethodGet, "/community/services", nil, &services)

		var bookable string
		for _, r := range services.Resources {
			if r.IsBookable {
				bookable = r.ID
				break
			}
		}
		if bookable == "" {
			t.Skip("no bookable resources seeded in this database")
		}

		start := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Second)
		end := start.Add(time.Hour)
		body := map[string]string{
			"start_time": start.Format(time.RFC3339),
			"end_time":   end.Format(time.RFC3339),
			"note":       "integration booking",
		}

		var booked struct {
			ID string `json:"id"`
		}
		c.decode(http.MethodPost, "/community/services/"+bookable+"/bookings", body, http.StatusCreated, &booked)
		if booked.ID == "" {
			t.Fatal("booking id missing from create response")
		}

		var mine struct {
			Bookings []struct {
				ID string `json:"id"`
			} `json:"bookings"`
		}
		c.decodeData(http.MethodGet, "/community/bookings", nil, &mine)
		if !containsID(mine.Bookings, booked.ID) {
			t.Fatal("booking not visible in /community/bookings")
		}

		// The PostgreSQL exclusion constraint must reject an overlapping slot.
		overlap := map[string]string{
			"start_time": start.Add(30 * time.Minute).Format(time.RFC3339),
			"end_time":   start.Add(90 * time.Minute).Format(time.RFC3339),
		}
		c.expectStatus(http.MethodPost, "/community/services/"+bookable+"/bookings", overlap, "", http.StatusConflict)

		c.decodeData(http.MethodDelete, "/community/bookings/"+booked.ID, nil, nil)

		pastStart := time.Now().UTC().Add(-48 * time.Hour)
		c.expectStatus(http.MethodPost, "/community/services/"+bookable+"/bookings", map[string]string{
			"start_time": pastStart.Format(time.RFC3339),
			"end_time":   pastStart.Add(time.Hour).Format(time.RFC3339),
		}, "", http.StatusBadRequest)
	})

	t.Run("notifications_and_search", func(t *testing.T) {
		var notes struct {
			Notifications []struct {
				ID string `json:"id"`
			} `json:"notifications"`
			Unread int `json:"unread"`
		}
		c.decodeData(http.MethodGet, "/community/notifications", nil, &notes)
		c.decodeData(http.MethodGet, "/me/notifications", nil, &notes)

		if len(notes.Notifications) > 0 {
			id := notes.Notifications[0].ID
			c.decodeData(http.MethodPost, "/community/notifications/"+id+"/read", nil, nil)
		}
		c.decodeData(http.MethodPost, "/community/notifications/read-all", nil, nil)

		c.expectStatus(http.MethodGet, "/community/search", nil, "", http.StatusBadRequest)
		c.expectStatus(http.MethodGet, "/community/search?q=zzzznomatch", nil, "", http.StatusOK)
	})

	t.Run("role_boundaries", func(t *testing.T) {
		// MEMBER must not reach admin analytics or organiser tooling.
		c.expectStatus(http.MethodGet, "/institutional/analytics", nil, "", http.StatusForbidden)
		c.expectStatus(http.MethodGet, "/institutional/graph", nil, "", http.StatusForbidden)
		c.expectStatus(http.MethodPost, "/organiser/sessions/"+refresh+"/check-in/open", map[string]int{"duration_minutes": 30}, "", http.StatusForbidden)
	})

	_ = refresh
}

func containsID(items []struct {
	ID string `json:"id"`
}, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func TestAdminAnalyticsWithPromotedUser(t *testing.T) {
	c := newAPI(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	pool, err := database.NewPostgresPool(cfg.Database)
	if err != nil || pool == nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer pool.Close()

	email, _, _ := registerTestUser(t, c, "admin")

	ctx := context.Background()
	if _, err := pool.Exec(ctx, "UPDATE users SET role = 'ADMIN' WHERE email = $1", email); err != nil {
		t.Fatalf("promote user to ADMIN: %v", err)
	}

	var login struct {
		AccessToken string `json:"access_token"`
	}
	c.decode(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": testPassword}, http.StatusOK, &login)
	c.token = login.AccessToken

	c.expectStatus(http.MethodGet, "/institutional/analytics", nil, "", http.StatusOK)
	c.expectStatus(http.MethodGet, "/institutional/graph", nil, "", http.StatusOK)
}
