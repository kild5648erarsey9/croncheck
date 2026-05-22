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

func newAuditMetricsSetup(t *testing.T) (*audit.Log, *api.AuditMetricsHandler) {
	t.Helper()
	log := audit.New()
	return log, api.NewAuditMetricsHandler(log)
}

func TestAuditMetricsHandler_Empty(t *testing.T) {
	_, h := newAuditMetricsSetup(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/metrics", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(result))
	}
}

func TestAuditMetricsHandler_AggregatesCounts(t *testing.T) {
	log, h := newAuditMetricsSetup(t)

	log.Record("job-a", true, 100*time.Millisecond)
	log.Record("job-a", false, 200*time.Millisecond)
	log.Record("job-a", true, 150*time.Millisecond)
	log.Record("job-b", false, 50*time.Millisecond)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/metrics", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	type summary struct {
		JobID     string `json:"job_id"`
		Total     int    `json:"total"`
		Successes int    `json:"successes"`
		Failures  int    `json:"failures"`
	}
	var result []summary
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(result))
	}
	// sorted by job_id: job-a first
	if result[0].JobID != "job-a" {
		t.Errorf("expected job-a first, got %s", result[0].JobID)
	}
	if result[0].Total != 3 || result[0].Successes != 2 || result[0].Failures != 1 {
		t.Errorf("job-a counts wrong: %+v", result[0])
	}
	if result[1].JobID != "job-b" || result[1].Total != 1 || result[1].Failures != 1 {
		t.Errorf("job-b counts wrong: %+v", result[1])
	}
}

func TestAuditMetricsHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditMetricsSetup(t)

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/audit/metrics", nil)
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rr.Code)
		}
	}
}

func TestRegisterAuditMetricsRoute(t *testing.T) {
	log := audit.New()
	log.Record("job-x", true, 10*time.Millisecond)
	h := api.NewAuditMetricsHandler(log)

	mux := http.NewServeMux()
	api.RegisterAuditMetricsRoute(mux, h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/metrics", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
