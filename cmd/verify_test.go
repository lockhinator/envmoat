package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

func setupVerifyTest(t *testing.T) (string, string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore() error: %v", err)
	}
	markerDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(markerDir, 0o755)
	os.WriteFile(filepath.Join(markerDir, resolver.MarkerName), nil, 0o644)
	markerDir, err = filepath.EvalSymlinks(markerDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error: %v", markerDir, err)
	}
	cfg, err := store.ReadConfig(s.ConfigPath)
	if err != nil {
		t.Fatalf("ReadConfig() error: %v", err)
	}
	password := "testpassword"
	luk, err := crypto.DeriveLUK(password, cfg.GlobalSalt)
	if err != nil {
		t.Fatalf("DeriveLUK() error: %v", err)
	}
	bundleFile := "test-bundle.enc"
	dek, err := crypto.DeriveDEK(luk, bundleFile)
	if err != nil {
		t.Fatalf("DeriveDEK() error: %v", err)
	}
	plaintext, err := store.MarshalBundle(map[string]string{"API_KEY": "sk-test123"})
	if err != nil {
		t.Fatalf("MarshalBundle() error: %v", err)
	}
	if err := s.WriteBundle(bundleFile, plaintext, dek); err != nil {
		t.Fatalf("WriteBundle() error: %v", err)
	}
	if err := s.AddAutoMapping(markerDir, bundleFile); err != nil {
		t.Fatalf("AddAutoMapping() error: %v", err)
	}
	auth.SetLUK(s.BasePath, luk)
	cleanup := func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}
	return tmpDir, bundleFile, cleanup
}

func TestVerifyCommandHealthyStore(t *testing.T) {
	tmpDir, _, cleanup := setupVerifyTest(t)
	defer cleanup()
	os.Setenv("HOME", tmpDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"verify"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd verify error: %v", err)
	}
	if !strings.Contains(stderr.String(), "healthy") {
		t.Errorf("Expected 'healthy' on stderr, got: %s", stderr.String())
	}
	_ = stdout
}

func TestVerifyCommandOrphanedBundles(t *testing.T) {
	tmpDir, _, cleanup := setupVerifyTest(t)
	defer cleanup()
	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	luk, _ := auth.GetLUK(s.BasePath)
	orphanFile := "orphan-bundle.enc"
	dek, _ := crypto.DeriveDEK(luk, orphanFile)
	plaintext, _ := store.MarshalBundle(map[string]string{"ORPHAN_KEY": "value"})
	s.WriteBundle(orphanFile, plaintext, dek)
	os.Setenv("HOME", tmpDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"verify"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd verify error: %v", err)
	}
	if !strings.Contains(stderr.String(), "orphaned") {
		t.Errorf("Expected 'orphaned' on stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), orphanFile) {
		t.Errorf("Expected orphan filename on stderr, got: %s", stderr.String())
	}
	_ = stdout
}

func TestVerifyCommandNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}()
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"verify"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when store not initialized")
	}
	_ = stdout
	_ = stderr
}

func TestVerifyCommandCorruptedBundle(t *testing.T) {
	tmpDir, bundleFile, cleanup := setupVerifyTest(t)
	defer cleanup()
	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	bundlePath := filepath.Join(s.BundlesPath, bundleFile)
	os.WriteFile(bundlePath, []byte("corrupted data"), 0o600)
	os.Setenv("HOME", tmpDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"verify"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for corrupted bundle")
	}
	_ = stdout
	_ = stderr
}
