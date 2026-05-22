package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/audit"
)

func newAuditJobSetup(t *testing.T) (*audit.Log, *api.AuditJobHandler) {
	t.Helper()
	log := audit.New()
	return log, api.NewAuditJobHandler(log)
}

func TestAuditJobHandler_Found(t *testing.T) {
	log, h := newAuditJobSetup(t)
	log.Record("job-alpha", "success", 0)
	log.Record("job-alpha", "failure", 1)

	req := httptest.NewRequest(http.MethodGet, "/audit/jobs/job-alpha", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []audit.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.JobID != "job-alpha" {
			t.Errorf("unexpected job id %q", e.JobID)
		}
	}
}

func TestAuditJobHandler_NotFound(t *testing.T) {
	_, h := newAuditJobSetup(t)

	req := httptest.NewRequest(http.MethodGet, "/audit/jobs/nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAuditJobHandler_MissingJobID(t *testing.T) {
	_, h := newAuditJobSetup(t)

	req := httptest.NewRequest(http.MethodGet, "/audit/jobs/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuditJobHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditJobSetup(t)

	req := httptest.NewRequest(http.MethodPost, "/audit/jobs/job-alpha", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRegisterAuditJobRoute(t *testing.T) {
	log, h := newAuditJobSetup(t)
	log.Record("job-beta", "success", 0)

	mux := http.NewServeMux()
	api.RegisterAuditJobRoute(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/audit/jobs/job-beta", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
