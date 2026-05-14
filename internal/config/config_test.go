package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "croncheck-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
listen_addr: ":9090"
alert_channels:
  - name: ops-slack
    type: slack
    options:
      webhook_url: https://hooks.slack.com/test
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    max_duration: 30m
    grace_period: 5m
    alert_channels: [ops-slack]
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected listen_addr :9090, got %s", cfg.ListenAddr)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].MaxDuration != 30*time.Minute {
		t.Errorf("expected max_duration 30m, got %v", cfg.Jobs[0].MaxDuration)
	}
}

func TestLoad_DefaultListenAddr(t *testing.T) {
	path := writeTempConfig(t, `
alert_channels: []
jobs: []
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default :8080, got %s", cfg.ListenAddr)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_UnknownAlertChannel(t *testing.T) {
	path := writeTempConfig(t, `
alert_channels: []
jobs:
  - name: nightly-sync
    schedule: "0 3 * * *"
    alert_channels: [ghost-channel]
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for unknown alert channel, got nil")
	}
}

func TestLoad_JobMissingSchedule(t *testing.T) {
	path := writeTempConfig(t, `
alert_channels: []
jobs:
  - name: broken-job
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing schedule, got nil")
	}
}
