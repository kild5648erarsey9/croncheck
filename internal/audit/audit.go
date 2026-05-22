package audit

import (
	"sync"
	"time"
)

// Entry records the outcome of a single cron job execution.
type Entry struct {
	JobID     string
	Success   bool
	Duration  time.Duration
	Timestamp time.Time
}

// clock is a replaceable time source for testing.
var clock = func() time.Time { return time.Now() }

// Log is a thread-safe, append-only store of audit entries.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
}

// New creates an empty audit Log.
func New() *Log {
	return &Log{}
}

// Record appends a new entry for the given job.
func (l *Log) Record(jobID string, success bool, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, Entry{
		JobID:     jobID,
		Success:   success,
		Duration:  duration,
		Timestamp: clock(),
	})
}

// All returns a shallow copy of all entries.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Entry, len(l.entries))
	copy(result, l.entries)
	return result
}

// FilterByJob returns entries matching the given job ID, or nil if none found.
func (l *Log) FilterByJob(jobID string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var result []Entry
	for _, e := range l.entries {
		if e.JobID == jobID {
			result = append(result, e)
		}
	}
	return result
}

// Len returns the total number of recorded entries.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
