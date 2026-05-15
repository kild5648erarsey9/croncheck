package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/croncheck/internal/alert"
	"github.com/croncheck/internal/job"
)

// JobHandler handles HTTP requests for job lifecycle events (start/finish).
type JobHandler struct {
	registry *job.Registry
	alertMgr *alert.Manager
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(registry *job.Registry, alertMgr *alert.Manager) *JobHandler {
	return &JobHandler{registry: registry, alertMgr: alertMgr}
}

// HandleStart processes a job start event: POST /jobs/{id}/start
func (h *JobHandler) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := extractJobID(r.URL.Path, "/start")
	if id == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}
	j, err := h.registry.Get(id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	j.Start()
	w.WriteHeader(http.StatusNoContent)
}

// HandleFinish processes a job finish event: POST /jobs/{id}/finish
func (h *JobHandler) HandleFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := extractJobID(r.URL.Path, "/finish")
	if id == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}
	var req struct {
		ExitCode int `json:"exit_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	j, err := h.registry.Get(id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	ev := j.Finish(req.ExitCode)
	if ev != nil && h.alertMgr != nil {
		_ = h.alertMgr.Notify(*ev)
	}
	w.WriteHeader(http.StatusNoContent)
}

// extractJobID pulls the job ID from a path like /jobs/{id}/suffix.
func extractJobID(path, suffix string) string {
	path = strings.TrimSuffix(path, suffix)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}
