package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/alert"
)

func TestWebhookSender_Name(t *testing.T) {
	s := NewWebhookSender("http://example.com/hook")
	if s.Name() != "webhook" {
		t.Errorf("expected name 'webhook', got %q", s.Name())
	}
}

func TestWebhookSender_Send_Success(t *testing.T) {
	var received alert.Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewWebhookSender(server.URL)
	event := alert.Event{
		JobID:   "job-1",
		Message: "job failed",
		SentAt:  time.Now(),
	}

	if err := s.Send(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.JobID != event.JobID {
		t.Errorf("expected JobID %q, got %q", event.JobID, received.JobID)
	}
	if received.Message != event.Message {
		t.Errorf("expected Message %q, got %q", event.Message, received.Message)
	}
}

func TestWebhookSender_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewWebhookSender(server.URL)
	event := alert.Event{JobID: "job-2", Message: "timeout"}

	if err := s.Send(event); err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

func TestWebhookSender_Send_InvalidURL(t *testing.T) {
	s := NewWebhookSender("http://127.0.0.1:0/no-server")
	event := alert.Event{JobID: "job-3", Message: "fail"}

	if err := s.Send(event); err == nil {
		t.Error("expected connection error, got nil")
	}
}

func TestWebhookSender_Send_NonSuccessStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"bad request", http.StatusBadRequest},
		{"unauthorized", http.StatusUnauthorized},
		{"not found", http.StatusNotFound},
		{"service unavailable", http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			s := NewWebhookSender(server.URL)
			event := alert.Event{JobID: "job-4", Message: "check"}

			if err := s.Send(event); err == nil {
				t.Errorf("expected error for status %d, got nil", tc.statusCode)
			}
		})
	}
}
