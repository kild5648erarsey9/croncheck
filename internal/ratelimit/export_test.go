// export_test.go exposes internal constructors for white-box testing.
package ratelimit

import "time"

// NewWithClock creates a Limiter with a custom clock function, enabling
// deterministic time-based tests without real sleeps.
func NewWithClock(cooldown time.Duration, now func() time.Time) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
		now:      now,
	}
}
