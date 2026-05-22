package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/audit"
)

func newAuditTimelineSetup() (*audit.Log, *AuditTimelineHandler) {
	log := audit.New()
	return log, NewAuditTimelineHandler(log)
}

func TestAuditTimelineHandler_Empty(t *testing.T) {
	_, h := newAuditTimelineSetup()
	req := httptest.NewRequest(http.MethodGet, "/audit/timeline", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty timeline, got %d buckets", len(result))
	}
}

func TestAuditTimelineHandler_AggregatesByHour(t *testing.T) {
	log, h := newAuditTimelineSetup()

	now := time.Now().UTC()
	log.Record("job-a", now.Add(-30*time.Minute), now.Add(-25*time.Minute), true)
	log.Record("job-b", now.Add(-20*time.Minute), now.Add(-15*time.Minute), false)
	log.Record("job-c", now.Add(-10*time.Minute), now.Add(-5*time.Minute), true)

	req := httptest.NewRequest(http.MethodGet, "/audit/timeline?hours=2", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one bucket")
	}
	total := 0
	for _, b := range result {
		total += int(b["total"].(float64))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestAuditTimelineHandler_MethodNotAllowed(t *testing.T) {
	_, h := newAuditTimelineSetup()
	req := httptest.NewRequest(http.MethodPost, "/audit/timeline", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestRegisterAuditTimelineRoute(t *testing.T) {
	log := audit.New()
	mux := http.NewServeMux()
	RegisterAuditTimelineRoute(mux, log)

	req := httptest.NewRequest(http.MethodGet, "/audit/timeline", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
