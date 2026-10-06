package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/campuscare/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// auditTimeout bounds the audit write so a slow database cannot add its latency
// to every mutating request.
const auditTimeout = 3 * time.Second

// AuditLog records state-mutating requests in the audit_logs table.
//
// It deliberately does not read the request body. An earlier version buffered
// every body into memory purely to restore it for the handler, which cost an
// unbounded allocation on every POST and bought nothing: the audit row records
// who did what to which route, not what they sent.
//
// The pool may be nil when PostgreSQL was unreachable at boot; auditing is then
// skipped rather than panicking inside the request pipeline.
func AuditLog(l *logger.Logger, db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			return
		}

		if db == nil {
			l.WithFields(logrus.Fields{
				"request_id": RequestIDOf(c),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
			}).Warn("audit skipped: database unavailable")
			return
		}

		userID := nullableString(UserID(c))

		const query = `INSERT INTO audit_logs (user_id, method, path, status, latency_ms, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`

		ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
		defer cancel()

		if _, err := db.Exec(ctx, query, userID, c.Request.Method,
			c.Request.URL.Path, c.Writer.Status(), time.Since(start).Milliseconds(), time.Now()); err != nil {
			l.WithError(err).WithFields(logrus.Fields{
				"request_id": RequestIDOf(c),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
			}).Warn("audit entry not persisted")
		}
	}
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
