package ratelimit_test

import (
	"testing"
	"time"

	"github.com/croncheck/internal/ratelimit"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if !l.Allow("job-1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinCooldown_Suppressed(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("job-1")
	if l.Allow("job-1") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestAllow_AfterCooldownExpires_Permitted(t *testing.T) {
	now := time.Now()
	l := ratelimit.New(5 * time.Minute)

	// Inject a fake clock so we can advance time.
	calls := 0
	l = ratelimitWithClock(5*time.Minute, func() time.Time {
		calls++
		if calls == 1 {
			return now
		}
		return now.Add(6 * time.Minute)
	})

	l.Allow("job-1") // first: recorded
	if !l.Allow("job-1") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_IndependentJobIDs(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("job-1")
	if !l.Allow("job-2") {
		t.Fatal("expected different job to be allowed independently")
	}
}

func TestReset_ClearsState(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	l.Allow("job-1") // record
	l.Reset("job-1")
	if !l.Allow("job-1") {
		t.Fatal("expected allow after reset")
	}
}

func TestLen_TracksJobs(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if l.Len() != 0 {
		t.Fatalf("expected 0, got %d", l.Len())
	}
	l.Allow("job-1")
	l.Allow("job-2")
	if l.Len() != 2 {
		t.Fatalf("expected 2, got %d", l.Len())
	}
	l.Reset("job-1")
	if l.Len() != 1 {
		t.Fatalf("expected 1 after reset, got %d", l.Len())
	}
}

// ratelimitWithClock is a test helper that creates a Limiter with an injected
// clock function by reaching into the exported New constructor and patching
// the unexported field via a thin wrapper defined in export_test.go.
func ratelimitWithClock(d time.Duration, fn func() time.Time) *ratelimit.Limiter {
	return ratelimit.NewWithClock(d, fn)
}
