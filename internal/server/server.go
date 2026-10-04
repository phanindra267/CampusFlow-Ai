package server

import (
	"context"
	"net/http"
	

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
