//go:build darwin

package backend

import "errors"

var (
	ErrNotAvailable = errors.New("keyring backend not available")
	ErrNotImplemented = errors.New("not implemented yet")
)

// darwinKeyringBackend is a placeholder for macOS Keychain implementation.
type darwinKeyringBackend struct{}

func (d *darwinKeyringBackend) StoreLUK(key []byte) error {
	return ErrNotImplemented
}

func (d *darwinKeyringBackend) GetLUK() ([]byte, error) {
	return nil, ErrNotImplemented
}

func (d *darwinKeyringBackend) DeleteLUK() error {
	return ErrNotImplemented
}

// darwinClipboardBackend is a placeholder for macOS clipboard implementation.
type darwinClipboardBackend struct{}

func (d *darwinClipboardBackend) Copy(text string) error {
	return ErrNotImplemented
}

func NewKeyringBackend() KeyringBackend {
	return &darwinKeyringBackend{}
}

func NewClipboardBackend() ClipboardBackend {
	return &darwinClipboardBackend{}
}