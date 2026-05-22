package audit

import (
	"sync"
	"time"
)

// RetentionPolicy defines how long audit entries are kept.
type RetentionPolicy struct {
	MaxAge time.Duration
	MaxEntries int
}

// Reaper periodically removes audit entries that exceed the retention policy.
type Reaper struct {
	log    *Log
	policy RetentionPolicy
	stop   chan struct{}
	wg     sync.WaitGroup
}

// NewReaper creates a Reaper that enforces the given policy against the audit log.
func NewReaper(log *Log, policy RetentionPolicy) *Reaper {
	return &Reaper{
		log:    log,
		policy: policy,
		stop:   make(chan struct{}),
	}
}

// Start begins the background retention loop, running every interval.
func (r *Reaper) Start(interval time.Duration) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.purge()
			case <-r.stop:
				return
			}
		}
	}()
}

// Stop halts the background retention loop.
func (r *Reaper) Stop() {
	close(r.stop)
	r.wg.Wait()
}

// purge removes entries that violate the retention policy.
func (r *Reaper) purge() {
	r.log.mu.Lock()
	defer r.log.mu.Unlock()

	now := r.log.clock()

	if r.policy.MaxAge > 0 {
		cutoff := now.Add(-r.policy.MaxAge)
		filtered := r.log.entries[:0]
		for _, e := range r.log.entries {
			if e.Timestamp.After(cutoff) {
				filtered = append(filtered, e)
			}
		}
		r.log.entries = filtered
	}

	if r.policy.MaxEntries > 0 && len(r.log.entries) > r.policy.MaxEntries {
		r.log.entries = r.log.entries[len(r.log.entries)-r.policy.MaxEntries:]
	}
}
