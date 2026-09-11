package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	requests map[string][]time.Time
}

var limiter = &rateLimiter{requests: map[string][]time.Time{}}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			limiter.mu.Lock()
			defer limiter.mu.Unlock()

			cutoff := time.Now().Add(-window)
			entries := limiter.requests[key]
			filtered := entries[:0]
			for _, ts := range entries {
				if ts.After(cutoff) {
					filtered = append(filtered, ts)
				}
			}
			if len(filtered) >= limit {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			filtered = append(filtered, time.Now())
			limiter.requests[key] = filtered
			next.ServeHTTP(w, r)
		})
	}
}
