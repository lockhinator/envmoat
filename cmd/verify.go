package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/store"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify integrity of envmoat store",
	Long: `Check the integrity of the envmoat store:
- All bundles decrypt successfully
- Index references valid bundles
- Detect orphaned bundles not referenced by any mapping

Orphaned bundles are listed but not automatically deleted.
Use 'envmoat verify --cleanup' to delete orphaned bundles.`,
	Args: cobra.NoArgs,
	RunE: runVerify,
}

var verifyCleanup bool

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().BoolVar(&verifyCleanup, "cleanup", false, "Delete orphaned bundles")
}

func runVerify(cmd *cobra.Command, _ []string) error {
	s, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	if !s.IsInitialized() {
		return fmt.Errorf("store not initialized")
	}

	idx, err := s.LoadIndex()
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}

	// Collect all referenced bundle filenames.
	referenced := make(map[string]bool)
	for _, fn := range idx.Auto {
		referenced[fn] = true
	}
	for _, fn := range idx.Profiles {
		referenced[fn] = true
	}

	hasErrors := false

	// Check all referenced bundles can be decrypted.
	luk, err := auth.GetLUK(s.BasePath)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if luk != nil {
		for fn := range referenced {
			dek, err := auth.DeriveDEK(luk, fn)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: error: derive DEK for %s: %v\n", fn, err)
				hasErrors = true
				continue
			}
			_, err = s.ReadBundle(fn, dek)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: error: decrypt %s: %v\n", fn, err)
				hasErrors = true
			}
		}
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "envmoat: info: LUK not available, skipping decryption check")
	}

	// Check all referenced bundles exist on disk.
	for fn := range referenced {
		path := filepath.Join(s.BundlesPath, fn)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Fprintf(cmd.ErrOrStderr(), "envmoat: error: referenced bundle %s not found\n", fn)
			hasErrors = true
		}
	}

	// Find orphaned bundles.
	entries, err := os.ReadDir(s.BundlesPath)
	if err != nil {
		return fmt.Errorf("read bundles directory: %w", err)
	}

	var orphans []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !referenced[name] {
			orphans = append(orphans, name)
		}
	}

	if len(orphans) > 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "envmoat: orphaned bundles:")
		for _, o := range orphans {
			fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", o)
		}

		if verifyCleanup {
			if err := cleanupOrphans(s, orphans); err != nil {
				fmt.Fprintf(os.Stderr, "envmoat: error: cleanup: %v\n", err)
				hasErrors = true
			}
		}
	}

	if !hasErrors && len(orphans) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "envmoat: store is healthy")
	} else if !hasErrors {
		fmt.Fprintln(cmd.ErrOrStderr(), "envmoat: store is healthy (orphaned bundles listed above)")
	} else {
		return fmt.Errorf("store has errors")
	}
	return nil
}

func cleanupOrphans(s *store.Store, orphans []string) error {
	for _, fn := range orphans {
		if !cmdutil.Confirm(fmt.Sprintf("Delete orphaned bundle %s?", fn)) {
			continue
		}
		if err := s.DeleteBundle(fn); err != nil {
			fmt.Fprintf(os.Stderr, "envmoat: warning: delete %s: %v\n", fn, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "envmoat: deleted orphan %s\n", fn)
	}
	return nil
}
