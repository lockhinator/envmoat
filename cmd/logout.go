package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/session"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the active session cache",
	Long: `Delete the cached LUK from the platform keyring.

After logout, the next envmoat command will prompt for your master password
(or Touch ID on macOS).

Example:
  envmoat logout`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, _ []string) error {
	sess := session.NewSession(keyringBackend)

	// Check if a session exists before clearing.
	if !sess.Exists() {
		fmt.Fprintln(os.Stderr, "No active session.")
		return nil
	}

	if err := sess.Clear(); err != nil {
		return fmt.Errorf("clear session: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Session cleared. Next command will prompt for Touch ID.")
	return nil
}
