import os
import shutil

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"

def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

# 1. Config with Secrets
write_file("internal/config/config.go", """
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Weaviate WeaviateConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
}

type LogConfig struct {
	Level  string
	Format string
}

type WeaviateConfig struct {
	Host   string
	Scheme string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error if .env doesn't exist

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		jwtSecret = "fallback-development-secret-change-in-prod"
	}

	return &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "campuscare"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Weaviate: WeaviateConfig{
			Host:   getEnv("WEAVIATE_HOST", "localhost:8080"),
			Scheme: getEnv("WEAVIATE_SCHEME", "http"),
		},
	}, nil
}

func (db DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
""")

# 2. Auth & Passwords
write_file("pkg/auth/password.go", """
package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
""")

write_file("pkg/auth/jwt.go", """
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, role, secret string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
}
""")

# 3. Middleware Update
write_file("internal/delivery/http/middleware/auth.go", """
package middleware

import (
	"net/http"
	"strings"

	"github.com/campuscare/api/pkg/auth"
	"github.com/gin-gonic/gin"
)

func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := auth.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
			return
		}
		
		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
""")

# 4. Repositories
write_file("internal/repository/postgres/user_repo.go", """
package postgres

import (
	"context"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password_hash, role, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Status)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role, status FROM users WHERE email = $1`
	var u domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
""")

write_file("internal/repository/postgres/event_repo.go", """
package postgres

import (
	"context"

	"github.com/campuscare/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, e *domain.Event) error {
	query := `INSERT INTO events (id, title, description, category, organizer_id, capacity, start_time, end_time, registration_deadline) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, e.ID, e.Title, e.Description, e.Category, e.OrganizerID, e.Capacity, e.StartTime, e.EndTime, e.RegistrationDeadline)
	return err
}

func (r *EventRepository) List(ctx context.Context) ([]domain.Event, error) {
	query := `SELECT id, title, description, category, organizer_id, capacity FROM events ORDER BY start_time DESC LIMIT 50`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Category, &e.OrganizerID, &e.Capacity); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
""")

# 5. Handlers Fix
write_file("internal/delivery/http/handler/auth.go", """
package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/auth"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	Register(ctx *gin.Context, req domain.RegisterRequest) (*domain.User, error)
	Login(ctx *gin.Context, req domain.LoginRequest) (string, error)
}

type AuthHandler struct {
	jwtSecret string
	db        UserRepository // For simplicity in this script, we'll interface this directly
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
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

# Update Event Handler
write_file("internal/delivery/http/handler/event.go", """
package handler

import (
	"net/http"

	"github.com/campuscare/api/internal/domain"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"context"
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
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	
	e.ID = uuid.New().String()
	userID, _ := c.Get("userID")
	e.OrganizerID = userID.(string)

	if err := h.repo.Create(c.Request.Context(), &e); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create event")
		return
	}
	response.Success(c, http.StatusCreated, e)
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	events, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list events")
		return
	}
	response.Success(c, http.StatusOK, events)
}
""")

# RAG & Agents (Stubs with structure)
write_file("internal/agent/agent.go", """
package agent

import (
	"context"
	"errors"
	"fmt"
)

type Tool func(ctx context.Context, args map[string]interface{}) (string, error)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(name string, t Tool) {
	r.tools[name] = t
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", errors.New("tool not found")
	}
	return t(ctx, args)
}

// Example tool
func GetAttendanceTool(ctx context.Context, args map[string]interface{}) (string, error) {
	return "Mock attendance: 85% for ML, 78% for OS.", nil
}
""")

print("Scripts written successfully. Ready to execute.")
