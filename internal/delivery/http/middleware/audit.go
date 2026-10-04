package middleware

import (
	"bytes"
	"io"
	"log"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuditLog(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		// Only log state-mutating requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			userID, _ := c.Get("userID")
			uidStr, _ := userID.(string)

			query := `INSERT INTO audit_logs (user_id, method, path, status, latency_ms, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
			_, err := db.Exec(context.Background(), query, uidStr, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start).Milliseconds(), time.Now())
			if err != nil {
				log.Printf("Failed to write audit log: %v", err)
			}
		}
	}
}
