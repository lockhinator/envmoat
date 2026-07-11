package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/auth"
	"github.com/lockinator/envmoat/internal/session"
	"github.com/lockinator/envmoat/internal/store"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show envmoat session and bundle status",
	Long: `Display the current envmoat status including:
- Active bundle and profile name
- Session TTL remaining
- Keychain state (protected item and cache)
- Debug mode hint

Example:
  envmoat status`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, _ []string) error {
	fmt.Println("envmoat status")

	// Try to resolve the active bundle.
	bundle, err := resolveBundle()
	if err != nil {
		// Not in a tracked directory — show partial status.
		fmt.Println("Bundle: none")
		fmt.Println("Profile: none")
		printSessionStatus(cmd)
		return nil
	}

	fmt.Printf("Bundle: %s\n", bundle.BundleFile)
	if bundle.ProfileName != "" {
		fmt.Printf("Profile: %s\n", bundle.ProfileName)
	} else {
		fmt.Println("Profile: default")
	}
	printSessionStatus(cmd)
	return nil
}

// printSessionStatus prints session and keychain diagnostics to stderr.
func printSessionStatus(c *cobra.Command) {
	sess := session.NewSession(keyringBackend)
	remaining := sess.GetRemainingTTL()
	if remaining > 0 {
		fmt.Fprintf(os.Stderr, "Session: active (%s remaining)\n", formatDuration(remaining))
	} else if sess.Exists() {
		// Exists() returns true but GetRemainingTTL returned 0 — handle gracefully.
		fmt.Fprintln(os.Stderr, "Session: active")
	} else {
		fmt.Fprintln(os.Stderr, "Session: expired")
	}

	// Keychain state.
	storePath := getStorePath()
	var protectedStr, cacheStr string
	if auth.HasLUK(storePath) {
		protectedStr = "yes"
	} else {
		protectedStr = "no"
	}
	lukData, err := keyringBackend.GetLUK()
	if err == nil && len(lukData) > 0 {
		cacheStr = "yes"
	} else {
		cacheStr = "no"
	}
	fmt.Fprintf(os.Stderr, "Keychain: protected=%s cache=%s\n", protectedStr, cacheStr)
	fmt.Fprintln(os.Stderr, "Debug: Set ENVMOAT_DEBUG=1 for verbose logging")
}

// getStorePath returns the path to the envmoat store directory.
func getStorePath() string {
	s, err := store.NewStore()
	if err != nil {
		return ""
	}
	return s.BasePath
}

// formatDuration formats a time.Duration as a human-readable string.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	if d >= 60*time.Minute {
		hours := d / time.Hour
		minutes := (d % time.Hour) / time.Minute
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", d/time.Minute)
}
