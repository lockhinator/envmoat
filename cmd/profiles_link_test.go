package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

func setupProfilesLinkTest(t *testing.T) (string, string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	os.Setenv("ENVMOAT_TEST_PASSWORD", "testpassword")
	// Reset flags that persist across tests.
	profilesLinkForce = false
	profilesUnlinkSkipConfirm = false

	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore() error: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", oldHome)
		os.Unsetenv("ENVMOAT_TEST_PASSWORD")
	}
	return tmpDir, s.ConfigPath, cleanup
}

func TestProfilesLinkValid(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a profile first.
	s, _ := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	cfg, _ := store.ReadConfig(s.ConfigPath)
	luk, _ := crypto.DeriveLUK("testpassword", cfg.GlobalSalt)
	dek, _ := crypto.DeriveDEK(luk, "profile-myprofile.enc")
	emptyBundle, _ := store.MarshalBundle(map[string]string{})
	s.WriteBundle("profile-myprofile.enc", emptyBundle, dek)
	s.AddProfileMapping("myprofile", "profile-myprofile.enc")

	// Run profiles link.
	rootCmd.SetArgs([]string{"profiles", "link", projectDir, "myprofile"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("profiles link error: %v", err)
	}

	// Verify marker file was created with correct content.
	markerPath := filepath.Join(projectDir, ".envmoat")
	content, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.Contains(string(content), "profile: myprofile") {
		t.Errorf("marker content = %q, want 'profile: myprofile'", string(content))
	}

	// Verify .gitignore was created.
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	content, err = os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(content), ".envmoat") {
		t.Error(".gitignore should contain .envmoat")
	}
}

func TestProfilesLinkForce(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a profile.
	s, _ := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	cfg, _ := store.ReadConfig(s.ConfigPath)
	luk, _ := crypto.DeriveLUK("testpassword", cfg.GlobalSalt)
	dek, _ := crypto.DeriveDEK(luk, "profile-existing.enc")
	emptyBundle, _ := store.MarshalBundle(map[string]string{})
	s.WriteBundle("profile-existing.enc", emptyBundle, dek)
	s.AddProfileMapping("existing", "profile-existing.enc")

	// Create an existing marker.
	markerPath := filepath.Join(projectDir, ".envmoat")
	os.WriteFile(markerPath, []byte("profile: oldprofile\n"), 0o600)

	// Run profiles link --force.
	rootCmd.SetArgs([]string{"profiles", "link", "--force", projectDir, "existing"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("profiles link --force error: %v", err)
	}

	// Verify marker was overwritten.
	content, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.Contains(string(content), "profile: existing") {
		t.Errorf("marker content = %q, want 'profile: existing'", string(content))
	}
}

func TestProfilesLinkAutoCreateProfile(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Run profiles link with a non-existent profile — should auto-create.
	rootCmd.SetArgs([]string{"profiles", "link", projectDir, "autocreated"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("profiles link error: %v", err)
	}

	// Verify marker was created.
	markerPath := filepath.Join(projectDir, ".envmoat")
	content, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.Contains(string(content), "profile: autocreated") {
		t.Errorf("marker content = %q, want 'profile: autocreated'", string(content))
	}

	// Verify profile was created in store.
	s, _ := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	bundleFile, ok := s.GetProfileBundle("autocreated")
	if !ok {
		t.Fatal("profile autocreated should exist in store")
	}
	if bundleFile == "" {
		t.Error("bundle file should not be empty")
	}

	// Verify bundle file exists on disk.
	bundlePath := filepath.Join(tmpDir, ".envmoat", "bundles", bundleFile)
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		t.Fatalf("bundle file %s should exist on disk", bundlePath)
	}
}

func TestProfilesLinkNoMarker(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a profile.
	s, _ := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	cfg, _ := store.ReadConfig(s.ConfigPath)
	luk, _ := crypto.DeriveLUK("testpassword", cfg.GlobalSalt)
	dek, _ := crypto.DeriveDEK(luk, "profile-test.enc")
	emptyBundle, _ := store.MarshalBundle(map[string]string{})
	s.WriteBundle("profile-test.enc", emptyBundle, dek)
	s.AddProfileMapping("test", "profile-test.enc")

	// Run profiles link without --force on existing marker.
	markerPath := filepath.Join(projectDir, ".envmoat")
	os.WriteFile(markerPath, []byte("old\n"), 0o600)

	rootCmd.SetArgs([]string{"profiles", "link", projectDir, "test"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("profiles link without --force should error on existing marker")
	}
}
