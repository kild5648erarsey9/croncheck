package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/metrics"
)

func newTestCollector() *metrics.Collector {
	c := metrics.NewCollector()
	c.Record("backup", 3*time.Second, true)
	c.Record("cleanup", 1*time.Second, false)
	return c
}

func TestMetricsHandler_ListAll(t *testing.T) {
	h := NewMetricsHandler(newTestCollector())
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.handleList(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var result []metrics.JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("got %d entries, want 2", len(result))
	}
}

func TestMetricsHandler_GetFound(t *testing.T) {
	h := NewMetricsHandler(newTestCollector())
	req := httptest.NewRequest(http.MethodGet, "/metrics/backup", nil)
	rec := httptest.NewRecorder()
	h.handleGet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var m metrics.JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if m.JobID != "backup" {
		t.Errorf("JobID = %q, want backup", m.JobID)
	}
}

func TestMetricsHandler_GetNotFound(t *testing.T) {
	h := NewMetricsHandler(newTestCollector())
	req := httptest.NewRequest(http.MethodGet, "/metrics/unknown", nil)
	rec := httptest.NewRecorder()
	h.handleGet(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	h := NewMetricsHandler(newTestCollector())
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.handleList(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestMetricsHandler_RegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	h := NewMetricsHandler(metrics.NewCollector())
	h.RegisterRoutes(mux)

	paths := []string{"/metrics", "/metrics/"}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Errorf("route %s not registered", p)
		}
	}
}
