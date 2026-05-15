package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/example/croncheck/internal/job"
)

// Schedule holds the expected interval for a job.
type Schedule struct {
	JobID    string
	Interval time.Duration
}

// Scheduler periodically checks whether jobs have been started
// within their expected intervals and logs warnings when they haven't.
type Scheduler struct {
	mu        sync.Mutex
	schedules map[string]Schedule
	registry  *job.Registry
	stopCh    chan struct{}
	tick      time.Duration
}

// New creates a Scheduler that evaluates schedules every tickInterval.
func New(registry *job.Registry, tickInterval time.Duration) *Scheduler {
	return &Scheduler{
		schedules: make(map[string]Schedule),
		registry:  registry,
		stopCh:    make(chan struct{}),
		tick:      tickInterval,
	}
}

// Add registers an expected schedule for a job.
func (s *Scheduler) Add(sc Schedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[sc.JobID] = sc
}

// Remove removes a schedule by job ID.
func (s *Scheduler) Remove(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.schedules, jobID)
}

// Start begins the evaluation loop in a background goroutine.
func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(s.tick)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.evaluate()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop halts the evaluation loop.
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) evaluate() {
	s.mu.Lock()
	scheds := make([]Schedule, 0, len(s.schedules))
	for _, sc := range s.schedules {
		scheds = append(scheds, sc)
	}
	s.mu.Unlock()

	for _, sc := range scheds {
		j, err := s.registry.Get(sc.JobID)
		if err != nil {
			log.Printf("[scheduler] job %q not found in registry", sc.JobID)
			continue
		}
		snap := j.Snapshot()
		if snap.LastStart.IsZero() {
			continue
		}
		if time.Since(snap.LastStart) > sc.Interval {
			log.Printf("[scheduler] job %q missed its expected interval of %s", sc.JobID, sc.Interval)
		}
	}
}
