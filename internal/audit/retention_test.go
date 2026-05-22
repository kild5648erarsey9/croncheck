package audit

import (
	"testing"
	"time"
)

func TestReaper_PurgeByMaxAge(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return base }

	log := New()
	SetClock(log, clock)

	// Record entries at different times
	SetClock(log, func() time.Time { return base.Add(-2 * time.Hour) })
	log.Record("job-a", "success", "")

	SetClock(log, func() time.Time { return base.Add(-30 * time.Minute) })
	log.Record("job-b", "success", "")

	SetClock(log, clock)
	log.Record("job-c", "success", "")

	policy := RetentionPolicy{MaxAge: 1 * time.Hour}
	reaper := NewReaper(log, policy)
	reaper.purge()

	entries := log.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after purge, got %d", len(entries))
	}
	if entries[0].JobID != "job-b" {
		t.Errorf("expected job-b, got %s", entries[0].JobID)
	}
	if entries[1].JobID != "job-c" {
		t.Errorf("expected job-c, got %s", entries[1].JobID)
	}
}

func TestReaper_PurgeByMaxEntries(t *testing.T) {
	log := New()
	for i := 0; i < 10; i++ {
		log.Record("job", "success", "")
	}

	policy := RetentionPolicy{MaxEntries: 5}
	reaper := NewReaper(log, policy)
	reaper.purge()

	if got := len(log.All()); got != 5 {
		t.Fatalf("expected 5 entries, got %d", got)
	}
}

func TestReaper_PurgeCombined(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	log := New()
	SetClock(log, func() time.Time { return base.Add(-2 * time.Hour) })
	for i := 0; i < 3; i++ {
		log.Record("old-job", "success", "")
	}
	SetClock(log, func() time.Time { return base })
	for i := 0; i < 8; i++ {
		log.Record("new-job", "success", "")
	}

	policy := RetentionPolicy{MaxAge: 1 * time.Hour, MaxEntries: 4}
	reaper := NewReaper(log, policy)
	reaper.purge()

	entries := log.All()
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.JobID != "new-job" {
			t.Errorf("expected only new-job entries, got %s", e.JobID)
		}
	}
}

func TestReaper_StartStop(t *testing.T) {
	log := New()
	for i := 0; i < 20; i++ {
		log.Record("job", "success", "")
	}

	policy := RetentionPolicy{MaxEntries: 5}
	reaper := NewReaper(log, policy)
	reaper.Start(10 * time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	reaper.Stop()

	if got := len(log.All()); got > 5 {
		t.Errorf("expected at most 5 entries after reaper ran, got %d", got)
	}
}
