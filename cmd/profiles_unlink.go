package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/cmdutil"
)

var profilesUnlinkSkipConfirm bool

var profilesUnlinkCmd = &cobra.Command{
	Use:   "unlink <path>",
	Short: "Unlink a project directory from its profile",
	Long: `Remove the .envmoat marker file from a project directory, unlinking it
from its associated profile.

The profile and its bundle are NOT deleted — only the marker is removed.
Other directories linked to the same profile remain unaffected.`,
	Args: cobra.ExactArgs(1),
	RunE: runProfilesUnlink,
}

func init() {
	profilesCmd.AddCommand(profilesUnlinkCmd)
	profilesUnlinkCmd.Flags().BoolVarP(&profilesUnlinkSkipConfirm, "yes", "y", false, "Skip confirmation prompt")
}

func runProfilesUnlink(cmd *cobra.Command, args []string) error {
	targetPath := args[0]

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

	markerPath := filepath.Join(absPath, ".envmoat")

	// Check marker exists.
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		cmdutil.Errorf("run 'envmoat profiles link' first", "no .envmoat marker found at %s", markerPath)
		return fmt.Errorf("no marker found at %s", markerPath)
	}

	// Read marker content to get profile name.
	content, err := os.ReadFile(markerPath)
	if err != nil {
		cmdutil.Errorf("read marker failed", "read marker: %v", err)
		return fmt.Errorf("read marker: %v", err)
	}

	markerContent := strings.TrimSpace(string(content))
	var profileName string
	if strings.HasPrefix(markerContent, "profile: ") {
		profileName = strings.TrimSpace(markerContent[len("profile: "):])
	} else if markerContent == "" || markerContent == "disabled" {
		cmdutil.Errorf("marker is not a profile marker", "expected 'profile: <name>', got %q", markerContent)
		return fmt.Errorf("not a profile marker")
	}

	// Confirmation prompt.
	if !profilesUnlinkSkipConfirm {
		prompt := fmt.Sprintf("Unlink %s from profile %s? (y/N) ", absPath, profileName)
		if !cmdutil.Confirm(prompt) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove the marker file.
	if err := os.Remove(markerPath); err != nil {
		cmdutil.Errorf("remove marker failed", "unlink: %v", err)
		return fmt.Errorf("unlink: %v", err)
	}

	fmt.Printf("Unlinked %s from profile %s\n", absPath, profileName)
	return nil
}
