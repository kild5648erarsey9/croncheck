package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/croncheck/internal/audit"
)

// AuditJobHandler serves per-job audit log entries via HTTP.
type AuditJobHandler struct {
	log *audit.Log
}

// NewAuditJobHandler creates a new AuditJobHandler backed by the given audit log.
func NewAuditJobHandler(log *audit.Log) *AuditJobHandler {
	return &AuditJobHandler{log: log}
}

// ServeHTTP dispatches to the appropriate sub-handler based on the URL path.
func (h *AuditJobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := extractAuditJobID(r.URL.Path)
	if jobID == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}

	entries := h.log.FilterByJob(jobID)
	if entries == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// RegisterAuditJobRoute registers the per-job audit route on the given mux.
func RegisterAuditJobRoute(mux *http.ServeMux, h *AuditJobHandler) {
	mux.Handle("/audit/jobs/", h)
}

// extractAuditJobID pulls the job ID from paths like /audit/jobs/{id}.
func extractAuditJobID(path string) string {
	const prefix = "/audit/jobs/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	id = strings.Trim(id, "/")
	return id
}
