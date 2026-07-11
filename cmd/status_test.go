package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestStatusShowsBundleName(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"API_KEY": "secret123"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, stderr, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status command failed: %v stderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "envmoat status") {
		t.Errorf("expected 'envmoat status' in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Bundle:") {
		t.Errorf("expected 'Bundle:' in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Profile:") {
		t.Errorf("expected 'Profile:' in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Session:") {
		t.Errorf("expected 'Session:' in stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Keychain:") {
		t.Errorf("expected 'Keychain:' in stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "Debug:") {
		t.Errorf("expected 'Debug:' in stderr, got:\n%s", stderr)
	}
}

func TestStatusNotInTrackedDir(t *testing.T) {
	_, _, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	// Use a temp dir without a marker.
	untrackedDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(untrackedDir)
	defer os.Chdir(origDir)

	stdout, stderr, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status command failed: %v stderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Bundle: none") {
		t.Errorf("expected 'Bundle: none' in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Profile: none") {
		t.Errorf("expected 'Profile: none' in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "Session:") {
		t.Errorf("expected 'Session:' in stderr, got:\n%s", stderr)
	}
}

func TestStatusKeychainState(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	// The mock keyring has a LUK set, so cache should be "yes".
	if !strings.Contains(stderr, "cache=yes") {
		t.Errorf("expected 'cache=yes' in stderr (mock keyring has LUK), got:\n%s", stderr)
	}
}
