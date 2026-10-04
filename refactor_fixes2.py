import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"

def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

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
""")

write_file("internal/delivery/http/handler/event.go", """
package handler

import (
	"context"
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventRepository interface {
	Create(ctx context.Context, e *domain.Event) error
	List(ctx context.Context) ([]domain.Event, error)
}

type EventHandler struct {
	repo EventRepository
}

func NewEventHandler(repo EventRepository) *EventHandler {
	return &EventHandler{repo: repo}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var e domain.Event
	if err := c.ShouldBindJSON(&e); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}
	
	e.ID = uuid.New().String()
	userID, _ := c.Get("userID")
	e.OrganizerID = userID.(string)

	if err := h.repo.Create(c.Request.Context(), &e); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create event", err)
		return
	}
	response.Success(c, http.StatusCreated, "event created", e)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list events", err)
		return
	}
	response.Success(c, http.StatusOK, "events retrieved", events)
}
""")
