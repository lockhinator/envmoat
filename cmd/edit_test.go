package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditNotFoundKey(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"EXISTING_KEY": "value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"edit", "NONEXISTENT_KEY"})
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' in stderr, got: %q", stderr)
	}
}

func TestEditCancelled(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"EDIT_KEY": "original-value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Set EDITOR to a command that exits non-zero (cancels edit).
	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "false") // 'false' exits with code 1
	defer os.Setenv("EDITOR", oldEditor)

	_, stderr, err := runCmd(t, []string{"edit", "EDIT_KEY"})
	if err != nil {
		t.Fatalf("edit command failed: %v stderr: %s", err, stderr)
	}

	// Verify value was NOT changed (cancelled = no changes).
	stdout, _, err := runCmd(t, []string{"get", "EDIT_KEY"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "original-value") {
		t.Errorf("expected 'original-value' after cancelled edit, got: %q", stdout)
	}
}

func TestEditRoundtrip(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, map[string]string{"EDIT_KEY": "original-value"})
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Create a fake editor that writes a new value to the temp file.
	editorScript := filepath.Join(t.TempDir(), "fake-editor")
	os.WriteFile(editorScript, []byte(`#!/bin/sh
# Read the temp file path from args and write a new value
echo "new-edited-value" > "$1"
`), 0755)

	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", editorScript)
	defer os.Setenv("EDITOR", oldEditor)

	stdout, stderr, err := runCmd(t, []string{"edit", "EDIT_KEY"})
	if err != nil {
		t.Fatalf("edit command failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Updated EDIT_KEY") {
		t.Errorf("expected 'Updated EDIT_KEY' in output, got: %q", stdout)
	}

	// Verify the value was updated.
	stdout, _, err = runCmd(t, []string{"get", "EDIT_KEY"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "new-edited-value") {
		t.Errorf("expected 'new-edited-value' after edit, got: %q", stdout)
	}
}
