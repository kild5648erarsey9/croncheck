package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/audit"
)

func newTestAuditLog() *audit.Log {
	return audit.New()
}

func TestAuditHandler_ListAll_Empty(t *testing.T) {
	l := newTestAuditLog()
	h := api.NewAuditHandler(l)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	h.RegisterAuditRoutes(http.NewServeMux())

	mux := http.NewServeMux()
	h.RegisterAuditRoutes(mux)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []audit.Entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestAuditHandler_ListAll_ReturnsEntries(t *testing.T) {
	l := newTestAuditLog()
	l.Record(audit.EventJobStarted, "job-1", "started", nil)
	l.Record(audit.EventJobFinished, "job-1", "finished", nil)

	h := api.NewAuditHandler(l)
	mux := http.NewServeMux()
	h.RegisterAuditRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []audit.Entry
	json.NewDecoder(rr.Body).Decode(&entries) //nolint:errcheck
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestAuditHandler_ByJob_Found(t *testing.T) {
	l := newTestAuditLog()
	l.Record(audit.EventJobStarted, "job-42", "started", nil)
	l.Record(audit.EventJobStarted, "job-99", "started", nil)

	h := api.NewAuditHandler(l)
	mux := http.NewServeMux()
	h.RegisterAuditRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/job-42", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []audit.Entry
	json.NewDecoder(rr.Body).Decode(&entries) //nolint:errcheck
	if len(entries) != 1 || entries[0].JobID != "job-42" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestAuditHandler_ByJob_NotFound(t *testing.T) {
	l := newTestAuditLog()
	h := api.NewAuditHandler(l)
	mux := http.NewServeMux()
	h.RegisterAuditRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/ghost", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestAuditHandler_MethodNotAllowed(t *testing.T) {
	l := newTestAuditLog()
	h := api.NewAuditHandler(l)
	mux := http.NewServeMux()
	h.RegisterAuditRoutes(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
