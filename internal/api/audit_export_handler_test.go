package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/audit"
)

func newAuditExportSetup(t *testing.T) (*audit.Log, *api.AuditExportHandler) {
	t.Helper()
	log := audit.New()
	return log, api.NewAuditExportHandler(log)
}

func TestAuditExportHandler_JSONFormat(t *testing.T) {
	log, h := newAuditExportSetup(t)
	log.Record("job-1", "success", 120)
	log.Record("job-2", "failure", 300)

	req := httptest.NewRequest(http.MethodGet, "/audit/export?format=json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var entries []audit.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestAuditExportHandler_CSVFormat(t *testing.T) {
	log, h := newAuditExportSetup(t)
	log.Record("job-1", "success", 50)

	req := httptest.NewRequest(http.MethodGet, "/audit/export?format=csv", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "job_id") {
		t.Error("expected CSV header row")
	}
	if !strings.Contains(body, "job-1") {
		t.Error("expected job-1 in CSV output")
	}
	_ = time.Now() // ensure time import used
}

func TestAuditExportHandler_DefaultFormat_IsJSON(t *testing.T) {
	_, h := newAuditExportSetup(t)
	req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected default json format, got %s", ct)
	}
}

func TestAuditExportHandler_UnsupportedFormat(t *testing.T) {
	_, h := newAuditExportSetup(t)
	req := httptest.NewRequest(http.MethodGet, "/audit/export?format=xml", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAuditExportHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditExportSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/audit/export", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestRegisterAuditExportRoute(t *testing.T) {
	_, h := newAuditExportSetup(t)
	mux := http.NewServeMux()
	api.RegisterAuditExportRoute(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 from registered route, got %d", rec.Code)
	}
}
