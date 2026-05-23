package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/audit"
	"github.com/yourorg/croncheck/internal/api"
	"github.com/yourorg/croncheck/internal/notifier"
	"github.com/yourorg/croncheck/internal/replay"
)

func newReplaySetup(t *testing.T) (*audit.Log, *api.ReplayHandler) {
	t.Helper()
	log := audit.New()
	n := notifier.New(nil)
	r := replay.New(log, n)
	return log, api.NewReplayHandler(r)
}

func TestReplayHandler_Success(t *testing.T) {
	log, h := newReplaySetup(t)
	log.Record(audit.Entry{JobID: "job-1", Status: "failure", Timestamp: time.Now()})
	log.Record(audit.Entry{JobID: "job-1", Status: "success", Timestamp: time.Now()})

	body, _ := json.Marshal(map[string]string{"job_id": "job-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/replay", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if int(resp["replayed"].(float64)) != 2 {
		t.Fatalf("expected replayed=2, got %v", resp["replayed"])
	}
}

func TestReplayHandler_MethodNotAllowed(t *testing.T) {
	_, h := newReplaySetup(t)
	req := httptest.NewRequest(http.MethodGet, "/api/replay", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestReplayHandler_InvalidJSON(t *testing.T) {
	_, h := newReplaySetup(t)
	req := httptest.NewRequest(http.MethodPost, "/api/replay", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestReplayHandler_InvalidSince(t *testing.T) {
	_, h := newReplaySetup(t)
	body, _ := json.Marshal(map[string]string{"since": "not-a-time"})
	req := httptest.NewRequest(http.MethodPost, "/api/replay", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRegisterReplayRoute(t *testing.T) {
	_, h := newReplaySetup(t)
	mux := http.NewServeMux()
	api.RegisterReplayRoute(mux, h)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/replay", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via mux, got %d", rec.Code)
	}
}
