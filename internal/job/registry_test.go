package job

import (
	"testing"
	"time"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	j := New("db-backup", "0 3 * * *", 20*time.Minute)

	if err := r.Register(j); err != nil {
		t.Fatalf("unexpected error registering job: %v", err)
	}

	got, err := r.Get("db-backup")
	if err != nil {
		t.Fatalf("unexpected error getting job: %v", err)
	}
	if got.Name != "db-backup" {
		t.Errorf("expected 'db-backup', got %q", got.Name)
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	r := NewRegistry()
	j := New("metrics", "*/1 * * * *", 0)

	if err := r.Register(j); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if err := r.Register(j); err == nil {
		t.Fatal("expected error on duplicate register, got nil")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("ghost")
	if err == nil {
		t.Fatal("expected error for unknown job, got nil")
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	names := []string{"job-a", "job-b", "job-c"}
	for _, n := range names {
		_ = r.Register(New(n, "@hourly", 0))
	}

	all := r.All()
	if len(all) != len(names) {
		t.Errorf("expected %d jobs, got %d", len(names), len(all))
	}
}

func TestRegistry_Len(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("expected 0, got %d", r.Len())
	}
	_ = r.Register(New("one", "@daily", 0))
	if r.Len() != 1 {
		t.Errorf("expected 1, got %d", r.Len())
	}
}
