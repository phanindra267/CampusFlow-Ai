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
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	httpServer  *http.Server
	tlsCertFile string
	tlsKeyFile  string
	db          *pgxpool.Pool
}

// maxRequestBodyBytes caps any single request body.
//
// Bodies here are JSON documents describing an event, a review comment or a
// ticket message; none is legitimately large. The cap exists so a single request
// cannot force the process to buffer an arbitrary amount of memory. It is
// generous enough that no legitimate client is affected: the largest real
// payload is an announcement body, and the largest text fields in the schema are
// well under this.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// securityHeaders sets response headers that cost nothing and close off
// embedding, sniffing and referrer leakage. The API serves JSON, so there is no
// inline script of its own to protect: these are defence in depth for the
// browser that renders the responses.
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		// This API never returns HTML, so the strictest possible policy costs
		// nothing. frame-ancestors is the CSP spelling of the same idea as
		// X-Frame-Options and is honoured by current browsers.
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Next()
	}
}

// New builds the HTTP server and registers every API route.
func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	router := gin.New()

	// Decide who may lie about the client address before anything reads it.
	//
	// Gin trusts every proxy by default, which means it believes an
	// X-Forwarded-For header from anyone. That matters because ClientIP() feeds
	// two things: the auth rate limiter's bucket key and the audit trail's idea
	// of who did something. With proxy trust left open, a caller can send a
	// fresh X-Forwarded-For on every request and get an unlimited number of rate
	// limiter buckets, which makes the limiter a no-op, while also attributing
	// their actions to invented addresses.
	//
	// The safe default is therefore to trust nobody and use the direct peer. A
	// deployment behind Caddy opts in explicitly via TRUSTED_PROXIES, listing
	// only the proxy's address, so that proxy's forwarded headers are honoured
	// and everyone else's are ignored.
	if err := router.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		// A malformed CIDR would otherwise leave the engine in its insecure
		// default state without anyone noticing. Refuse to start instead.
		log.WithError(err).Error("TRUSTED_PROXIES is not a valid list of CIDR ranges; refusing to start")
		return nil
	}

	// RequestID runs first so every later middleware and every handler can log
	// under the same correlation ID.
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(cors(cfg))
	router.Use(securityHeaders())
	router.Use(middleware.BodyLimit(maxRequestBodyBytes))
	router.Use(middleware.AuditLog(log, db))

	userRepo := postgres.NewUserRepository(db)

	registerRoutes(router, cfg, newHandlers(cfg, db, userRepo))

	handler := http.Handler(router)

	// HTTP/2 needs two different mechanisms depending on whether there is TLS.
	//
	// With a certificate, net/http negotiates "h2" over ALPN by itself as long
	// as TLSNextProto is left nil, which is what enables h2 in the first place.
	//
	// Without a certificate there is nothing to negotiate: the connection is
	// cleartext HTTP/1.1 until the client sends the RFC 7540 preface or an
	// Upgrade header. h2c.NewHandler sniffs for that preface and switches the
	// connection over, which is the only way to serve h2 over plain HTTP.
	// Wrapping the router rather than replacing it keeps the h1 path intact for
	// browsers, which do not speak h2c.
	if cfg.Server.HTTP2 && !cfg.Server.TLSConfigured() {
		handler = h2c.NewHandler(router, &http2.Server{
			IdleTimeout:          60 * time.Second,
			MaxConcurrentStreams: 250,
			MaxReadFrameSize:     1 << 20,
		})
		log.WithField("mode", "h2c").Info(
			"HTTP/2 enabled without TLS: h2c is available for clients that speak it, " +
				"HTTP/1.1 remains for everyone else")
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Server.Port,
			Handler:           handler,
			ReadTimeout:       cfg.Server.ReadTimeout,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       60 * time.Second,
		},
		tlsCertFile: cfg.Server.TLSCertFile,
		tlsKeyFile:  cfg.Server.TLSKeyFile,
		db:          db,
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

// Start serves until the server is shut down. When a certificate has been
// configured it terminates TLS and lets net/http negotiate HTTP/2 over ALPN;
// otherwise the handler is already h2c-aware from New.
func (s *Server) Start() error {
	if s.tlsCertFile != "" && s.tlsKeyFile != "" {
		err := s.httpServer.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

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
