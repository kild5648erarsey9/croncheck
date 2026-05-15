package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/croncheck/internal/job"
	"github.com/yourorg/croncheck/internal/metrics"
)

// MetricsJobHandler combines job registry and metrics collector
// to expose a unified view of job status with metrics.
type MetricsJobHandler struct {
	registry  *job.Registry
	collector *metrics.Collector
}

// NewMetricsJobHandler creates a new MetricsJobHandler.
func NewMetricsJobHandler(r *job.Registry, c *metrics.Collector) *MetricsJobHandler {
	return &MetricsJobHandler{registry: r, collector: c}
}

type jobWithMetrics struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	TotalRuns   int     `json:"total_runs"`
	Failures    int     `json:"failures"`
	AvgDuration float64 `json:"avg_duration_ms"`
}

// ServeHTTP handles GET /jobs/{id}/metrics and GET /jobs/metrics.
func (h *MetricsJobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if a specific job ID is requested: /jobs/{id}/metrics
	path := strings.TrimSuffix(r.URL.Path, "/metrics")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 2 && parts[0] == "jobs" {
		h.handleSingle(w, r, parts[1])
		return
	}

	h.handleAll(w, r)
}

func (h *MetricsJobHandler) handleSingle(w http.ResponseWriter, r *http.Request, id string) {
	j, ok := h.registry.Get(id)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	snap := j.Snapshot()
	m, err := h.collector.Get(id)
	if err != nil {
		http.Error(w, "metrics not found", http.StatusNotFound)
		return
	}

	result := jobWithMetrics{
		ID:          snap.ID,
		Status:      snap.Status,
		TotalRuns:   m.TotalRuns,
		Failures:    m.Failures,
		AvgDuration: m.AvgDurationMs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *MetricsJobHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	jobs := h.registry.All()
	results := make([]jobWithMetrics, 0, len(jobs))

	for _, j := range jobs {
		snap := j.Snapshot()
		m, err := h.collector.Get(snap.ID)
		if err != nil {
			continue
		}
		results = append(results, jobWithMetrics{
			ID:          snap.ID,
			Status:      snap.Status,
			TotalRuns:   m.TotalRuns,
			Failures:    m.Failures,
			AvgDuration: m.AvgDurationMs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
