package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/croncheck/internal/audit"
)

// AuditMetricsHandler exposes aggregated audit statistics over HTTP.
type AuditMetricsHandler struct {
	log *audit.Log
}

// NewAuditMetricsHandler creates a new AuditMetricsHandler.
func NewAuditMetricsHandler(log *audit.Log) *AuditMetricsHandler {
	return &AuditMetricsHandler{log: log}
}

type auditJobSummary struct {
	JobID      string `json:"job_id"`
	Total      int    `json:"total"`
	Successes  int    `json:"successes"`
	Failures   int    `json:"failures"`
}

// ServeHTTP handles GET /audit/metrics — returns per-job success/failure counts.
func (h *AuditMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := h.log.All()

	counts := make(map[string]*auditJobSummary)
	for _, e := range entries {
		s, ok := counts[e.JobID]
		if !ok {
			s = &auditJobSummary{JobID: e.JobID}
			counts[e.JobID] = s
		}
		s.Total++
		if e.Success {
			s.Successes++
		} else {
			s.Failures++
		}
	}

	summaries := make([]auditJobSummary, 0, len(counts))
	for _, s := range counts {
		summaries = append(summaries, *s)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].JobID < summaries[j].JobID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// RegisterAuditMetricsRoute registers the audit metrics route on the given mux.
func RegisterAuditMetricsRoute(mux *http.ServeMux, h *AuditMetricsHandler) {
	mux.Handle("/audit/metrics", h)
}
