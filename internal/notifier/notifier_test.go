package notifier_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/croncheck/internal/notifier"
)

// mockSender captures sent events for assertion.
type mockSender struct {
	name   string
	events []notifier.Event
	errOn  bool
}

func (m *mockSender) Name() string { return m.name }
func (m *mockSender) Send(e notifier.Event) error {
	if m.errOn {
		return errors.New("mock error")
	}
	m.events = append(m.events, e)
	return nil
}

func makeEvent(status string) notifier.Event {
	return notifier.Event{
		JobID:    "job-1",
		Status:   status,
		Duration: 2 * time.Second,
		Message:  "test",
	}
}

func TestNotifier_Notify_Delivered(t *testing.T) {
	s := &mockSender{name: "s1"}
	n := notifier.New(s)
	if err := n.Notify(makeEvent("success")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(s.events))
	}
	if s.events[0].Status != "success" {
		t.Errorf("expected status success, got %s", s.events[0].Status)
	}
}

func TestNotifier_Notify_SetsTimestamp(t *testing.T) {
	s := &mockSender{name: "s1"}
	n := notifier.New(s)
	e := makeEvent("failure")
	_ = n.Notify(e)
	if s.events[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestNotifier_Notify_MultiplesSenders(t *testing.T) {
	s1 := &mockSender{name: "s1"}
	s2 := &mockSender{name: "s2"}
	n := notifier.New(s1, s2)
	_ = n.Notify(makeEvent("timeout"))
	if len(s1.events) != 1 || len(s2.events) != 1 {
		t.Error("expected each sender to receive one event")
	}
}

func TestNotifier_Notify_CollectsErrors(t *testing.T) {
	s := &mockSender{name: "bad", errOn: true}
	n := notifier.New(s)
	err := n.Notify(makeEvent("failure"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("error should mention sender name, got: %v", err)
	}
}

func TestNotifier_SenderCount(t *testing.T) {
	n := notifier.New(&mockSender{name: "a"}, &mockSender{name: "b"})
	if n.SenderCount() != 2 {
		t.Errorf("expected 2 senders, got %d", n.SenderCount())
	}
}
