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

func setupDeinitTest(t *testing.T) (string, string, string, func()) {
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
	return tmpDir, markerDir, bundleFile, cleanup
}

func TestDeinitCommandBasic(t *testing.T) {
	tmpDir, markerDir, bundleFile, cleanup := setupDeinitTest(t)
	defer cleanup()
	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	os.Setenv("HOME", tmpDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"deinit", "-y", markerDir})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd deinit error: %v", err)
	}
	markerPath := filepath.Join(markerDir, resolver.MarkerName)
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("Marker file should have been removed")
	}
	bundlePath := filepath.Join(s.BundlesPath, bundleFile)
	if _, err := os.Stat(bundlePath); !os.IsNotExist(err) {
		t.Error("Bundle file should have been removed")
	}
	idx, err := s.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex() error: %v", err)
	}
	if fn, exists := idx.Auto[markerDir]; exists {
		t.Errorf("Auto mapping should be removed, still present: %s", fn)
	}
	if !strings.Contains(stderr.String(), "deactivated") {
		t.Errorf("Expected 'deactivated' on stderr, got: %s", stderr.String())
	}
	_ = stdout
}

func TestDeinitCommandNoMarker(t *testing.T) {
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
	rootCmd.SetArgs([]string{"deinit", "-y", filepath.Join(tmpDir, "project")})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no marker found")
	}
	_ = stdout
	_ = stderr
}

func TestDeinitCommandDisabledMarker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}()
	markerDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(markerDir, 0o755)
	os.WriteFile(filepath.Join(markerDir, resolver.MarkerName), []byte("disabled"), 0o644)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"deinit", "-y", markerDir})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd deinit error: %v", err)
	}
	markerPath := filepath.Join(markerDir, resolver.MarkerName)
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("Marker file should have been removed")
	}
	_ = stdout
	_ = stderr
}

func TestDeinitCommandNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"deinit"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no args provided")
	}
	_ = stdout
	_ = stderr
}

func TestDeinitCommandWrongDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}()
	markerDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(markerDir, 0o755)
	os.WriteFile(filepath.Join(markerDir, resolver.MarkerName), nil, 0o644)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"deinit", "-y", filepath.Join(tmpDir, "other")})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when marker not in specified directory")
	}
	_ = stdout
	_ = stderr
}
