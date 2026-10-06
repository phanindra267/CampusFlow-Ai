package store

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type MemoryStore struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	enabled bool
	rate    rate.Limit
	burst   int
	idleTTL time.Duration
	gcInt   time.Duration
	lastGC  time.Time
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewMemoryStore(perMinute, burst int) *MemoryStore {
	if perMinute <= 0 {
		return &MemoryStore{enabled: false, buckets: map[string]*bucket{}}
	}
	b := burst
	if b <= 0 {
		b = perMinute
	}
	return &MemoryStore{
		enabled: true,
		buckets: make(map[string]*bucket),
		rate:    rate.Limit(float64(perMinute) / 60.0),
		burst:   b,
		idleTTL: 10 * time.Minute,
		gcInt:   time.Minute,
	}
}

func (s *MemoryStore) Enabled() bool {
	return s != nil && s.enabled
}

func (s *MemoryStore) Allow(key string, _ rate.Limit, _ int, now time.Time) (bool, time.Duration) {
	if !s.Enabled() {
		return true, 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweep(now)
	entry, ok := s.buckets[key]
	if !ok {
		entry = &bucket{limiter: rate.NewLimiter(s.rate, s.burst)}
		s.buckets[key] = entry
	}
	entry.lastSeen = now
	res := entry.limiter.ReserveN(now, 1)
	if !res.OK() {
		return false, s.burstInterval()
	}
	if delay := res.DelayFrom(now); delay > 0 {
		res.CancelAt(now)
		return false, delay
	}
	return true, 0
}

func (s *MemoryStore) Size() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.buckets)
}

func (s *MemoryStore) burstInterval() time.Duration {
	if s.rate <= 0 {
		return time.Second
	}
	return time.Duration(float64(s.burst) / float64(s.rate) * float64(time.Second))
}

func (s *MemoryStore) sweep(now time.Time) {
	if now.Sub(s.lastGC) < s.gcInt {
		return
	}
	s.lastGC = now
	for k, e := range s.buckets {
		if now.Sub(e.lastSeen) > s.idleTTL {
			delete(s.buckets, k)
		}
	}
}
