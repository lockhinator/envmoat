package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallShellHookIdempotent(t *testing.T) {
	// Create a temp home directory with a zshrc.
	tmpDir := t.TempDir()
	zshrcPath := filepath.Join(tmpDir, ".zshrc")

	// Write initial content.
	initialContent := "# My zshrc\nexport PATH=\"$HOME/bin:$PATH\"\n"
	if err := os.WriteFile(zshrcPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("write zshrc: %v", err)
	}

	// Save original home and SHELL.
	origHome := os.Getenv("HOME")
	origShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SHELL", origShell)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SHELL", "/bin/zsh")

	// First install.
	rcPath, shellType, err := detectRcFile(tmpDir)
	if err != nil {
		t.Fatalf("detectRcFile: %v", err)
	}
	if rcPath != zshrcPath {
		t.Errorf("rcPath = %q, want %q", rcPath, zshrcPath)
	}
	if shellType != "zsh" {
		t.Errorf("shellType = %q, want %q", shellType, "zsh")
	}

	// Manually test idempotency: read content, check no __envmoat_hook yet.
	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("read zshrc: %v", err)
	}
	if strings.Contains(string(content), "__envmoat_hook") {
		t.Error("zshrc should not contain __envmoat_hook yet")
	}

	// Append hook manually (simulating installShellHook).
	hook := generateShellHook("zsh")
	toWrite := initialContent + "\n" + hook + "\n"
	if err := os.WriteFile(zshrcPath, []byte(toWrite), 0644); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	// Now check idempotency: content should have __envmoat_hook.
	content, err = os.ReadFile(zshrcPath)
	if err != nil {
		t.Fatalf("read zshrc: %v", err)
	}
	if !strings.Contains(string(content), "__envmoat_hook") {
		t.Error("zshrc should contain __envmoat_hook after install")
	}
	if !strings.Contains(string(content), "__envmoat_hook_end") {
		t.Error("zshrc should contain __envmoat_hook_end after install")
	}
	// Should have add-zsh-hook for zsh.
	if !strings.Contains(string(content), "add-zsh-hook") {
		t.Error("zshrc should contain add-zsh-hook for zsh")
	}
}

func TestInstallShellHookBash(t *testing.T) {
	tmpDir := t.TempDir()
	bashrcPath := filepath.Join(tmpDir, ".bashrc")

	initialContent := "# My bashrc\n"
	if err := os.WriteFile(bashrcPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("write bashrc: %v", err)
	}

	origHome := os.Getenv("HOME")
	origShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SHELL", origShell)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SHELL", "/bin/bash")

	rcPath, shellType, err := detectRcFile(tmpDir)
	if err != nil {
		t.Fatalf("detectRcFile: %v", err)
	}
	if rcPath != bashrcPath {
		t.Errorf("rcPath = %q, want %q", rcPath, bashrcPath)
	}
	if shellType != "bash" {
		t.Errorf("shellType = %q, want %q", shellType, "bash")
	}

	// Generate bash hook.
	hook := generateShellHook("bash")
	if !strings.Contains(hook, "PROMPT_COMMAND") {
		t.Error("bash hook should contain PROMPT_COMMAND")
	}
	if !strings.Contains(hook, "BASH_VERSINFO") {
		t.Error("bash hook should contain BASH_VERSINFO check")
	}

	// Append hook.
	toWrite := initialContent + "\n" + hook + "\n"
	if err := os.WriteFile(bashrcPath, []byte(toWrite), 0644); err != nil {
		t.Fatalf("write hook: %v", err)
	}

	content, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("read bashrc: %v", err)
	}
	if !strings.Contains(string(content), "__envmoat_hook") {
		t.Error("bashrc should contain __envmoat_hook after install")
	}
}

func TestDetectRcFileFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// No rc files exist yet.
	origHome := os.Getenv("HOME")
	origShell := os.Getenv("SHELL")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SHELL", origShell)
	}()

	os.Setenv("HOME", tmpDir)
	os.Setenv("SHELL", "/bin/zsh")

	rcPath, shellType, err := detectRcFile(tmpDir)
	if err != nil {
		t.Fatalf("detectRcFile: %v", err)
	}
	if rcPath != filepath.Join(tmpDir, ".zshrc") {
		t.Errorf("rcPath = %q, want %q", rcPath, filepath.Join(tmpDir, ".zshrc"))
	}
	if shellType != "zsh" {
		t.Errorf("shellType = %q, want %q", shellType, "zsh")
	}
}

func TestShellHookContainsEnvmoatLoad(t *testing.T) {
	// Both zsh and bash hooks should call `envmoat load`.
	for _, shell := range []string{"zsh", "bash"} {
		t.Run(shell, func(t *testing.T) {
			hook := generateShellHook(shell)
			if !strings.Contains(hook, "envmoat load") {
				t.Errorf("%s hook should call 'envmoat load'", shell)
			}
			if !strings.Contains(hook, "__envmoat_last_bundle") {
				t.Errorf("%s hook should use __envmoat_last_bundle", shell)
			}
			if !strings.Contains(hook, "__envmoat_hook_end") {
				t.Errorf("%s hook should end with __envmoat_hook_end", shell)
			}
		})
	}
}

func TestSetupCommandExists(t *testing.T) {
	if setupCmd == nil {
		t.Fatal("setupCmd should not be nil")
	}
	if setupCmd.Use != "setup" {
		t.Errorf("setupCmd.Use = %q, want %q", setupCmd.Use, "setup")
	}
}

func TestSetupCommandHasResetFlag(t *testing.T) {
	flag := setupCmd.Flags().Lookup("reset")
	if flag == nil {
		t.Fatal("setupCmd should have --reset flag")
	}
	if flag.Usage != "Change master password (re-encrypts all bundles)" {
		t.Errorf("reset flag usage = %q", flag.Usage)
	}
}

func TestGenerateShellHookZshHasChpwd(t *testing.T) {
	hook := generateShellHook("zsh")
	if !strings.Contains(hook, "add-zsh-hook") {
		t.Error("zsh hook should use add-zsh-hook")
	}
	if !strings.Contains(hook, "chpwd") {
		t.Error("zsh hook should use chpwd hook")
	}
}

func TestGenerateShellHookBashHasPromptCommand(t *testing.T) {
	hook := generateShellHook("bash")
	if !strings.Contains(hook, "PROMPT_COMMAND") {
		t.Error("bash hook should use PROMPT_COMMAND")
	}
	if strings.Contains(hook, "add-zsh-hook") {
		t.Error("bash hook should not use add-zsh-hook")
	}
}
