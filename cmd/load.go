package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Emit shell-safe export lines for the current project",
	Long: `Shell hook command: resolves the nearest .envmoat marker, decrypts
the bundle, and emits "export K='V'" lines to stdout.

First line is a comment with the bundle hash for change detection.
Errors go to stderr. Exits 0 with no output when no bundle is found.`,
	Args: cobra.NoArgs,
	RunE: runLoad,
}

func init() {
	rootCmd.AddCommand(loadCmd)
}

func runLoad(cmd *cobra.Command, _ []string) error {
	result, err := resolver.ResolveFromPWD()
	if err != nil {
		return fmt.Errorf("resolve marker: %w", err)
	}
	if result == nil {
		// No marker found — not an envmoat project.
		return nil
	}

	if result.Marker == resolver.MarkerDisabled {
		// Marker explicitly disabled.
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
		fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: no bundle found for %s\n", result.MarkerDir)
		return nil
	}

	luk, err := auth.GetLUK(s.BasePath)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if luk == nil {
		// Session expired or not authenticated — exit silently.
		return nil
	}

	dek, err := auth.DeriveDEK(luk, bundleFilename)
	if err != nil {
		return fmt.Errorf("derive DEK: %w", err)
	}

	plaintext, err := s.ReadBundle(bundleFilename, dek)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: bundle may be corrupted: %v\n", err)
		return nil
	}

	hash := sha256.Sum256(plaintext)
	hashHex := hex.EncodeToString(hash[:])

	var bundleData map[string]json.RawMessage
	if err := json.Unmarshal(plaintext, &bundleData); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: warning: invalid bundle: %v\n", err)
		return nil
	}

	secrets := make(map[string]string)
	for k, v := range bundleData {
		if k == "_meta" {
			continue
		}
		var val string
		if err := json.Unmarshal(v, &val); err != nil {
			continue
		}
		secrets[k] = val
	}

	if err := cmdutil.EmitLoadOutput(cmd.OutOrStdout(), cmd.ErrOrStderr(), hashHex, secrets); err != nil {
		return fmt.Errorf("emit output: %w", err)
	}
	return nil
}
