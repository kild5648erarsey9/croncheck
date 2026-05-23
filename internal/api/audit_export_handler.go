package api

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/croncheck/internal/audit"
)

// AuditExportHandler serves audit log entries in downloadable formats (JSON or CSV).
type AuditExportHandler struct {
	log *audit.Log
}

// NewAuditExportHandler creates a new AuditExportHandler.
func NewAuditExportHandler(log *audit.Log) *AuditExportHandler {
	return &AuditExportHandler{log: log}
}

func (h *AuditExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	entries := h.log.All()

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=audit_export.json")
		_ = json.NewEncoder(w).Encode(entries)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=audit_export.csv")
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"job_id", "event", "duration_ms", "timestamp"})
		for _, e := range entries {
			_ = cw.Write([]string{
				e.JobID,
				e.Event,
				strconv.FormatInt(e.DurationMs, 10),
				e.Timestamp.Format(time.RFC3339),
			})
		}
		cw.Flush()
	default:
		http.Error(w, "unsupported format: use 'json' or 'csv'", http.StatusBadRequest)
	}
}

// RegisterAuditExportRoute registers the export endpoint on the given mux.
func RegisterAuditExportRoute(mux *http.ServeMux, h *AuditExportHandler) {
	mux.Handle("/audit/export", h)
}
