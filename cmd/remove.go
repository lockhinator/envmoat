package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

var (
	removeSkipConfirm bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <KEY>",
	Short: "Remove a secret from the current project bundle",
	Long: `Remove a secret from the current project's envmoat bundle.

The bundle is re-encrypted and written atomically after removal.
Use -y to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&removeSkipConfirm, "yes", "y", false, "Skip confirmation prompt")
}

func runRemove(cmd *cobra.Command, args []string) error {
	key := args[0]

	result, err := resolver.ResolveFromPWD()
	if err != nil {
		return fmt.Errorf("resolve marker: %w", err)
	}
	if result == nil {
		return fmt.Errorf("no .envmoat marker found")
	}

	if result.Marker == resolver.MarkerDisabled {
		return fmt.Errorf("marker is disabled")
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

	if !removeSkipConfirm {
		if !cmdutil.Confirm(fmt.Sprintf("Remove secret %q from bundle?", key)) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	luk, err := auth.GetLUK(s.BasePath)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if luk == nil {
		return fmt.Errorf("not authenticated; run 'envmoat setup'")
	}

	dek, err := auth.DeriveDEK(luk, bundleFilename)
	if err != nil {
		return fmt.Errorf("derive DEK: %w", err)
	}

	plaintext, err := s.ReadBundle(bundleFilename, dek)
	if err != nil {
		return fmt.Errorf("read bundle: %w", err)
	}

	var bundleData map[string]json.RawMessage
	if err := json.Unmarshal(plaintext, &bundleData); err != nil {
		return fmt.Errorf("parse bundle: %w", err)
	}

	delete(bundleData, key)

	// Update _meta timestamp.
	meta := map[string]string{
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	metaRaw, _ := json.Marshal(meta)
	bundleData["_meta"] = metaRaw

	newPlaintext, _ := json.MarshalIndent(bundleData, "", "  ")

	if err := s.WriteBundle(bundleFilename, newPlaintext, dek); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: removed secret %q\n", key)
	return nil
}
