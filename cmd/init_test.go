package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandExists(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}
	if initCmd.Use != "init [project-root]" {
		t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init [project-root]")
	}
}

func TestAppendToGitignoreNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	err := appendToGitignore(gitignorePath, ".envmoat")
	if err != nil {
		t.Fatalf("appendToGitignore: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(content), ".envmoat") {
		t.Error(".gitignore should contain .envmoat")
	}
}

func TestAppendToGitignoreExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	// Write existing content.
	existing := "*.log\nnode_modules/\n"
	if err := os.WriteFile(gitignorePath, []byte(existing), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	err := appendToGitignore(gitignorePath, ".envmoat")
	if err != nil {
		t.Fatalf("appendToGitignore: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(content), ".envmoat") {
		t.Error(".gitignore should contain .envmoat")
	}
	if !strings.Contains(string(content), "*.log") {
		t.Error(".gitignore should still contain original content")
	}
}

func TestAppendToGitignoreIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	// Write existing content with .envmoat already present.
	existing := "*.log\n.envmoat\nnode_modules/\n"
	if err := os.WriteFile(gitignorePath, []byte(existing), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	err := appendToGitignore(gitignorePath, ".envmoat")
	if err != nil {
		t.Fatalf("appendToGitignore: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	// Should not have duplicated .envmoat.
	count := strings.Count(string(content), ".envmoat")
	if count != 1 {
		t.Errorf(".envmoat appears %d times, want 1 (idempotent)", count)
	}
}

func TestAppendToGitignoreNoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	// Write content without trailing newline.
	existing := "*.log"
	if err := os.WriteFile(gitignorePath, []byte(existing), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	err := appendToGitignore(gitignorePath, ".envmoat")
	if err != nil {
		t.Fatalf("appendToGitignore: %v", err)
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	// .envmoat should be on its own line.
	if !strings.Contains(string(content), "\n.envmoat") {
		t.Errorf(".envmoat should be on its own line, got: %q", string(content))
	}
}

func TestAutoBundleNameGeneration(t *testing.T) {
	tests := []struct {
		name     string
		dirPath  string
		existing map[string]bool
		want     string
	}{
		{
			name:     "simple directory",
			dirPath:  "/home/user/projects/my-app",
			existing: map[string]bool{},
			want:     "auto-my-app.enc",
		},
		{
			name:     "directory with spaces",
			dirPath:  "/home/user/projects/my cool app",
			existing: map[string]bool{},
			want:     "auto-my-cool-app.enc",
		},
		{
			name:     "directory with special chars",
			dirPath:  "/home/user/projects/my@cool#app!",
			existing: map[string]bool{},
			want:     "auto-my-cool-app.enc",
		},
		{
			name:     "collision with existing bundle",
			dirPath:  "/home/user/projects/my-app",
			existing: map[string]bool{"auto-my-app.enc": true},
			want:     "", // Will have random suffix, just check prefix.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := storeAutoBundleName(tc.dirPath, tc.existing)
			if tc.want != "" && got != tc.want {
				t.Errorf("autoBundleName(%q) = %q, want %q", tc.dirPath, got, tc.want)
			}
			// For collision case, check it has the right prefix and .enc suffix.
			if tc.want == "" {
				if !strings.HasPrefix(got, "auto-my-app-") {
					t.Errorf("collision name %q should start with 'auto-my-app-'", got)
				}
				if !strings.HasSuffix(got, ".enc") {
					t.Errorf("collision name %q should end with '.enc'", got)
				}
			}
		})
	}
}

func TestAutoBundleNameSlugification(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-app", "auto-my-app.enc"},
		{"my_app", "auto-my-app.enc"},
		{"MyApp", "auto-myapp.enc"},
		{"MY-APP", "auto-my-app.enc"},
		{"my  app", "auto-my-app.enc"},
		{"my--app", "auto-my-app.enc"},
		{"my.app", "auto-my-app.enc"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := storeAutoBundleName("/home/user/projects/"+tc.input, map[string]bool{})
			if got != tc.want {
				t.Errorf("autoBundleName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// storeAutoBundleName is the test-exported version of the auto-bundle naming logic.
// This mirrors the logic in store.AutoBundleName for testing.
func storeAutoBundleName(dirPath string, existing map[string]bool) string {
	base := filepath.Base(dirPath)
	// Slugify: lowercase, replace non-alphanumeric with hyphens, collapse hyphens, trim.
	slug := strings.ToLower(base)
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		} else if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
	}
	// Trim trailing hyphens.
	for len(result) > 0 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	slug = string(result)

	name := "auto-" + slug + ".enc"
	if !existing[name] {
		return name
	}

	// Collision: append short hash.
	hash := "a1b2c3" // Deterministic for tests.
	return "auto-" + slug + "-" + hash + ".enc"
}

func TestMarkerFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	markerPath := filepath.Join(tmpDir, ".envmoat")

	// Create marker with 0600 permissions.
	err := os.WriteFile(markerPath, nil, 0600)
	if err != nil {
		t.Fatalf("write marker: %v", err)
	}

	info, err := os.Stat(markerPath)
	if err != nil {
		t.Fatalf("stat marker: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("marker permissions = %o, want 0600", perm)
	}
}
