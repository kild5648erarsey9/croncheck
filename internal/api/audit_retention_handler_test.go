package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/audit"
)

func newAuditRetentionSetup(t *testing.T) (*audit.Log, *audit.Reaper, http.Handler) {
	t.Helper()
	log := audit.New()
	reaper := audit.NewReaper(log, 0, 0)
	h := api.NewAuditRetentionHandler(reaper)
	return log, reaper, h
}

func TestAuditRetentionHandler_GetConfig(t *testing.T) {
	_, _, h := newAuditRetentionSetup(t)
	req := httptest.NewRequest(http.MethodGet, "/audit/retention", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := body["max_age_seconds"]; !ok {
		t.Error("expected max_age_seconds field")
	}
	if _, ok := body["max_entries"]; !ok {
		t.Error("expected max_entries field")
	}
}

func TestAuditRetentionHandler_UpdateConfig(t *testing.T) {
	_, _, h := newAuditRetentionSetup(t)
	body := map[string]interface{}{
		"max_age_seconds": 3600,
		"max_entries":     500,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/audit/retention", bytes.NewReader(b))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuditRetentionHandler_PurgeNow(t *testing.T) {
	log, _, h := newAuditRetentionSetup(t)
	log.Record("job-1", "success", time.Second)
	log.Record("job-2", "failure", 0)
	req := httptest.NewRequest(http.MethodDelete, "/audit/retention", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := resp["purged"]; !ok {
		t.Error("expected purged field in response")
	}
}

func TestAuditRetentionHandler_MethodNotAllowed(t *testing.T) {
	_, _, h := newAuditRetentionSetup(t)
	req := httptest.NewRequest(http.MethodPost, "/audit/retention", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestRegisterAuditRetentionRoute(t *testing.T) {
	_, reaper, _ := newAuditRetentionSetup(t)
	mux := http.NewServeMux()
	api.RegisterAuditRetentionRoute(mux, reaper)
	req := httptest.NewRequest(http.MethodGet, "/audit/retention", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
