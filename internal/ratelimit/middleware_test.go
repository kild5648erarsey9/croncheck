package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/ratelimit"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func TestHTTPMiddleware_FirstRequest_PassesThrough(t *testing.T) {
	rl := ratelimit.New()
	handler := ratelimit.HTTPMiddleware(rl, 5*time.Minute, http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodPost, "/jobs/my-job/finish", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHTTPMiddleware_SecondRequest_WithinCooldown_RateLimited(t *testing.T) {
	rl := ratelimit.New()
	handler := ratelimit.HTTPMiddleware(rl, 5*time.Minute, http.HandlerFunc(okHandler))

	for i, wantCode := range []int{http.StatusOK, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodPost, "/jobs/my-job/finish", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != wantCode {
			t.Fatalf("request %d: expected %d, got %d", i+1, wantCode, rec.Code)
		}
	}
}

func TestHTTPMiddleware_EmptyJobID_PassesThrough(t *testing.T) {
	rl := ratelimit.New()
	handler := ratelimit.HTTPMiddleware(rl, 5*time.Minute, http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHTTPMiddleware_DifferentJobs_IndependentLimits(t *testing.T) {
	rl := ratelimit.New()
	handler := ratelimit.HTTPMiddleware(rl, 5*time.Minute, http.HandlerFunc(okHandler))

	paths := []string{"/jobs/job-a/finish", "/jobs/job-b/finish"}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200, got %d", path, rec.Code)
		}
	}
}
