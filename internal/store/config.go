package store

import (
	"crypto/rand"
	"os"

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
	SessionTTLMinutes int `yaml:"session_ttl_minutes,omitempty"`
}

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
	}, nil
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
