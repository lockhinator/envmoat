package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

func TestProfilesUnlinkValid(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a profile and marker.
	s, _ := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	cfg, _ := store.ReadConfig(s.ConfigPath)
	luk, _ := crypto.DeriveLUK("testpassword", cfg.GlobalSalt)
	dek, _ := crypto.DeriveDEK(luk, "profile-unlinktest.enc")
	emptyBundle, _ := store.MarshalBundle(map[string]string{})
	s.WriteBundle("profile-unlinktest.enc", emptyBundle, dek)
	s.AddProfileMapping("unlinktest", "profile-unlinktest.enc")

	markerPath := filepath.Join(projectDir, ".envmoat")
	os.WriteFile(markerPath, []byte("profile: unlinktest\n"), 0o600)

	// Run profiles unlink with --yes.
	rootCmd.SetArgs([]string{"profiles", "unlink", "--yes", projectDir})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("profiles unlink error: %v", err)
	}

	// Verify marker was removed.
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Fatal("marker should be removed after unlink")
	}

	// Verify profile still exists in store (bundle not deleted).
	_, ok := s.GetProfileBundle("unlinktest")
	if !ok {
		t.Error("profile should still exist after unlink")
	}
}

func TestProfilesUnlinkNoMarker(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Run profiles unlink on directory without marker.
	rootCmd.SetArgs([]string{"profiles", "unlink", "--yes", projectDir})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("profiles unlink should error when no marker exists")
	}
}

func TestProfilesUnlinkNonProfileMarker(t *testing.T) {
	tmpDir, _, cleanup := setupProfilesLinkTest(t)
	defer cleanup()

	projectDir := filepath.Join(tmpDir, "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Create a non-profile marker.
	markerPath := filepath.Join(projectDir, ".envmoat")
	os.WriteFile(markerPath, []byte(""), 0o600) // empty marker = auto bundle

	rootCmd.SetArgs([]string{"profiles", "unlink", "--yes", projectDir})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("profiles unlink should error on non-profile marker")
	}
}
