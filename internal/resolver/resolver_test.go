package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

// --- ParseMarker tests ---

func TestParseMarkerEmpty(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte(""), 0600)

	content, name, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerDefault {
		t.Errorf("expected MarkerDefault, got %v", content)
	}
	if name != "" {
		t.Errorf("expected empty profile name, got %q", name)
	}
}

func TestParseMarkerWhitespaceOnly(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("  \n\n  "), 0600)

	content, _, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerDefault {
		t.Errorf("expected MarkerDefault, got %v", content)
	}
}

func TestParseMarkerDisabled(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("disabled\n"), 0600)

	content, _, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerDisabled {
		t.Errorf("expected MarkerDisabled, got %v", content)
	}
}

func TestParseMarkerDisabledWithWhitespace(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("  disabled  \n"), 0600)

	content, _, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerDisabled {
		t.Errorf("expected MarkerDisabled, got %v", content)
	}
}

func TestParseMarkerProfile(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("profile: myapp-dev\n"), 0600)

	content, name, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerProfile {
		t.Errorf("expected MarkerProfile, got %v", content)
	}
	if name != "myapp-dev" {
		t.Errorf("expected profile name 'myapp-dev', got %q", name)
	}
}

func TestParseMarkerProfileWithExtraWhitespace(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("  profile:   staging  \n"), 0600)

	content, name, err := ParseMarker(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != MarkerProfile {
		t.Errorf("expected MarkerProfile, got %v", content)
	}
	if name != "staging" {
		t.Errorf("expected profile name 'staging', got %q", name)
	}
}

func TestParseMarkerEmptyProfileName(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("profile: \n"), 0600)

	_, _, err := ParseMarker(path)
	if err == nil {
		t.Fatal("expected error for empty profile name")
	}
}

func TestParseMarkerUnrecognized(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("some random text\n"), 0600)

	_, _, err := ParseMarker(path)
	if err == nil {
		t.Fatal("expected error for unrecognized content")
	}
}

func TestParseMarkerCaseSensitive(t *testing.T) {
	dir := tmpDir(t)
	path := filepath.Join(dir, MarkerName)
	os.WriteFile(path, []byte("DISABLED\n"), 0600)

	_, _, err := ParseMarker(path)
	if err == nil {
		t.Fatal("expected error for uppercase 'DISABLED' — should be case-sensitive")
	}
}

func TestParseMarkerNonexistent(t *testing.T) {
	_, _, err := ParseMarker("/nonexistent/.envmoat")
	if err == nil {
		t.Fatal("expected error for nonexistent marker")
	}
}

// --- Resolve tests ---

func TestResolveFindsMarker(t *testing.T) {
	dir := tmpDir(t)
	os.WriteFile(filepath.Join(dir, MarkerName), []byte(""), 0600)

	result, err := Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.MarkerDir != dir {
		t.Errorf("expected marker dir %s, got %s", dir, result.MarkerDir)
	}
	if result.Marker != MarkerDefault {
		t.Errorf("expected MarkerDefault, got %v", result.Marker)
	}
}

func TestResolveWalksUp(t *testing.T) {
	root := tmpDir(t)
	child := filepath.Join(root, "child", "grandchild")
	os.MkdirAll(child, 0755)
	os.WriteFile(filepath.Join(root, MarkerName), []byte("profile: test-profile\n"), 0600)

	result, err := Resolve(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.MarkerDir != root {
		t.Errorf("expected marker dir %s, got %s", root, result.MarkerDir)
	}
	if result.Marker != MarkerProfile {
		t.Errorf("expected MarkerProfile, got %v", result.Marker)
	}
	if result.ProfileName != "test-profile" {
		t.Errorf("expected profile name 'test-profile', got %q", result.ProfileName)
	}
}

func TestResolveStopsAtDisabled(t *testing.T) {
	root := tmpDir(t)
	child := filepath.Join(root, "child")
	os.MkdirAll(child, 0755)
	os.WriteFile(filepath.Join(root, MarkerName), []byte("profile: parent\n"), 0600)
	os.WriteFile(filepath.Join(child, MarkerName), []byte("disabled\n"), 0600)

	result, err := Resolve(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.MarkerDir != child {
		t.Errorf("expected marker dir %s (disabled), got %s", child, result.MarkerDir)
	}
	if result.Marker != MarkerDisabled {
		t.Errorf("expected MarkerDisabled, got %v", result.Marker)
	}
}

func TestResolveNoMarker(t *testing.T) {
	dir := tmpDir(t)
	// Set a walk root so we don't walk to /.
	orig := os.Getenv("ENVMOAT_WALK_ROOT")
	os.Setenv("ENVMOAT_WALK_ROOT", dir)
	defer os.Setenv("ENVMOAT_WALK_ROOT", orig)

	result, err := Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result (no marker), got %+v", result)
	}
}

// --- ResolveFromPWD test ---

func TestResolveFromPWD(t *testing.T) {
	dir := tmpDir(t)
	os.WriteFile(filepath.Join(dir, MarkerName), []byte("profile: pwd-test\n"), 0600)

	origPwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origPwd)

	result, err := ResolveFromPWD()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Marker != MarkerProfile {
		t.Errorf("expected MarkerProfile, got %v", result.Marker)
	}
	if result.ProfileName != "pwd-test" {
		t.Errorf("expected profile name 'pwd-test', got %q", result.ProfileName)
	}
}

// --- FindWalkRoot tests ---

func TestFindWalkRootDefault(t *testing.T) {
	orig := os.Getenv("ENVMOAT_WALK_ROOT")
	os.Unsetenv("ENVMOAT_WALK_ROOT")
	defer os.Setenv("ENVMOAT_WALK_ROOT", orig)

	root := FindWalkRoot()
	if root != "/" {
		t.Errorf("expected walk root '/', got %q", root)
	}
}

func TestFindWalkRootEnvVar(t *testing.T) {
	orig := os.Getenv("ENVMOAT_WALK_ROOT")
	os.Setenv("ENVMOAT_WALK_ROOT", "/tmp")
	defer os.Setenv("ENVMOAT_WALK_ROOT", orig)

	root := FindWalkRoot()
	if root != "/tmp" {
		t.Errorf("expected walk root '/tmp', got %q", root)
	}
}

// --- ResolveResult marker content values ---

func TestMarkerContentValues(t *testing.T) {
	if MarkerDefault != 0 {
		t.Errorf("expected MarkerDefault == 0, got %d", MarkerDefault)
	}
	if MarkerDisabled != 1 {
		t.Errorf("expected MarkerDisabled == 1, got %d", MarkerDisabled)
	}
	if MarkerProfile != 2 {
		t.Errorf("expected MarkerProfile == 2, got %d", MarkerProfile)
	}
}

// tmpDir creates a temporary directory and returns its absolute path.
// If passed to a *testing.T, it will be cleaned up after the test.
func tmpDir(t ...*testing.T) string {
	dir, err := os.MkdirTemp("", "envmoat-resolver-*")
	if err != nil {
		panic(err)
	}
	if len(t) > 0 && t[0] != nil {
		t[0].Cleanup(func() { os.RemoveAll(dir) })
	}
	return dir
}
