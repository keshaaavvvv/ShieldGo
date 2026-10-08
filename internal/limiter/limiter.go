package limiter

import (
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastUpdate time.Time
}

type Limiter struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	buckets map[string]*bucket
}

func New(rate, burst float64) *Limiter {
	return &Limiter{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*bucket),
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{
			tokens:     l.burst - 1,
			lastUpdate: now,
		}
		return true
	}

	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * l.rate

	if b.tokens > l.burst {
		b.tokens = l.burst
	}

	b.lastUpdate = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

func (l *Limiter) Cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for key, b := range l.buckets {
		if b.lastUpdate.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}
