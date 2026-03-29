package middleware

import (
	"net/http"
	"sync"

	"github.com/tunabsdrmz/boxing-gym-management/internal/utils"
	"golang.org/x/time/rate"
)

// RateLimitPerIP enforces an average of rpm requests per minute per client IP.
func RateLimitPerIP(rpm int) func(http.Handler) http.Handler {
	if rpm < 1 {
		rpm = 60
	}
	burst := rpm
	if burst > 60 {
		burst = 60
	}
	lim := rate.Limit(float64(rpm) / 60.0)

	var mu sync.Mutex
	clients := make(map[string]*rate.Limiter)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = xff
			}
			mu.Lock()
			limiter, ok := clients[ip]
			if !ok {
				limiter = rate.NewLimiter(lim, burst)
				clients[ip] = limiter
			}
			mu.Unlock()
			if !limiter.Allow() {
				utils.RateLimitExceededResponse(w, r, "60")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
