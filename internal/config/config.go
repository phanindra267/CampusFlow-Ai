package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Weaviate WeaviateConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CORSOrigins  []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	Issuer     string
	TTL        time.Duration
	RefreshTTL time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type WeaviateConfig struct {
	Host   string
	Scheme string
}

// FallbackDevJWTSecret is only ever used for local development, and only when
// the operator has not supplied JWT_SECRET. Load refuses to return a
// configuration in production that relies on it.
const FallbackDevJWTSecret = "fallback-development-secret-change-in-prod"

func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error if .env doesn't exist

	env := getEnv("APP_ENV", "development")

	jwtSecret := getEnv("JWT_SECRET", "")
	usingFallbackSecret := false
	if jwtSecret == "" {
		jwtSecret = FallbackDevJWTSecret
		usingFallbackSecret = true
	}

	if usingFallbackSecret && !isDevelopment(env) {
		return nil, fmt.Errorf(
			"JWT_SECRET must be set when APP_ENV=%q; refusing to start with the built-in development secret",
			env,
		)
	}
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(jwtSecret))
	}

	return &Config{
		Env: env,
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			CORSOrigins:  getEnvList("CORS_ORIGINS", []string{"http://localhost:3000"}),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "campuscare"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			Issuer:     getEnv("JWT_ISSUER", "campuscare-api"),
			TTL:        getEnvDuration("JWT_TTL", 24*time.Hour),
			RefreshTTL: getEnvDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Weaviate: WeaviateConfig{
			// 8081 is the host port Compose publishes for Weaviate; 8080 is
			// this API's own port, so defaulting there would make the server
			// probe itself for a vector index.
			Host:   getEnv("WEAVIATE_HOST", "localhost:8081"),
			Scheme: getEnv("WEAVIATE_SCHEME", "http"),
		},
	}, nil
}

func isDevelopment(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}

func (db DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvDuration parses a Go duration string, falling back on unset or invalid
// values so a typo cannot take the process down at start-up.
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

// getEnvList parses a comma-separated environment variable into a string slice,
// falling back to the supplied defaults when the variable is unset or empty.
func getEnvList(key string, fallback []string) []string {
	raw := getEnv(key, "")
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}
	return values
}
