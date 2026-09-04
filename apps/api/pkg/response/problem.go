package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"arsen/pkg/cqrs"
)

// ProblemDetail represents an RFC 9457 problem details object.
type ProblemDetail struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail"`
	Instance string       `json:"instance"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// FieldError describes a validation error on a single field.
type FieldError struct {
	Field  string `json:"field"`
	Detail string `json:"detail"`
}

// WriteProblem writes an arbitrary ProblemDetail as an RFC 9457 response.
func WriteProblem(w http.ResponseWriter, p *ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	json.NewEncoder(w).Encode(p) //nolint:errcheck // best-effort write to response
}

// BadRequest writes a 400 problem detail response.
func BadRequest(w http.ResponseWriter, r *http.Request, detail string, errs []FieldError) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Validation Failed",
		Status:   http.StatusBadRequest,
		Detail:   detail,
		Instance: r.URL.Path,
		Errors:   errs,
	})
}

// Unauthorized writes a 401 problem detail response.
func Unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Authentication Failed",
		Status:   http.StatusUnauthorized,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// Forbidden writes a 403 problem detail response.
func Forbidden(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Forbidden",
		Status:   http.StatusForbidden,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// NotFound writes a 404 problem detail response.
func NotFound(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Not Found",
		Status:   http.StatusNotFound,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// Conflict writes a 409 problem detail response.
func Conflict(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Conflict",
		Status:   http.StatusConflict,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// TooManyRequests writes a 429 problem detail response.
func TooManyRequests(w http.ResponseWriter, r *http.Request, detail string) {
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Too Many Requests",
		Status:   http.StatusTooManyRequests,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// InternalError writes a 500 problem detail response. The correlationID is
// included in the detail message so operators can trace the failure.
func InternalError(w http.ResponseWriter, r *http.Request, correlationID string) {
	detail := "an unexpected error occurred"
	if correlationID != "" {
		detail = "an unexpected error occurred (correlation_id: " + correlationID + ")"
	}
	WriteProblem(w, &ProblemDetail{
		Type:     "about:blank",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}

// HandleError translates domain errors from the cqrs package into the
// appropriate RFC 9457 problem detail response.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var notFoundErr *cqrs.NotFoundError
	if errors.As(err, &notFoundErr) {
		NotFound(w, r, err.Error())
		return
	}

	var conflictErr *cqrs.ConflictError
	if errors.As(err, &conflictErr) {
		Conflict(w, r, err.Error())
		return
	}

	var validationErr *cqrs.ValidationError
	if errors.As(err, &validationErr) {
		fieldErrors := make([]FieldError, len(validationErr.Fields))
		for i, fe := range validationErr.Fields {
			fieldErrors[i] = FieldError{Field: fe.Field, Detail: fe.Detail}
		}
		BadRequest(w, r, err.Error(), fieldErrors)
		return
	}

	var unauthorizedErr *cqrs.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		Unauthorized(w, r, err.Error())
		return
	}

	var forbiddenErr *cqrs.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		Forbidden(w, r, err.Error())
		return
	}

	InternalError(w, r, "")
}
