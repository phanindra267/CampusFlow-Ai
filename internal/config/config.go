package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env       string
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Log       LogConfig
	Weaviate  WeaviateConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CORSOrigins  []string
	// TrustedProxies lists the networks whose X-Forwarded-For may be believed.
	// It is empty by default, which makes the client IP the direct peer: safe
	// against bucket-flooding through a spoofed header, and the thing an
	// operator must configure explicitly when a load balancer is introduced.
	TrustedProxies []string
	// HTTP2 enables HTTP/2 on the listener.
	//
	// With TLS configured it means ALPN negotiation for "h2", which is what a
	// real deployment wants. Without TLS it means h2c, the cleartext upgrade
	// path, which exists so the API can serve multiplexed requests to clients
	// such as gRPC gateways and Go's own client without forcing a self-signed
	// certificate into local development.
	HTTP2 bool
	// TLSCertFile and TLSKeyFile point at a PEM certificate and its private key.
	// When both are set the server terminates TLS itself, which is what enables
	// ALPN-negotiated HTTP/2. In production the presence of TLS is required; see
	// the validation in Load.
	TLSCertFile string
	TLSKeyFile  string
}

// TLSConfigured reports whether the server has been given a certificate to
// terminate TLS with.
func (s ServerConfig) TLSConfigured() bool {
	return s.TLSCertFile != "" && s.TLSKeyFile != ""
}

// RateLimitConfig bounds how fast unauthenticated authentication endpoints may
// be called.
type RateLimitConfig struct {
	// AuthPerMinute is the sustained rate for the credential endpoints. Zero
	// disables limiting.
	AuthPerMinute int
	// AuthBurst is how many authentication requests may arrive at once.
	AuthBurst int
}

type DatabaseConfig struct {
	Host             string
	Port             string
	User             string
	Password         string
	Name             string
	SSLMode          string
	MaxConns         int
	MinConns         int
	MaxConnLifetime  time.Duration
	MaxConnIdleTime  time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout   time.Duration
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

	certFile := strings.TrimSpace(getEnv("TLS_CERT_FILE", ""))
	keyFile := strings.TrimSpace(getEnv("TLS_KEY_FILE", ""))

	// Half-configured TLS is the dangerous state: an operator who sets only the
	// certificate would otherwise start a server that silently serves plain HTTP
	// while believing it is serving TLS. Reject the mismatch instead.
	if (certFile == "") != (keyFile == "") {
		return nil, fmt.Errorf(
			"TLS_CERT_FILE and TLS_KEY_FILE must be set together (got cert=%q key=%q)",
			certFile, keyFile)
	}

	server := ServerConfig{
		Port:           getEnv("PORT", "8080"),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		CORSOrigins:    getEnvList("CORS_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxies: getEnvList("TRUSTED_PROXIES", nil),
		// HTTP/2 is on by default. Turning it off is only useful when a proxy in
		// front insists on HTTP/1.1, which is rare enough to need an explicit
		// opt-out rather than an opt-in.
		HTTP2:       getEnvBool("HTTP2_ENABLED", true),
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
	}

	if !isDevelopment(env) && !server.TLSConfigured() {
		return nil, fmt.Errorf(
			"TLS_CERT_FILE and TLS_KEY_FILE must be set when APP_ENV=%q; "+
				"serving this API over plain HTTP in production would put credentials "+
				"on the wire in the clear", env)
	}

	return &Config{
		Env:    env,
		Server: server,
		Database: DatabaseConfig{
			Host:             getEnv("DB_HOST", "localhost"),
			Port:             getEnv("DB_PORT", "5432"),
			User:             getEnv("DB_USER", "postgres"),
			Password:         getEnv("DB_PASSWORD", "postgres"),
			Name:             getEnv("DB_NAME", "campuscare"),
			SSLMode:          getEnv("DB_SSLMODE", "disable"),
			MaxConns:         getEnvInt("DB_MAX_CONNS", 20),
			MinConns:         getEnvInt("DB_MIN_CONNS", 2),
			MaxConnLifetime:  getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:  getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvDuration("DB_HEALTHCHECK_PERIOD", time.Minute),
			ConnectTimeout:   getEnvDuration("DB_CONNECT_TIMEOUT", 5*time.Second),
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
		RateLimit: RateLimitConfig{
			// bcrypt makes every login attempt cost real CPU, so an unthrottled
			// login endpoint is both a brute-force surface and a cheap way to
			// saturate the CPU. Ten attempts a minute from one address is
			// generous for a person typing a password and unusable for a
			// guessing script.
			AuthPerMinute: getEnvInt("AUTH_RATE_LIMIT_PER_MINUTE", 10),
			AuthBurst:     getEnvInt("AUTH_RATE_LIMIT_BURST", 5),
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

// getEnv returns the environment variable's value, or fallback when it is unset
// or blank.
//
// Treating a blank value as unset matters more than it looks. A variable that is
// declared but empty is common: `APP_ENV= docker compose run ...`, a .env line
// with nothing after the equals sign, a CI step that exports a variable it
// could not resolve. Reading that as the literal empty string would make
// isDevelopment("") false and push the process into the production branch, where
// it demands a TLS certificate and a full-length JWT secret and then refuses to
// start. The other helpers in this file already fold blank into unset, so this
// makes the behaviour uniform rather than special.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
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

// getEnvInt parses a non-negative integer environment variable, falling back on
// unset, unparseable or negative values. A negative rate is meaningless and
// would otherwise disable limiting by accident, so it is treated as unset.
func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

// getEnvBool parses a boolean environment variable, falling back on unset or
// unparseable values. Anything other than the explicit affirmative strings
// counts as false, so a typo such as HTTP2_ENABLED=enbale disables the feature
// rather than being silently read as "on" by strconv.ParseBool's tolerance for
// "1"/"t"/"yes".
func getEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
