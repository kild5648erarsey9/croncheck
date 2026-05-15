// Package scheduler provides interval-based monitoring for cron jobs.
//
// A Scheduler holds a set of Schedule entries, each associating a job ID
// with an expected execution interval. On every tick the Scheduler queries
// the job registry and logs a warning for any job whose last start time
// exceeds its configured interval.
//
// Typical usage:
//
//	s := scheduler.New(registry, 30*time.Second)
//	s.Add(scheduler.Schedule{JobID: "backup", Interval: 24*time.Hour})
//	s.Start()
//	// …
//	s.Stop()
package scheduler
