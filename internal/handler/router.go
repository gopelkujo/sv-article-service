package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gopelkujo/sv-article-service/internal/middleware"
)

// NewRouter builds the chi router with middleware, health probes, and article routes.
// corsOrigins is the allowlist used by CORS middleware (e.g. the Vercel frontend URL).
func NewRouter(
	articleHandler *ArticleHandler,
	healthHandler *HealthHandler,
	logger *slog.Logger,
	corsOrigins []string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.CORS(corsOrigins))
	r.Use(middleware.Recover(logger))
	r.Use(middleware.Logger(logger))

	r.Get("/healthz", healthHandler.Liveness)
	r.Get("/readyz", healthHandler.Readiness)

	r.Route("/article", func(r chi.Router) {
		r.Post("/", articleHandler.Create)
		r.Get("/{limit}/{offset}", articleHandler.List)
		r.Get("/{id}", articleHandler.GetByID)
		r.Put("/{id}", articleHandler.Update)
		r.Patch("/{id}", articleHandler.Update)
		r.Delete("/{id}", articleHandler.Delete)
	})

	return r
}
