package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/campuscare/api/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter is an in-process token bucket keyed by an arbitrary string.
//
// Deliberately in-process rather than backed by Redis. At this scale it needs to
// do exactly one job: make online password guessing and credential stuffing
// impractical on a handful of unauthenticated routes. A shared store would add
// a network hop and an availability dependency to the login path, which is the
// one path that must keep working when everything else is down. The cost of this
// choice is stated plainly: limits are per instance, so a deployment fronted by
// N replicas allows N times the configured rate. That is an acceptable trade for
// brute-force resistance and is documented as such.
//
// Every method is safe for concurrent use, and the bucket map is bounded so a
// flood from many source addresses cannot grow it without limit.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket

	// enabled is false when constructed with a non-positive rate, which is how
	// an operator disables limiting deliberately rather than by accident.
	enabled bool

	rate  rate.Limit
	burst int

	// idleTTL is how long a bucket survives without being used. A client that
	// stops sending is forgotten rather than remembered forever, which bounds
	// memory against a rotating-source-address flood.
	idleTTL time.Duration
	// gcInterval is the minimum gap between sweeps, so the sweep itself cannot
	// become a hot spot under load.
	gcInterval time.Duration
	lastGC     time.Time
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
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
// returns true and no buckets are ever created.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.PerMinute <= 0 {
		return &RateLimiter{enabled: false, buckets: map[string]*bucket{}}
	}

	burst := cfg.Burst
	if burst <= 0 {
		// Default burst is a small multiple of the minute rate: high enough for a
		// legitimate burst (a page firing several parallel requests on load), low
		// enough that a brute-force run cannot get far ahead of the token flow.
		burst = cfg.PerMinute
	}

	return &RateLimiter{
		enabled:    true,
		buckets:    make(map[string]*bucket),
		rate:       rate.Limit(float64(cfg.PerMinute) / 60.0),
		burst:      burst,
		idleTTL:    10 * time.Minute,
		gcInterval: time.Minute,
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

	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.sweep(now)

	entry, ok := rl.buckets[key]
	if !ok {
		entry = &bucket{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.buckets[key] = entry
	}
	entry.lastSeen = now

	reservation := entry.limiter.ReserveN(now, 1)
	if !reservation.OK() {
		// The burst is at least 1, so ReserveN can only fail if the caller
		// asked for more tokens than the bucket can ever hold.
		return false, rl.burstInterval()
	}
		if delay := reservation.DelayFrom(now); delay > 0 {
			// Give the token back: a rejected request must not also spend a token,
			// otherwise a client that retries quickly keeps itself locked out.
			reservation.CancelAt(now)
			return false, delay
		}

	return true, 0
}

// burstInterval is the time one full burst would take to drain, used when
// ReserveN cannot report a useful delay.
func (rl *RateLimiter) burstInterval() time.Duration {
	if rl.rate <= 0 {
		return time.Second
	}
	return time.Duration(float64(rl.burst) / float64(rl.rate) * float64(time.Second))
}

// sweep drops buckets that have been idle longer than idleTTL. The caller must
// hold rl.mu.
func (rl *RateLimiter) sweep(now time.Time) {
	if now.Sub(rl.lastGC) < rl.gcInterval {
		return
	}
	rl.lastGC = now

	for key, entry := range rl.buckets {
		if now.Sub(entry.lastSeen) > rl.idleTTL {
			delete(rl.buckets, key)
		}
	}
}

// Size reports how many buckets are currently tracked. Exists for tests and for
// the memory-bound assertion; not called on the request path.
func (rl *RateLimiter) Size() int {
	if rl == nil {
		return 0
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
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
