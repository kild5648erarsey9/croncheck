package notifier_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/croncheck/internal/notifier"
)

func TestLogSender_Name(t *testing.T) {
	ls := notifier.NewLogSender(nil)
	if ls.Name() != "log" {
		t.Errorf("expected name 'log', got %q", ls.Name())
	}
}

func TestLogSender_Send_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	ls := notifier.NewLogSender(&buf)
	e := notifier.Event{
		JobID:     "backup",
		Status:    "failure",
		Duration:  3 * time.Second,
		Message:   "exit code 1",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	if err := ls.Send(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"backup", "failure", "3s", "exit code 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q, got: %s", want, out)
		}
	}
}
