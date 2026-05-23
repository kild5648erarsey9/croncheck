package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/croncheck/internal/audit"
)

// AuditRetentionHandler exposes GET/PUT/DELETE endpoints for managing
// audit log retention policy and triggering manual purges.
type AuditRetentionHandler struct {
	reaper *audit.Reaper
}

// NewAuditRetentionHandler constructs a handler backed by the given Reaper.
func NewAuditRetentionHandler(r *audit.Reaper) *AuditRetentionHandler {
	return &AuditRetentionHandler{reaper: r}
}

func (h *AuditRetentionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuditRetentionHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	maxAge, maxEntries := h.reaper.Config()
	resp := map[string]interface{}{
		"max_age_seconds": maxAge.Seconds(),
		"max_entries":     maxEntries,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (h *AuditRetentionHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MaxAgeSeconds float64 `json:"max_age_seconds"`
		MaxEntries    int     `json:"max_entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	maxAge := time.Duration(body.MaxAgeSeconds * float64(time.Second))
	h.reaper.SetConfig(maxAge, body.MaxEntries)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"}) //nolint:errcheck
}

func (h *AuditRetentionHandler) handleDelete(w http.ResponseWriter, _ *http.Request) {
	purged := h.reaper.Purge()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"purged": purged}) //nolint:errcheck
}

// RegisterAuditRetentionRoute mounts the handler at /audit/retention.
func RegisterAuditRetentionRoute(mux *http.ServeMux, r *audit.Reaper) {
	h := NewAuditRetentionHandler(r)
	mux.Handle("/audit/retention", h)
}
