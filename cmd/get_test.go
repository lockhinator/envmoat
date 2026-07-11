package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGetExistingKey(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"API_KEY": "secret123"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"get", "API_KEY"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "secret123") {
		t.Errorf("expected 'secret123' in output, got: %q", stdout)
	}
}

func TestGetMissingKey(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"OTHER_KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"get", "MISSING_KEY"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' in stderr, got: %q", stderr)
	}
}

func TestGetNoKey(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"get"})
	if err == nil {
		t.Fatal("expected error when no key provided")
	}
	if !strings.Contains(stderr, "provide a KEY argument") {
		t.Errorf("expected 'provide a KEY argument' in stderr, got: %q", stderr)
	}
}

func TestGetNotInTrackedDir(t *testing.T) {
	_, _, cleanup := testEnv(t, map[string]string{"KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	untrackedDir := t.TempDir()
	os.Chdir(untrackedDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"get", "KEY"})
	if err == nil {
		t.Fatal("expected error when not in tracked directory")
	}
	if !strings.Contains(stderr, "tracked directory") {
		t.Errorf("expected 'tracked directory' in stderr, got: %q", stderr)
	}
}

func TestGetSpecialChars(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"SPECIAL": "val=ue with spaces"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"get", "SPECIAL"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "val=ue with spaces") {
		t.Errorf("expected 'val=ue with spaces' in output, got: %q", stdout)
	}
}

func TestGetEmptyBundle(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"get", "NONEXISTENT"})
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' in stderr, got: %q", stderr)
	}
}

func TestGetJSON(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"API_KEY": "secret123"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"get", "--json", "API_KEY"})
	if err != nil {
		t.Fatalf("get --json failed: %v", err)
	}

	var result struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON output: %v (got: %q)", err, stdout)
	}
	if result.Key != "API_KEY" {
		t.Errorf("JSON key = %q, want %q", result.Key, "API_KEY")
	}
	if result.Value != "secret123" {
		t.Errorf("JSON value = %q, want %q", result.Value, "secret123")
	}
}

func TestGetJSONSpecialChars(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"SPECIAL": `val"ue\twith$pecial`})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, _, err := runCmd(t, []string{"get", "--json", "SPECIAL"})
	if err != nil {
		t.Fatalf("get --json failed: %v", err)
	}

	var result struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON output: %v (got: %q)", err, stdout)
	}
	if result.Value != `val"ue\twith$pecial` {
		t.Errorf("JSON value = %q, want %q", result.Value, `val"ue\twith$pecial`)
	}
}

func TestGetDefaultVsJSON(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"KEY": "value123"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Default mode: raw value.
	stdoutDefault, _, err := runCmd(t, []string{"get", "KEY"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdoutDefault, "value123") || strings.Contains(stdoutDefault, "{") {
		t.Errorf("default mode should print raw value, got: %q", stdoutDefault)
	}

	// JSON mode: structured output.
	stdoutJSON, _, err := runCmd(t, []string{"get", "--json", "KEY"})
	if err != nil {
		t.Fatalf("get --json failed: %v", err)
	}
	var result struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(stdoutJSON), &result); err != nil {
		t.Fatalf("invalid JSON output: %v (got: %q)", err, stdoutJSON)
	}
	if result.Value != "value123" {
		t.Errorf("JSON value = %q, want %q", result.Value, "value123")
	}
}
