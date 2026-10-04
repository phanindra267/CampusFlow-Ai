package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ErrOAuthNotConfigured is returned when the Google OAuth environment
// variables are absent. Callers should surface 501 rather than attempting an
// exchange that cannot succeed.
var ErrOAuthNotConfigured = errors.New("google oauth is not configured")

// GetGoogleOAuthConfig builds the Google OAuth client from the environment.
func GetGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		RedirectURL:  strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GoogleOAuthConfigured reports whether every required variable is present.
func GoogleOAuthConfigured() bool {
	cfg := GetGoogleOAuthConfig()
	return cfg.ClientID != "" && cfg.ClientSecret != "" && cfg.RedirectURL != ""
}

// OAuthStateTTL bounds how long an authorisation "state" value stays valid.
const OAuthStateTTL = 10 * time.Minute
