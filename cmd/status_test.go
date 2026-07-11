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

	stdout, _, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	if !strings.Contains(stdout, "Bundle:") {
		t.Errorf("expected 'Bundle:' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Profile:") {
		t.Errorf("expected 'Profile:' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Session:") {
		t.Errorf("expected 'Session:' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Keychain:") {
		t.Errorf("expected 'Keychain:' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "ENVMOAT_DEBUG") {
		t.Errorf("expected debug hint in output, got: %q", stdout)
	}
}

func TestStatusNotInTrackedDir(t *testing.T) {
	_, _, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	untrackedDir := t.TempDir()
	os.Chdir(untrackedDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(stdout, "Bundle: none") {
		t.Errorf("expected 'Bundle: none' in output, got: %q", stdout)
	}
}

func TestStatusSessionActive(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	// The mock keyring has a LUK set, so session should be active.
	if !strings.Contains(stdout, "Session:") {
		t.Fatalf("expected 'Session:' in output, got: %q", stdout)
	}
}

func TestStatusKeychainState(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"status"})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	// Mock keyring has LUK set, so both protected and cache should be true.
	if !strings.Contains(stdout, "Keychain: protected=true cache=true") {
		t.Errorf("expected 'Keychain: protected=true cache=true', got: %q", stdout)
	}
}
