package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/auth"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	authCfg AuthConfig
	db      UserRepository
}

func NewOAuthHandler(authCfg AuthConfig, repo UserRepository) *OAuthHandler {
	return &OAuthHandler{authCfg: authCfg, db: repo}
}

// oauthState is the payload carried through the provider round-trip. Signing it
// with the server secret means no server-side session store is required while
// still binding the callback to the browser that started the flow.
type oauthState struct {
	Nonce       string `json:"nonce"`
	RedirectURI string `json:"redirect_uri"`
}

func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	oauthCfg := config.GetGoogleOAuthConfig()
	if !config.GoogleOAuthConfigured() {
		response.Error(c, http.StatusNotImplemented, "google sign-in is not configured", config.ErrOAuthNotConfigured)
		return
	}

	redirectURI := sanitizeRedirect(c.Query("redirect_uri"))

	state, err := h.signState(oauthState{
		Nonce:       uuid.NewString(),
		RedirectURI: redirectURI,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not start sign-in", err)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline))
}

func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	if !config.GoogleOAuthConfigured() {
		response.Error(c, http.StatusNotImplemented, "google sign-in is not configured", config.ErrOAuthNotConfigured)
		return
	}

	// The provider reports user denial through a query parameter rather than a
	// non-2xx status.
	if providerErr := c.Query("error"); providerErr != "" {
		response.Error(c, http.StatusBadRequest, "google sign-in was not completed", fmt.Errorf("%s", providerErr))
		return
	}

	state, err := h.verifyState(c.Query("state"))
	if err != nil {
		// A missing or forged state means the callback did not originate from
		// this server's sign-in request.
		response.Error(c, http.StatusBadRequest, "invalid oauth state", err)
		return
	}

	code := c.Query("code")
	if code == "" {
		response.Error(c, http.StatusBadRequest, "authorization code missing", nil)
		return
	}

	oauthCfg := config.GetGoogleOAuthConfig()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	tok, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "failed to exchange token", err)
		return
	}

	client := oauthCfg.Client(ctx, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		response.Error(c, http.StatusBadGateway, "failed to get user info", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.Error(c, http.StatusBadGateway, "failed to get user info",
			fmt.Errorf("provider returned %d", resp.StatusCode))
		return
	}

	var userInfo struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"verified_email"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		response.Error(c, http.StatusBadGateway, "failed to parse user info", err)
		return
	}

	if userInfo.Email == "" || !userInfo.EmailVerified {
		response.Error(c, http.StatusBadRequest, "google account email is missing or unverified", nil)
		return
	}

	email := strings.ToLower(strings.TrimSpace(userInfo.Email))

	user, err := h.db.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, postgres.ErrNotFound) {
			// A genuine database failure must not be mistaken for "no such
			// user", which would otherwise create a duplicate account.
			response.Error(c, http.StatusInternalServerError, "failed to load user", err)
			return
		}

		if err := h.provisionOAuthUser(ctx, email); err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to create user", err)
			return
		}
		if user, err = h.db.GetByEmail(ctx, email); err != nil || user == nil {
			response.Error(c, http.StatusInternalServerError, "failed to load user", err)
			return
		}
	}

	tokens, err := issueTokenPair(h.authCfg, user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not generate token", err)
		return
	}

	// Returning tokens in a fragment keeps them out of server access logs and
	// out of the Referer header of any subsequent navigation.
	fragment := url.Values{}
	fragment.Set("access_token", tokens.Access)
	fragment.Set("refresh_token", tokens.Refresh)
	target := state.RedirectURI
	if target == "" {
		target = "/dashboard"
	}
	c.Redirect(http.StatusFound, target+"#"+fragment.Encode())
}

func (h *OAuthHandler) provisionOAuthUser(ctx context.Context, email string) error {
	// OAuth-only accounts get an unusable random hash rather than an empty
	// string so they can never be signed into with a password.
	unusable, err := auth.HashPassword(uuid.NewString() + uuid.NewString())
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: unusable,
		Role:         domain.DefaultRegistrationRole,
		Status:       "ACTIVE",
	}

	if err := h.db.Create(ctx, user); err != nil && !errors.Is(err, postgres.ErrDuplicateKey) {
		return err
	}
	return nil
}

func (h *OAuthHandler) signState(state oauthState) (string, error) {
	return auth.SignTransient(state.Nonce, state.RedirectURI, h.authCfg.Secret, h.authCfg.Issuer, config.OAuthStateTTL)
}

func (h *OAuthHandler) verifyState(raw string) (oauthState, error) {
	if raw == "" {
		return oauthState{}, errors.New("state parameter missing")
	}

	claims, err := auth.VerifyTransient(raw, h.authCfg.Secret, h.authCfg.Issuer)
	if err != nil {
		return oauthState{}, err
	}
	if claims.Nonce == "" {
		return oauthState{}, errors.New("state nonce missing")
	}
	return oauthState{Nonce: claims.Nonce, RedirectURI: claims.RedirectURI}, nil
}

// sanitizeRedirect only permits same-origin relative paths, so the OAuth
// callback cannot be turned into an open redirect.
func sanitizeRedirect(raw string) string {
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return ""
	}
	if strings.ContainsAny(raw, "\r\n") {
		return ""
	}
	return raw
}
