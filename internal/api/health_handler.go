package api

import (
	"encoding/json"
	"net/http"

	"github.com/croncheck/internal/healthcheck"
)

// HealthHandler serves health check status over HTTP.
type HealthHandler struct {
	checker *healthcheck.Checker
}

// NewHealthHandler creates a new HealthHandler backed by the given Checker.
func NewHealthHandler(checker *healthcheck.Checker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP handles GET /health requests.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	results := h.checker.Run()

	type response struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks,omitempty"`
	}

	resp := response{
		Status: "ok",
		Checks: make(map[string]string),
	}

	overall := http.StatusOK
	for name, err := range results {
		if err != nil {
			resp.Status = "degraded"
			resp.Checks[name] = err.Error()
			overall = http.StatusServiceUnavailable
		} else {
			resp.Checks[name] = "ok"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(overall)
	_ = json.NewEncoder(w).Encode(resp)
}

// RegisterHealthRoute registers the health endpoint on the provided mux.
func (h *HealthHandler) RegisterHealthRoute(mux *http.ServeMux) {
	mux.Handle("/health", h)
}
