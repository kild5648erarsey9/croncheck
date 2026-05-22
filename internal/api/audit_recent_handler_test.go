package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/audit"
)

func newAuditRecentSetup(t *testing.T) (*audit.Log, *AuditRecentHandler) {
	t.Helper()
	log := audit.New()
	return log, NewAuditRecentHandler(log)
}

func TestAuditRecentHandler_Empty(t *testing.T) {
	_, h := newAuditRecentSetup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/recent", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []audit.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestAuditRecentHandler_DefaultLimit(t *testing.T) {
	log, h := newAuditRecentSetup(t)
	for i := 0; i < 15; i++ {
		log.Record(audit.Entry{JobID: fmt.Sprintf("job-%d", i), Status: "success", At: time.Now()})
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/recent", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []audit.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != defaultRecentLimit {
		t.Errorf("expected %d entries, got %d", defaultRecentLimit, len(entries))
	}
}

func TestAuditRecentHandler_CustomLimit(t *testing.T) {
	log, h := newAuditRecentSetup(t)
	for i := 0; i < 20; i++ {
		log.Record(audit.Entry{JobID: fmt.Sprintf("job-%d", i), Status: "success", At: time.Now()})
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/recent?limit=5", nil)
	h.ServeHTTP(rec, req)

	var entries []audit.Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

func TestAuditRecentHandler_InvalidLimit(t *testing.T) {
	_, h := newAuditRecentSetup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/recent?limit=abc", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAuditRecentHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditRecentSetup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit/recent", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestRegisterAuditRecentRoute(t *testing.T) {
	_, h := newAuditRecentSetup(t)
	mux := http.NewServeMux()
	RegisterAuditRecentRoute(mux, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/recent", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
