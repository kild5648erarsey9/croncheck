package api

import (
	"encoding/json"
	"net/http"

	"github.com/croncheck/internal/healthcheck"
)

// HealthHandler serves the /health endpoint.
type HealthHandler struct {
	checker *healthcheck.Checker
}

// NewHealthHandler creates a HealthHandler backed by the given Checker.
func NewHealthHandler(c *healthcheck.Checker) *HealthHandler {
	return &HealthHandler{checker: c}
}

// ServeHTTP handles GET /health and returns a JSON health status.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.checker.Run()

	code := http.StatusOK
	if !status.Healthy {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// RegisterHealthRoute mounts the health handler on the given mux.
func RegisterHealthRoute(mux *http.ServeMux, c *healthcheck.Checker) {
	h := NewHealthHandler(c)
	mux.Handle("/health", h)
}
