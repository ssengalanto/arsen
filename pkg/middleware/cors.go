package middleware

import (
	"net/http"
)

// CORS returns a middleware that sets Cross-Origin Resource Sharing headers.
// If allowedOrigins contains "*", every origin is permitted. Otherwise the
// request's Origin header is checked against the provided list and the
// Access-Control-Allow-Origin header is set only when a match is found.
//
// Preflight OPTIONS requests are handled by returning a 204 with the
// appropriate CORS headers.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := false
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		origins[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, ok := origins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
