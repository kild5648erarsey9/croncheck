package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/croncheck/internal/healthcheck"
)

func newTestHealthChecker() *healthcheck.Checker {
	return healthcheck.New()
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	checker := newTestHealthChecker()
	checker.Register("db", func() error { return nil })
	checker.Register("cache", func() error { return nil })

	h := NewHealthHandler(checker)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
}

func TestHealthHandler_OneFailing(t *testing.T) {
	checker := newTestHealthChecker()
	checker.Register("db", func() error { return nil })
	checker.Register("broken", func() error { return fmt.Errorf("connection refused") })

	h := NewHealthHandler(checker)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "degraded" {
		t.Errorf("expected status degraded, got %v", body["status"])
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	checker := newTestHealthChecker()
	h := NewHealthHandler(checker)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/health", nil)
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s: expected 405, got %d", method, rr.Code)
		}
	}
}

func TestRegisterHealthRoute(t *testing.T) {
	checker := newTestHealthChecker()
	h := NewHealthHandler(checker)
	mux := http.NewServeMux()
	h.RegisterHealthRoute(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 via mux, got %d", rr.Code)
	}
}
