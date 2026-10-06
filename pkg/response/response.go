package response

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// requestIDKey mirrors the middleware context key. It is duplicated rather than
// imported so that this package stays a leaf with no dependency on internal/,
// which keeps the dependency graph acyclic.
const requestIDKey = "requestID"

// Success writes a successful response in the shape every endpoint uses.
func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Error writes a failure response.
//
// The distinction that matters here is between the message and the cause.
// `message` is written by a developer, is safe to show a member, and is the only
// error text that reaches the client. `err` is whatever the failure actually was:
// a pgx error carrying table, column and constraint names, a Weaviate URL, a
// token-exchange response body.
//
// Those causes are logged with the request's correlation ID and never
// serialised. Sending them turned every 500 into a description of the schema and
// gave an attacker a free map of the database, plus a way to fingerprint
// accounts through the wording of constraint violations.
//
// Callers therefore pass the real error for the log and keep writing their own
// human-readable message for the client.
func Error(c *gin.Context, code int, message string, err error) {
	if err != nil {
		fields := logrus.Fields{
			"status":         code,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"client_message": message,
		}
		if id, ok := c.Get(requestIDKey); ok {
			if requestID, ok := id.(string); ok && requestID != "" {
				fields["request_id"] = requestID
			}
		}

		entry := logrus.WithFields(fields)
		// A 5xx is our fault and deserves Error; a 4xx is the caller's and is
		// logged at Warn so alerting can separate the two.
		if code >= 500 {
			entry.WithError(err).Error("request failed")
		} else {
			entry.WithError(err).Warn("request rejected")
		}
	}

	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
	})
}
