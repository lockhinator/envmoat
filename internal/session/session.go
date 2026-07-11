// Package session provides a thin wrapper around the platform keyring backend
// for caching the LUK (Login Unlock Key) across CLI invocations.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lockinator/envmoat/internal/backend"
)

// DefaultTTL is the default time-to-live for a cached LUK.
const DefaultTTL = 15 * time.Minute

// ErrExpired is returned when the cached LUK has expired.
var ErrExpired = errors.New("session cache expired")

// sessionValue is the wire format stored in the keyring:
// {"expiry": unix_nano, "luk": base64(luk_bytes), "config_salt": base64(salt_bytes)}
type sessionValue struct {
	Expiry     int64  `json:"expiry"`
	LUK        []byte `json:"luk"`
	ConfigSalt []byte `json:"config_salt,omitempty"`
}

// ErrConfigChanged is returned when the config salt has changed since the session was created.
var ErrConfigChanged = errors.New("config changed, session invalidated")

// Session wraps the keyring backend to store and retrieve the cached LUK.
type Session struct {
	keyring backend.KeyringBackend
	ttl     time.Duration
}

// NewSession creates a new Session using the given keyring backend with the default TTL.
func NewSession(keyring backend.KeyringBackend) *Session {
	return &Session{
		keyring: keyring,
		ttl:     DefaultTTL,
	}
}

// NewSessionWithTTL creates a new Session with an explicit TTL.
func NewSessionWithTTL(keyring backend.KeyringBackend, ttl time.Duration) *Session {
	return &Session{
		keyring: keyring,
		ttl:     ttl,
	}
}

// SetTTL updates the TTL at runtime.
func (s *Session) SetTTL(ttl time.Duration) {
	s.ttl = ttl
}

// GetLUK retrieves the cached LUK from the platform keyring.
// Returns ErrExpired if the cached entry has exceeded its TTL.
func (s *Session) GetLUK() ([]byte, error) {
	data, err := s.keyring.GetLUK()
	if err != nil {
		return nil, err
	}

	var sv sessionValue
	if err := json.Unmarshal(data, &sv); err != nil {
		// Old format (raw LUK bytes stored directly in keyring).
		// Treat as valid for backward compatibility — no TTL was enforced.
		return data, nil
	}

	if time.Now().UnixNano() > sv.Expiry {
		_ = s.keyring.DeleteLUK()
		return nil, ErrExpired
	}

	// Sliding window: reset TTL on each successful access.
	if err := s.SetLUKWithSalt(sv.LUK, sv.ConfigSalt); err != nil {
		return nil, fmt.Errorf("reset session TTL: %w", err)
	}

	return sv.LUK, nil
}

// SetLUK stores the LUK in the platform keyring with an expiry timestamp.
// Deprecated: use SetLUKWithSalt for new code that includes config salt tracking.
func (s *Session) SetLUK(luk []byte) error {
	return s.SetLUKWithSalt(luk, nil)
}

// SetLUKWithSalt stores the LUK in the platform keyring with an expiry timestamp
// and the current config salt for change detection.
func (s *Session) SetLUKWithSalt(luk []byte, configSalt []byte) error {
	sv := sessionValue{
		Expiry:     time.Now().Add(s.ttl).UnixNano(),
		LUK:        luk,
		ConfigSalt: configSalt,
	}
	data, err := json.Marshal(sv)
	if err != nil {
		return err
	}
	return s.keyring.StoreLUK(data)
}

// Exists checks whether a cached LUK is present and not expired.
func (s *Session) Exists() bool {
	_, err := s.GetLUK()
	return err == nil
}

// Clear removes the cached LUK from the keyring.
func (s *Session) Clear() error {
	return s.keyring.DeleteLUK()
}

// Invalidate clears the cached LUK and its expiry timestamp from the keyring,
// effectively ending the session. This is called when the config has changed
// (e.g., password rotation) so the user must re-authenticate.
func (s *Session) Invalidate() error {
	return s.keyring.DeleteLUK()
}

// GetLUKWithSalt retrieves the cached LUK from the platform keyring,
// comparing the stored config salt with the provided currentConfigSalt.
// Returns ErrExpired if the cached entry has exceeded its TTL.
// Returns ErrConfigChanged if the config salt has changed since the session was created.
func (s *Session) GetLUKWithSalt(currentConfigSalt []byte) ([]byte, error) {
	data, err := s.keyring.GetLUK()
	if err != nil {
		return nil, err
	}

	var sv sessionValue
	if err := json.Unmarshal(data, &sv); err != nil {
		// Old format (raw LUK bytes stored directly in keyring).
		// Treat as valid for backward compatibility — no TTL was enforced.
		return data, nil
	}

	if time.Now().UnixNano() > sv.Expiry {
		_ = s.keyring.DeleteLUK()
		return nil, ErrExpired
	}

	// Compare config salt to detect password rotation or config changes.
	if len(sv.ConfigSalt) > 0 && len(currentConfigSalt) > 0 {
		if !equalBytes(sv.ConfigSalt, currentConfigSalt) {
			_ = s.keyring.DeleteLUK()
			return nil, ErrConfigChanged
		}
	}

	return sv.LUK, nil
}

// equalBytes compares two byte slices for equality in constant time.
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GetRemainingTTL returns the remaining time until the cached LUK expires.
// It reads the raw keyring data and parses the expiry timestamp without
// resetting the TTL (no sliding window). Returns 0 if no cache exists or
// if the cache has already expired.
func (s *Session) GetRemainingTTL() time.Duration {
	data, err := s.keyring.GetLUK()
	if err != nil {
		return 0
	}

	var sv sessionValue
	if err := json.Unmarshal(data, &sv); err != nil {
		// Old format (raw LUK bytes) — treat as always valid.
		return s.ttl
	}

	remaining := time.Duration(sv.Expiry-time.Now().UnixNano())
	if remaining <= 0 {
		return 0
	}
	return remaining
}
