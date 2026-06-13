package middleware

import (
	"sync"
	"time"

	"qvarkk/kvault/internal/httpx"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipRateLimiter keeps a token-bucket limiter per client IP. Stale entries are
// reaped so the map cannot grow unbounded. Suitable for a single-instance,
// self-hosted deployment (state is in-process, not shared across replicas).
type ipRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    rate.Limit
	burst   int
	ttl     time.Duration
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		buckets: make(map[string]*bucket),
		rate:    r,
		burst:   burst,
		ttl:     10 * time.Minute,
	}
	go l.reapLoop()
	return l
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[ip]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.buckets[ip] = b
	}
	b.lastSeen = time.Now()
	return b.limiter.Allow()
}

func (l *ipRateLimiter) reapLoop() {
	ticker := time.NewTicker(l.ttl)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-l.ttl)
		l.mu.Lock()
		for ip, b := range l.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(l.buckets, ip)
			}
		}
		l.mu.Unlock()
	}
}

// AuthRateLimit is the throttle applied to credential endpoints (login,
// register): ~5 requests/minute per IP with a small burst. Returns a single
// shared limiter so callers can apply it to several routes with one budget.
func AuthRateLimit() gin.HandlerFunc {
	return RateLimit(rate.Every(12*time.Second), 5)
}

func RateLimit(r rate.Limit, burst int) gin.HandlerFunc {
	limiter := newIPRateLimiter(r, burst)
	return func(ctx *gin.Context) {
		if !limiter.allow(ctx.ClientIP()) {
			_ = ctx.Error(&httpx.PublicError{
				Err: httpx.ErrTooManyRequests,
				Key: "err.rate_limited",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
