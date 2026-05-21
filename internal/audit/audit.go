// Package audit provides a simple append-only audit log for cron job lifecycle events.
package audit

import (
	"sync"
	"time"
)

// EventKind describes the type of audit event.
type EventKind string

const (
	EventJobStarted  EventKind = "job.started"
	EventJobFinished EventKind = "job.finished"
	EventJobTimeout  EventKind = "job.timeout"
	EventAlertSent   EventKind = "alert.sent"
)

// Entry is a single immutable audit log record.
type Entry struct {
	Timestamp time.Time
	Kind      EventKind
	JobID     string
	Message   string
	Meta      map[string]string
}

// Log holds an in-memory sequence of audit entries.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	clock   func() time.Time
}

// New creates a new audit Log.
func New() *Log {
	return &Log{clock: time.Now}
}

// Record appends a new entry to the log.
func (l *Log) Record(kind EventKind, jobID, message string, meta map[string]string) {
	entry := Entry{
		Timestamp: l.clock(),
		Kind:      kind,
		JobID:     jobID,
		Message:   message,
		Meta:      meta,
	}
	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
}

// All returns a snapshot of all audit entries.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// FilterByJob returns all entries for a given job ID.
func (l *Log) FilterByJob(jobID string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.JobID == jobID {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the total number of recorded entries.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
