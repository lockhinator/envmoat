//go:build !linux && !darwin

package backend

import "fmt"

// stubClipboardBackend is a no-op clipboard backend for unsupported platforms.
type stubClipboardBackend struct{}

func (s *stubClipboardBackend) Copy(text string) error {
	return fmt.Errorf("clipboard not supported on this platform")
}

func NewClipboardBackend() ClipboardBackend {
	return &stubClipboardBackend{}
}
