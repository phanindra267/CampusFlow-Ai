package postgres

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// generateCheckInToken returns an unguessable token for an event check-in
// window. It is generated server-side rather than supplied by the client so a
// token cannot be chosen, predicted or reused across sessions.
func generateCheckInToken() (string, error) {
	// 24 random bytes gives 192 bits of entropy, which is far more than needed
	// for a token that also expires.
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
