import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"

def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

write_file("internal/domain/user.go", """
package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never expose
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}
""")

write_file("pkg/response/response.go", """
package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
	})
}
""")

write_file("internal/delivery/http/handler/auth.go", """
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
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.db.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not generate token")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
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
		response.Error(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"user_id": user.ID})
}
""")
