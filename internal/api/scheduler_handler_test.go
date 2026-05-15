package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/croncheck/internal/api"
	"github.com/example/croncheck/internal/job"
	"github.com/example/croncheck/internal/scheduler"
)

func newTestScheduler(t *testing.T) *scheduler.Scheduler {
	t.Helper()
	reg := job.NewRegistry()
	return scheduler.New(reg, time.Minute)
}

func TestSchedulerHandler_Add_Success(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	body := bytes.NewBufferString(`{"interval":"5m"}`)
	req := httptest.NewRequest(http.MethodPut, "/schedules/myjob", body)
	rw := httptest.NewRecorder()
	h.RegisterRoutes(http.NewServeMux())

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
}

func TestSchedulerHandler_Add_InvalidInterval(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	body := bytes.NewBufferString(`{"interval":"not-a-duration"}`)
	req := httptest.NewRequest(http.MethodPut, "/schedules/myjob", body)
	rw := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestSchedulerHandler_Add_InvalidJSON(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	body := bytes.NewBufferString(`{bad json`)
	req := httptest.NewRequest(http.MethodPut, "/schedules/myjob", body)
	rw := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestSchedulerHandler_Remove_Success(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	req := httptest.NewRequest(http.MethodDelete, "/schedules/myjob", nil)
	rw := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
}

func TestSchedulerHandler_MethodNotAllowed(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	req := httptest.NewRequest(http.MethodGet, "/schedules/myjob", nil)
	rw := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestSchedulerHandler_MissingJobID(t *testing.T) {
	h := api.NewSchedulerHandler(newTestScheduler(t))
	req := httptest.NewRequest(http.MethodPut, "/schedules/", bytes.NewBufferString(`{"interval":"1m"}`))
	rw := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}
