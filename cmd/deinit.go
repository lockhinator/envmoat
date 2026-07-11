package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

var (
	deinitSkipConfirm bool
)

var deinitCmd = &cobra.Command{
	Use:   "deinit <project-root>",
	Short: "Remove envmoat from a project",
	Long: `Remove the .envmoat marker file and its associated bundle from the store.
Also removes the mapping from index.json.

Use -y to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runDeinit,
}

func init() {
	rootCmd.AddCommand(deinitCmd)
	deinitCmd.Flags().BoolVarP(&deinitSkipConfirm, "yes", "y", false, "Skip confirmation prompt")
}

func runDeinit(cmd *cobra.Command, args []string) error {
	dir := args[0]

	result, err := resolver.Resolve(dir)
	if err != nil {
		return fmt.Errorf("resolve marker: %w", err)
	}
	if result == nil {
		return fmt.Errorf("no .envmoat marker found at %s", dir)
	}

	// Ensure marker is in the specified directory.
	if result.MarkerDir != filepath.Clean(dir) {
		return fmt.Errorf("marker found at %s, not at %s", result.MarkerDir, dir)
	}

	if result.Marker == resolver.MarkerDisabled {
		markerPath := filepath.Join(dir, resolver.MarkerName)
		if !deinitSkipConfirm {
			if !cmdutil.Confirm(fmt.Sprintf("Remove .envmoat marker from %s?", dir)) {
				fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
				return nil
			}
		}
		if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove marker: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: removed marker from %s\n", dir)
		return nil
	}

	s, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	var bundleFilename string
	switch result.Marker {
	case resolver.MarkerProfile:
		bundleFilename, _ = s.GetProfileBundle(result.ProfileName)
	case resolver.MarkerDefault:
		bundleFilename, _ = s.GetAutoBundle(result.MarkerDir)
	default:
		return fmt.Errorf("unknown marker type")
	}

	if bundleFilename == "" {
		return fmt.Errorf("no bundle found for %s", result.MarkerDir)
	}

	if !deinitSkipConfirm {
		if !cmdutil.Confirm(fmt.Sprintf("Remove envmoat from %s?", dir)) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	// Remove marker file.
	markerPath := filepath.Join(dir, resolver.MarkerName)
	if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove marker: %w", err)
	}

	// Remove from index.
	switch result.Marker {
	case resolver.MarkerDefault:
		if err := s.RemoveAutoMapping(result.MarkerDir); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: remove auto mapping: %v\n", err)
		}
	case resolver.MarkerProfile:
		if err := s.RemoveProfileMapping(result.ProfileName); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: remove profile mapping: %v\n", err)
		}
	}

	// Delete bundle.
	if err := s.DeleteBundle(bundleFilename); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: delete bundle: %v\n", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: deactivated envmoat for %s\n", dir)
	return nil
}
