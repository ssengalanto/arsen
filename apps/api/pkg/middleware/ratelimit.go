package middleware

import (
	"net"
	"net/http"
	"sync"

	"arsen/pkg/response"

	"golang.org/x/time/rate"
)

// RateLimit returns a middleware that enforces a per-IP token-bucket rate
// limit. Each unique client IP gets its own limiter configured with the
// given requests-per-second rate and burst size.
//
// When the limit is exceeded the middleware responds with a 429 status and
// an RFC 9457 problem detail body via response.TooManyRequests.
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	var clients sync.Map // map[string]*rate.Limiter

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			val, _ := clients.LoadOrStore(ip, rate.NewLimiter(rate.Limit(rps), burst))
			limiter := val.(*rate.Limiter)

			if !limiter.Allow() {
				response.TooManyRequests(w, r, "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the client IP from r.RemoteAddr, stripping the port.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might already be a bare IP (no port).
		return r.RemoteAddr
	}
	return host
}
