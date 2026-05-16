package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/croncheck/internal/healthcheck"
	"github.com/yourorg/croncheck/internal/metrics"
)

func newHealthMetricsSetup() (*healthcheck.Checker, *metrics.Collector) {
	return healthcheck.New(), metrics.NewCollector()
}

func TestHealthMetricsHandler_AllHealthy_NoJobs(t *testing.T) {
	checker, collector := newHealthMetricsSetup()
	h := NewHealthMetricsHandler(checker, collector)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
	if int(body["jobs_total"].(float64)) != 0 {
		t.Errorf("expected 0 jobs, got %v", body["jobs_total"])
	}
}

func TestHealthMetricsHandler_DegradedWithJobs(t *testing.T) {
	checker, collector := newHealthMetricsSetup()
	checker.Register("db", func() error { return errors.New("connection refused") })
	collector.Record("job-a", true, 100)
	collector.Record("job-b", false, 200)

	h := NewHealthMetricsHandler(checker, collector)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "degraded" {
		t.Errorf("expected degraded, got %v", body["status"])
	}
	if int(body["jobs_total"].(float64)) != 2 {
		t.Errorf("expected 2 jobs, got %v", body["jobs_total"])
	}
}

func TestHealthMetricsHandler_MethodNotAllowed(t *testing.T) {
	checker, collector := newHealthMetricsSetup()
	h := NewHealthMetricsHandler(checker, collector)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health/metrics", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRegisterHealthMetricsRoute(t *testing.T) {
	checker, collector := newHealthMetricsSetup()
	h := NewHealthMetricsHandler(checker, collector)
	mux := http.NewServeMux()
	RegisterHealthMetricsRoute(mux, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
