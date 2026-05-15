package watcher

import (
	"log"
	"time"

	"github.com/croncheck/internal/alert"
	"github.com/croncheck/internal/job"
)

// Watcher periodically checks all registered jobs for timeouts
// and emits alerts when a job exceeds its expected duration.
type Watcher struct {
	registry *job.Registry
	manager  *alert.Manager
	interval time.Duration
	stop     chan struct{}
}

// New creates a new Watcher that polls at the given interval.
func New(registry *job.Registry, manager *alert.Manager, interval time.Duration) *Watcher {
	return &Watcher{
		registry: registry,
		manager:  manager,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start begins the watch loop in a background goroutine.
func (w *Watcher) Start() {
	go w.run()
}

// Stop signals the watch loop to exit.
func (w *Watcher) Stop() {
	close(w.stop)
}

func (w *Watcher) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkJobs()
		case <-w.stop:
			log.Println("watcher: stopped")
			return
		}
	}
}

func (w *Watcher) checkJobs() {
	for _, j := range w.registry.All() {
		snap := j.Snapshot()
		if !snap.Running {
			continue
		}
		if snap.MaxDuration <= 0 {
			continue
		}
		elapsed := time.Since(snap.StartedAt)
		if elapsed > snap.MaxDuration {
			log.Printf("watcher: job %q exceeded max duration (%s > %s)",
				snap.ID, elapsed.Round(time.Second), snap.MaxDuration)
			ev := alert.Event{
				JobID:   snap.ID,
				Kind:    alert.KindTimeout,
				Message: "job exceeded max duration",
				Elapsed: elapsed,
			}
			if err := w.manager.Notify(ev); err != nil {
				log.Printf("watcher: alert error for job %q: %v", snap.ID, err)
			}
		}
	}
}
