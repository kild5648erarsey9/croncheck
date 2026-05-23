// Package replay provides functionality to re-trigger alert notifications
// for past audit log entries that match a given filter.
package replay

import (
	"fmt"
	"time"

	"github.com/yourorg/croncheck/internal/audit"
	"github.com/yourorg/croncheck/internal/notifier"
)

// Request describes the parameters for a replay operation.
type Request struct {
	JobID  string
	Since  time.Time
	Before time.Time
}

// Result summarises the outcome of a replay operation.
type Result struct {
	Replayed int
	Errors   []string
}

// Replayer re-sends notifications for historical audit entries.
type Replayer struct {
	log      *audit.Log
	notifier *notifier.Notifier
}

// New creates a Replayer backed by the given audit log and notifier.
func New(log *audit.Log, n *notifier.Notifier) *Replayer {
	return &Replayer{log: log, notifier: n}
}

// Run executes the replay for the given request and returns a summary.
func (r *Replayer) Run(req Request) Result {
	var entries []audit.Entry

	if req.JobID != "" {
		entries = r.log.FilterByJob(req.JobID)
	} else {
		entries = r.log.All()
	}

	var result Result
	for _, e := range entries {
		if !req.Since.IsZero() && e.Timestamp.Before(req.Since) {
			continue
		}
		if !req.Before.IsZero() && !e.Timestamp.Before(req.Before) {
			continue
		}
		event := notifier.Event{
			JobID:    e.JobID,
			Status:   e.Status,
			Duration: e.Duration,
			Message:  fmt.Sprintf("[replay] %s", e.Message),
		}
		if err := r.notifier.Notify(event); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", e.JobID, err))
			continue
		}
		result.Replayed++
	}
	return result
}
