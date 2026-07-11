package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/session"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the active session and require re-authentication",
	Long: `Delete the cached LUK from the platform keyring so that the next
envmoat command will prompt for your master password (Touch ID).

Example:
  envmoat logout`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	sess := session.NewSession(keyringBackend)

	if !sess.Exists() {
		fmt.Println("No active session.")
		return nil
	}

	if err := sess.Clear(); err != nil {
		return fmt.Errorf("clear session: %w", err)
	}

	fmt.Println("Session cleared. Next command will prompt for Touch ID.")
	return nil
}
