package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/campuscare/api/pkg/auth"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	UpdateDisplayName(ctx context.Context, id, displayName string) error
	GetProfile(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, id string, req *domain.UpdateProfileRequest) (*domain.User, error)
	LoadProfile(ctx context.Context, id string) (*domain.Profile, error)
}

// writeAuthRepoError maps repository sentinels onto HTTP responses for the auth
// handlers.
func writeAuthRepoError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		response.Error(c, http.StatusNotFound, "user not found", err)
	case errors.Is(err, postgres.ErrDuplicateKey):
		response.Error(c, http.StatusConflict, "email already registered", err)
	case errors.Is(err, postgres.ErrNoDatabase):
		response.Error(c, http.StatusServiceUnavailable, "database unavailable", err)
	default:
		response.Error(c, http.StatusInternalServerError, message, err)
	}
}

type AuthConfig struct {
	Secret    string
	Issuer    string
	AccessTTL time.Duration
}

func (c AuthConfig) tokenOptions() auth.TokenOptions {
	return auth.TokenOptions{Issuer: c.Issuer, AccessTTL: c.AccessTTL}
}

type AuthHandler struct {
	cfg AuthConfig
	db  UserRepository
}

func NewAuthHandler(cfg AuthConfig, repo UserRepository) *AuthHandler {
	return &AuthHandler{cfg: cfg, db: repo}
}

// Login exchanges credentials for an access + refresh token pair.
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Normalise so "Student@LPU.in" and "student@lpu.in" are the same account.
	email := domain.NormalizeEmail(req.Email)
	if err := domain.ValidateEmail(email); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	user, err := h.db.GetByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		// Same response for unknown account and wrong password so the endpoint
		// cannot be used to enumerate registered email addresses.
		response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}
	if user.Status != "" && !strings.EqualFold(user.Status, "ACTIVE") {
		response.Error(c, http.StatusForbidden, "account is not active", nil)
		return
	}

	tokens, err := h.issueTokens(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not generate token", err)
		return
	}

	response.Success(c, http.StatusOK, "login successful", gin.H{
		"access_token":  tokens.Access,
		"refresh_token": tokens.Refresh,
		"expires_in":    int(h.cfg.AccessTTL.Seconds()),
		"user":          user,
	})
}

// Register provisions a new account. Public registration always yields a
// member account; elevated roles are granted by an administrator.
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	email := domain.NormalizeEmail(req.Email)
	if err := domain.ValidateEmail(email); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not hash password", err)
		return
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		Role:         domain.DefaultRegistrationRole,
		Status:       "ACTIVE",
	}

	if err := h.db.Create(c.Request.Context(), user); err != nil {
		if errors.Is(err, postgres.ErrDuplicateKey) {
			response.Error(c, http.StatusConflict, "email already registered", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	response.Success(c, http.StatusCreated, "user registered", gin.H{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})
}

// Refresh exchanges a valid refresh token for a new token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	claims, err := auth.ValidateRefreshToken(req.RefreshToken, h.cfg.Secret, h.cfg.Issuer)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token", nil)
		return
	}

	// Re-read the account so a deleted or deactivated user cannot keep
	// refreshing their session indefinitely.
	user, err := h.db.GetByID(c.Request.Context(), claims.UserID)
	if err != nil || user == nil {
		response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token", nil)
		return
	}
	if user.Status != "" && !strings.EqualFold(user.Status, "ACTIVE") {
		response.Error(c, http.StatusForbidden, "account is not active", nil)
		return
	}

	tokens, err := h.issueTokens(user.ID, user.Role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not generate token", err)
		return
	}

	response.Success(c, http.StatusOK, "token refreshed", gin.H{
		"access_token":  tokens.Access,
		"refresh_token": tokens.Refresh,
		"expires_in":    int(h.cfg.AccessTTL.Seconds()),
		"user":          user,
	})
}

// Me returns the authenticated account. It doubles as a client-side session
// validity check.
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}

	user, err := h.db.GetProfile(c.Request.Context(), userID)
	if err != nil || user == nil {
		response.Error(c, http.StatusNotFound, "user not found", nil)
		return
	}

	response.Success(c, http.StatusOK, "current user", user)
}

// Profile returns the whole profile page in one response: the member's own
// fields plus the clubs, saved items, registrations, applications and support
// tickets they have accumulated. Keeping it in one call is deliberate, since the
// page needs all of it and a member should not wait on eight round trips to see
// their own profile.
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}

	profile, err := h.db.LoadProfile(c.Request.Context(), userID)
	if err != nil {
		writeAuthRepoError(c, err, "could not load profile")
		return
	}

	response.Success(c, http.StatusOK, "profile retrieved", profile)
}

// UpdateProfile changes how the member appears across CampusCare AI, and the
// optional context that makes recommendations and matches useful: program,
// branch, year of study, headline, bio and career interests. Program and year are
// self-reported and used for relevance only; they are not an academic record.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err)
		return
	}

	if strings.TrimSpace(req.DisplayName) == "" {
		response.Error(c, http.StatusBadRequest, "DISPLAY_NAME_REQUIRED", nil)
		return
	}

	user, err := h.db.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		writeAuthRepoError(c, err, "could not update profile")
		return
	}

	response.Success(c, http.StatusOK, "profile updated", user)
}

type tokenPair struct {
	Access  string
	Refresh string
}

func (h *AuthHandler) issueTokens(userID, role string) (tokenPair, error) {
	return issueTokenPair(h.cfg, userID, role)
}

// issueTokenPair mints an access/refresh pair from an auth configuration. It is
// shared by the password and OAuth sign-in paths so both issue identical claims.
func issueTokenPair(cfg AuthConfig, userID, role string) (tokenPair, error) {
	opts := cfg.tokenOptions()

	access, err := auth.GenerateAccessToken(userID, role, cfg.Secret, opts)
	if err != nil {
		return tokenPair{}, err
	}
	refresh, err := auth.GenerateRefreshToken(userID, role, cfg.Secret, opts)
	if err != nil {
		return tokenPair{}, err
	}
	return tokenPair{Access: access, Refresh: refresh}, nil
}
