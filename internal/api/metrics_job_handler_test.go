package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/job"
	"github.com/yourorg/croncheck/internal/metrics"
)

func setupMetricsJobHandler(t *testing.T) (*MetricsJobHandler, *job.Registry, *metrics.Collector) {
	t.Helper()
	reg := job.NewRegistry()
	col := metrics.NewCollector()

	j, err := job.New("backup", 5*time.Second)
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	if err := reg.Register(j); err != nil {
		t.Fatalf("Register: %v", err)
	}

	j.Start()
	j.Finish(true)
	col.Record(j.Snapshot())

	return NewMetricsJobHandler(reg, col), reg, col
}

func TestMetricsJobHandler_GetSingle(t *testing.T) {
	h, _, _ := setupMetricsJobHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/jobs/backup/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result jobWithMetrics
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.ID != "backup" {
		t.Errorf("expected id=backup, got %s", result.ID)
	}
	if result.TotalRuns != 1 {
		t.Errorf("expected TotalRuns=1, got %d", result.TotalRuns)
	}
}

func TestMetricsJobHandler_GetAll(t *testing.T) {
	h, _, _ := setupMetricsJobHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/jobs/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []jobWithMetrics
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestMetricsJobHandler_NotFound(t *testing.T) {
	h, _, _ := setupMetricsJobHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/jobs/unknown/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestMetricsJobHandler_MethodNotAllowed(t *testing.T) {
	h, _, _ := setupMetricsJobHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/jobs/backup/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMetricsJobHandler_ContentTypeJSON(t *testing.T) {
	h, _, _ := setupMetricsJobHandler(t)

	for _, path := range []string{"/jobs/backup/metrics", "/jobs/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		ct := rec.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("path %s: expected Content-Type application/json, got %s", path, ct)
		}
	}
}
