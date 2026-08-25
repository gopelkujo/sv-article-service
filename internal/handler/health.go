package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// HealthHandler exposes liveness and readiness probes.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler constructs a HealthHandler that can ping db for readiness.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness handles GET /healthz — process is up (no dependency checks).
func (h *HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Readiness handles GET /readyz — process can serve traffic (MySQL reachable).
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable", nil)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}
