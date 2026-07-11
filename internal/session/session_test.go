package session

import (
	"errors"
	"testing"
	"time"

	"github.com/lockinator/envmoat/internal/backend"
)

// Compile-time check: mockKeyringBackend must implement backend.KeyringBackend.
var _ backend.KeyringBackend = (*mockKeyringBackend)(nil)

// mockKeyringBackend implements backend.KeyringBackend for testing.
type mockKeyringBackend struct {
	luk  []byte
	err  error
}

func (m *mockKeyringBackend) StoreLUK(key []byte) error {
	m.luk = key
	return m.err
}

func (m *mockKeyringBackend) GetLUK() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.luk == nil {
		return nil, errors.New("not found")
	}
	result := make([]byte, len(m.luk))
	copy(result, m.luk)
	return result, nil
}

func (m *mockKeyringBackend) DeleteLUK() error {
	m.luk = nil
	return m.err
}

// mockKeyringBackendWithErr is a variant that returns an error on the next GetLUK call.
type mockKeyringBackendWithErr struct {
	luk    []byte
	errOnGet error
}

func (m *mockKeyringBackendWithErr) StoreLUK(key []byte) error {
	m.luk = key
	return nil
}

func (m *mockKeyringBackendWithErr) GetLUK() ([]byte, error) {
	if m.errOnGet != nil {
		return nil, m.errOnGet
	}
	if m.luk == nil {
		return nil, errors.New("not found")
	}
	result := make([]byte, len(m.luk))
	copy(result, m.luk)
	return result, nil
}

func (m *mockKeyringBackendWithErr) DeleteLUK() error {
	m.luk = nil
	return nil
}

func TestSetAndGetLUK(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	testLUK := []byte("test-luk-bytes-1234567890")
	if err := s.SetLUK(testLUK); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	got, err := s.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK failed: %v", err)
	}
	if string(got) != string(testLUK) {
		t.Errorf("GetLUK = %q, want %q", got, testLUK)
	}
}

func TestExistsBeforeAndAfterSet(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	if s.Exists() {
		t.Error("Exists should be false before SetLUK")
	}

	if err := s.SetLUK([]byte("data")); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	if !s.Exists() {
		t.Error("Exists should be true after SetLUK")
	}
}

func TestClear(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	if err := s.SetLUK([]byte("data")); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	if !s.Exists() {
		t.Fatal("Exists should be true before Clear")
	}

	if err := s.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if s.Exists() {
		t.Error("Exists should be false after Clear")
	}
}

func TestTTLExpiration(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSessionWithTTL(mock, 100*time.Millisecond)

	if err := s.SetLUK([]byte("test-luk")); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	// Should succeed immediately after setting.
	got, err := s.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK should succeed immediately after SetLUK: %v", err)
	}
	if string(got) != "test-luk" {
		t.Errorf("GetLUK = %q, want %q", got, "test-luk")
	}

	// Wait for TTL to expire.
	time.Sleep(150 * time.Millisecond)

	// Should now return ErrExpired.
	_, err = s.GetLUK()
	if !errors.Is(err, ErrExpired) {
		t.Errorf("GetLUK after expiry = %v, want ErrExpired", err)
	}

	// Verify the expired entry was cleared from the keyring.
	if mock.luk != nil {
		t.Error("keyring should be cleared after TTL expiration")
	}
}

func TestTTLNotExpired(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSessionWithTTL(mock, 5*time.Second)

	if err := s.SetLUK([]byte("test-luk")); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	// Wait a short time (well within TTL).
	time.Sleep(100 * time.Millisecond)

	// Should still succeed.
	got, err := s.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK should succeed within TTL: %v", err)
	}
	if string(got) != "test-luk" {
		t.Errorf("GetLUK = %q, want %q", got, "test-luk")
	}
}

func TestSetTTL(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	// Default TTL should be 15 minutes.
	if s.ttl != DefaultTTL {
		t.Errorf("default TTL = %v, want %v", s.ttl, DefaultTTL)
	}

	// Update TTL.
	s.SetTTL(30 * time.Second)
	if s.ttl != 30*time.Second {
		t.Errorf("SetTTL failed: ttl = %v, want %v", s.ttl, 30*time.Second)
	}

	// Verify the new TTL is used for expiry.
	if err := s.SetLUK([]byte("test-luk")); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Should still succeed (30s TTL > 100ms sleep).
	got, err := s.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK should succeed within 30s TTL: %v", err)
	}
	if string(got) != "test-luk" {
		t.Errorf("GetLUK = %q, want %q", got, "test-luk")
	}
}

