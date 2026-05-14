package job

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	j := New("backup", "0 2 * * *", 30*time.Minute)
	if j.Name != "backup" {
		t.Errorf("expected name 'backup', got %q", j.Name)
	}
	if j.LastStatus != StatusPending {
		t.Errorf("expected status pending, got %q", j.LastStatus)
	}
}

func TestJob_StartAndFinishSuccess(t *testing.T) {
	j := New("sync", "*/5 * * * *", time.Minute)
	j.Start()
	if j.LastStatus != StatusRunning {
		t.Errorf("expected running after Start, got %q", j.LastStatus)
	}
	j.Finish(true)
	if j.LastStatus != StatusSuccess {
		t.Errorf("expected success, got %q", j.LastStatus)
	}
	if j.FailCount != 0 {
		t.Errorf("expected FailCount 0, got %d", j.FailCount)
	}
}

func TestJob_FinishFailure(t *testing.T) {
	j := New("report", "0 8 * * *", 0)
	j.Start()
	j.Finish(false)
	if j.LastStatus != StatusFailed {
		t.Errorf("expected failed, got %q", j.LastStatus)
	}
	if j.FailCount != 1 {
		t.Errorf("expected FailCount 1, got %d", j.FailCount)
	}
}

func TestJob_Timeout(t *testing.T) {
	j := New("slow", "0 * * * *", 1*time.Millisecond)
	j.Start()
	time.Sleep(5 * time.Millisecond)
	j.Finish(true)
	if j.LastStatus != StatusTimeout {
		t.Errorf("expected timeout, got %q", j.LastStatus)
	}
	if j.FailCount != 1 {
		t.Errorf("expected FailCount 1, got %d", j.FailCount)
	}
}

func TestJob_Snapshot_Isolation(t *testing.T) {
	j := New("snap", "@daily", 0)
	j.Start()
	snap := j.Snapshot()
	j.Finish(false)
	if snap.LastStatus != StatusRunning {
		t.Errorf("snapshot should reflect state at time of call, got %q", snap.LastStatus)
	}
}
