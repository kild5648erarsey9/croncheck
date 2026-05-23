// Package replay provides the ability to re-send alert notifications for
// historical audit log entries.
//
// A Replayer is constructed with an audit.Log and a notifier.Notifier.
// Callers submit a Request specifying an optional job ID and/or time window;
// the Replayer iterates over matching entries and dispatches a notification
// for each one, prefixing the message with "[replay]" to distinguish
// replayed events from live ones.
//
// Typical use-case: an operator wants to re-alert on-call channels for all
// failures recorded in the past hour without restarting any cron jobs.
//
//	result := replayer.Run(replay.Request{
//	    JobID: "nightly-backup",
//	    Since: time.Now().Add(-time.Hour),
//	})
//	fmt.Printf("replayed %d events\n", result.Replayed)
package replay
