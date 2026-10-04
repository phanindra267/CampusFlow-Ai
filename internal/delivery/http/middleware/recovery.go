package middleware

import (
	"net/http"

	"github.com/campuscare/api/internal/logger"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
)

func Recovery(l *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err any) {
		l.Errorf("Panic recovered: %v", err)
		response.Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
		c.Abort()
	})
}
