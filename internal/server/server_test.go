package server

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/pkg/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
)

const (
	testJWTSecret = "test-secret-for-campuscare-route-tests"
	testJWTIssuer = "campuscare-test"
	testUserID    = "11111111-1111-1111-1111-111111111111"
)

func testConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.JWT.Secret = testJWTSecret
	cfg.JWT.Issuer = testJWTIssuer
	cfg.Env = "test"
	cfg.Server.Port = "0"
	cfg.Server.CORSOrigins = []string{"https://campuscare.example"}
	return cfg
}

// newTestServer must not panic: gin panics on duplicate or conflicting routes,
// so building the router here validates the whole registration table.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return New(testConfig(), nil)
}

func doRequest(t *testing.T, srv *Server, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)
	return w
}

func accessToken(t *testing.T, role string) string {
	t.Helper()

	token, err := auth.GenerateAccessToken(testUserID, role, testJWTSecret, auth.TokenOptions{Issuer: testJWTIssuer, AccessTTL: time.Hour})
	if err != nil {
		t.Fatalf("could not build access token: %v", err)
	}
	return token
}

func TestHealthEndpointsArePublic(t *testing.T) {
	srv := newTestServer(t)

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/health/live", "", ""); w.Code != http.StatusOK {
		t.Fatalf("health/live: expected 200, got %d", w.Code)
	}

	// With no database pool configured the readiness probe must report DOWN
	// instead of panicking.
	if w := doRequest(t, srv, http.MethodGet, "/api/v1/health/ready", "", ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("health/ready: expected 503, got %d", w.Code)
	}
}

func TestProtectedRoutesRequireToken(t *testing.T) {
	srv := newTestServer(t)

	protected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/clubs"},
		{http.MethodGet, "/api/v1/events"},
		{http.MethodGet, "/api/v1/opportunities"},
		{http.MethodGet, "/api/v1/community/discussions"},
		{http.MethodPost, "/api/v1/community/discussions"},
		{http.MethodGet, "/api/v1/me/notifications"},
		{http.MethodGet, "/api/v1/community/search?q=ai"},
	}

	for _, tc := range protected {
		if w := doRequest(t, srv, tc.method, tc.path, "", "{}"); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401 without token, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestMalformedTokenIsRejected(t *testing.T) {
	srv := newTestServer(t)

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/events", "not-a-jwt", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed token, got %d", w.Code)
	}
}

// A refresh token must not be usable as an access token, otherwise a stolen
// refresh token would grant API access for its full lifetime.
func TestRefreshTokenIsNotAcceptedAsAccessToken(t *testing.T) {
	srv := newTestServer(t)

	refresh, err := auth.GenerateRefreshToken(testUserID, "MEMBER", testJWTSecret, auth.TokenOptions{Issuer: testJWTIssuer, AccessTTL: time.Hour})
	if err != nil {
		t.Fatalf("could not build refresh token: %v", err)
	}

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/events", refresh, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when using a refresh token as an access token, got %d", w.Code)
	}
}

func TestTokenSignedWithAnotherSecretIsRejected(t *testing.T) {
	srv := newTestServer(t)

	token, err := auth.GenerateAccessToken(testUserID, "MEMBER", "a-completely-different-secret", auth.TokenOptions{Issuer: testJWTIssuer, AccessTTL: time.Hour})
	if err != nil {
		t.Fatalf("could not build token: %v", err)
	}

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/events", token, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a token signed with another secret, got %d", w.Code)
	}
}

func TestRoleGuardBlocksInsufficientRole(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	adminOnly := []string{
		"/api/v1/institutional/analytics",
		"/api/v1/institutional/graph?query=ai",
		"/api/v1/admin/users",
		"/api/v1/admin/approvals/clubs",
	}

	for _, path := range adminOnly {
		if w := doRequest(t, srv, http.MethodGet, path, memberToken, ""); w.Code != http.StatusForbidden {
			t.Errorf("GET %s: expected 403 for member, got %d", path, w.Code)
		}
	}

	organiserOnly := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/organiser/sessions/22222222-2222-2222-2222-222222222222/check-in/open"},
		{http.MethodGet, "/api/v1/organiser/sessions/22222222-2222-2222-2222-222222222222/attendance"},
	}

	for _, tc := range organiserOnly {
		if w := doRequest(t, srv, tc.method, tc.path, memberToken, "{}"); w.Code != http.StatusForbidden {
			t.Errorf("%s %s: expected 403 for member, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestAuthenticatedMemberReachesAllowedRoute(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	// These handlers need a database pool, which is not available here, so the
	// response must be a handled server error rather than a panic or a 401.
	allowed := []string{
		"/api/v1/events",
		"/api/v1/clubs",
		"/api/v1/opportunities",
		"/api/v1/community/discussions",
		"/api/v1/me/notifications",
	}

	for _, path := range allowed {
		w := doRequest(t, srv, http.MethodGet, path, memberToken, "")
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Errorf("GET %s: member was rejected from an allowed route: %d", path, w.Code)
		}
	}
}

func TestSearchRequiresQueryParameter(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	// The handler validates before touching the database, so this is a real
	// assertion rather than a side effect of the missing pool.
	if w := doRequest(t, srv, http.MethodGet, "/api/v1/community/search", memberToken, ""); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a search with no query, got %d", w.Code)
	}
}

