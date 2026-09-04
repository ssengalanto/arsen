package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code
// written by downstream handlers.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code before delegating to the
// underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write ensures the status code is recorded even when WriteHeader is
// never called explicitly (defaults to 200).
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter so callers that need access
// to optional interfaces (http.Flusher, http.Hijacker, etc.) can type-assert
// through.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logging is a middleware that emits a structured log line for every
// completed HTTP request. It logs method, path, status, duration (ms),
// and request_id (via the logger enriched by the RequestID middleware).
//
// Log levels are chosen by status code range:
//   - 2xx/3xx -> Info
//   - 4xx     -> Warn
//   - 5xx     -> Error
//
// The Authorization header value is redacted to avoid leaking credentials.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		logger := LoggerFromContext(r.Context())

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.statusCode),
			slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
		}

		// Redact Authorization header: log its presence but mask the value.
		if r.Header.Get("Authorization") != "" {
			attrs = append(attrs, slog.String("authorization", "[REDACTED]"))
		}

		msg := r.Method + " " + r.URL.Path
		args := attrsToArgs(attrs)

		switch {
		case rw.statusCode >= 500:
			logger.Error(msg, args...)
		case rw.statusCode >= 400:
			logger.Warn(msg, args...)
		default:
			logger.Info(msg, args...)
		}
	})
}

// attrsToArgs converts a slice of slog.Attr to a []any suitable for passing
// to slog.Logger methods.
func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return args
}
