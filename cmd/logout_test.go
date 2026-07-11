package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/backend"
	"github.com/lockinator/envmoat/internal/session"
)

func TestLogoutClearsSession(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Verify session exists before logout.
	sess := newSessionForTest(keyringBackend)
	if !sess.Exists() {
		t.Fatal("expected session to exist before logout")
	}

	stdout, _, err := runCmd(t, []string{"logout"})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if !strings.Contains(stdout, "Session cleared") {
		t.Errorf("expected 'Session cleared' in output, got: %q", stdout)
	}

	// Verify session is gone after logout.
	sess = newSessionForTest(keyringBackend)
	if sess.Exists() {
		t.Fatal("expected session to be cleared after logout")
	}
}

func TestLogoutNoSession(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Clear the session first so there's no active session.
	keyringBackend = &mockKeyringBackend{luk: nil}

	stdout, _, err := runCmd(t, []string{"logout"})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	if !strings.Contains(stdout, "No active session") {
		t.Errorf("expected 'No active session' in output, got: %q", stdout)
	}

	// Restore the original keyring backend for other tests.
	keyringBackend = backend.NewKeyringBackend()
}

func newSessionForTest(kb backend.KeyringBackend) *session.Session {
	return session.NewSession(kb)
}
