// Package ratelimit provides a simple per-job alert rate limiter to prevent
// alert storms when a job repeatedly fails in a short window.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks the last alert time per job and suppresses duplicate alerts
// within a configurable cooldown window.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	now      func() time.Time
}

// New creates a Limiter with the given cooldown duration.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for the given jobID should be sent,
// i.e. no alert has been sent within the cooldown window. Calling Allow
// with a permitted jobID records the current time as the last sent time.
func (l *Limiter) Allow(jobID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if last, ok := l.lastSent[jobID]; ok {
		if now.Sub(last) < l.cooldown {
			return false
		}
	}
	l.lastSent[jobID] = now
	return true
}

// Reset clears the rate-limit state for a specific job. Useful when a job
// recovers and we want the next failure to alert immediately.
func (l *Limiter) Reset(jobID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastSent, jobID)
}

// Len returns the number of jobs currently tracked by the limiter.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.lastSent)
}
