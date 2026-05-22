package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/croncheck/internal/audit"
)

const defaultRecentLimit = 10
const maxRecentLimit = 100

// AuditRecentHandler serves the most recent audit log entries.
type AuditRecentHandler struct {
	log *audit.Log
}

// NewAuditRecentHandler creates a new AuditRecentHandler.
func NewAuditRecentHandler(log *audit.Log) *AuditRecentHandler {
	return &AuditRecentHandler{log: log}
}

func (h *AuditRecentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := defaultRecentLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}
		if n > maxRecentLimit {
			n = maxRecentLimit
		}
		limit = n
	}

	all := h.log.All()
	start := len(all) - limit
	if start < 0 {
		start = 0
	}
	recent := all[start:]

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(recent); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// RegisterAuditRecentRoute registers the /audit/recent endpoint.
func RegisterAuditRecentRoute(mux *http.ServeMux, h *AuditRecentHandler) {
	mux.Handle("/audit/recent", h)
}
