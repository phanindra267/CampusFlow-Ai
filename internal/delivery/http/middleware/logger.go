package middleware

import (
	"time"

	"github.com/campuscare/api/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		l.WithFields(logrus.Fields{
			"status":  statusCode,
			"latency": latency,
			"ip":      clientIP,
			"method":  method,
			"path":    path,
		}).Info("HTTP Request")
	}
}
