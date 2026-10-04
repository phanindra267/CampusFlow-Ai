package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/delivery/http/middleware"
	"github.com/campuscare/api/internal/logger"
	"github.com/campuscare/api/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
	db         *pgxpool.Pool
}

// New builds the HTTP server and registers every API route.
func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	router := gin.New()
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(cors(cfg))
	router.Use(middleware.AuditLog(db))

	userRepo := postgres.NewUserRepository(db)

	registerRoutes(router, cfg, newHandlers(cfg, db, userRepo))

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Server.Port,
			Handler:           router,
			ReadTimeout:       cfg.Server.ReadTimeout,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       60 * time.Second,
		},
		db: db,
	}
}

func cors(cfg *config.Config) gin.HandlerFunc {
	allowed := cfg.Server.CORSOrigins

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origin, allowed) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, candidate := range allowed {
		if candidate == "*" || candidate == origin {
			return true
		}
	}
	return false
}

// Handler exposes the fully wired router so tests can drive the real route graph
// in-process instead of binding a TCP port.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server and releases the database pool.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.db != nil {
		s.db.Close()
	}
	return s.httpServer.Shutdown(ctx)
}
