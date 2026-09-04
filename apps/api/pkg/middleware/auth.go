package middleware

import (
	"context"
	"net/http"
	"strings"

	"arsen/pkg/jwt"
	"arsen/pkg/response"
)

// userIDKey is an unexported context key for storing the authenticated user ID.
type userIDKey struct{}

// GetUserID retrieves the authenticated user ID from the context.
// It returns the user ID and true if present, or an empty string and false otherwise.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey{}).(string)
	return id, ok
}

// Auth returns an HTTP middleware that validates JWT bearer tokens.
// On success it injects the token's subject (user ID) into the request context.
func Auth(jwtService *jwt.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.Unauthorized(w, r, "Missing or malformed authorization token.")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtService.ValidateToken(tokenStr)
			if err != nil {
				response.Unauthorized(w, r, "Invalid or expired token.")
				return
			}

			userID, err := claims.GetSubject()
			if err != nil || userID == "" {
				response.Unauthorized(w, r, "Invalid or expired token.")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
