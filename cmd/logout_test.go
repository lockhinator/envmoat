package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/session"
)

func TestLogoutClearsSession(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Verify session exists before logout (mock keyring has LUK).
	sess := session.NewSession(keyringBackend)
	if !sess.Exists() {
		t.Fatal("expected session to exist before logout")
	}

	_, stderr, err := runCmd(t, []string{"logout"})
	if err != nil {
		t.Fatalf("logout command failed: %v stderr: %s", err, stderr)
	}

	if !strings.Contains(stderr, "Session cleared") {
		t.Errorf("expected 'Session cleared' in stderr, got:\n%s", stderr)
	}

	// Verify session is cleared after logout.
	if sess.Exists() {
		t.Error("expected session to be cleared after logout")
	}
}

func TestLogoutNoSession(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Clear the session first.
	sess := session.NewSession(keyringBackend)
	if err := sess.Clear(); err != nil {
		t.Fatalf("clear session: %v", err)
	}

	stdout, stderr, err := runCmd(t, []string{"logout"})
	if err != nil {
		t.Fatalf("logout command failed: %v stderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "No active session") && !strings.Contains(stderr, "No active session") {
		t.Errorf("expected 'No active session' in output, got stdout=%q stderr=%q", stdout, stderr)
	}
}
