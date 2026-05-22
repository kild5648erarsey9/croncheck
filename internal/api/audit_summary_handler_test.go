package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/audit"
)

func newAuditSummarySetup(t *testing.T) (*audit.Log, *api.AuditSummaryHandler) {
	t.Helper()
	log := audit.New()
	return log, api.NewAuditSummaryHandler(log)
}

func TestAuditSummaryHandler_Empty(t *testing.T) {
	_, h := newAuditSummarySetup(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/summary", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var out []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(out))
	}
}

func TestAuditSummaryHandler_AggregatesCorrectly(t *testing.T) {
	log, h := newAuditSummarySetup(t)

	now := time.Now()
	log.Record("job-a", true, now.Add(-2*time.Minute))
	log.Record("job-a", false, now.Add(-1*time.Minute))
	log.Record("job-a", true, now)
	log.Record("job-b", false, now)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/summary", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	type summary struct {
		JobID     string `json:"job_id"`
		TotalRuns int    `json:"total_runs"`
		Failures  int    `json:"failures"`
		LastEvent string `json:"last_event"`
	}
	var out []summary
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(out))
	}
	// sorted by job_id: job-a first
	if out[0].JobID != "job-a" {
		t.Errorf("expected job-a first, got %s", out[0].JobID)
	}
	if out[0].TotalRuns != 3 {
		t.Errorf("expected 3 total runs for job-a, got %d", out[0].TotalRuns)
	}
	if out[0].Failures != 1 {
		t.Errorf("expected 1 failure for job-a, got %d", out[0].Failures)
	}
	if out[1].JobID != "job-b" {
		t.Errorf("expected job-b second, got %s", out[1].JobID)
	}
	if out[1].Failures != 1 {
		t.Errorf("expected 1 failure for job-b, got %d", out[1].Failures)
	}
}

func TestAuditSummaryHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditSummarySetup(t)

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/audit/summary", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestRegisterAuditSummaryRoute(t *testing.T) {
	_, h := newAuditSummarySetup(t)
	mux := http.NewServeMux()
	api.RegisterAuditSummaryRoute(mux, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/summary", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 via mux, got %d", rec.Code)
	}
}
