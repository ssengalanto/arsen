package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"arsen/pkg/cqrs"
	"arsen/pkg/middleware"
	"arsen/pkg/response"
)

// Request/response DTOs.

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyRequest struct {
	Token string `json:"token"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type userResponse struct {
	Self          string `json:"self"`
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	CreatedAt     string `json:"createdAt"`
	Message       string `json:"message,omitempty"`
}

type acknowledgmentResponse struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Handler holds CQRS command buses for user endpoints.
type Handler struct {
	registerBus           *cqrs.CommandBus[RegisterCommand, *RegisterResult]
	verifyEmailBus        *cqrs.CommandBus[VerifyEmailCommand, *VerifyEmailResult]
	resendVerificationBus *cqrs.CommandBus[ResendVerificationCommand, cqrs.Unit]
}

// NewHandler creates a new user Handler.
func NewHandler(
	registerBus *cqrs.CommandBus[RegisterCommand, *RegisterResult],
	verifyEmailBus *cqrs.CommandBus[VerifyEmailCommand, *VerifyEmailResult],
	resendVerificationBus *cqrs.CommandBus[ResendVerificationCommand, cqrs.Unit],
) *Handler {
	return &Handler{
		registerBus:           registerBus,
		verifyEmailBus:        verifyEmailBus,
		resendVerificationBus: resendVerificationBus,
	}
}

// RegisterRoutes mounts user endpoints onto the given router with per-endpoint
// rate limiting.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/users", func(r chi.Router) {
		r.With(middleware.RateLimit(10.0/60.0, 20)).Post("/", h.handleRegister)
		r.With(middleware.RateLimit(10.0/60.0, 20)).Post("/verify", h.handleVerifyEmail)
		r.With(middleware.RateLimit(5.0/60.0, 10)).Post("/resend-verification", h.handleResendVerification)
	})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body.", nil)
		return
	}

	result, err := h.registerBus.Dispatch(r.Context(), RegisterCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Normalize ConflictError to BadRequest to prevent email enumeration.
		var conflictErr *cqrs.ConflictError
		if errors.As(err, &conflictErr) {
			response.BadRequest(w, r, "Registration failed.", nil)
			return
		}
		response.HandleError(w, r, err)
		return
	}

	userResp := userResponse{
		Self:          "/api/users/" + result.User.ID,
		Kind:          "User",
		ID:            result.User.ID,
		Email:         result.User.Email,
		EmailVerified: false,
		CreatedAt:     result.User.CreatedAt.UTC().Format(time.RFC3339),
		Message:       "Account created. Please check your email to verify your address.",
	}
	response.Created(w, r, "/api/users/"+result.User.ID, userResp)
}

func (h *Handler) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body.", nil)
		return
	}

	result, err := h.verifyEmailBus.Dispatch(r.Context(), VerifyEmailCommand{
		Token: req.Token,
	})
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	userResp := userResponse{
		Self:          "/api/users/" + result.User.ID,
		Kind:          "User",
		ID:            result.User.ID,
		Email:         result.User.Email,
		EmailVerified: true,
		CreatedAt:     result.User.CreatedAt.UTC().Format(time.RFC3339),
		Message:       "Email verified successfully.",
	}
	response.JSON(w, r, http.StatusOK, userResp)
}

func (h *Handler) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body.", nil)
		return
	}

	_, err := h.resendVerificationBus.Dispatch(r.Context(), ResendVerificationCommand{
		Email: req.Email,
	})
	if err != nil {
		// Log the error but always return success to prevent email enumeration.
		slog.Error("resend verification failed", "error", err.Error())
	}

	response.JSON(w, r, http.StatusOK, acknowledgmentResponse{
		Kind:    "Acknowledgment",
		Message: "If an account exists with this email and is not yet verified, a new verification email has been sent.",
	})
}
