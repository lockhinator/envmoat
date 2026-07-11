package backend

import (
	"errors"
	"testing"
)

// Compile-time interface conformance checks.
var (
	_ KeyringBackend   = (*mockKeyringBackend)(nil)
	_ ClipboardBackend = (*mockClipboardBackend)(nil)
)

// --- Mock implementations ---

type mockKeyringBackend struct {
	store   map[string]string
	getErr  error
	setErr  error
	delErr  error
}

func (m *mockKeyringBackend) StoreLUK(key []byte) error {
	if m.setErr != nil {
		return m.setErr
	}
	if m.store == nil {
		m.store = make(map[string]string)
	}
	m.store[kcKey] = string(key)
	return nil
}

func (m *mockKeyringBackend) GetLUK() ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	val, ok := m.store[kcKey]
	if !ok {
		return nil, ErrNotAvailable
	}
	return []byte(val), nil
}

func (m *mockKeyringBackend) DeleteLUK() error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.store, kcKey)
	return nil
}

type mockClipboardBackend struct {
	lastCopied string
	copyErr    error
}

func (m *mockClipboardBackend) Copy(text string) error {
	if m.copyErr != nil {
		return m.copyErr
	}
	m.lastCopied = text
	return nil
}

// --- Tests ---

func TestMockKeyringImplementsKeyringBackend(t *testing.T) {
	// Compile-time check ensures mockKeyringBackend satisfies KeyringBackend.
	var _ KeyringBackend = &mockKeyringBackend{}
}

func TestMockClipboardImplementsClipboardBackend(t *testing.T) {
	// Compile-time check ensures mockClipboardBackend satisfies ClipboardBackend.
	var _ ClipboardBackend = &mockClipboardBackend{}
}

func TestMockKeyringStoreAndGetLUK(t *testing.T) {
	mk := &mockKeyringBackend{}
	key := []byte("test-luk-key-12345")

	if err := mk.StoreLUK(key); err != nil {
		t.Fatalf("StoreLUK failed: %v", err)
	}

	retrieved, err := mk.GetLUK()
	if err != nil {
		t.Fatalf("GetLUK failed: %v", err)
	}

	if string(retrieved) != string(key) {
		t.Fatalf("GetLUK returned wrong key: got %q, want %q", retrieved, key)
	}
}

func TestMockKeyringGetLUKNotAvailable(t *testing.T) {
	mk := &mockKeyringBackend{}

	_, err := mk.GetLUK()
	if !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("expected ErrNotAvailable, got: %v", err)
	}
}

func TestMockKeyringDeleteLUK(t *testing.T) {
	mk := &mockKeyringBackend{}
	key := []byte("test-luk-key-12345")

	if err := mk.StoreLUK(key); err != nil {
		t.Fatalf("StoreLUK failed: %v", err)
	}

	if err := mk.DeleteLUK(); err != nil {
		t.Fatalf("DeleteLUK failed: %v", err)
	}

	_, err := mk.GetLUK()
	if !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("expected ErrNotAvailable after delete, got: %v", err)
	}
}

func TestMockKeyringStoreError(t *testing.T) {
	expectedErr := errors.New("keyring unavailable")
	mk := &mockKeyringBackend{setErr: expectedErr}

	err := mk.StoreLUK([]byte("key"))
	if err == nil {
		t.Fatal("expected error from StoreLUK")
	}
	if err.Error() != expectedErr.Error() {
		t.Fatalf("StoreLUK error mismatch: got %v, want %v", err, expectedErr)
	}
}

func TestMockKeyringGetError(t *testing.T) {
	expectedErr := errors.New("keyring access denied")
	mk := &mockKeyringBackend{getErr: expectedErr}

	_, err := mk.GetLUK()
	if err == nil {
		t.Fatal("expected error from GetLUK")
	}
	if err.Error() != expectedErr.Error() {
		t.Fatalf("GetLUK error mismatch: got %v, want %v", err, expectedErr)
	}
}

func TestMockKeyringDeleteError(t *testing.T) {
	expectedErr := errors.New("keyring locked")
	mk := &mockKeyringBackend{delErr: expectedErr}

	err := mk.DeleteLUK()
	if err == nil {
		t.Fatal("expected error from DeleteLUK")
	}
	if err.Error() != expectedErr.Error() {
		t.Fatalf("DeleteLUK error mismatch: got %v, want %v", err, expectedErr)
	}
}

func TestMockClipboardCopy(t *testing.T) {
	mc := &mockClipboardBackend{}
	text := "secret-value-123"

	if err := mc.Copy(text); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if mc.lastCopied != text {
		t.Fatalf("clipboard text mismatch: got %q, want %q", mc.lastCopied, text)
	}
}

func TestMockClipboardCopyError(t *testing.T) {
	expectedErr := errors.New("clipboard unavailable")
	mc := &mockClipboardBackend{copyErr: expectedErr}

	err := mc.Copy("text")
	if err == nil {
		t.Fatal("expected error from Copy")
	}
	if err.Error() != expectedErr.Error() {
		t.Fatalf("Copy error mismatch: got %v, want %v", err, expectedErr)
	}
}

func TestNewKeyringBackend(t *testing.T) {
	kb := NewKeyringBackend()
	if kb == nil {
		t.Fatal("NewKeyringBackend returned nil")
	}
}

func TestNewClipboardBackend(t *testing.T) {
	cb := NewClipboardBackend()
	if cb == nil {
		t.Fatal("NewClipboardBackend returned nil")
	}
}

func TestErrNotAvailable(t *testing.T) {
	if ErrNotAvailable == nil {
		t.Fatal("ErrNotAvailable is nil")
	}
	if ErrNotAvailable.Error() == "" {
		t.Fatal("ErrNotAvailable has empty error message")
	}
}

func TestKeychainConstants(t *testing.T) {
	if kcService == "" {
		t.Fatal("kcService is empty")
	}
	if kcKey == "" {
		t.Fatal("kcKey is empty")
	}
}
