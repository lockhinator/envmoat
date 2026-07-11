package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

var profilesLinkForce bool

var profilesLinkCmd = &cobra.Command{
	Use:   "link <path> <name>",
	Short: "Link a project directory to a named profile",
	Long: `Link a project directory to a named profile by creating a .envmoat marker.

The marker file will contain "profile: <name>" so that all envmoat commands
in that directory use the specified profile's bundle.

If the profile doesn't exist yet, it will be created automatically with an
empty encrypted bundle. Use --force to overwrite an existing marker file.`,
	Args: cobra.ExactArgs(2),
	RunE: runProfilesLink,
}

func init() {
	profilesCmd.AddCommand(profilesLinkCmd)
	profilesLinkCmd.Flags().BoolVarP(&profilesLinkForce, "force", "f", false, "Overwrite existing marker file")
}

func runProfilesLink(cmd *cobra.Command, args []string) error {
	targetPath := args[0]
	profileName := args[1]

	// Validate profile name.
	if !profileNameRegex.MatchString(profileName) {
		cmdutil.Errorf("use alphanumeric, underscore, and hyphen only", "invalid profile name %q", profileName)
		return fmt.Errorf("invalid profile name")
	}

	// Resolve target path to absolute.
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		cmdutil.Errorf("provide a valid path", "resolve path: %v", err)
		return fmt.Errorf("resolve path: %v", err)
	}

	// Verify directory exists.
	info, err := os.Stat(absPath)
	if err != nil {
		cmdutil.Errorf("directory does not exist", "stat %s: %v", absPath, err)
		return fmt.Errorf("stat %s: %v", absPath, err)
	}
	if !info.IsDir() {
		cmdutil.Errorf("not a directory", "%s is not a directory", absPath)
		return fmt.Errorf("%s is not a directory", absPath)
	}

	// Check store is initialized.
	st, err := store.NewStore()
	if err != nil {
		cmdutil.Errorf("run 'envmoat setup' first", "create store: %v", err)
		return fmt.Errorf("create store: %v", err)
	}
	if !st.IsInitialized() {
		cmdutil.Errorf("run 'envmoat setup' first", "envmoat not initialized")
		return fmt.Errorf("envmoat not initialized")
	}

	markerPath := filepath.Join(absPath, ".envmoat")

	// Check if marker already exists.
	if _, err := os.Stat(markerPath); err == nil {
		if !profilesLinkForce {
			cmdutil.Errorf("use --force to overwrite", "marker already exists at %s", markerPath)
			return fmt.Errorf("marker already exists")
		}
	}

	// Create profile if it doesn't exist yet.
	_, exists := st.GetProfileBundle(profileName)
	if !exists {
		if err := createProfileInline(st, profileName); err != nil {
			cmdutil.Errorf("create profile failed", "%v", err)
			return fmt.Errorf("create profile: %v", err)
		}
	}

	// Write marker file with "profile: <name>" content.
	markerContent := "profile: " + profileName + "\n"
	if err := os.WriteFile(markerPath, []byte(markerContent), 0600); err != nil {
		cmdutil.Errorf("create marker failed", "write marker: %v", err)
		return fmt.Errorf("write marker: %v", err)
	}

	// Auto-append .envmoat to .gitignore.
	gitignorePath := filepath.Join(absPath, ".gitignore")
	if err := appendToGitignore(gitignorePath, ".envmoat"); err != nil {
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
	}

	fmt.Printf("Linked %s to profile %s\n", absPath, profileName)
	return nil
}

// createProfileInline creates a new profile with an empty encrypted bundle.
// This is called from profiles link when the profile doesn't exist yet.
func createProfileInline(st *store.Store, name string) error {
	bundleFile := fmt.Sprintf("profile-%s.enc", name)

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

	// Handle collision by appending a short hash.
	if existingBundles[bundleFile] {
		hash := shortHash(name)
		bundleFile = fmt.Sprintf("profile-%s-%s.enc", name, hash)
	}

	// Read config for LUK derivation.
	cfg, err := store.ReadConfig(st.ConfigPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// Prompt for password to derive DEK.
	password := os.Getenv("ENVMOAT_TEST_PASSWORD")
	if password == "" {
		password, err = promptPassword()
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
	}

	// Derive LUK and then DEK.
	luk, err := crypto.DeriveLUK(password, cfg.GlobalSalt)
	if err != nil {
		return fmt.Errorf("derive LUK: %w", err)
	}

	dek, err := crypto.DeriveDEK(luk, bundleFile)
	if err != nil {
		return fmt.Errorf("derive DEK: %w", err)
	}

	// Create initial empty bundle.
	emptyBundle := make(map[string]string)
	emptyJSON, err := store.MarshalBundle(emptyBundle)
	if err != nil {
		return fmt.Errorf("marshal empty bundle: %w", err)
	}

	if err := st.WriteBundle(bundleFile, emptyJSON, dek); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}

	// Add profile mapping to index.
	if err := st.AddProfileMapping(name, bundleFile); err != nil {
		return fmt.Errorf("add profile mapping: %w", err)
	}

	return nil
}
