package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/lcmetzger/rate_limiter/internal/rl"
)

func RateLimiterMiddleware(rateLimiter *rl.RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || ip == "" {
			ip = r.RemoteAddr
		}
		if ip == "" {
			http.Error(w, "IP address not found", http.StatusBadRequest)
			return
		}
		token := r.Header.Get("API_KEY")

		var key string
		var limit int

		if strings.ToUpper(rateLimiter.RateLimitType) == "IP" {
			key = "ip:" + ip
			limit = rateLimiter.IpRateLimit
		}

		if strings.ToUpper(rateLimiter.RateLimitType) == "TOKEN" {
			key = "token:" + token
			limit = rateLimiter.TokenRateLimit
		}

		if rateLimiter.IsBlocked(key) || !rateLimiter.Allow(key, limit) {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
