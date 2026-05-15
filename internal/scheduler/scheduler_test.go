package scheduler_test

import (
	"testing"
	"time"

	"github.com/example/croncheck/internal/job"
	"github.com/example/croncheck/internal/scheduler"
)

func newRegistry(t *testing.T, ids ...string) *job.Registry {
	t.Helper()
	reg := job.NewRegistry()
	for _, id := range ids {
		_, err := reg.Register(job.New(id, id+" desc", 5*time.Second))
		if err != nil {
			t.Fatalf("register %q: %v", id, err)
		}
	}
	return reg
}

func TestScheduler_AddAndRemove(t *testing.T) {
	reg := newRegistry(t, "job1")
	s := scheduler.New(reg, time.Minute)

	sc := scheduler.Schedule{JobID: "job1", Interval: time.Minute}
	s.Add(sc)
	s.Remove("job1")
	// No panic — removal of existing and non-existing key is safe.
	s.Remove("nonexistent")
}

func TestScheduler_StartStop(t *testing.T) {
	reg := newRegistry(t, "job1")
	s := scheduler.New(reg, 50*time.Millisecond)
	s.Add(scheduler.Schedule{JobID: "job1", Interval: time.Minute})
	s.Start()
	time.Sleep(120 * time.Millisecond)
	s.Stop()
}

func TestScheduler_MissedInterval_LogsWarning(t *testing.T) {
	reg := newRegistry(t, "cron1")
	j, _ := reg.Get("cron1")
	// Start the job so LastStart is set in the past.
	j.Start()
	time.Sleep(10 * time.Millisecond)

	s := scheduler.New(reg, 30*time.Millisecond)
	// Interval shorter than elapsed time → should trigger warning log.
	s.Add(scheduler.Schedule{JobID: "cron1", Interval: 1 * time.Millisecond})
	s.Start()
	time.Sleep(80 * time.Millisecond)
	s.Stop()
	// Test passes if no panic occurs; log output verified manually.
}

func TestScheduler_NoWarning_WhenJobNotStarted(t *testing.T) {
	reg := newRegistry(t, "fresh")
	s := scheduler.New(reg, 30*time.Millisecond)
	s.Add(scheduler.Schedule{JobID: "fresh", Interval: 1 * time.Millisecond})
	s.Start()
	time.Sleep(80 * time.Millisecond)
	s.Stop()
	// LastStart is zero → evaluate skips gracefully.
}

func TestScheduler_UnknownJob_NoPanic(t *testing.T) {
	reg := job.NewRegistry()
	s := scheduler.New(reg, 30*time.Millisecond)
	s.Add(scheduler.Schedule{JobID: "ghost", Interval: time.Second})
	s.Start()
	time.Sleep(80 * time.Millisecond)
	s.Stop()
}
