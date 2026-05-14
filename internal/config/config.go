package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a single monitored cron job.
type Job struct {
	Name            string        `yaml:"name"`
	Schedule        string        `yaml:"schedule"`
	MaxDuration     time.Duration `yaml:"max_duration"`
	GracePeriod     time.Duration `yaml:"grace_period"`
	AlertChannels   []string      `yaml:"alert_channels"`
}

// AlertChannel defines a notification target.
type AlertChannel struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"` // "slack", "email", "webhook"
	Options map[string]string `yaml:"options"`
}

// Config is the top-level configuration structure.
type Config struct {
	ListenAddr    string         `yaml:"listen_addr"`
	Jobs          []Job          `yaml:"jobs"`
	AlertChannels []AlertChannel `yaml:"alert_channels"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	knownChannels := make(map[string]struct{}, len(c.AlertChannels))
	for _, ch := range c.AlertChannels {
		if ch.Name == "" {
			return fmt.Errorf("alert channel missing name")
		}
		knownChannels[ch.Name] = struct{}{}
	}

	for _, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job missing name")
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q missing schedule", job.Name)
		}
		for _, ch := range job.AlertChannels {
			if _, ok := knownChannels[ch]; !ok {
				return fmt.Errorf("job %q references unknown alert channel %q", job.Name, ch)
			}
		}
	}
	return nil
}
