package alert

import (
	"fmt"
	"log"
	"time"
)

// Channel represents an alerting backend.
type Channel interface {
	Send(event Event) error
}

// EventKind describes what triggered the alert.
type EventKind string

const (
	EventJobFailed  EventKind = "job_failed"
	EventJobTimeout EventKind = "job_timeout"
)

// Event carries the information sent to an alert channel.
type Event struct {
	Kind      EventKind
	JobName   string
	OccuredAt time.Time
	Message   string
}

// String returns a human-readable representation of the event.
func (e Event) String() string {
	return fmt.Sprintf("[%s] job=%q at=%s msg=%s",
		e.Kind, e.JobName, e.OccuredAt.Format(time.RFC3339), e.Message)
}

// LogChannel is a simple Channel that writes alerts to the standard logger.
type LogChannel struct{}

// Send implements Channel for LogChannel.
func (l *LogChannel) Send(event Event) error {
	log.Printf("ALERT %s", event)
	return nil
}

// Manager dispatches events to one or more channels.
type Manager struct {
	channels []Channel
}

// NewManager creates a Manager with the provided channels.
func NewManager(channels ...Channel) *Manager {
	return &Manager{channels: channels}
}

// Notify sends the event to all registered channels, collecting errors.
func (m *Manager) Notify(event Event) []error {
	var errs []error
	for _, ch := range m.channels {
		if err := ch.Send(event); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
