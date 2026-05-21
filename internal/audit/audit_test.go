package audit_test

import (
	"testing"
	"time"

	"github.com/croncheck/internal/audit"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRecord_AppendsEntry(t *testing.T) {
	l := audit.New()
	l.Record(audit.EventJobStarted, "job-1", "started", nil)
	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := audit.New()
	l.Record(audit.EventJobStarted, "job-1", "started", nil)
	l.Record(audit.EventJobFinished, "job-1", "finished", nil)

	entries := l.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Mutating the returned slice must not affect the log.
	entries[0].JobID = "mutated"
	if l.All()[0].JobID != "job-1" {
		t.Error("audit log was mutated through returned slice")
	}
}

func TestFilterByJob_OnlyMatchingEntries(t *testing.T) {
	l := audit.New()
	l.Record(audit.EventJobStarted, "job-1", "started", nil)
	l.Record(audit.EventJobStarted, "job-2", "started", nil)
	l.Record(audit.EventJobFinished, "job-1", "finished", nil)

	result := l.FilterByJob("job-1")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries for job-1, got %d", len(result))
	}
	for _, e := range result {
		if e.JobID != "job-1" {
			t.Errorf("unexpected job ID %q in filtered results", e.JobID)
		}
	}
}

func TestFilterByJob_NotFound_ReturnsNil(t *testing.T) {
	l := audit.New()
	result := l.FilterByJob("nonexistent")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestRecord_StoresTimestamp(t *testing.T) {
	l := audit.New()
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	// Use export_test helper to inject clock.
	audit.SetClock(l, fixedClock(now))
	l.Record(audit.EventAlertSent, "job-1", "alert sent", map[string]string{"channel": "webhook"})

	entries := l.All()
	if !entries[0].Timestamp.Equal(now) {
		t.Errorf("expected timestamp %v, got %v", now, entries[0].Timestamp)
	}
}

func TestRecord_MetaStored(t *testing.T) {
	l := audit.New()
	meta := map[string]string{"exit_code": "1", "duration_ms": "3200"}
	l.Record(audit.EventJobFinished, "job-3", "finished with error", meta)

	e := l.All()[0]
	if e.Meta["exit_code"] != "1" {
		t.Errorf("expected exit_code=1, got %q", e.Meta["exit_code"])
	}
}
