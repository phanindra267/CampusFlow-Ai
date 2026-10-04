package middleware

import (
	"net/http"
	"strings"

	"github.com/campuscare/api/pkg/auth"
	"github.com/gin-gonic/gin"
)

// Context keys for values published by RequireAuth.
const (
	ContextUserID = "userID"
	ContextRole   = "role"
)

func RequireAuth(jwtSecret, jwtIssuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c)
		if !ok {
			return
		}

		claims, err := auth.ValidateAccessToken(token, jwtSecret, jwtIssuer)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// bearerToken extracts and shape-checks the Authorization header, aborting the
// request with 401 when it is absent or malformed.
func bearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
		return "", false
	}

	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return "", false
	}
	return strings.TrimSpace(token), true
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(ContextRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
			return
		}

		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// UserID returns the authenticated user's id, or "" when RequireAuth did not run.
func UserID(c *gin.Context) string {
	if value, ok := c.Get(ContextUserID); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

// Role returns the authenticated user's role, or "" when RequireAuth did not run.
func Role(c *gin.Context) string {
	if value, ok := c.Get(ContextRole); ok {
		if role, ok := value.(string); ok {
			return role
		}
	}
	return ""
}
