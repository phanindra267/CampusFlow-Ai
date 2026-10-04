import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI"
def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

# CTE Graph Repo
write_file("internal/repository/postgres/graph_repo.go", """
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphRepository struct {
	db *pgxpool.Pool
}

func NewGraphRepository(db *pgxpool.Pool) *GraphRepository {
	return &GraphRepository{db: db}
}

// GetUserNetwork uses a recursive CTE to find connected users (e.g. friends of friends)
func (r *GraphRepository) GetUserNetwork(ctx context.Context, userID string, depth int) ([]string, error) {
	query := `
	WITH RECURSIVE network AS (
		SELECT friend_id, 1 as current_depth
		FROM user_connections
		WHERE user_id = $1
		
		UNION
		
		SELECT uc.friend_id, n.current_depth + 1
		FROM user_connections uc
		INNER JOIN network n ON n.friend_id = uc.user_id
		WHERE n.current_depth < $2
	)
	SELECT DISTINCT friend_id FROM network WHERE friend_id != $1;
	`
	rows, err := r.db.Query(ctx, query, userID, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		connections = append(connections, id)
	}
	return connections, nil
}
""")

# Analytics Repo
write_file("internal/repository/postgres/analytics_repo.go", """
package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

type EventStats struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func (r *AnalyticsRepository) GetEventStats(ctx context.Context, since time.Time) ([]EventStats, error) {
	query := `
		SELECT category, COUNT(*) as count 
		FROM events 
		WHERE start_time >= $1 
		GROUP BY category 
		ORDER BY count DESC
	`
	rows, err := r.db.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		if err := rows.Scan(&s.Category, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
""")

# Fix server.go wiring
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
	router.Use(gin.Recovery())
	router.Use(middleware.AuditLog(db)) // Audit Logger

	// CORS
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

	// Dependency Injection
	userRepo := postgres.NewUserRepository(db)
	eventRepo := postgres.NewEventRepository(db)

	authHandler := handler.NewAuthHandler(cfg.JWT.Secret, userRepo)
	eventHandler := handler.NewEventHandler(eventRepo)
	oauthHandler := handler.NewOAuthHandler(cfg.JWT.Secret, userRepo)
	otpHandler := handler.NewOTPHandler()

	// API Routing
	api := router.Group("/api/v1")
	{
		api.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
		
		// Auth
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/register", authHandler.Register)
		api.GET("/auth/google", oauthHandler.GoogleLogin)
		api.GET("/auth/google/callback", oauthHandler.GoogleCallback)
		
		// OTP Auth
		api.POST("/auth/otp/request", otpHandler.RequestOTP)
		api.POST("/auth/otp/verify", otpHandler.VerifyOTP)

		// Protected
		protected := api.Group("/")
		protected.Use(middleware.RequireAuth(cfg.JWT.Secret))
		{
			campus := protected.Group("/campus")
			campus.Use(middleware.RequireRole("STUDENT", "SUPER_ADMIN", "ADMIN"))
			{
				campus.POST("/events", eventHandler.CreateEvent)
				campus.GET("/events", eventHandler.ListEvents)
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
