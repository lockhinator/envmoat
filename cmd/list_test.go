package cmd

import (
	"os"
	"sort"
	"strings"
	"testing"
)

func TestListShowsKeys(t *testing.T) {
	secrets := map[string]string{"API_KEY": "secret1", "DB_PASS": "secret2", "TOKEN": "secret3"}
	_, markerDir, cleanup := testEnv(t, secrets)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, stderr, err := runCmd(t, []string{"list"})
	if err != nil {
		t.Fatalf("list failed: %v stderr: %s", err, stderr)
	}
	for key := range secrets {
		if !strings.Contains(stdout, key) {
			t.Errorf("expected key %q in output, got: %q", key, stdout)
		}
	}
	// Verify values are NOT shown.
	for _, val := range secrets {
		if strings.Contains(stdout, val) {
			t.Errorf("value %q should not appear in output, got: %q", val, stdout)
		}
	}
	// Verify bundle name shown on stderr.
	if !strings.Contains(stderr, "bundle:") {
		t.Errorf("expected 'bundle:' in stderr, got: %q", stderr)
	}
}

func TestListSortedKeys(t *testing.T) {
	secrets := map[string]string{"ZEBRA": "1", "ALPHA": "2", "MIDDLE": "3"}
	_, markerDir, cleanup := testEnv(t, secrets)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"list"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	lines := strings.Fields(stdout)
	expected := []string{"ALPHA", "MIDDLE", "ZEBRA"}
	if !sort.StringsAreSorted(lines) {
		t.Errorf("expected sorted keys, got: %v", lines)
	}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(lines))
	}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: got %q, want %q", i, lines[i], exp)
		}
	}
}

func TestListEmptyBundle(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, stderr, err := runCmd(t, []string{"list"})
	if err != nil {
		t.Fatalf("list failed: %v stderr: %s", err, stderr)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected no keys in output, got: %q", stdout)
	}
	// Bundle name should still be shown.
	if !strings.Contains(stderr, "bundle:") {
		t.Errorf("expected 'bundle:' in stderr, got: %q", stderr)
	}
}

func TestListNotInTrackedDir(t *testing.T) {
	_, _, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	untrackedDir := t.TempDir()
	os.Chdir(untrackedDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"list"})
	if err == nil {
		t.Fatal("expected error when not in tracked directory")
	}
	if !strings.Contains(stderr, "tracked directory") {
		t.Errorf("expected 'tracked directory' in stderr, got: %q", stderr)
	}
}

func TestListSetRoundtrip(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Set a key.
	_, _, err := runCmd(t, []string{"set", "NEW_KEY", "new-value"})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}

	// List keys.
	stdout, _, err := runCmd(t, []string{"list"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(stdout, "NEW_KEY") {
		t.Errorf("expected 'NEW_KEY' in list output, got: %q", stdout)
	}
	// Verify value is NOT shown.
	if strings.Contains(stdout, "new-value") {
		t.Errorf("value should not appear in list output")
	}
}
