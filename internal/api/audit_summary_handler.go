package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/croncheck/internal/audit"
)

// AuditSummaryHandler serves a summarised view of the audit log grouped by job,
// showing the last event time, total runs, and failure count per job.
type AuditSummaryHandler struct {
	log *audit.Log
}

// NewAuditSummaryHandler constructs an AuditSummaryHandler backed by the given log.
func NewAuditSummaryHandler(log *audit.Log) *AuditSummaryHandler {
	return &AuditSummaryHandler{log: log}
}

type jobAuditSummary struct {
	JobID     string `json:"job_id"`
	TotalRuns int    `json:"total_runs"`
	Failures  int    `json:"failures"`
	LastEvent string `json:"last_event"` // RFC3339
}

func (h *AuditSummaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := h.log.All()

	type accumulator struct {
		total    int
		failures int
		last     string
	}

	acc := make(map[string]*accumulator)
	for _, e := range entries {
		a, ok := acc[e.JobID]
		if !ok {
			a = &accumulator{}
			acc[e.JobID] = a
		}
		a.total++
		if !e.Success {
			a.failures++
		}
		ts := e.Timestamp.Format("2006-01-02T15:04:05Z07:00")
		if a.last == "" || ts > a.last {
			a.last = ts
		}
	}

	summaries := make([]jobAuditSummary, 0, len(acc))
	for id, a := range acc {
		summaries = append(summaries, jobAuditSummary{
			JobID:     id,
			TotalRuns: a.total,
			Failures:  a.failures,
			LastEvent: a.last,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].JobID < summaries[j].JobID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// RegisterAuditSummaryRoute mounts the summary handler under GET /audit/summary.
func RegisterAuditSummaryRoute(mux *http.ServeMux, h *AuditSummaryHandler) {
	mux.Handle("/audit/summary", h)
}
