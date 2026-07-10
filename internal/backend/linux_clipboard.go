//go:build linux

package backend

import (
	"bytes"
	"os/exec"
)

// linuxClipboardBackend copies text to the Linux clipboard via wl-copy or xclip.
type linuxClipboardBackend struct{}

// Copy writes the given text to the Linux clipboard.
// Tries wl-copy first (Wayland), falls back to xclip (X11).
func (l *linuxClipboardBackend) Copy(text string) error {
	if err := wlCopy(text); err == nil {
		return nil
	}
	return xclipCopy(text)
}

func wlCopy(text string) error {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

func xclipCopy(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

func NewClipboardBackend() ClipboardBackend {
	return &linuxClipboardBackend{}
}
