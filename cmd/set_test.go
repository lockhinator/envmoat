package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lockinator/envmoat/internal/backend"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

// mockKeyringBackend implements backend.KeyringBackend for testing.
type mockKeyringBackend struct {
	luk []byte
}

func (m *mockKeyringBackend) StoreLUK(key []byte) error {
	m.luk = append([]byte(nil), key...)
	return nil
}

func (m *mockKeyringBackend) GetLUK() ([]byte, error) {
	if m.luk == nil {
		return nil, backend.ErrNotAvailable
	}
	return append([]byte(nil), m.luk...), nil
}

func (m *mockKeyringBackend) DeleteLUK() error {
	m.luk = nil
	return nil
}

// testEnv sets up a temporary envmoat store with a bundle for testing.
// Returns the temp dir, marker dir, and cleanup function.
// Sets HOME to tmpDir so store.NewStore() finds the test store at tmpDir/.envmoat.
func testEnv(t *testing.T, secrets map[string]string) (tmpDir, markerDir string, cleanup func()) {
	t.Helper()
	tmpDir = t.TempDir()
	os.Chmod(tmpDir, 0700)

	// Set HOME so store.NewStore() finds the test store at tmpDir/.envmoat.
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	s, err := store.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitStore(); err != nil {
		t.Fatal(err)
	}

	// Create marker dir with .envmoat marker.
	markerDir = t.TempDir()
	markerDir, err = filepath.EvalSymlinks(markerDir)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(markerDir, ".envmoat"), nil, 0644)

	// Register marker dir in index.
	bundleFile := store.AutoBundleName(markerDir, nil)
	if err := s.AddAutoMapping(markerDir, bundleFile); err != nil {
		t.Fatal(err)
	}

	// Write bundle with secrets.
	luk := make([]byte, 32)
	for i := range luk {
		luk[i] = byte(i)
	}
	dek, err := crypto.DeriveDEK(luk, bundleFile)
	if err != nil {
		t.Fatal(err)
	}

	plaintext, _ := json.Marshal(secrets)
	if err := s.WriteBundle(bundleFile, plaintext, dek); err != nil {
		t.Fatal(err)
	}

	// Set up keyring with LUK.
	keyringBackend = &mockKeyringBackend{luk: luk}

	cleanup = func() {
		os.Setenv("HOME", oldHome)
		keyringBackend = backend.NewKeyringBackend()
		// Reset global flags that persist across tests.
		setFile = ""
		setStdin = false
		getJSON = false
	}

	return
}

// runCmd executes rootCmd with the given args, capturing stdout and stderr.
// Returns stdout, stderr strings and any error.
// Uses ExecuteC() to avoid os.Exit on error.
func runCmd(t *testing.T, args []string) (string, string, error) {
	t.Helper()

	// Redirect os.Stdout/os.Stderr via pipes since commands write directly to them.
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	rootCmd.SetArgs(args)
	_, err := rootCmd.ExecuteC()

	stdoutW.Close()
	stderrW.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR) //nolint:errcheck
	io.Copy(&stderrBuf, stderrR) //nolint:errcheck
	stdoutR.Close()
	stderrR.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestSetWithValue(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	stdout, stderr, err := runCmd(t, []string{"set", "API_KEY", "secret123"})
	if err != nil {
		t.Fatalf("set command failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Set API_KEY") {
		t.Errorf("expected 'Set API_KEY' in output, got: %q", stdout)
	}
}

func TestSetStdin(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Set up stdin.
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte("stdin-value\n"))
		w.Close()
	}()
	defer func() {
		os.Stdin = oldStdin
	}()

	stdout, stderr, err := runCmd(t, []string{"set", "STDIN_KEY", "--stdin"})
	if err != nil {
		t.Fatalf("set --stdin failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Set STDIN_KEY") {
		t.Errorf("expected 'Set STDIN_KEY' in output, got: %q", stdout)
	}
}

func TestSetFileImport(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	envFile := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(envFile, []byte("KEY_A=value_a\nKEY_B=value_b\n"), 0644)

	stdout, stderr, err := runCmd(t, []string{"set", "--file", envFile})
	if err != nil {
		t.Fatalf("set --file failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Imported 2 secrets") {
		t.Errorf("expected 'Imported 2 secrets' in output, got: %q", stdout)
	}
}

func TestSetFileImportSkipsComments(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	envFile := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(envFile, []byte("# comment\nKEY_A=value_a\n\n# another\nKEY_B=value_b\n"), 0644)

	stdout, stderr, err := runCmd(t, []string{"set", "--file", envFile})
	if err != nil {
		t.Fatalf("set --file failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Imported 2 secrets") {
		t.Errorf("expected 'Imported 2 secrets' in output, got: %q", stdout)
	}
}

func TestSetFileImportHandlesQuotes(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	envFile := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(envFile, []byte(`KEY_A="quoted value"
KEY_B='single quoted'
`), 0644)

	stdout, stderr, err := runCmd(t, []string{"set", "--file", envFile})
	if err != nil {
		t.Fatalf("set --file failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Imported 2 secrets") {
		t.Errorf("expected 'Imported 2 secrets' in output, got: %q", stdout)
	}
}

func TestSetKeyValidation(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"set", "invalid key!", "value"})
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if !strings.Contains(stderr, "invalid key") {
		t.Errorf("expected 'invalid key' in stderr, got: %q", stderr)
	}
}

func TestSetSizeWarning(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	largeValue := strings.Repeat("x", SecretSizeWarning+1)

	_, stderr, err := runCmd(t, []string{"set", "LARGE_KEY", largeValue})
	if err != nil {
		t.Fatalf("set failed: %v stderr: %s", err, stderr)
	}
	if !strings.Contains(stderr, "warning") {
		t.Errorf("expected size warning in stderr, got: %q", stderr)
	}
}

func TestSetNotInTrackedDir(t *testing.T) {
	_, _, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	untrackedDir := t.TempDir()
	os.Chdir(untrackedDir)
	defer os.Chdir(origDir)

	_, stderr, err := runCmd(t, []string{"set", "KEY", "value"})
	if err == nil {
		t.Fatal("expected error when not in tracked directory")
	}
	if !strings.Contains(stderr, "tracked directory") {
		t.Errorf("expected 'tracked directory' in stderr, got: %q", stderr)
	}
}

func TestSetRoundtrip(t *testing.T) {
	_, markerDir, cleanup := testEnv(t, nil)
	defer cleanup()

	origDir, _ := os.Getwd()
	os.Chdir(markerDir)
	defer os.Chdir(origDir)

	// Set a key.
	_, stderr, err := runCmd(t, []string{"set", "ROUNDTRIP_KEY", "roundtrip-value"})
	if err != nil {
		t.Fatalf("set failed: %v stderr: %s", err, stderr)
	}

	// Get the key back.
	stdout, _, err := runCmd(t, []string{"get", "ROUNDTRIP_KEY"})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "roundtrip-value") {
		t.Errorf("expected 'roundtrip-value' in output, got: %q", stdout)
	}
}
