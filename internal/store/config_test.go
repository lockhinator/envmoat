package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfigHasSessionTTL(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error: %v", err)
	}
	if cfg.SessionTTL != "15m" {
		t.Errorf("SessionTTL = %q, want %q", cfg.SessionTTL, "15m")
	}
	if cfg.SessionTTLMinutes != DefaultSessionTTLMinutes {
		t.Errorf("SessionTTLMinutes = %d, want %d", cfg.SessionTTLMinutes, DefaultSessionTTLMinutes)
	}
}

func TestParseSessionTTL(t *testing.T) {
	tests := []struct {
		name     string
		sessionTTL string
		want    time.Duration
		wantErr bool
	}{
		{"15m", "15m", 15 * time.Minute, false},
		{"30m", "30m", 30 * time.Minute, false},
		{"1h", "1h", time.Hour, false},
		{"2h", "2h", 2 * time.Hour, false},
		{"45m", "45m", 45 * time.Minute, false},
		{"invalid", "abc", 0, true},
		{"empty string with no minutes defaults to 15m", "", 15 * time.Minute, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{SessionTTL: tt.sessionTTL}
			if tt.sessionTTL == "" {
				cfg.SessionTTLMinutes = DefaultSessionTTLMinutes
			}
			dur, err := cfg.ParseSessionTTL()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseSessionTTL() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSessionTTL() error: %v", err)
			}
			if dur != tt.want {
				t.Errorf("ParseSessionTTL() = %v, want %v", dur, tt.want)
			}
		})
	}
}

func TestParseSessionTTLFallbackToMinutes(t *testing.T) {
	cfg := &Config{
		SessionTTL:        "",
		SessionTTLMinutes: 30,
	}
	dur, err := cfg.ParseSessionTTL()
	if err != nil {
		t.Fatalf("ParseSessionTTL() error: %v", err)
	}
	if dur != 30*time.Minute {
		t.Errorf("ParseSessionTTL() fallback = %v, want %v", dur, 30*time.Minute)
	}
}

func TestParseSessionTTLDefaultWhenBothEmpty(t *testing.T) {
	cfg := &Config{}
	dur, err := cfg.ParseSessionTTL()
	if err != nil {
		t.Fatalf("ParseSessionTTL() error: %v", err)
	}
	if dur != 15*time.Minute {
		t.Errorf("ParseSessionTTL() default = %v, want %v", dur, 15*time.Minute)
	}
}

func TestWriteReadConfigRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error: %v", err)
	}
	cfg.SessionTTL = "30m"

	if err := WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig() error: %v", err)
	}

	readCfg, err := ReadConfig(configPath)
	if err != nil {
		t.Fatalf("ReadConfig() error: %v", err)
	}

	if readCfg.SessionTTL != "30m" {
		t.Errorf("SessionTTL roundtrip = %q, want %q", readCfg.SessionTTL, "30m")
	}
	if readCfg.Version != cfg.Version {
		t.Errorf("Version roundtrip = %d, want %d", readCfg.Version, cfg.Version)
	}
	if len(readCfg.GlobalSalt) != GlobalSaltSize {
		t.Errorf("GlobalSalt length = %d, want %d", len(readCfg.GlobalSalt), GlobalSaltSize)
	}
}

func TestWriteReadConfigPreservesMinutes(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Version:           1,
		SessionTTLMinutes: 45,
		SessionTTL:        "",
	}
	// Write a minimal config with only salt and minutes
	salt := make([]byte, GlobalSaltSize)
	for i := range salt {
		salt[i] = byte(i)
	}
	cfg.GlobalSalt = salt

	if err := WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig() error: %v", err)
	}

	readCfg, err := ReadConfig(configPath)
	if err != nil {
		t.Fatalf("ReadConfig() error: %v", err)
	}

	if readCfg.SessionTTLMinutes != 45 {
		t.Errorf("SessionTTLMinutes roundtrip = %d, want %d", readCfg.SessionTTLMinutes, 45)
	}
}

func TestWriteConfigPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error: %v", err)
	}

	if err := WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig() error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Errorf("config file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestReadConfigNonexistent(t *testing.T) {
	_, err := ReadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("ReadConfig() expected error for nonexistent file, got nil")
	}
}
