package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextRequestID is the gin context key holding this request's correlation ID.
const ContextRequestID = "requestID"

// HeaderRequestID is the canonical header used to propagate a correlation ID
// across service boundaries.
const HeaderRequestID = "X-Request-ID"

// maxInboundRequestIDLen bounds how much of a caller-supplied ID is trusted. The
// value is echoed back in a response header and written to every log line, so an
// unbounded inbound string would let a client inject newlines into logs or blow
// up their size.
const maxInboundRequestIDLen = 64

// RequestID assigns every request a correlation ID, echoes it back to the
// caller, and makes it available to handlers and loggers.
//
// Without one, a report of "the AI page failed at 14:03" can only be correlated
// with the server logs by reading every line around that time. With it, the ID
// from the client's error report appears verbatim on every log line the request
// produced, including the ones written deep in the repository layer.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" || len(id) > maxInboundRequestIDLen {
			id = uuid.NewString()
		}

		c.Set(ContextRequestID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}

// RequestIDOf returns the correlation ID for the current request, or an empty
// string when the middleware is not installed.
func RequestIDOf(c *gin.Context) string {
	value, _ := c.Get(ContextRequestID)
	id, _ := value.(string)
	return id
}
