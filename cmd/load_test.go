package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

func TestLoadShellEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"plain", "sk-1234567890abcdef", "'sk-1234567890abcdef'"},
		{"empty", "", "''"},
		{"spaces", "hello world", "'hello world'"},
		{"single quote", "it's", `'it'\''s'`},
		{"multiple single quotes", "a'b'c", `'a'\''b'\''c'`},
		{"double quotes", `he said "hello"`, `'he said "hello"'`},
		{"dollar sign", "$HOME", "'$HOME'"},
		{"backtick", "`command`", "'`command`'"},
		{"backslash", `C:\Users\test`, `'C:\Users\test'`},
		{"exclamation", "don't!", `'don'\''t!'`},
		{"ampersand", "foo & bar", "'foo & bar'"},
		{"pipe", "cat file | grep", "'cat file | grep'"},
		{"semicolon", "echo a; echo b", "'echo a; echo b'"},
		{"newline", "line1\nline2", "'line1\nline2'"},
		{"tab", "col1\tcol2", "'col1\tcol2'"},
		{"mixed", `it's a "complex" value with $pecial & chars`, `'it'\''s a "complex" value with $pecial & chars'`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cmdutil.ShellEscapeSingleQuote(tc.value)
			if result != tc.expected {
				t.Errorf("ShellEscapeSingleQuote(%q) = %q, want %q", tc.value, result, tc.expected)
			}
		})
	}
}

func TestEmitLoadOutput(t *testing.T) {
	secrets := map[string]string{
		"API_KEY": "sk-123", "DB_PASS": "p@$$w0rd", "Z_SORT_LAST": "z",
	}
	var stdout, stderr bytes.Buffer
	err := cmdutil.EmitLoadOutput(&stdout, &stderr, "abc123", secrets)
	if err != nil {
		t.Fatalf("EmitLoadOutput error: %v", err)
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "#bundle_hash:sha256:abc123\n") {
		t.Errorf("Missing hash line. Got:\n%s", out)
	}
	for _, key := range []string{"API_KEY", "DB_PASS", "Z_SORT_LAST"} {
		expected := "export " + key + "=" + cmdutil.ShellEscapeSingleQuote(secrets[key])
		if !strings.Contains(out, expected+"\n") {
			t.Errorf("Missing export line for %s. Got:\n%s", key, out)
		}
	}
	apiIdx := strings.Index(out, "export API_KEY")
	dbIdx := strings.Index(out, "export DB_PASS")
	zIdx := strings.Index(out, "export Z_SORT_LAST")
	if apiIdx >= dbIdx || dbIdx >= zIdx {
		t.Errorf("Exports not sorted: API_KEY@%d DB_PASS@%d Z_SORT_LAST@%d", apiIdx, dbIdx, zIdx)
	}
	if stderr.Len() > 0 {
		t.Errorf("Expected empty stderr, got: %s", stderr.String())
	}
}

func TestEmitLoadOutputEmptyBundle(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := cmdutil.EmitLoadOutput(&stdout, &stderr, "def456", map[string]string{})
	if err != nil {
		t.Fatalf("EmitLoadOutput error: %v", err)
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "#bundle_hash:sha256:def456\n") {
		t.Errorf("Missing hash line. Got:\n%s", out)
	}
	if strings.Contains(out, "export") {
		t.Errorf("Expected no exports for empty bundle. Got:\n%s", out)
	}
}

func setupLoadTest(t *testing.T, secrets map[string]string) (string, string, string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	// Canonicalize tmpDir for macOS symlink resolution
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error: %v", tmpDir, err)
	}
	oldHome := os.Getenv("HOME")
	oldEnvLUK := os.Getenv("ENVMOAT_LUK")
	oldPwd, _ := os.Getwd()
	os.Setenv("ENVMOAT_LUK", "") // Clear any inherited LUK
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
	// Canonicalize path for macOS symlink resolution (/var/folders -> /private/var/folders)
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
		os.Setenv("ENVMOAT_LUK", oldEnvLUK)
		os.Chdir(oldPwd)
	}
	return tmpDir, markerDir, bundleFile, cleanup
}

func TestLoadCommandBasic(t *testing.T) {
	tmpDir, markerDir, bundleFile, cleanup := setupLoadTest(t, map[string]string{"API_KEY": "sk-test123"})
	defer cleanup()
	os.Setenv("HOME", tmpDir)
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"load"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "export API_KEY=") {
		t.Fatalf("Expected export API_KEY in output. Got:\nstdout: %s\nstderr: %s", out, stderr.String())
	}
	if !strings.Contains(out, "sk-test123") {
		t.Fatalf("Expected secret value in output. Got:\n%s", out)
	}
	if !strings.HasPrefix(out, "#bundle_hash:sha256:") {
		t.Fatalf("Expected hash line. Got:\n%s", out)
	}
	_ = bundleFile
	_ = stderr
}

func TestLoadCommandNoMarker(t *testing.T) {
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
	rootCmd.SetArgs([]string{"load"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout, got: %s", stdout.String())
	}
	_ = stderr
}

func TestLoadCommandDisabledMarker(t *testing.T) {
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
	rootCmd.SetArgs([]string{"load"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout, got: %s", stdout.String())
	}
	_ = stderr
}

func TestLoadCommandMissingBundle(t *testing.T) {
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
	markerDir, err = filepath.EvalSymlinks(markerDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q) error: %v", markerDir, err)
	}
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"load"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "no bundle found") {
		t.Errorf("Expected warning on stderr, got: %s", stderr.String())
	}
}

func TestLoadCommandCorruptedBundle(t *testing.T) {
	tmpDir, markerDir, bundleFile, cleanup := setupLoadTest(t, map[string]string{"API_KEY": "sk-test123"})
	defer cleanup()
	storeDir := filepath.Join(tmpDir, ".envmoat")
	bundlePath := filepath.Join(storeDir, "bundles", bundleFile)
	os.WriteFile(bundlePath, []byte("corrupted data"), 0o600)
	os.Setenv("HOME", tmpDir)
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"load"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("Expected no stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "corrupted") {
		t.Errorf("Expected warning about corruption on stderr, got: %s", stderr.String())
	}
}

func TestLoadCommandSingleQuoteEscaping(t *testing.T) {
	tmpDir, markerDir, _, cleanup := setupLoadTest(t, map[string]string{"SECRET": "it's a secret"})
	defer cleanup()
	os.Setenv("HOME", tmpDir)
	os.Chdir(markerDir)
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"load"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd load error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "'\\''") {
		t.Errorf("Expected escaped single quote in output. Got:\n%s", out)
	}
	_ = stderr
}
