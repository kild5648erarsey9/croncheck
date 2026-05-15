package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/job"
)

func newTestRegistry(t *testing.T, ids ...string) *job.Registry {
	t.Helper()
	r := job.NewRegistry()
	for _, id := range ids {
		_, err := r.Register(id, 5*time.Second)
		if err != nil {
			t.Fatalf("failed to register job %q: %v", id, err)
		}
	}
	return r
}

func TestListJobs_Empty(t *testing.T) {
	r := newTestRegistry(t)
	h := NewHandler(r)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	h.ListJobs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []jobSnapshotResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp))
	}
}

func TestListJobs_ReturnAll(t *testing.T) {
	r := newTestRegistry(t, "job-a", "job-b")
	h := NewHandler(r)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	h.ListJobs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []jobSnapshotResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(resp))
	}
}

func TestListJobs_MethodNotAllowed(t *testing.T) {
	r := newTestRegistry(t)
	h := NewHandler(r)

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		req := httptest.NewRequest(method, "/jobs", nil)
		rec := httptest.NewRecorder()
		h.ListJobs(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestRegisterRoutes(t *testing.T) {
	r := newTestRegistry(t, "job-x")
	h := NewHandler(r)
	mux := http.NewServeMux()
	RegisterRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via mux, got %d", rec.Code)
	}
}
