package middleware

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/campuscare/api/internal/delivery/http/middleware/store"
	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter supports in-memory and pluggable stores (e.g., Redis) for
// distributed rate limiting across replicas. Defaults to memory for
// backward compatibility.
type RateLimiter struct {
	enabled bool
	rate    rate.Limit
	burst   int
	store   rateLimitStore
}

type rateLimitStore interface {
	Allow(key string, r rate.Limit, burst int, now time.Time) (bool, time.Duration)
	Enabled() bool
	Size() int
}

// RateLimitConfig describes one limiter.
type RateLimitConfig struct {
	// PerMinute is the sustained rate. PerSecond is derived from it.
	PerMinute int
	// Burst is how many requests may arrive at once before the sustained rate
	// applies. It also sets the ceiling on how much a single client can
	// front-load.
	Burst int
}

// NewRateLimiter builds a limiter allowing perMinute sustained requests with the
// given burst. Non-positive values disable limiting, in which case Allow always
// returns true.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	return NewRateLimiterWithStore(cfg, nil)
}

// NewRateLimiterWithStore builds a limiter with an explicit store. If store is
// nil, an in-memory store is created.
func NewRateLimiterWithStore(cfg RateLimitConfig, s rateLimitStore) *RateLimiter {
	if cfg.PerMinute <= 0 {
		return &RateLimiter{enabled: false}
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = cfg.PerMinute
	}
	r := rate.Limit(float64(cfg.PerMinute) / 60.0)
	if s == nil {
		s = store.NewMemoryStore(cfg.PerMinute, burst)
	}
	return &RateLimiter{
		enabled: s.Enabled(),
		rate:    r,
		burst:   burst,
		store:   s,
	}
}

// Enabled reports whether this limiter does anything.
func (rl *RateLimiter) Enabled() bool {
	return rl != nil && rl.enabled
}

// Allow consumes one token for key, reporting whether the request may proceed
// and, when it may not, how long the caller should wait.
func (rl *RateLimiter) Allow(key string) (bool, time.Duration) {
	if !rl.Enabled() {
		return true, 0
	}
	if rl.store != nil {
		return rl.store.Allow(key, rl.rate, rl.burst, time.Now())
	}
	return true, 0
}

// Size reports how many buckets are currently tracked. Exists for tests and for
// the memory-bound assertion; not called on the request path.
func (rl *RateLimiter) Size() int {
	if rl == nil || rl.store == nil {
		return 0
	}
	return rl.store.Size()
}

// KeyFunc derives the rate-limit bucket for a request.
type KeyFunc func(c *gin.Context) string

// KeyByClientIP buckets by the direct peer address.
//
// This deliberately does not read X-Forwarded-For. Unless the deployment's
// trusted-proxy list is configured correctly, honouring that header lets a
// caller mint an unlimited number of buckets simply by varying it, which turns
// the limiter into a no-op. Configuring SetTrustedProxies in server.go and using
// c.ClientIP() is the way to be correct behind a load balancer.
func KeyByClientIP(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimit rejects requests that exceed the limiter's budget.
//
// A rejected request gets 429 and a Retry-After header so a well-behaved client
// backs off on its own rather than retrying into the wall.
func RateLimit(rl *RateLimiter, keyFunc KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Enabled() {
			c.Next()
			return
		}

		allowed, retryAfter := rl.Allow(keyFunc(c))
		if allowed {
			c.Next()
			return
		}

		seconds := int(math.Ceil(retryAfter.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		c.Writer.Header().Set("Retry-After", strconv.Itoa(seconds))

		response.Error(c, http.StatusTooManyRequests,
			"too many requests, please retry shortly", nil)
		c.Abort()
	}
}
