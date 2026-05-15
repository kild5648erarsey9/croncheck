package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/croncheck/internal/job"
)

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	registry *job.Registry
}

// NewHandler creates a new Handler with the given registry.
func NewHandler(r *job.Registry) *Handler {
	return &Handler{registry: r}
}

// jobSnapshotResponse is the JSON representation of a job snapshot.
type jobSnapshotResponse struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	ExitCode  int        `json:"exit_code"`
	Error     string     `json:"error,omitempty"`
}

// ListJobs handles GET /jobs and returns all registered job snapshots.
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	all := h.registry.All()
	resp := make([]jobSnapshotResponse, 0, len(all))
	for _, j := range all {
		snap := j.Snapshot()
		entry := jobSnapshotResponse{
			ID:       snap.ID,
			Status:   snap.Status,
			ExitCode: snap.ExitCode,
		}
		if snap.Error != nil {
			entry.Error = snap.Error.Error()
		}
		if !snap.StartedAt.IsZero() {
			t := snap.StartedAt
			entry.StartedAt = &t
		}
		if !snap.FinishedAt.IsZero() {
			t := snap.FinishedAt
			entry.FinishedAt = &t
		}
		resp = append(resp, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterRoutes attaches handler methods to the given mux.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/jobs", h.ListJobs)
}
