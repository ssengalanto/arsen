package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"arsen/pkg/cqrs"
	"arsen/pkg/response"
)

type healthResponse struct {
	Self   string `json:"self"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

type readinessResponse struct {
	Self   string            `json:"self"`
	Kind   string            `json:"kind"`
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

type Handler struct {
	readinessBus *cqrs.QueryBus[ReadinessQuery, *ReadinessResult]
}

func NewHandler(readinessBus *cqrs.QueryBus[ReadinessQuery, *ReadinessResult]) *Handler {
	return &Handler{readinessBus: readinessBus}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.handleHealth)
	r.Get("/readyz", h.handleReadiness)
}

// @Summary Liveness check
// @Description Returns 200 if the server process is running
// @Tags health
// @Produce json
// @Success 200 {object} healthResponse
// @Router /healthz [get]
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, r, http.StatusOK, healthResponse{
		Self:   "/healthz",
		Kind:   "Health",
		Status: "ok",
	})
}

// @Summary Readiness check
// @Description Returns 200 if the server is ready to serve traffic (database is reachable)
// @Tags health
// @Produce json
// @Success 200 {object} readinessResponse
// @Failure 503 {object} readinessResponse
// @Router /readyz [get]
func (h *Handler) handleReadiness(w http.ResponseWriter, r *http.Request) {
	result, err := h.readinessBus.Ask(r.Context(), ReadinessQuery{})
	if err != nil {
		response.JSON(w, r, http.StatusServiceUnavailable, readinessResponse{
			Self:   "/readyz",
			Kind:   "Readiness",
			Status: "unavailable",
			Checks: map[string]string{"database": "error"},
		})
		return
	}

	if !result.Ready {
		response.JSON(w, r, http.StatusServiceUnavailable, readinessResponse{
			Self:   "/readyz",
			Kind:   "Readiness",
			Status: "unavailable",
			Checks: map[string]string{"database": result.Database},
		})
		return
	}

	response.JSON(w, r, http.StatusOK, readinessResponse{
		Self:   "/readyz",
		Kind:   "Readiness",
		Status: "ready",
	})
}
