package notifier

import (
	"fmt"
	"io"
	"os"
)

// LogSender writes events to an io.Writer (defaults to os.Stderr).
type LogSender struct {
	out io.Writer
}

// NewLogSender creates a LogSender that writes to w.
// If w is nil, os.Stderr is used.
func NewLogSender(w io.Writer) *LogSender {
	if w == nil {
		w = os.Stderr
	}
	return &LogSender{out: w}
}

// Name returns the sender identifier.
func (l *LogSender) Name() string {
	return "log"
}

// Send formats the event and writes it to the configured writer.
func (l *LogSender) Send(event Event) error {
	_, err := fmt.Fprintf(
		l.out,
		"[%s] job=%s status=%s duration=%s message=%q\n",
		event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		event.JobID,
		event.Status,
		event.Duration,
		event.Message,
	)
	return err
}