func TestOldFormatTreatedAsValid(t *testing.T) {
	// Simulate an old-format LUK stored directly in the keyring (no JSON wrapper).
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	// Store raw bytes (simulating old format before TTL was added).
	if err := s.keyring.StoreLUK([]byte("old-luk-bytes")); err != nil {
		t.Fatalf("StoreLUK failed: %v", err)
	}

	// GetLUK should treat old format as valid for backward compatibility.
	got, err := s.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK on old format = %v, want nil", err)
	}
	if string(got) != "old-luk-bytes" {
		t.Errorf("GetLUK = %q, want %q", got, "old-luk-bytes")
	}
}

func TestSaltMatchAllowsCachedLUK(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	testLUK := []byte("test-luk-bytes")
	configSalt := []byte("config-salt-12345678901234567890123456")

	// Store LUK with config salt.
	if err := s.SetLUKWithSalt(testLUK, configSalt); err != nil {
		t.Fatalf("SetLUKWithSalt failed: %v", err)
	}

	// GetLUKWithSalt with matching salt should succeed.
	got, err := s.GetLUKWithSalt(configSalt)
	if err != nil {
		t.Fatalf("GetLUKWithSalt with matching salt = %v, want nil", err)
	}
	if string(got) != string(testLUK) {
		t.Errorf("GetLUKWithSalt = %q, want %q", got, testLUK)
	}
}

func TestSaltMismatchInvalidatesCache(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	testLUK := []byte("test-luk-bytes")
	oldSalt := []byte("old-salt-12345678901234567890123456")
	newSalt := []byte("new-salt-12345678901234567890123456")

	// Store LUK with old config salt.
	if err := s.SetLUKWithSalt(testLUK, oldSalt); err != nil {
		t.Fatalf("SetLUKWithSalt failed: %v", err)
	}

	// GetLUKWithSalt with different salt should return ErrConfigChanged.
	_, err := s.GetLUKWithSalt(newSalt)
	if !errors.Is(err, ErrConfigChanged) {
		t.Errorf("GetLUKWithSalt with mismatched salt = %v, want ErrConfigChanged", err)
	}

	// Verify the expired entry was cleared from the keyring.
	if mock.luk != nil {
		t.Error("keyring should be cleared after config salt mismatch")
	}
}

func TestFirstRunNoStoredSalt(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	// No LUK stored — GetLUKWithSalt should return the "not found" error.
	configSalt := []byte("config-salt-12345678901234567890123456")
	_, err := s.GetLUKWithSalt(configSalt)

	// Should get the "not found" error (backend.ErrNotAvailable or similar).
	if err == nil {
		t.Fatal("GetLUKWithSalt with no stored LUK should return error")
	}
	if errors.Is(err, ErrConfigChanged) {
		t.Error("should not return ErrConfigChanged when no LUK is stored")
	}

	// Now store a LUK and verify it works.
	testLUK := []byte("test-luk-bytes")
	if err := s.SetLUKWithSalt(testLUK, configSalt); err != nil {
		t.Fatalf("SetLUKWithSalt failed: %v", err)
	}

	got, err := s.GetLUKWithSalt(configSalt)
	if err != nil {
		t.Fatalf("GetLUKWithSalt after SetLUKWithSalt = %v, want nil", err)
	}
	if string(got) != string(testLUK) {
		t.Errorf("GetLUKWithSalt = %q, want %q", got, testLUK)
	}
}

func TestInvalidate(t *testing.T) {
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	testLUK := []byte("test-luk-bytes")
	configSalt := []byte("config-salt-12345678901234567890123456")

	if err := s.SetLUKWithSalt(testLUK, configSalt); err != nil {
		t.Fatalf("SetLUKWithSalt failed: %v", err)
	}

	if !s.Exists() {
		t.Fatal("Exists should be true before Invalidate")
	}

	if err := s.Invalidate(); err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	if s.Exists() {
		t.Error("Exists should be false after Invalidate")
	}
}

func TestNoStoredSaltTreatedAsMatch(t *testing.T) {
	// When the stored session has no config salt (e.g., from old SetLUK call),
	// it should not trigger ErrConfigChanged — treat as match.
	mock := &mockKeyringBackend{}
	s := NewSession(mock)

	testLUK := []byte("test-luk-bytes")
	configSalt := []byte("config-salt-12345678901234567890123456")

	// Use SetLUK (no salt) to simulate old-format session.
	if err := s.SetLUK(testLUK); err != nil {
		t.Fatalf("SetLUK failed: %v", err)
	}

	// GetLUKWithSalt with any salt should succeed (no stored salt = no mismatch).
	got, err := s.GetLUKWithSalt(configSalt)
	if err != nil {
		t.Fatalf("GetLUKWithSalt with no stored salt = %v, want nil", err)
	}
	if string(got) != string(testLUK) {
		t.Errorf("GetLUKWithSalt = %q, want %q", got, testLUK)
	}
}
