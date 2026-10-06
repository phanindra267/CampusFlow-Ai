package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps how much of a request body the server will read.
//
// Without it, a single unauthenticated POST declaring a multi-gigabyte
// Content-Length can make the process allocate until it is killed. The limit is
// enforced with http.MaxBytesReader, which fails the read as soon as the cap is
// passed rather than after buffering the whole thing.
//
// Binding a body that exceeds the cap then fails with a size error, which the
// handler reports as INVALID_REQUEST.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && maxBytes > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
