import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"
def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

# Rewrite server.go
write_file("internal/server/server.go", """
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/delivery/http/handler"
	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
	db         *pgxpool.Pool
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Global Middleware
	router.Use(gin.Recovery())
	// CORS for Next.js frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Dependency Injection Setup
	userRepo := postgres.NewUserRepository(db)
	eventRepo := postgres.NewEventRepository(db)

	authHandler := handler.NewAuthHandler(cfg.JWT.Secret, userRepo)
	eventHandler := handler.NewEventHandler(eventRepo)

	// API Routing
	api := router.Group("/api/v1")
	{
		// Health & Public
		api.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/register", authHandler.Register)

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middleware.RequireAuth(cfg.JWT.Secret))
		{
			// Student Scope
			campus := protected.Group("/campus")
			campus.Use(middleware.RequireRole("STUDENT", "SUPER_ADMIN"))
			{
				campus.POST("/events", eventHandler.CreateEvent)
				campus.GET("/events", eventHandler.ListEvents)
				// other mock handlers could be injected here
			}
		}
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Server.Port,
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		db: db,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.db.Close()
	return s.httpServer.Shutdown(ctx)
}
""")

# Fix main.go
write_file("cmd/api/main.go", """
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbPool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	srv := server.New(cfg, dbPool)

	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
""")

# Create Docker & CI
write_file("Dockerfile", """
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]
""")

write_file("docker-compose.yml", """
version: '3.8'
services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgrespassword
      POSTGRES_DB: campuscare
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgrespassword
      - DB_NAME=campuscare
      - DB_SSLMODE=disable
      - JWT_SECRET=production-secret-replace-me
    depends_on:
      - db

  weaviate:
    image: semitechnologies/weaviate:1.24.1
    ports:
      - "8081:8080"
    environment:
      QUERY_DEFAULTS_LIMIT: 25
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      DEFAULT_VECTORIZER_MODULE: 'none'

volumes:
  postgres_data:
""")

write_file(".github/workflows/ci.yml", """
name: CI
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Build
      run: go build -v ./...
    - name: Test
      run: go test -v ./...
""")

print("Server injection, Docker, and CI files written.")
