package api

import (
	"encoding/json"
	"net/http"

	"github.com/croncheck/internal/metrics"
)

// MetricsHandler exposes job metrics over HTTP.
type MetricsHandler struct {
	collector *metrics.Collector
}

// NewMetricsHandler creates a MetricsHandler backed by the given Collector.
func NewMetricsHandler(c *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{collector: c}
}

// RegisterRoutes attaches metrics endpoints to mux.
func (h *MetricsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", h.handleList)
	mux.HandleFunc("/metrics/", h.handleGet)
}

// handleList returns metrics for all jobs as a JSON array.
func (h *MetricsHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := h.collector.All()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(all); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// handleGet returns metrics for a single job identified by the URL suffix.
func (h *MetricsHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobID := r.URL.Path[len("/metrics/"):]
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	m, ok := h.collector.Get(jobID)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
