package notifier

import (
	"fmt"
	"time"
)

// Event represents a cron job lifecycle event to be reported.
type Event struct {
	JobID     string
	Status    string // "success", "failure", "timeout"
	Duration  time.Duration
	Message   string
	Timestamp time.Time
}

// Sender defines the interface for sending notifications.
type Sender interface {
	Send(event Event) error
	Name() string
}

// Notifier dispatches events to one or more Sender implementations.
type Notifier struct {
	senders []Sender
}

// New creates a Notifier with the provided senders.
func New(senders ...Sender) *Notifier {
	return &Notifier{senders: senders}
}

// Notify dispatches the event to all registered senders.
// It collects and returns all errors encountered.
func (n *Notifier) Notify(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	var errs []error
	for _, s := range n.senders {
		if err := s.Send(event); err != nil {
			errs = append(errs, fmt.Errorf("sender %q: %w", s.Name(), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notifier errors: %v", errs)
	}
	return nil
}

// SenderCount returns the number of registered senders.
func (n *Notifier) SenderCount() int {
	return len(n.senders)
}
