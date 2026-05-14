package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultListenAddr = ":8080"

// JobConfig defines the configuration for a single monitored cron job.
type JobConfig struct {
	Name        string        `yaml:"name"`
	Schedule    string        `yaml:"schedule"`
	MaxDuration time.Duration `yaml:"max_duration"`
}

// AlertConfig holds settings for a notification channel.
type AlertConfig struct {
	Channel string            `yaml:"channel"`
	Options map[string]string `yaml:"options"`
}

// Config is the top-level application configuration.
type Config struct {
	ListenAddr string        `yaml:"listen_addr"`
	Jobs       []JobConfig   `yaml:"jobs"`
	Alerts     []AlertConfig `yaml:"alerts"`
}

// Load reads and parses a YAML configuration file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = defaultListenAddr
	}

	return &cfg, nil
}

var validChannels = map[string]bool{
	"slack":   true,
	"email":   true,
	"webhook": true,
}

func validate(cfg *Config) error {
	for _, a := range cfg.Alerts {
		if !validChannels[a.Channel] {
			return fmt.Errorf("unknown alert channel %q", a.Channel)
		}
	}
	for _, j := range cfg.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job is missing a name")
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q is missing a schedule", j.Name)
		}
	}
	return nil
}
