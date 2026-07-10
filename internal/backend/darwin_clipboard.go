//go:build darwin

package backend

import (
	"bytes"
	"os/exec"
)

// darwinClipboardBackend copies text to the macOS clipboard via pbcopy.
type darwinClipboardBackend struct{}

// Copy writes the given text to the macOS clipboard using pbcopy.
func (d *darwinClipboardBackend) Copy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

func NewClipboardBackend() ClipboardBackend {
	return &darwinClipboardBackend{}
}
