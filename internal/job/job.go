package job

import (
	"sync"
	"time"
)

// Status represents the current state of a cron job.
type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusTimeout Status = "timeout"
)

// Job holds the runtime state of a monitored cron job.
type Job struct {
	mu          sync.RWMutex
	Name        string
	Schedule    string
	MaxDuration time.Duration
	LastStart   time.Time
	LastEnd     time.Time
	LastStatus  Status
	FailCount   int
}

// New creates a new Job with the given name, schedule, and max duration.
func New(name, schedule string, maxDuration time.Duration) *Job {
	return &Job{
		Name:        name,
		Schedule:    schedule,
		MaxDuration: maxDuration,
		LastStatus:  StatusPending,
	}
}

// Start records the start time of a job run.
func (j *Job) Start() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.LastStart = time.Now()
	j.LastStatus = StatusRunning
}

// Finish records the end time and status of a job run.
func (j *Job) Finish(success bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.LastEnd = time.Now()
	if !success {
		j.LastStatus = StatusFailed
		j.FailCount++
		return
	}
	if j.MaxDuration > 0 && j.LastEnd.Sub(j.LastStart) > j.MaxDuration {
		j.LastStatus = StatusTimeout
		j.FailCount++
		return
	}
	j.LastStatus = StatusSuccess
	j.FailCount = 0
}

// Snapshot returns a read-safe copy of the job's current state.
func (j *Job) Snapshot() Job {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return Job{
		Name:        j.Name,
		Schedule:    j.Schedule,
		MaxDuration: j.MaxDuration,
		LastStart:   j.LastStart,
		LastEnd:     j.LastEnd,
		LastStatus:  j.LastStatus,
		FailCount:   j.FailCount,
	}
}