// The AI retrieval endpoint is part of the protected surface: campus records
// must not be readable without a session.
func TestAIContextRequiresAuthentication(t *testing.T) {
	srv := newTestServer(t)

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/ai/context?q=cloud", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a token, got %d", w.Code)
	}
}

func TestAIContextRequiresQueryParameter(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	if w := doRequest(t, srv, http.MethodGet, "/api/v1/ai/context", memberToken, ""); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with no query, got %d", w.Code)
	}
	if w := doRequest(t, srv, http.MethodGet, "/api/v1/ai/context?q=%20%20", memberToken, ""); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with a blank query, got %d", w.Code)
	}
}

// Retrieval is a read, not a generation proxy. The route surface must not
// accept a prompt to forward, because inference runs client-side.
func TestAIDoesNotExposeGenerationEndpoint(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	// A POST to the retrieval route must not exist at all.
	if w := doRequest(t, srv, http.MethodPost, "/api/v1/ai/context", memberToken, `{"prompt":"hello"}`); w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for POST /ai/context, got %d", w.Code)
	}

	for _, path := range []string{"/api/v1/ai/chat", "/api/v1/ai/complete", "/api/v1/ai/generate"} {
		if w := doRequest(t, srv, http.MethodPost, path, memberToken, `{"prompt":"hello"}`); w.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, w.Code)
		}
	}
}

func TestDiscussionCreationRejectsInvalidBody(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	cases := []string{
		`{}`,
		`{"title":"ab","body":"x"}`,
		`{"title":"valid title","body":""}`,
		`{"title":"valid title","body":"x","category":"NONSENSE"}`,
	}

	for _, body := range cases {
		w := doRequest(t, srv, http.MethodPost, "/api/v1/community/discussions", memberToken, body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("body %s: expected 400, got %d", body, w.Code)
		}
	}
}

func TestBookingRejectsInvertedWindow(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	// end before start, and in the past: both are rejected before the database.
	inverted := `{"start_time":"2027-01-01T12:00:00Z","end_time":"2027-01-01T11:00:00Z"}`
	if w := doRequest(t, srv, http.MethodPost,
		"/api/v1/community/services/33333333-3333-3333-3333-333333333333/bookings",
		memberToken, inverted); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an inverted booking window, got %d", w.Code)
	}

	past := `{"start_time":"2020-01-01T10:00:00Z","end_time":"2020-01-01T11:00:00Z"}`
	if w := doRequest(t, srv, http.MethodPost,
		"/api/v1/community/services/33333333-3333-3333-3333-333333333333/bookings",
		memberToken, past); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a booking in the past, got %d", w.Code)
	}
}

func TestFeedbackRejectsUnknownActivityType(t *testing.T) {
	srv := newTestServer(t)
	memberToken := accessToken(t, "MEMBER")

	w := doRequest(t, srv, http.MethodPost,
		"/api/v1/feedback/COURSE/44444444-4444-4444-4444-444444444444",
		memberToken, `{"rating":5}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unsupported activity type, got %d", w.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/events", nil)
	req.Header.Set("Origin", "https://campuscare.example")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight: expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://campuscare.example" {
		t.Fatalf("preflight: expected the origin to be echoed, got %q", got)
	}
}

// serveH2C binds the server's handler to a real loopback port and serves it, so
// the protocol negotiation can be exercised over an actual connection rather than
// simulated through httptest. That distinction is the whole point: a test that
// calls Handler.ServeHTTP directly never touches the h2c sniffer, so it would
// pass whether or not HTTP/2 was wired in at all.
func serveH2C(t *testing.T, handler http.Handler) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { _ = srv.Close() })

	return "http://" + listener.Addr().String()
}

// h2cClient dials without TLS, which is what an h2c connection looks like on the
// wire: the HTTP/2 preface over a plain TCP socket.
func h2cClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: 10 * time.Second,
	}
}

func TestServerSpeaksH2CWhenHTTP2EnabledWithoutTLS(t *testing.T) {
	srv := newTestServer(t)
	base := serveH2C(t, srv.httpServer.Handler)

	resp, err := h2cClient().Get(base + "/api/v1/health/live")
	if err != nil {
		t.Fatalf("h2c request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.ProtoMajor != 2 {
		t.Fatalf("expected the connection to be negotiated as HTTP/2, got %s", resp.Proto)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health/live over h2c: expected 200, got %d", resp.StatusCode)
	}
	// The response must still travel through the whole middleware chain; a
	// protocol upgrade that bypassed the router would return 404 or an empty
	// body rather than the health payload.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), `"status":"UP"`) {
		t.Fatalf("health/live over h2c returned %q, expected the UP payload", body)
	}
}

func TestServerFallsBackToHTTP1WhenHTTP2Disabled(t *testing.T) {
	cfg := testConfig()
	cfg.Server.HTTP2 = false

	gin.SetMode(gin.TestMode)
	srv := New(cfg, nil)
	if srv == nil {
		t.Fatal("New returned nil")
	}

	base := serveH2C(t, srv.httpServer.Handler)

	// A default client has no h2c support, so this is the HTTP/1.1 path a
	// browser would take.
	resp, err := http.Get(base + "/api/v1/health/live")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.ProtoMajor != 1 {
		t.Fatalf("with HTTP/2 disabled the server should answer HTTP/1.1, got %s", resp.Proto)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health/live over HTTP/1.1: expected 200, got %d", resp.StatusCode)
	}
}
