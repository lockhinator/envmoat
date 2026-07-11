package store

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// GlobalSaltSize is the size of the global salt in bytes.
	GlobalSaltSize = 32

	// DefaultSessionTTLMinutes is the default session TTL.
	DefaultSessionTTLMinutes = 15
)

// Config holds global envmoat settings stored in config.yaml.
type Config struct {
	// Version is the config schema version.
	Version int `yaml:"version"`

	// GlobalSalt is the salt used for LUK derivation via scrypt.
	GlobalSalt []byte `yaml:"global_salt"`

	// SessionTTLMinutes is the duration (in minutes) a session stays unlocked.
	// Deprecated: use SessionTTL instead.
	SessionTTLMinutes int `yaml:"session_ttl_minutes,omitempty"`

	// SessionTTL is the session TTL as a Go duration string (e.g. "15m", "30m", "1h").
	SessionTTL string `yaml:"session_ttl,omitempty"`
}

// defaultSessionTTL is the default session TTL duration.
const defaultSessionTTL = "15m"

// DefaultConfig returns a Config with default values and a fresh random global salt.
func DefaultConfig() (*Config, error) {
	salt := make([]byte, GlobalSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return &Config{
		Version:           1,
		GlobalSalt:        salt,
		SessionTTLMinutes: DefaultSessionTTLMinutes,
		SessionTTL:        defaultSessionTTL,
	}, nil
}

// ParseSessionTTL parses the SessionTTL string field into a time.Duration.
// If SessionTTL is empty, it falls back to SessionTTLMinutes (in minutes).
// Returns an error if the duration string is invalid.
func (c *Config) ParseSessionTTL() (time.Duration, error) {
	if c.SessionTTL != "" {
		d, err := time.ParseDuration(c.SessionTTL)
		if err != nil {
			return 0, fmt.Errorf("parse session_ttl %q: %w", c.SessionTTL, err)
		}
		return d, nil
	}
	// Fallback to SessionTTLMinutes for backward compatibility.
	if c.SessionTTLMinutes > 0 {
		return time.Duration(c.SessionTTLMinutes) * time.Minute, nil
	}
	// Default if neither is set.
	return time.ParseDuration(defaultSessionTTL)
}

// WriteConfig writes the config to a YAML file with 0600 permissions.
func WriteConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return atomicWrite(path, data, 0600)
}

// ReadConfig reads and unmarshals a config YAML file.
func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
