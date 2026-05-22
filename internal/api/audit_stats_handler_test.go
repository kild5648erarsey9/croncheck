package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/croncheck/internal/audit"
)

func newAuditStatsSetup() (*audit.Log, *AuditStatsHandler) {
	log := audit.New()
	h := NewAuditStatsHandler(log)
	return log, h
}

func TestAuditStatsHandler_Empty(t *testing.T) {
	_, h := newAuditStatsSetup()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

func TestAuditStatsHandler_AggregatesStats(t *testing.T) {
	log, h := newAuditStatsSetup()
	log.Record("job-a", true, 0)
	log.Record("job-a", true, 0)
	log.Record("job-a", false, 0)
	log.Record("job-b", false, 0)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	// sorted by job_id: job-a first
	if result[0]["job_id"] != "job-a" {
		t.Errorf("expected job-a first, got %v", result[0]["job_id"])
	}
	if result[0]["total"].(float64) != 3 {
		t.Errorf("expected total=3 for job-a, got %v", result[0]["total"])
	}
	if result[0]["successes"].(float64) != 2 {
		t.Errorf("expected successes=2 for job-a, got %v", result[0]["successes"])
	}
	rate := result[0]["success_rate"].(float64)
	if rate < 0.666 || rate > 0.667 {
		t.Errorf("expected success_rate~0.667 for job-a, got %v", rate)
	}
}

func TestAuditStatsHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditStatsSetup()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit/stats", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestRegisterAuditStatsRoute(t *testing.T) {
	mux := http.NewServeMux()
	log := audit.New()
	RegisterAuditStatsRoute(mux, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
