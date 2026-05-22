package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/croncheck/internal/audit"
)

// AuditRetentionHandler exposes an endpoint to query and update the audit retention policy.
type AuditRetentionHandler struct {
	reaper *audit.Reaper
	policy *audit.RetentionPolicy
}

// NewAuditRetentionHandler creates a handler backed by the given reaper and policy.
func NewAuditRetentionHandler(reaper *audit.Reaper, policy *audit.RetentionPolicy) *AuditRetentionHandler {
	return &AuditRetentionHandler{reaper: reaper, policy: policy}
}

func (h *AuditRetentionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuditRetentionHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	type response struct {
		MaxAgeSeconds int `json:"max_age_seconds"`
		MaxEntries    int `json:"max_entries"`
	}
	resp := response{
		MaxAgeSeconds: int(h.policy.MaxAge.Seconds()),
		MaxEntries:    h.policy.MaxEntries,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuditRetentionHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	type request struct {
		MaxAgeSeconds int `json:"max_age_seconds"`
		MaxEntries    int `json:"max_entries"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.MaxAgeSeconds < 0 || req.MaxEntries < 0 {
		http.Error(w, "values must be non-negative", http.StatusBadRequest)
		return
	}
	h.policy.MaxAge = time.Duration(req.MaxAgeSeconds) * time.Second
	h.policy.MaxEntries = req.MaxEntries
	w.WriteHeader(http.StatusNoContent)
}

// RegisterAuditRetentionRoute registers the retention policy endpoint.
func RegisterAuditRetentionRoute(mux *http.ServeMux, reaper *audit.Reaper, policy *audit.RetentionPolicy) {
	handler := NewAuditRetentionHandler(reaper, policy)
	mux.Handle("/audit/retention", handler)
}
