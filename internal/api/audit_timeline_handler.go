package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/croncheck/internal/audit"
)

// AuditTimelineHandler returns audit entries grouped by time bucket.
type AuditTimelineHandler struct {
	log *audit.Log
}

// NewAuditTimelineHandler creates a new AuditTimelineHandler.
func NewAuditTimelineHandler(log *audit.Log) *AuditTimelineHandler {
	return &AuditTimelineHandler{log: log}
}

type timelineBucket struct {
	Bucket   string `json:"bucket"`
	Success  int    `json:"success"`
	Failure  int    `json:"failure"`
	Total    int    `json:"total"`
}

func (h *AuditTimelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse optional window query param (hours, default 24)
	windowHours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 168 {
			windowHours = n
		}
	}

	cutoff := time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour)
	entries := h.log.All()

	buckets := make(map[string]*timelineBucket)
	for _, e := range entries {
		if e.StartedAt.Before(cutoff) {
			continue
		}
		key := e.StartedAt.UTC().Format("2006-01-02T15")
		b, ok := buckets[key]
		if !ok {
			b = &timelineBucket{Bucket: key}
			buckets[key] = b
		}
		b.Total++
		if e.Success {
			b.Success++
		} else {
			b.Failure++
		}
	}

	result := make([]timelineBucket, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, *b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RegisterAuditTimelineRoute registers the timeline endpoint on the given mux.
func RegisterAuditTimelineRoute(mux *http.ServeMux, log *audit.Log) {
	h := NewAuditTimelineHandler(log)
	mux.Handle("/audit/timeline", h)
}
