package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// ReadinessCheck reports whether dependencies are reachable. A nil pool means
// Postgres was unavailable at boot, which is itself a not-ready condition.
func ReadinessCheck(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"error":  "Database unavailable",
			})
			return
		}

		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
				"error":  "Database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	}
}
