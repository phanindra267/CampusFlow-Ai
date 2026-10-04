package middleware

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLog records state-mutating requests in the audit_logs table.
// The pool may be nil when Postgres is unreachable at boot; auditing is then
// skipped instead of panicking inside the request pipeline.
func AuditLog(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		c.Next()

		switch c.Request.Method {
		case "POST", "PUT", "PATCH", "DELETE":
		default:
			return
		}

		if db == nil {
			log.Printf("audit: database unavailable, skipped %s %s", c.Request.Method, c.Request.URL.Path)
			return
		}

		uid, _ := c.Get("userID")
		userID, _ := uid.(string)

		const query = `INSERT INTO audit_logs (user_id, method, path, status, latency_ms, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := db.Exec(ctx, query, nullableString(userID), c.Request.Method,
			c.Request.URL.Path, c.Writer.Status(), time.Since(start).Milliseconds(), time.Now()); err != nil {
			log.Printf("audit: failed to persist entry for %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		}
	}
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
