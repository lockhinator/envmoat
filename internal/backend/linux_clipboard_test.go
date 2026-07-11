//go:build linux

package backend

import (
	"os"
	"testing"
)

func TestIsWayland(t *testing.T) {
	// Save and restore WAYLAND_DISPLAY
	original := os.Getenv("WAYLAND_DISPLAY")
	defer os.Setenv("WAYLAND_DISPLAY", original)

	// Wayland session
	os.Setenv("WAYLAND_DISPLAY", "wayland-0")
	if !isWayland() {
		t.Error("isWayland() should return true when WAYLAND_DISPLAY is set")
	}

	// X11 session
	os.Setenv("WAYLAND_DISPLAY", "")
	if isWayland() {
		t.Error("isWayland() should return false when WAYLAND_DISPLAY is empty")
	}
}

func TestNewClipboardBackendDetectsWayland(t *testing.T) {
	original := os.Getenv("WAYLAND_DISPLAY")
	defer os.Setenv("WAYLAND_DISPLAY", original)

	os.Setenv("WAYLAND_DISPLAY", "wayland-0")
	be := NewClipboardBackend()
	l, ok := be.(*linuxClipboardBackend)
	if !ok {
		t.Fatal("expected *linuxClipboardBackend")
	}
	if l.method != "wl-clipboard" {
		t.Errorf("expected wl-clipboard method, got %s", l.method)
	}
}

func TestNewClipboardBackendDetectsX11(t *testing.T) {
	original := os.Getenv("WAYLAND_DISPLAY")
	defer os.Setenv("WAYLAND_DISPLAY", original)

	os.Setenv("WAYLAND_DISPLAY", "")
	be := NewClipboardBackend()
	l, ok := be.(*linuxClipboardBackend)
	if !ok {
		t.Fatal("expected *linuxClipboardBackend")
	}
	if l.method != "xclip" {
		t.Errorf("expected xclip method, got %s", l.method)
	}
}

func TestCopyReturnsErrorWhenToolMissing(t *testing.T) {
	// On CI, neither xclip nor wl-clipboard is typically available.
	// We can't easily simulate this without mocking exec.Command,
	// so we just verify the error type when calling Copy with a missing tool.
	be := &linuxClipboardBackend{method: "xclip"}
	err := be.Copy("test")
	if err == nil {
		t.Skip("xclip is available on this system; skipping missing-tool test")
	}
	// If xclip is not found, exec.Command.Run returns *exec.ExitError or similar.
	// We don't assert the exact error since it varies by platform.
}
