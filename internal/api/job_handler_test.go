package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/job"
)

func registerTestJob(t *testing.T, reg *job.Registry, id string) {
	t.Helper()
	j, err := job.New(id, "test job", 5*time.Second)
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	if err := reg.Register(j); err != nil {
		t.Fatalf("Register: %v", err)
	}
}

func TestHandleStart_Success(t *testing.T) {
	reg := newTestRegistry(t)
	registerTestJob(t, reg, "backup")
	h := NewJobHandler(reg, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/backup/start", nil)
	rw := httptest.NewRecorder()
	h.HandleStart(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rw.Code)
	}
}

func TestHandleStart_NotFound(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewJobHandler(reg, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/missing/start", nil)
	rw := httptest.NewRecorder()
	h.HandleStart(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rw.Code)
	}
}

func TestHandleStart_MethodNotAllowed(t *testing.T) {
	reg := newTestRegistry(t)
	h := NewJobHandler(reg, nil)

	req := httptest.NewRequest(http.MethodGet, "/jobs/backup/start", nil)
	rw := httptest.NewRecorder()
	h.HandleStart(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rw.Code)
	}
}

func TestHandleFinish_Success(t *testing.T) {
	reg := newTestRegistry(t)
	registerTestJob(t, reg, "cleanup")
	h := NewJobHandler(reg, nil)

	body := bytes.NewBufferString(`{"exit_code":0}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs/cleanup/finish", body)
	rw := httptest.NewRecorder()
	h.HandleFinish(rw, req)

	if rw.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rw.Code)
	}
}

func TestHandleFinish_BadBody(t *testing.T) {
	reg := newTestRegistry(t)
	registerTestJob(t, reg, "cleanup")
	h := NewJobHandler(reg, nil)

	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/jobs/cleanup/finish", body)
	rw := httptest.NewRecorder()
	h.HandleFinish(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rw.Code)
	}
}

func TestExtractJobID(t *testing.T) {
	cases := []struct {
		path, suffix, want string
	}{
		{"/jobs/backup/start", "/start", "backup"},
		{"/jobs/my-job/finish", "/finish", "my-job"},
		{"/jobs", "/start", ""},
	}
	for _, c := range cases {
		got := extractJobID(c.path, c.suffix)
		if got != c.want {
			t.Errorf("extractJobID(%q, %q) = %q; want %q", c.path, c.suffix, got, c.want)
		}
	}
}
