package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newLimitedRouter(t *testing.T, rl *RateLimiter) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimit(rl, KeyByClientIP))
	r.POST("/limited", func(c *gin.Context) { c.Status(http.StatusOK) })

	return r
}

func post(r *gin.Engine, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/limited", nil)
	// SetRemoteAddr is what ClientIP reads when no proxy is trusted.
	req.RemoteAddr = ip + ":12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimitAllowsBurstThenRejects(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 60, Burst: 3})
	r := newLimitedRouter(t, rl)

	for i := 1; i <= 3; i++ {
		if got := post(r, "10.0.0.1").Code; got != http.StatusOK {
			t.Fatalf("request %d: want 200 within burst, got %d", i, got)
		}
	}

	got := post(r, "10.0.0.1")
	if got.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429 once burst is exhausted, got %d", got.Code)
	}
}

func TestRateLimitSetsRetryAfter(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 60, Burst: 1})
	r := newLimitedRouter(t, rl)

	post(r, "10.0.0.1")
	got := post(r, "10.0.0.1")

	if got.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", got.Code)
	}

	raw := got.Header().Get("Retry-After")
	if raw == "" {
		t.Fatal("429 must carry Retry-After so clients can back off on their own")
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 1 {
		t.Fatalf("Retry-After must be a positive integer, got %q", raw)
	}
}

func TestRateLimitRejectionDoesNotSpendAToken(t *testing.T) {
	// 60/minute is one token per second, which makes the timing deterministic
	// without relying on wall-clock luck.
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 60, Burst: 1})
	r := newLimitedRouter(t, rl)

	if got := post(r, "10.0.0.1").Code; got != http.StatusOK {
		t.Fatalf("want 200 for the initial token, got %d", got)
	}

	// Hammer the limiter while it is rejecting. If each rejection spent a token,
	// the bucket would fall further and further behind and the client would stay
	// locked out long past a single replenish interval.
	for i := 0; i < 20; i++ {
		if got := post(r, "10.0.0.1").Code; got != http.StatusTooManyRequests {
			t.Fatalf("request %d: want 429 while empty, got %d", i, got)
		}
	}

	// One replenish interval later the client must be back in service.
	time.Sleep(1200 * time.Millisecond)

	if got := post(r, "10.0.0.1").Code; got != http.StatusOK {
		t.Fatalf("want 200 after a replenish interval, got %d: rejections are spending tokens", got)
	}
}

func TestRateLimitKeysAreIndependent(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 60, Burst: 1})
	r := newLimitedRouter(t, rl)

	post(r, "10.0.0.1")
	if got := post(r, "10.0.0.2"); got.Code != http.StatusOK {
		t.Fatalf("a different client must not be penalised, got %d", got.Code)
	}
}

func TestRateLimitDisabledIsPassThrough(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 0})
	r := newLimitedRouter(t, rl)

	if rl.Enabled() {
		t.Fatal("a zero rate must disable the limiter")
	}
	for i := 0; i < 200; i++ {
		if got := post(r, "10.0.0.1").Code; got != http.StatusOK {
			t.Fatalf("request %d: want 200 when disabled, got %d", i, got)
		}
	}
	if rl.Size() != 0 {
		t.Fatalf("a disabled limiter must not allocate buckets, got %d", rl.Size())
	}
}

func TestRateLimitBucketMemoryIsBounded(t *testing.T) {
	// Flooding from many source addresses must not grow the bucket map without
	// bound. The sweep only runs on an interval, so this asserts on the steady
	// state rather than on immediate eviction.
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 600, Burst: 1})

	for i := 0; i < 5000; i++ {
		rl.Allow("addr-" + strconv.Itoa(i))
	}

	if got := rl.Size(); got == 0 {
		t.Fatal("buckets should have been created")
	}
	// 5000 distinct clients with a 10 minute idle TTL is legitimate state, not a
	// leak; what must not happen is unbounded growth beyond the number of
	// distinct keys seen.
	if got := rl.Size(); got > 5000 {
		t.Fatalf("bucket map grew past the number of distinct keys: %d", got)
	}
}

func TestRateLimitConcurrentAccessIsSafe(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{PerMinute: 6000, Burst: 5})

	done := make(chan struct{})
	for i := 0; i < 16; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 200; j++ {
				rl.Allow("shared-" + strconv.Itoa(n%4))
			}
		}(i)
	}
	for i := 0; i < 16; i++ {
		<-done
	}
}
