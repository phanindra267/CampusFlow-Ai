package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set an environment variable
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Server.Port != "9999" {
		t.Errorf("Expected port to be 9999, got %s", cfg.Server.Port)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "host",
		Port:     "5432",
		User:     "user",
		Password: "password",
		Name:     "name",
		SSLMode:  "disable",
	}
	expected := "host=host port=5432 user=user password=password dbname=name sslmode=disable"
	if cfg.DSN() != expected {
		t.Errorf("Expected DSN %s, got %s", expected, cfg.DSN())
	}
}

// loadWithEnv points Load at an isolated set of environment variables. Load reads
// the process environment directly, so every variable it consults has to be
// neutralised or an earlier test's value leaks into the next one.
func loadWithEnv(t *testing.T, env map[string]string) (*Config, error) {
	t.Helper()

	for _, key := range []string{
		"APP_ENV", "JWT_SECRET", "PORT", "HTTP2_ENABLED",
		"TLS_CERT_FILE", "TLS_KEY_FILE", "TRUSTED_PROXIES",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("JWT_SECRET", strings.Repeat("k", 40))

	for key, value := range env {
		t.Setenv(key, value)
	}

	// godotenv.Load() inside Load would otherwise pull the developer's .env back
	// in and undo the isolation this helper just established.
	t.Chdir(t.TempDir())

	return Load()
}

func TestLoadHTTP2DefaultsOn(t *testing.T) {
	cfg, err := loadWithEnv(t, nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.Server.HTTP2 {
		t.Error("HTTP/2 should be enabled by default; a feature that is safer on by " +
			"default should not need an opt-in")
	}
	if cfg.Server.TLSConfigured() {
		t.Error("no certificate was configured, so TLSConfigured should be false")
	}
}

func TestLoadHTTP2OptOut(t *testing.T) {
	cfg, err := loadWithEnv(t, map[string]string{"HTTP2_ENABLED": "false"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Server.HTTP2 {
		t.Error("HTTP2_ENABLED=false should disable HTTP/2")
	}
}

func TestLoadHTTP2TreatsTyposAsOff(t *testing.T) {
	// An unrecognised value must not be read as "on". strconv.ParseBool would
	// accept "1"/"t" and error here; this codebase's parser is stricter still,
	// so a typo fails closed rather than quietly enabling a feature nobody asked
	// for.
	for _, value := range []string{"enbale", "2", "yess", "on"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := loadWithEnv(t, map[string]string{"HTTP2_ENABLED": value})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.Server.HTTP2 != (value == "on") {
				t.Errorf("HTTP2_ENABLED=%q produced HTTP2=%v", value, cfg.Server.HTTP2)
			}
		})
	}
}

func TestLoadRejectsHalfConfiguredTLS(t *testing.T) {
	// Setting only the certificate is the dangerous state: an operator would
	// believe the server is serving TLS while it silently serves plain HTTP.
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", strings.Repeat("k", 40))
	t.Setenv("TLS_CERT_FILE", "/etc/ssl/server.crt")
	t.Chdir(t.TempDir())

	if _, err := Load(); err == nil {
		t.Fatal("Load accepted a certificate with no private key; startup must fail " +
			"rather than fall back to cleartext")
	}
}

func TestLoadRequiresTLSOutsideDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("k", 40))
	t.Chdir(t.TempDir())

	if _, err := Load(); err == nil {
		t.Fatal("Load accepted production without TLS; that would put login " +
			"credentials on the wire in the clear")
	}

	t.Setenv("TLS_CERT_FILE", "/etc/ssl/server.crt")
	t.Setenv("TLS_KEY_FILE", "/etc/ssl/server.key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load with a certificate pair: %v", err)
	}
	if !cfg.Server.TLSConfigured() {
		t.Error("TLSConfigured should report true when both paths are set")
	}
}
