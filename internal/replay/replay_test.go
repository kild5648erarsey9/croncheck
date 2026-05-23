package replay_test

import (
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/audit"
	"github.com/yourorg/croncheck/internal/notifier"
	"github.com/yourorg/croncheck/internal/replay"
)

func makeLog(t *testing.T) *audit.Log {
	t.Helper()
	return audit.New()
}

func makeNotifier(t *testing.T) *notifier.Notifier {
	t.Helper()
	return notifier.New(nil) // no senders — notifications are no-ops
}

func TestReplayer_ReplayAll(t *testing.T) {
	log := makeLog(t)
	log.Record(audit.Entry{JobID: "job-a", Status: "success", Timestamp: time.Now()})
	log.Record(audit.Entry{JobID: "job-b", Status: "failure", Timestamp: time.Now()})

	r := replay.New(log, makeNotifier(t))
	res := r.Run(replay.Request{})

	if res.Replayed != 2 {
		t.Fatalf("expected 2 replayed, got %d", res.Replayed)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
}

func TestReplayer_FilterByJobID(t *testing.T) {
	log := makeLog(t)
	log.Record(audit.Entry{JobID: "job-a", Status: "success", Timestamp: time.Now()})
	log.Record(audit.Entry{JobID: "job-b", Status: "success", Timestamp: time.Now()})

	r := replay.New(log, makeNotifier(t))
	res := r.Run(replay.Request{JobID: "job-a"})

	if res.Replayed != 1 {
		t.Fatalf("expected 1 replayed, got %d", res.Replayed)
	}
}

func TestReplayer_FilterBySince(t *testing.T) {
	log := makeLog(t)
	now := time.Now()
	log.Record(audit.Entry{JobID: "old", Status: "success", Timestamp: now.Add(-2 * time.Hour)})
	log.Record(audit.Entry{JobID: "new", Status: "success", Timestamp: now})

	r := replay.New(log, makeNotifier(t))
	res := r.Run(replay.Request{Since: now.Add(-time.Hour)})

	if res.Replayed != 1 {
		t.Fatalf("expected 1 replayed, got %d", res.Replayed)
	}
}

func TestReplayer_FilterBefore(t *testing.T) {
	log := makeLog(t)
	now := time.Now()
	log.Record(audit.Entry{JobID: "old", Status: "success", Timestamp: now.Add(-2 * time.Hour)})
	log.Record(audit.Entry{JobID: "new", Status: "success", Timestamp: now})

	r := replay.New(log, makeNotifier(t))
	res := r.Run(replay.Request{Before: now.Add(-time.Hour)})

	if res.Replayed != 1 {
		t.Fatalf("expected 1 replayed, got %d", res.Replayed)
	}
}
