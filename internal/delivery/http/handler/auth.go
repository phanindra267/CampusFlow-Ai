package handler

import (
	"context"
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/auth"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type AuthHandler struct {
	jwtSecret string
	db        UserRepository
}

func NewAuthHandler(jwtSecret string, repo UserRepository) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, db: repo}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	user, err := h.db.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not generate token", err)
		return
	}

	response.Success(c, http.StatusOK, "login successful", gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	hash, _ := auth.HashPassword(req.Password)
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		Status:       "ACTIVE",
	}

	if err := h.db.Create(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	response.Success(c, http.StatusCreated, "user registered", gin.H{"user_id": user.ID})
}
