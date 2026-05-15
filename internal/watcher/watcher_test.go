package watcher_test

import (
	"testing"
	"time"

	"github.com/croncheck/internal/alert"
	"github.com/croncheck/internal/job"
	"github.com/croncheck/internal/watcher"
)

// captureChannel records every event it receives.
type captureChannel struct {
	events []alert.Event
}

func (c *captureChannel) Name() string { return "capture" }
func (c *captureChannel) Send(ev alert.Event) error {
	c.events = append(c.events, ev)
	return nil
}

func newTestSetup(t *testing.T) (*job.Registry, *alert.Manager, *captureChannel) {
	t.Helper()
	reg := job.NewRegistry()
	cap := &captureChannel{}
	mgr := alert.NewManager([]alert.Channel{cap})
	return reg, mgr, cap
}

func TestWatcher_DetectsTimeout(t *testing.T) {
	reg, mgr, cap := newTestSetup(t)

	j, err := job.New("slow-job", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	if err := reg.Register(j); err != nil {
		t.Fatalf("Register: %v", err)
	}
	j.Start()

	w := watcher.New(reg, mgr, 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	// Wait long enough for the watcher to fire at least once after the job times out.
	time.Sleep(120 * time.Millisecond)

	if len(cap.events) == 0 {
		t.Fatal("expected at least one timeout alert, got none")
	}
	ev := cap.events[0]
	if ev.JobID != "slow-job" {
		t.Errorf("expected job ID %q, got %q", "slow-job", ev.JobID)
	}
	if ev.Kind != alert.KindTimeout {
		t.Errorf("expected kind %q, got %q", alert.KindTimeout, ev.Kind)
	}
}

func TestWatcher_NoAlertWhenNotRunning(t *testing.T) {
	reg, mgr, cap := newTestSetup(t)

	j, _ := job.New("idle-job", 10*time.Millisecond)
	_ = reg.Register(j)
	// job never started — should not trigger a timeout alert

	w := watcher.New(reg, mgr, 20*time.Millisecond)
	w.Start()
	defer w.Stop()

	time.Sleep(60 * time.Millisecond)

	if len(cap.events) != 0 {
		t.Errorf("expected no alerts, got %d", len(cap.events))
	}
}

func TestWatcher_StopHaltsChecks(t *testing.T) {
	reg, mgr, cap := newTestSetup(t)

	j, _ := job.New("bg-job", 10*time.Millisecond)
	_ = reg.Register(j)
	j.Start()

	w := watcher.New(reg, mgr, 15*time.Millisecond)
	w.Start()
	w.Stop() // stop immediately

	before := len(cap.events)
	time.Sleep(50 * time.Millisecond)
	after := len(cap.events)

	if after > before+1 {
		t.Errorf("watcher kept firing after Stop(): before=%d after=%d", before, after)
	}
}
