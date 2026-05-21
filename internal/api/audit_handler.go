package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/croncheck/internal/audit"
)

// AuditHandler serves audit log entries over HTTP.
type AuditHandler struct {
	log *audit.Log
}

// NewAuditHandler creates a new AuditHandler backed by the provided audit.Log.
func NewAuditHandler(l *audit.Log) *AuditHandler {
	return &AuditHandler{log: l}
}

// RegisterAuditRoutes wires the audit endpoints into the given mux.
func (h *AuditHandler) RegisterAuditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/audit", h.handleList)
	mux.HandleFunc("/audit/", h.handleByJob)
}

// handleList returns all audit entries as JSON.
func (h *AuditHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	entries := h.log.All()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries) //nolint:errcheck
}

// handleByJob returns audit entries filtered by job ID extracted from the URL path.
func (h *AuditHandler) handleByJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/audit/")
	if jobID == "" {
		http.Error(w, "missing job ID", http.StatusBadRequest)
		return
	}
	entries := h.log.FilterByJob(jobID)
	if len(entries) == 0 {
		http.Error(w, "no audit entries found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries) //nolint:errcheck
}
