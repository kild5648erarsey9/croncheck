package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/example/croncheck/internal/scheduler"
)

// SchedulerHandler exposes HTTP endpoints to manage job schedules.
type SchedulerHandler struct {
	sched *scheduler.Scheduler
}

// NewSchedulerHandler creates a SchedulerHandler backed by the given Scheduler.
func NewSchedulerHandler(s *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{sched: s}
}

// RegisterRoutes attaches scheduler routes to the given mux.
func (h *SchedulerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/schedules/", h.dispatch)
}

type scheduleRequest struct {
	Interval string `json:"interval"`
}

func (h *SchedulerHandler) dispatch(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/schedules/")
	if jobID == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		h.handleAdd(w, r, jobID)
	case http.MethodDelete:
		h.handleRemove(w, jobID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SchedulerHandler) handleAdd(w http.ResponseWriter, r *http.Request, jobID string) {
	var req scheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	dur, err := time.ParseDuration(req.Interval)
	if err != nil || dur <= 0 {
		http.Error(w, "invalid interval", http.StatusBadRequest)
		return
	}
	h.sched.Add(scheduler.Schedule{JobID: jobID, Interval: dur})
	w.WriteHeader(http.StatusNoContent)
}

func (h *SchedulerHandler) handleRemove(w http.ResponseWriter, jobID string) {
	h.sched.Remove(jobID)
	w.WriteHeader(http.StatusNoContent)
}
