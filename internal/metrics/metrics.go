package metrics

import (
	"sync"
	"time"
)

// JobMetrics holds runtime statistics for a single cron job.
type JobMetrics struct {
	JobID        string
	TotalRuns    int64
	SuccessRuns  int64
	FailureRuns  int64
	LastDuration time.Duration
	AvgDuration  time.Duration
	LastRunAt    time.Time
	totalNanos   int64
}

// Collector aggregates metrics across all tracked jobs.
type Collector struct {
	mu   sync.RWMutex
	jobs map[string]*JobMetrics
}

// NewCollector creates an empty Collector.
func NewCollector() *Collector {
	return &Collector{
		jobs: make(map[string]*JobMetrics),
	}
}

// Record updates metrics for a job after a run completes.
func (c *Collector) Record(jobID string, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.jobs[jobID]
	if !ok {
		m = &JobMetrics{JobID: jobID}
		c.jobs[jobID] = m
	}

	m.TotalRuns++
	m.LastDuration = duration
	m.LastRunAt = time.Now()
	m.totalNanos += duration.Nanoseconds()
	m.AvgDuration = time.Duration(m.totalNanos / m.TotalRuns)

	if success {
		m.SuccessRuns++
	} else {
		m.FailureRuns++
	}
}

// Get returns a copy of the metrics for the given job ID.
// Returns false if the job has not been recorded yet.
func (c *Collector) Get(jobID string) (JobMetrics, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.jobs[jobID]
	if !ok {
		return JobMetrics{}, false
	}
	return *m, true
}

// All returns a snapshot of metrics for every tracked job.
func (c *Collector) All() []JobMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]JobMetrics, 0, len(c.jobs))
	for _, m := range c.jobs {
		out = append(out, *m)
	}
	return out
}
