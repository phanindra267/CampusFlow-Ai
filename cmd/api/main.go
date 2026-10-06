package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/campuscare/api/internal/config"
	"github.com/campuscare/api/internal/database"
	"github.com/campuscare/api/internal/logger"
	"github.com/campuscare/api/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// shutdownTimeout bounds how long in-flight requests may take to drain before
// the process exits anyway.
const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		// The logger may not exist yet if configuration failed, so this one
		// message goes to stderr in plain text.
		os.Stderr.WriteString("campuscare-api: " + err.Error() + "\n")
		os.Exit(1)
	}
}

func run() error {
	// A nil *Config with a non-nil error is a valid Load outcome: it refuses to
	// hand back a configuration that would use a fallback signing secret or a
	// too-short JWT_SECRET. That refusal has to end the process here. Treating
	// it as a warning and continuing would dereference the nil config one line
	// later, turning a deliberate security guard into an opaque panic.
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg == nil {
		return errNilConfig
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	dbPool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		// PostgreSQL is the source of truth for every domain in this service, so
		// outside development an instance without it cannot serve traffic and
		// must not start. In development the server still comes up so the
		// frontend and the health endpoints remain reachable while a developer
		// starts Postgres; /health/ready reports DOWN and DB-backed routes fail
		// closed rather than corrupting anything.
		if !isDevelopment(cfg.Env) {
			return errDatabaseUnavailable(err)
		}
		log.WithError(err).Warn(
			"PostgreSQL is unreachable. Starting in degraded development mode: " +
				"health endpoints respond, but every database-backed route will fail. " +
				"Start Postgres and restart to restore full function.")
	}

	srv := server.New(cfg, dbPool)

	serveErr := make(chan error, 1)
	go func() {
		log.WithFields(logrus.Fields{
			"port": cfg.Server.Port,
			// Stating the negotiated-protocol story up front turns "why is my
			// client still on HTTP/1.1" into something answered by the start-up
			// log rather than by guessing at the server's configuration.
			"protocol": protocolSummary(cfg),
		}).Info("HTTP server listening")
		if err := srv.Start(); err != nil {
			serveErr <- err
		}
		close(serveErr)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		if err != nil {
			closePool(log, dbPool)
			return errListenFailed(err)
		}
		// Start returned without an error, which means the server closed on its
		// own. Fall through to shutdown so resources are still released.
	case sig := <-signals:
		log.WithField("signal", sig.String()).Info("Shutdown signal received; draining")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := srv.Shutdown(ctx)
	closePool(log, dbPool)

	if shutdownErr != nil {
		return errDrainFailed(shutdownErr)
	}

	log.Info("Shutdown complete")
	return nil
}

// protocolSummary describes, in a few words, which wire protocols this listener
// will speak. HTTP/2 needs TLS to be negotiated properly and h2c to be reached
// at all, so the same HTTP2 switch means different things depending on whether a
// certificate was supplied.
func protocolSummary(cfg *config.Config) string {
	switch {
	case !cfg.Server.HTTP2:
		return "HTTP/1.1"
	case cfg.Server.TLSConfigured():
		return "HTTP/2 over TLS (ALPN), HTTP/1.1 fallback"
	default:
		return "HTTP/2 over cleartext (h2c), HTTP/1.1 fallback"
	}
}

// closePool releases database connections, reporting a failure without masking
// whatever error is already on its way out of run.
func closePool(log *logger.Logger, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	pool.Close()
	log.Info("Database connection pool closed")
}

func isDevelopment(env string) bool {
	switch env {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}
