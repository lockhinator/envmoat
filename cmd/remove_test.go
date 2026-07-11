package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

func setupRemoveTest(t *testing.T, secrets map[string]string) (string, string, string, func()) {
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
	plaintext, err := store.MarshalBundle(secrets)
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

func TestRemoveCommandBasic(t *testing.T) {
	tmpDir, markerDir, bundleFile, cleanup := setupRemoveTest(t, map[string]string{
		"API_KEY": "sk-test123", "DB_PASS": "password123",
	})
	defer cleanup()
	s, err := store.NewStoreAt(filepath.Join(tmpDir, ".envmoat"))
	if err != nil {
		t.Fatalf("NewStoreAt() error: %v", err)
	}
	os.Setenv("HOME", tmpDir)
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"remove", "-y", "API_KEY"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd remove error: %v", err)
	}
	if !strings.Contains(stderr.String(), "removed secret") {
		t.Errorf("Expected 'removed secret' on stderr, got: %s", stderr.String())
	}
	luk, _ := auth.GetLUK(s.BasePath)
	dek, _ := crypto.DeriveDEK(luk, bundleFile)
	plaintext, _ := s.ReadBundle(bundleFile, dek)
	var bundleData map[string]json.RawMessage
	json.Unmarshal(plaintext, &bundleData)
	if _, exists := bundleData["API_KEY"]; exists {
		t.Error("API_KEY should have been removed from bundle")
	}
	if _, exists := bundleData["DB_PASS"]; !exists {
		t.Error("DB_PASS should still exist in bundle")
	}
	_ = stdout
}

func TestRemoveCommandNoMarker(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}()
	os.Chdir(tmpDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"remove", "-y", "API_KEY"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no marker found")
	}
	_ = stdout
	_ = stderr
}

func TestRemoveCommandDisabledMarker(t *testing.T) {
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
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"remove", "-y", "API_KEY"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error for disabled marker")
	}
	_ = stdout
	_ = stderr
}

func TestRemoveCommandNoBundle(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	oldPwd, _ := os.Getwd()
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Chdir(oldPwd)
	}()
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
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"remove", "-y", "API_KEY"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no bundle found")
	}
	_ = stdout
	_ = stderr
}

func TestRemoveCommandNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"remove"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected error when no args provided")
	}
	_ = stdout
	_ = stderr
}
