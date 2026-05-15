// Package healthcheck provides a simple health status tracker for croncheck.
package healthcheck

import (
	"sync"
	"time"
)

// Status represents the overall health of the daemon.
type Status struct {
	Healthy   bool              `json:"healthy"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Check  `json:"checks"`
}

// Check holds the result of a single named health check.
type Check struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// Checker evaluates and reports system health.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]func() Check
}

// New creates a new Checker with no registered checks.
func New() *Checker {
	return &Checker{
		checks: make(map[string]func() Check),
	}
}

// Register adds a named health check function.
func (c *Checker) Register(name string, fn func() Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = fn
}

// Run executes all registered checks and returns an aggregated Status.
func (c *Checker) Run() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]Check, len(c.checks))
	healthy := true

	for name, fn := range c.checks {
		result := fn()
		results[name] = result
		if !result.OK {
			healthy = false
		}
	}

	return Status{
		Healthy:   healthy,
		Timestamp: time.Now().UTC(),
		Checks:    results,
	}
}
