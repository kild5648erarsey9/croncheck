package metrics

import (
	"testing"
	"time"
)

func TestCollector_RecordSuccess(t *testing.T) {
	c := NewCollector()
	c.Record("job1", 2*time.Second, true)

	m, ok := c.Get("job1")
	if !ok {
		t.Fatal("expected metrics for job1")
	}
	if m.TotalRuns != 1 {
		t.Errorf("TotalRuns = %d, want 1", m.TotalRuns)
	}
	if m.SuccessRuns != 1 {
		t.Errorf("SuccessRuns = %d, want 1", m.SuccessRuns)
	}
	if m.FailureRuns != 0 {
		t.Errorf("FailureRuns = %d, want 0", m.FailureRuns)
	}
	if m.LastDuration != 2*time.Second {
		t.Errorf("LastDuration = %v, want 2s", m.LastDuration)
	}
}

func TestCollector_RecordFailure(t *testing.T) {
	c := NewCollector()
	c.Record("job2", 500*time.Millisecond, false)

	m, ok := c.Get("job2")
	if !ok {
		t.Fatal("expected metrics for job2")
	}
	if m.FailureRuns != 1 {
		t.Errorf("FailureRuns = %d, want 1", m.FailureRuns)
	}
}

func TestCollector_AvgDuration(t *testing.T) {
	c := NewCollector()
	c.Record("job3", 2*time.Second, true)
	c.Record("job3", 4*time.Second, true)

	m, _ := c.Get("job3")
	if m.TotalRuns != 2 {
		t.Errorf("TotalRuns = %d, want 2", m.TotalRuns)
	}
	if m.AvgDuration != 3*time.Second {
		t.Errorf("AvgDuration = %v, want 3s", m.AvgDuration)
	}
}

func TestCollector_GetNotFound(t *testing.T) {
	c := NewCollector()
	_, ok := c.Get("missing")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestCollector_All(t *testing.T) {
	c := NewCollector()
	c.Record("a", time.Second, true)
	c.Record("b", time.Second, false)

	all := c.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d entries, want 2", len(all))
	}
}

func TestCollector_Snapshot_Isolation(t *testing.T) {
	c := NewCollector()
	c.Record("job", time.Second, true)

	m, _ := c.Get("job")
	m.TotalRuns = 999

	original, _ := c.Get("job")
	if original.TotalRuns == 999 {
		t.Error("Get returned a reference instead of a copy")
	}
}
