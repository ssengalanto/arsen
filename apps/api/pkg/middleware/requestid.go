package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
)

// contextKey is an unexported type used for context keys in this package
// to avoid collisions with keys defined in other packages.
type contextKey string

const requestIDKey contextKey = "request_id"

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestID is a middleware that assigns a unique identifier to each request.
// If the incoming request already carries an X-Request-ID header, that value
// is reused; otherwise a new v4 UUID is generated from crypto/rand.
// The ID is stored in the request context, added to the slog logger in
// context, and set as the X-Request-ID response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newUUID()
		}

		// Store request ID in context.
		ctx := context.WithValue(r.Context(), requestIDKey, id)

		// Add request_id to the slog logger in context so downstream
		// handlers that use slog.InfoContext (etc.) include it automatically.
		logger := slog.Default().With(slog.String("request_id", id))
		ctx = context.WithValue(ctx, slogLoggerKey, logger)

		// Set response header.
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// slogLoggerKey is the context key for the enriched slog.Logger.
const slogLoggerKey contextKey = "slog_logger"

// LoggerFromContext returns the slog.Logger stored in the context by the
// RequestID middleware. Falls back to slog.Default() if none is present.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(slogLoggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// newUUID generates a version 4 UUID string from 16 random bytes.
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand should never fail on supported platforms; if it does
		// we have bigger problems, so panic is appropriate.
		panic(fmt.Sprintf("middleware: failed to generate UUID: %v", err))
	}
	// Set version (4) and variant (RFC 4122) bits.
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
