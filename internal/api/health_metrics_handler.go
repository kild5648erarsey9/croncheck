package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/croncheck/internal/healthcheck"
	"github.com/yourorg/croncheck/internal/metrics"
)

// HealthMetricsHandler combines health status with job metrics summary.
type HealthMetricsHandler struct {
	checker   *healthcheck.Checker
	collector *metrics.Collector
}

// NewHealthMetricsHandler creates a new HealthMetricsHandler.
func NewHealthMetricsHandler(checker *healthcheck.Checker, collector *metrics.Collector) *HealthMetricsHandler {
	return &HealthMetricsHandler{checker: checker, collector: collector}
}

type healthMetricsSummary struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	JobsTotal int               `json:"jobs_total"`
	Timestamp time.Time         `json:"timestamp"`
}

func (h *HealthMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := h.checker.RunAll()

	overall := "ok"
	checks := make(map[string]string, len(result))
	for name, err := range result {
		if err != nil {
			overall = "degraded"
			checks[name] = err.Error()
		} else {
			checks[name] = "ok"
		}
	}

	all := h.collector.All()

	summary := healthMetricsSummary{
		Status:    overall,
		Checks:    checks,
		JobsTotal: len(all),
		Timestamp: time.Now().UTC(),
	}

	statusCode := http.StatusOK
	if overall != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(summary)
}

// RegisterHealthMetricsRoute registers the combined health+metrics route.
func RegisterHealthMetricsRoute(mux *http.ServeMux, h *HealthMetricsHandler) {
	mux.Handle("/health/metrics", h)
}
