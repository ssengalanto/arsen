package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"arsen/pkg/config"
	"arsen/pkg/jwt"
	"arsen/pkg/middleware"
	"arsen/pkg/response"
)

// Server wraps a chi router and the underlying http.Server.
type Server struct {
	Router     chi.Router
	HTTP       *http.Server
	JWTService *jwt.Service
}

// New creates a Server with global middleware and error handlers wired up.
func New(cfg *config.Config, jwtService *jwt.Service) *Server {
	r := chi.NewRouter()

	// Global middleware — order matters.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS([]string{"*"})) // TODO: make configurable per environment

	// Serve Swagger UI in non-production environments.
	if cfg.Env != "prod" {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	// RFC 9457 problem detail for unmatched routes / methods.
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, r, "The requested resource was not found.")
	}))

	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.WriteProblem(w, response.ProblemDetail{
			Type:     "about:blank",
			Title:    "Method Not Allowed",
			Status:   http.StatusMethodNotAllowed,
			Detail:   fmt.Sprintf("Method %s is not allowed for this resource.", r.Method),
			Instance: r.URL.Path,
		})
	}))

	httpServer := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: r,
	}

	return &Server{
		Router:     r,
		HTTP:       httpServer,
		JWTService: jwtService,
	}
}
