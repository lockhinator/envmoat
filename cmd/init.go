// Package cmd implements the envmoat CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

// initCmd — create marker + auto-named bundle
var initCmd = &cobra.Command{
	Use:   "init [project-root]",
	Short: "Initialize envmoat in a project directory",
	Long: `Create a .envmoat marker file and auto-named secret bundle for the project.
Defaults to current directory. Auto-adds .envmoat to .gitignore.`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	// Determine project root.
	projectRoot := "."
	if len(args) > 0 {
		projectRoot = args[0]
	}

	// Resolve to absolute path.
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		cmdutil.Errorf("run 'envmoat init <valid-path>'", "resolve path: %v", err)
		return
	}

	// Verify directory exists.
	info, err := os.Stat(absPath)
	if err != nil {
		cmdutil.Errorf("run 'envmoat init <existing-path>'", "stat directory: %v", err)
		return
	}
	if !info.IsDir() {
		cmdutil.Errorf("run 'envmoat init <directory>'", "%s is not a directory", absPath)
		return
	}

	// Check store is initialized.
	st, err := store.NewStore()
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' first", "create store: %v", err)
		return
	}

	if !st.IsInitialized() {
		cmdutil.Errorf("run 'envmoat setup' first", "envmoat not initialized")
		return
	}

	// Check marker doesn't already exist.
	markerPath := filepath.Join(absPath, ".envmoat")
	if _, err := os.Stat(markerPath); err == nil {
		cmdutil.Errorf("run 'envmoat init' in a different directory or remove existing .envmoat", "marker already exists at %s", markerPath)
		return
	}

	// Check if this directory already has an auto mapping.
	existingBundle, ok := st.GetAutoBundle(absPath)
	if ok {
		fmt.Printf("Directory already initialized with bundle %s.\n", existingBundle)
		return
	}

	// Read existing bundles to avoid naming collisions.
	existingBundles := make(map[string]bool)
	entries, err := os.ReadDir(st.BundlesPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				existingBundles[entry.Name()] = true
			}
		}
	}

	// Generate auto-bundle name.
	bundleName := store.AutoBundleName(absPath, existingBundles)

	// Read config for LUK derivation.
	cfg, err := store.ReadConfig(st.ConfigPath)
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' first", "read config: %v", err)
		return
	}

	// Prompt for password to derive DEK.
	password, err := promptPassword()
	if err != nil {
		cmdutil.Errorf("", "read password: %v", err)
		return
	}

	// Derive LUK and then DEK.
	luk, err := crypto.DeriveLUK(password, cfg.GlobalSalt)
	if err != nil {
		cmdutil.Errorf("run 'envmoat init' again", "derive LUK: %v", err)
		return
	}

	dek, err := crypto.DeriveDEK(luk, bundleName)
	if err != nil {
		cmdutil.Errorf("run 'envmoat init' again", "derive DEK: %v", err)
		return
	}

	// Create initial empty bundle.
	emptyJSON := []byte(`{"secrets":{}}`)
	if err := st.WriteBundle(bundleName, emptyJSON, dek); err != nil {
		cmdutil.Errorf("run 'envmoat init' again", "write bundle: %v", err)
		return
	}

	// Add auto mapping to index.
	if err := st.AddAutoMapping(absPath, bundleName); err != nil {
		cmdutil.Errorf("run 'envmoat init' again", "add index mapping: %v", err)
		return
	}

	// Create .envmoat marker file (empty, 0600 permissions).
	if err := os.WriteFile(markerPath, nil, 0600); err != nil {
		cmdutil.Errorf("run 'envmoat init' again", "create marker: %v", err)
		return
	}

	// Auto-append .envmoat to .gitignore.
	gitignorePath := filepath.Join(absPath, ".gitignore")
	if err := appendToGitignore(gitignorePath, ".envmoat"); err != nil {
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
		fmt.Println("Manually add .envmoat to your .gitignore.")
	}

	fmt.Printf("Initialized envmoat in %s\n", absPath)
	fmt.Printf("Bundle: %s\n", bundleName)
	fmt.Printf("Marker: %s\n", markerPath)
	fmt.Println("Run 'envmoat set KEY VALUE' to add secrets.")
}

// appendToGitignore adds a pattern to .gitignore if not already present.
func appendToGitignore(path string, pattern string) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if pattern already exists.
	if len(content) > 0 {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line == pattern {
				return nil // Already present.
			}
		}
	}

	// Append pattern.
	var toWrite string
	if len(content) > 0 {
		if !strings.HasSuffix(string(content), "\n") {
			toWrite = string(content) + "\n"
		} else {
			toWrite = string(content)
		}
	}
	toWrite += pattern + "\n"

	return os.WriteFile(path, []byte(toWrite), 0644)
}
