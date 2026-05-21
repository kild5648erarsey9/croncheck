// export_test.go exposes internal helpers for white-box testing.
package audit

import "time"

// SetClock replaces the clock function on a Log for testing purposes.
func SetClock(l *Log, fn func() time.Time) {
	l.clock = fn
}
