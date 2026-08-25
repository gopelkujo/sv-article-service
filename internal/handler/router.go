package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gopelkujo/sv-article-service/internal/middleware"
)

// NewRouter builds the chi router with middleware, health probes, and article routes.
func NewRouter(articleHandler *ArticleHandler, healthHandler *HealthHandler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
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
