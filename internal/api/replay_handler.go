package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/croncheck/internal/replay"
)

// ReplayHandler exposes an HTTP endpoint to trigger audit-log replays.
type ReplayHandler struct {
	replayer *replay.Replayer
}

// NewReplayHandler creates a ReplayHandler backed by the given Replayer.
func NewReplayHandler(r *replay.Replayer) *ReplayHandler {
	return &ReplayHandler{replayer: r}
}

type replayRequest struct {
	JobID  string `json:"job_id"`
	Since  string `json:"since"`
	Before string `json:"before"`
}

type replayResponse struct {
	Replayed int      `json:"replayed"`
	Errors   []string `json:"errors,omitempty"`
}

func (h *ReplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req replayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	rr := replay.Request{JobID: req.JobID}

	if req.Since != "" {
		t, err := time.Parse(time.RFC3339, req.Since)
		if err != nil {
			http.Error(w, "invalid since timestamp", http.StatusBadRequest)
			return
		}
		rr.Since = t
	}
	if req.Before != "" {
		t, err := time.Parse(time.RFC3339, req.Before)
		if err != nil {
			http.Error(w, "invalid before timestamp", http.StatusBadRequest)
			return
		}
		rr.Before = t
	}

	result := h.replayer.Run(rr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replayResponse{
		Replayed: result.Replayed,
		Errors:   result.Errors,
	})
}

// RegisterReplayRoute mounts the replay handler under /api/replay.
func RegisterReplayRoute(mux *http.ServeMux, h *ReplayHandler) {
	mux.Handle("/api/replay", h)
}
