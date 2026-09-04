package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"arsen/pkg/cqrs"
	"arsen/pkg/middleware"
	"arsen/pkg/response"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type sessionResponse struct {
	Self         string `json:"self"`
	Kind         string `json:"kind"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
}

// Handler holds the CQRS command buses for auth endpoints.
type Handler struct {
	loginBus   *cqrs.CommandBus[LoginCommand, *LoginResult]
	refreshBus *cqrs.CommandBus[RefreshTokenCommand, *RefreshTokenResult]
}

// NewHandler creates a new auth Handler.
func NewHandler(
	loginBus *cqrs.CommandBus[LoginCommand, *LoginResult],
	refreshBus *cqrs.CommandBus[RefreshTokenCommand, *RefreshTokenResult],
) *Handler {
	return &Handler{loginBus: loginBus, refreshBus: refreshBus}
}

// RegisterRoutes mounts auth endpoints onto the given router with per-endpoint
// rate limiting.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.With(middleware.RateLimit(10.0/60.0, 20)).Post("/api/sessions", h.handleLogin)
	r.With(middleware.RateLimit(10.0/60.0, 20)).Post("/api/tokens", h.handleRefresh)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body.", nil)
		return
	}

	result, err := h.loginBus.Dispatch(r.Context(), LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	sessionResp := sessionResponse{
		Self:         "/api/sessions/current",
		Kind:         "Session",
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
	}
	response.JSON(w, r, http.StatusOK, sessionResp)
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, "Invalid request body.", nil)
		return
	}

	result, err := h.refreshBus.Dispatch(r.Context(), RefreshTokenCommand{
		Token: req.RefreshToken,
	})
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	resp := sessionResponse{
		Self:         "/api/tokens",
		Kind:         "TokenPair",
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
	}
	response.JSON(w, r, http.StatusOK, resp)
}
