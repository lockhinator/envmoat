//go:build linux

package backend

import "errors"

var (
	ErrNotAvailable = errors.New("keyring backend not available")
	ErrNotImplemented = errors.New("not implemented yet")
)

// linuxKeyringBackend is a placeholder for Linux Keyring implementation.
type linuxKeyringBackend struct{}

func (l *linuxKeyringBackend) StoreLUK(key []byte) error {
	return ErrNotImplemented
}

func (l *linuxKeyringBackend) GetLUK() ([]byte, error) {
	return nil, ErrNotImplemented
}

func (l *linuxKeyringBackend) DeleteLUK() error {
	return ErrNotImplemented
}

// linuxClipboardBackend is a placeholder for Linux clipboard implementation.
type linuxClipboardBackend struct{}

func (l *linuxClipboardBackend) Copy(text string) error {
	return ErrNotImplemented
}

func NewKeyringBackend() KeyringBackend {
	return &linuxKeyringBackend{}
}

func NewClipboardBackend() ClipboardBackend {
	return &linuxClipboardBackend{}
}