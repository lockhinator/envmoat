package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/cmdutil"
)

var editCmd = &cobra.Command{
	Use:   "edit <KEY>",
	Short: "Edit a secret in-place using $EDITOR",
	Long: `Open a secret's value in your $EDITOR for in-place editing.

Creates a temporary file in ~/.envmoat/ (not /tmp) so the file survives
if the editor crashes. The temp file is deleted on exit regardless of outcome.

If the editor exits with a non-zero code, no changes are made and no error
is returned — this allows cancelling by simply closing the editor.

Examples:
  envmoat edit API_KEY
  EDITOR=nano envmoat edit DB_PASS`,
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmdutil.Errorf("provide a KEY argument", "usage: envmoat edit <KEY>")
		return fmt.Errorf("no key specified")
	}

	key := args[0]
	bundle, err := resolveBundle()
	if err != nil {
		cmdutil.Errorf("", "%v", err)
		return err
	}

	// Load existing secrets.
	plaintext, err := bundle.Store.ReadBundle(bundle.BundleFile, bundle.DEK)
	if err != nil {
		cmdutil.Errorf("", "read bundle: %v", err)
		return err
	}

	secrets := make(map[string]string)
	if len(plaintext) > 0 {
		if err := json.Unmarshal(plaintext, &secrets); err != nil {
			cmdutil.Errorf("", "parse bundle: %v", err)
			return err
		}
	}
	if secrets == nil {
		secrets = make(map[string]string)
	}

	// Check if key exists.
	currentValue, exists := secrets[key]
	if !exists {
		cmdutil.Errorf("key %q not found in bundle", "secret not found: %q", key)
		return fmt.Errorf("secret not found: %q", key)
	}

	// Create temp file in ~/.envmoat/ (not /tmp).
	envmoatDir := filepath.Dir(bundle.Store.ConfigPath)
	tempFile, err := os.CreateTemp(envmoatDir, "edit-*")
	if err != nil {
		cmdutil.Errorf("", "create temp file: %v", err)
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Write current value to temp file.
	if _, err := tempFile.WriteString(currentValue); err != nil {
		tempFile.Close()
		cmdutil.Errorf("", "write temp file: %v", err)
		return err
	}
	tempFile.Close()

	// Open $EDITOR (default to "vim" if not set).
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmdExec := exec.Command(editor, tempPath)
	cmdExec.Stdin = os.Stdin
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr

	if err := cmdExec.Run(); err != nil {
		// Non-zero exit: user cancelled. No changes, no error.
		return nil
	}

	// Read new value from temp file.
	newValue, err := os.ReadFile(tempPath)
	if err != nil {
		cmdutil.Errorf("", "read edited value: %v", err)
		return err
	}

	// Update secrets map and write bundle atomically.
	secrets[key] = string(newValue)
	newPlaintext, _ := json.Marshal(secrets)
	if err := bundle.Store.WriteBundle(bundle.BundleFile, newPlaintext, bundle.DEK); err != nil {
		cmdutil.Errorf("", "write bundle: %v", err)
		return err
	}

	fmt.Printf("Updated %s\n", key)
	return nil
}
