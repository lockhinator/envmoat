package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

func testStore(t *testing.T) (*Store, func()) {
	t.Helper()
	tmpDir := t.TempDir()

	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
	}
	// Ensure temp dir has correct permissions for ValidatePermissions tests.
	os.Chmod(tmpDir, 0700)
	s.indexLock = flock.New(filepath.Join(tmpDir, ".index.lock"))

	return s, func() {}
}

func TestInitStore(t *testing.T) {
	tmpDir := t.TempDir()
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}

	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore failed: %v", err)
	}

	// Check directories exist.
	if _, err := os.Stat(s.BundlesPath); os.IsNotExist(err) {
		t.Fatal("bundles directory not created")
	}

	// Check config exists.
	if _, err := os.Stat(s.ConfigPath); os.IsNotExist(err) {
		t.Fatal("config.yaml not created")
	}

	// Check index exists.
	if _, err := os.Stat(s.IndexPath); os.IsNotExist(err) {
		t.Fatal("index.json not created")
	}

	// Idempotent: running again should not fail.
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore not idempotent: %v", err)
	}
}

func TestIsInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}

	if s.IsInitialized() {
		t.Fatal("should not be initialized before InitStore")
	}

	if err := s.InitStore(); err != nil {
		t.Fatal(err)
	}

	if !s.IsInitialized() {
		t.Fatal("should be initialized after InitStore")
	}
}

func TestWriteReadBundle(t *testing.T) {
	s, _ := testStore(t)
	if err := s.InitStore(); err != nil {
		t.Fatal(err)
	}

	dek := make([]byte, 32)
	for i := range dek {
		dek[i] = byte(i)
	}

	plaintext := []byte(`{"_meta":{"created_at":"2025-01-01T00:00:00Z"},"KEY":"value"}`)
	filename := "test.enc"

	if err := s.WriteBundle(filename, plaintext, dek); err != nil {
		t.Fatalf("WriteBundle failed: %v", err)
	}

	decrypted, err := s.ReadBundle(filename, dek)
	if err != nil {
		t.Fatalf("ReadBundle failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDeleteBundle(t *testing.T) {
	s, _ := testStore(t)
	if err := s.InitStore(); err != nil {
		t.Fatal(err)
	}

	dek := make([]byte, 32)
	plaintext := []byte(`{}`)
	filename := "delete-me.enc"

	if err := s.WriteBundle(filename, plaintext, dek); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteBundle(filename); err != nil {
		t.Fatalf("DeleteBundle failed: %v", err)
	}

	// Reading deleted bundle should return ErrBundleNotFound.
	_, err := s.ReadBundle(filename, dek)
	if err != ErrBundleNotFound {
		t.Fatalf("expected ErrBundleNotFound, got: %v", err)
	}
}

func TestAutoBundleName(t *testing.T) {
	tests := []struct {
		dir      string
		existing map[string]bool
		want     string
	}{
		{"/Users/test/myproject", nil, "auto-myproject.enc"},
		{"/Users/test/My Project", nil, "auto-my-project.enc"},
		{"/Users/test/my_project", nil, "auto-my-project.enc"},
		{"/Users/test/My--Project", nil, "auto-my-project.enc"},
		{"/Users/test/myproject", map[string]bool{"auto-myproject.enc": true}, ""}, // collision, non-empty
	}

	for _, tt := range tests {
		got := AutoBundleName(tt.dir, tt.existing)
		if tt.want != "" && got != tt.want {
			t.Errorf("AutoBundleName(%q) = %q, want %q", tt.dir, got, tt.want)
		}
		// For collision test, just verify it's different from the existing name.
		if tt.want == "" && len(tt.existing) > 0 {
			if got == "auto-myproject.enc" {
				t.Errorf("collision not resolved for %q", tt.dir)
			}
		}
	}
}

func TestConfigRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteConfig(cfgPath, cfg); err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	read, err := ReadConfig(cfgPath)
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	if read.Version != cfg.Version {
		t.Errorf("version mismatch: got %d, want %d", read.Version, cfg.Version)
	}

	if string(read.GlobalSalt) != string(cfg.GlobalSalt) {
		t.Error("global salt mismatch")
	}

	if read.SessionTTLMinutes != cfg.SessionTTLMinutes {
		t.Errorf("TTL mismatch: got %d, want %d", read.SessionTTLMinutes, cfg.SessionTTLMinutes)
	}
}

func TestValidatePermissions(t *testing.T) {
	s, _ := testStore(t)
	if err := s.InitStore(); err != nil {
		t.Fatal(err)
	}

	// Permissions should be fine after InitStore.
	if err := s.ValidatePermissions(); err != nil {
		t.Fatalf("ValidatePermissions failed after init: %v", err)
	}
}

func TestValidateBundleFilename(t *testing.T) {
	tests := []struct {
		name    string
		filename string
		wantErr bool
	}{
		{"simple name", "test.enc", false},
		{"with dash", "my-bundle.enc", false},
		{"with underscore", "my_bundle.enc", false},
		{"empty", "", true},
		{"contains slash", "../etc/passwd", true},
		{"contains dotdot", "foo/../../bar.enc", true},
		{"backslash", "foo\\bar.enc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBundleFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBundleFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			}
		})
	}
}
