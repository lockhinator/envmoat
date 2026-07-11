//go:build linux

package backend

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

// linuxClipboardBackend copies text to the Linux clipboard via xclip or wl-clipboard.
type linuxClipboardBackend struct {
	method string // "wl-clipboard" or "xclip"
}

var errNoClipboardTool = errors.New("neither xclip nor wl-clipboard is available")

func (l *linuxClipboardBackend) Copy(text string) error {
	switch l.method {
	case "wl-clipboard":
		return copyViaWlClipboard(text)
	default:
		return copyViaXclip(text)
	}
}

// copyViaWlClipboard copies text to the Wayland clipboard using wl-clipboard.
func copyViaWlClipboard(text string) error {
	cmd := exec.Command("wl-clipboard", "-i")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

// copyViaXclip copies text to the X11 clipboard using xclip.
func copyViaXclip(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

// NewClipboardBackend returns a ClipboardBackend for Linux, auto-detecting the display server.
func NewClipboardBackend() ClipboardBackend {
	if isWayland() {
		return &linuxClipboardBackend{method: "wl-clipboard"}
	}
	return &linuxClipboardBackend{method: "xclip"}
}

// isWayland returns true if the current session appears to be Wayland.
func isWayland() bool {
	return len(getenv("WAYLAND_DISPLAY")) > 0
}

// getenv is a testable wrapper around os.Getenv.
func getenv(key string) string {
	return os.Getenv(key)
}
