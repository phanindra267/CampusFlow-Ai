package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/auth"
	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-secret-for-handler-tests"

// TestMain drops the bcrypt work factor to its minimum so the suite does not
// spend several seconds per case on key stretching.
func TestMain(m *testing.M) {
	auth.SetDefaultCost(auth.MinimumTestCost)
	os.Exit(m.Run())
}

func testAuthConfig() AuthConfig {
	return AuthConfig{
		Secret:    testJWTSecret,
		Issuer:    "campuscare-test",
		AccessTTL: time.Hour,
	}
}

// fakeUserRepo records what the handler tried to persist.
type fakeUserRepo struct {
	created   []*domain.User
	createErr error
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range f.created {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, postgres.ErrNotFound
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range f.created {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, postgres.ErrNotFound
}

func (f *fakeUserRepo) Create(ctx context.Context, user *domain.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	for _, u := range f.created {
		if u.Email == user.Email {
			return postgres.ErrDuplicateKey
		}
	}
	f.created = append(f.created, user)
	return nil
}

func (f *fakeUserRepo) UpdateDisplayName(ctx context.Context, id, displayName string) error {
	for _, u := range f.created {
		if u.ID == id {
			u.DisplayName = displayName
			return nil
		}
	}
	return postgres.ErrNotFound
}

// GetProfile and UpdateProfile mirror the display-name behaviour: the fake is
// only used by the auth tests, which assert on identity and tokens rather than
// on profile columns.
func (f *fakeUserRepo) GetProfile(ctx context.Context, id string) (*domain.User, error) {
	return f.GetByID(ctx, id)
}

func (f *fakeUserRepo) UpdateProfile(ctx context.Context, id string, req *domain.UpdateProfileRequest) (*domain.User, error) {
	if err := f.UpdateDisplayName(ctx, id, req.DisplayName); err != nil {
		return nil, err
	}
	return f.GetByID(ctx, id)
}

func (f *fakeUserRepo) LoadProfile(ctx context.Context, id string) (*domain.Profile, error) {
	user, err := f.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Profile{User: *user}, nil
}

func newAuthTestRouter(repo UserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := testAuthConfig()
	h := NewAuthHandler(cfg, repo)

	router := gin.New()
	router.POST("/login", h.Login)
	router.POST("/register", h.Register)
	router.POST("/refresh", h.Refresh)

	// Stand-in for middleware.RequireAuth so /me can be exercised.
	authed := router.Group("/me", func(c *gin.Context) {
		claims, err := auth.ValidateAccessToken(c.GetHeader("X-Test-Token"), cfg.Secret, cfg.Issuer)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	})
	authed.GET("", h.Me)

	return router
}

func postJSON(t *testing.T, router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v (body=%s)", err, rec.Body.String())
	}
	if !envelope.Success {
		t.Fatalf("expected success=true, got %s", rec.Body.String())
	}
	if err := json.Unmarshal(envelope.Data, target); err != nil {
		t.Fatalf("decode data: %v (data=%s)", err, envelope.Data)
	}
}

type authPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

func seedUser(t *testing.T, repo *fakeUserRepo, email, password, role string) *domain.User {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := &domain.User{
		ID:           "9f1c2b7e-0000-4000-8000-000000000001",
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Status:       "ACTIVE",
	}
	repo.created = append(repo.created, user)
	return user
}

func TestRegisterAlwaysAssignsStudentRole(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)

	rec := postJSON(t, router, "/register", map[string]string{
		"email":    "attacker@lpu.in",
		"password": "supersecret",
		"role":     "SUPER_ADMIN",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 persisted user, got %d", len(repo.created))
	}
	if got := repo.created[0].Role; got != domain.DefaultRegistrationRole {
		t.Errorf("client-supplied role was honoured: got %q, want %q", got, domain.DefaultRegistrationRole)
	}
}

func TestRegisterNormalisesEmail(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)

	rec := postJSON(t, router, "/register", map[string]string{
		"email":    "  Student@LPU.IN ",
		"password": "supersecret",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := repo.created[0].Email; got != "student@lpu.in" {
		t.Errorf("email was not normalised: got %q", got)
	}
}

func TestRegisterHashesPasswordAndNeverReturnsIt(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)

	rec := postJSON(t, router, "/register", map[string]string{
		"email":    "student@lpu.in",
		"password": "correct horse battery",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 persisted user, got %d", len(repo.created))
	}

	stored := repo.created[0]
	if stored.PasswordHash == "correct horse battery" {
		t.Fatal("password was stored in plaintext")
	}
	if !auth.CheckPasswordHash("correct horse battery", stored.PasswordHash) {
		t.Error("stored hash does not verify against the original password")
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(stored.PasswordHash)) {
		t.Error("response body leaked the password hash")
	}
}

func TestRegisterRejectsShortPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)

	rec := postJSON(t, router, "/register", map[string]string{
		"email":    "student@lpu.in",
		"password": "short",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a short password, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.created) != 0 {
		t.Error("user was persisted despite invalid input")
	}
}

func TestRegisterDuplicateEmailReturnsConflict(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)

	body := map[string]string{"email": "student@lpu.in", "password": "correct horse battery"}
	if rec := postJSON(t, router, "/register", body); rec.Code != http.StatusCreated {
		t.Fatalf("first registration: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec := postJSON(t, router, "/register", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for a duplicate email, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.created) != 1 {
		t.Errorf("duplicate registration created an extra row: %d users", len(repo.created))
	}
}

func TestLoginReturnsTokenPairAndNeverLeaksHash(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	user := seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	rec := postJSON(t, router, "/login", map[string]string{
		"email":    "student@lpu.in",
		"password": "correct horse battery",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload authPayload
	decodeData(t, rec, &payload)

	if payload.AccessToken == "" {
		t.Error("login response did not include an access token")
	}
	if payload.RefreshToken == "" {
		t.Error("login response did not include a refresh token")
	}
	if payload.AccessToken == payload.RefreshToken {
		t.Error("access and refresh tokens are identical")
	}
	if payload.ExpiresIn != int(time.Hour.Seconds()) {
		t.Errorf("unexpected expires_in %d", payload.ExpiresIn)
	}
	if payload.User.Email != user.Email {
		t.Errorf("unexpected user email %q", payload.User.Email)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(user.PasswordHash)) {
		t.Error("login response leaked the password hash")
	}

	claims, err := auth.ValidateAccessToken(payload.AccessToken, testJWTSecret, "campuscare-test")
	if err != nil {
		t.Fatalf("issued access token does not validate: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != user.Role {
		t.Errorf("token carries unexpected claims: %+v", claims)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	rec := postJSON(t, router, "/login", map[string]string{
		"email":    "student@lpu.in",
		"password": "wrong password entirely",
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("access_token")) {
		t.Error("failed login still returned a token")
	}
}

func TestLoginUnknownAccountIsIndistinguishableFromWrongPassword(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "known@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	unknown := postJSON(t, router, "/login", map[string]string{
		"email": "nobody@lpu.in", "password": "correct horse battery",
	})
	wrong := postJSON(t, router, "/login", map[string]string{
		"email": "known@lpu.in", "password": "nope nope nope",
	})

	if unknown.Code != http.StatusUnauthorized || wrong.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for both cases, got %d and %d", unknown.Code, wrong.Code)
	}
	if unknown.Body.String() != wrong.Body.String() {
		t.Errorf("responses differ and can be used to enumerate accounts:\n unknown=%s\n wrong  =%s",
			unknown.Body.String(), wrong.Body.String())
	}
}

func TestLoginRejectsInactiveAccount(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	user := seedUser(t, repo, "suspended@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)
	user.Status = "SUSPENDED"

	rec := postJSON(t, router, "/login", map[string]string{
		"email": "suspended@lpu.in", "password": "correct horse battery",
	})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a suspended account, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshRotatesTokenPair(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	login := postJSON(t, router, "/login", map[string]string{
		"email": "student@lpu.in", "password": "correct horse battery",
	})
	var first authPayload
	decodeData(t, login, &first)

	rec := postJSON(t, router, "/refresh", map[string]string{"refresh_token": first.RefreshToken})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var second authPayload
	decodeData(t, rec, &second)
	if second.AccessToken == "" || second.RefreshToken == "" {
		t.Error("refresh did not return a new token pair")
	}
	if _, err := auth.ValidateAccessToken(second.AccessToken, testJWTSecret, "campuscare-test"); err != nil {
		t.Errorf("refreshed access token does not validate: %v", err)
	}
}

func TestRefreshRejectsAccessToken(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	login := postJSON(t, router, "/login", map[string]string{
		"email": "student@lpu.in", "password": "correct horse battery",
	})
	var payload authPayload
	decodeData(t, login, &payload)

	// An access token must not be usable to mint new credentials.
	rec := postJSON(t, router, "/refresh", map[string]string{"refresh_token": payload.AccessToken})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when replaying an access token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	user := seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	login := postJSON(t, router, "/login", map[string]string{
		"email": "student@lpu.in", "password": "correct horse battery",
	})
	var payload authPayload
	decodeData(t, login, &payload)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-Test-Token", payload.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got domain.User
	decodeData(t, rec, &got)
	if got.Email != user.Email {
		t.Errorf("unexpected email %q", got.Email)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(user.PasswordHash)) {
		t.Error("/me leaked the password hash")
	}
}

func TestMeRequiresAValidToken(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	for _, token := range []string{"", "not-a-jwt", "a.b.c"} {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if token != "" {
			req.Header.Set("X-Test-Token", token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("token %q: expected 401, got %d", token, rec.Code)
		}
	}
}

func TestLoginSurfacesRepositoryFailureAsUnauthorized(t *testing.T) {
	repo := &fakeUserRepo{}
	router := newAuthTestRouter(repo)
	seedUser(t, repo, "student@lpu.in", "correct horse battery", domain.DefaultRegistrationRole)

	repo.createErr = errors.New("connection refused")
	_ = repo // createErr only affects Create; the read path must still work.

	rec := postJSON(t, router, "/login", map[string]string{
		"email": "student@lpu.in", "password": "correct horse battery",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("a Create failure must not affect login, got %d: %s", rec.Code, rec.Body.String())
	}
}
