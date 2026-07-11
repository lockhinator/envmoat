package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "envmoat" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "envmoat")
	}
}

func TestRootCommandVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd --version failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "envmoat version") {
		t.Errorf("expected version output to contain 'envmoat version', got: %q", output)
	}
}

func TestRootCommandWelcome(t *testing.T) {
	// Just verify the command runs without error and has the right structure
	rootCmd.SetArgs([]string{})
	
	// Check that Run function is set
	if rootCmd.Run == nil {
		t.Fatal("rootCmd.Run should not be nil")
	}
	
	// Execute should not fail
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd (no args) failed: %v", err)
	}
}

func TestSubcommandsExist(t *testing.T) {
	expectedCmds := []string{
		"setup",
		"init",
		"set",
		"get",
		"list",
		"load",
		"remove",
		"deinit",
		"verify",
	}

	for _, name := range expectedCmds {
		cmd, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Errorf("subcommand %q not found: %v", name, err)
			continue
		}
		if cmd.Name() != name {
			t.Errorf("subcommand %q has Name = %q", name, cmd.Name())
		}
	}
}

func TestSubcommandsReachable(t *testing.T) {
	// Test that each subcommand is registered and can be found.
	// Commands that require interactive input or specific context may fail,
	// but they should not return "unknown command" errors.
	expectedCmds := []string{
		"setup",
		"init",
		"set",
		"get",
		"list",
		"load",
		"remove",
		"deinit",
		"verify",
	}

	for _, name := range expectedCmds {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{name})
			if err != nil {
				t.Fatalf("subcommand %q not found: %v", name, err)
			}
			if cmd == nil {
				t.Fatalf("subcommand %q returned nil command", name)
			}
		})
	}
}

func TestSubcommandHelp(t *testing.T) {
	// Test that --help works on subcommands
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"setup", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("setup --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Create a master password") {
		t.Errorf("expected setup help to contain 'Create a master password', got: %q", output)
	}
}