package alert_test

import (
	"errors"
	"testing"
	"time"

	"croncheck/internal/alert"
)

// recordingChannel captures sent events for assertions.
type recordingChannel struct {
	events []alert.Event
	errOnSend error
}

func (r *recordingChannel) Send(event alert.Event) error {
	r.events = append(r.events, event)
	return r.errOnSend
}

func makeEvent(kind alert.EventKind) alert.Event {
	return alert.Event{
		Kind:      kind,
		JobName:   "backup",
		OccuredAt: time.Now(),
		Message:   "something went wrong",
	}
}

func TestManager_NotifyDelivered(t *testing.T) {
	ch := &recordingChannel{}
	m := alert.NewManager(ch)

	event := makeEvent(alert.EventJobFailed)
	errs := m.Notify(event)

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if len(ch.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ch.events))
	}
	if ch.events[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", ch.events[0].JobName)
	}
}

func TestManager_NotifyMultipleChannels(t *testing.T) {
	ch1 := &recordingChannel{}
	ch2 := &recordingChannel{}
	m := alert.NewManager(ch1, ch2)

	m.Notify(makeEvent(alert.EventJobTimeout))

	if len(ch1.events) != 1 || len(ch2.events) != 1 {
		t.Error("expected each channel to receive exactly one event")
	}
}

func TestManager_NotifyCollectsErrors(t *testing.T) {
	sentinel := errors.New("channel unavailable")
	ch := &recordingChannel{errOnSend: sentinel}
	m := alert.NewManager(ch)

	errs := m.Notify(makeEvent(alert.EventJobFailed))

	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !errors.Is(errs[0], sentinel) {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

func TestEvent_String(t *testing.T) {
	e := alert.Event{
		Kind:      alert.EventJobFailed,
		JobName:   "cleanup",
		OccuredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Message:   "exit code 1",
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}
