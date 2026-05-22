package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/croncheck/internal/audit"
)

// AuditStatsHandler serves aggregated per-job audit statistics.
type AuditStatsHandler struct {
	log *audit.Log
}

// NewAuditStatsHandler creates a new AuditStatsHandler.
func NewAuditStatsHandler(log *audit.Log) *AuditStatsHandler {
	return &AuditStatsHandler{log: log}
}

type jobAuditStats struct {
	JobID        string  `json:"job_id"`
	Total        int     `json:"total"`
	Successes    int     `json:"successes"`
	Failures     int     `json:"failures"`
	SuccessRate  float64 `json:"success_rate"`
}

func (h *AuditStatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := h.log.All()

	counts := make(map[string]*jobAuditStats)
	for _, e := range entries {
		s, ok := counts[e.JobID]
		if !ok {
			s = &jobAuditStats{JobID: e.JobID}
			counts[e.JobID] = s
		}
		s.Total++
		if e.Success {
			s.Successes++
		} else {
			s.Failures++
		}
	}

	result := make([]jobAuditStats, 0, len(counts))
	for _, s := range counts {
		if s.Total > 0 {
			s.SuccessRate = float64(s.Successes) / float64(s.Total)
		}
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].JobID < result[j].JobID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RegisterAuditStatsRoute mounts the handler under /audit/stats.
func RegisterAuditStatsRoute(mux *http.ServeMux, log *audit.Log) {
	h := NewAuditStatsHandler(log)
	mux.Handle("/audit/stats", h)
}
