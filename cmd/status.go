package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/session"
	"github.com/lockinator/envmoat/internal/store"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show envmoat session and bundle status",
	Long: `Show current envmoat status including bundle, profile, session TTL,
and keychain state.

Example:
  envmoat status`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Resolve bundle info (without requiring decryption).
	bundleName := "none"
	profileName := "none"

	s, err := store.NewStore()
	if err == nil && s.IsInitialized() {
		result, resolveErr := resolver.ResolveFromPWD()
		if resolveErr == nil && result != nil {
			switch result.Marker {
			case resolver.MarkerDefault:
				bundleName, _ = s.GetAutoBundle(result.MarkerDir)
				if bundleName == "" {
					bundleName = "none"
				}
			case resolver.MarkerProfile:
				profileName = result.ProfileName
				bundleName, _ = s.GetProfileBundle(profileName)
				if bundleName == "" {
					bundleName = "none"
				}
			default:
				bundleName = "none"
			}
		} else {
			bundleName = "none"
		}
	} else {
		bundleName = "none"
	}

	// Session state.
	sessionStatus := "inactive"
	sess := session.NewSession(keyringBackend)
	if sess.Exists() {
		// Try to get remaining TTL by reading the cache item directly.
		data, err := keyringBackend.GetLUK()
		if err == nil && len(data) > 0 {
			sessionStatus = formatSessionStatus(data, sess)
		} else {
			sessionStatus = "active"
		}
	}

	// Keychain state: check if LUK exists in the keyring (both protected item and cache).
	_, protectedErr := keyringBackend.GetLUK()
	protectedExists := protectedErr == nil
	_, cacheErr := keyringBackend.GetLUK()
	cacheExists := cacheErr == nil

	// Print status.
	fmt.Printf("Bundle: %s\n", bundleName)
	fmt.Printf("Profile: %s\n", profileName)
	fmt.Printf("Session: %s\n", sessionStatus)
	fmt.Printf("Keychain: protected=%v cache=%v\n", protectedExists, cacheExists)
	fmt.Println("Debug: Set ENVMOAT_DEBUG=1 for verbose logging")

	return nil
}

// formatSessionStatus reads the cached session value and returns a human-readable
// status string with remaining TTL or "expired".
func formatSessionStatus(data []byte, sess *session.Session) string {
	_ = sess // used for TTL reference if needed in future
	// Try to parse as JSON sessionValue.
	var sv struct {
		Expiry int64 `json:"expiry"`
	}
	if err := json.Unmarshal(data, &sv); err == nil && sv.Expiry > 0 {
		remaining := time.Until(time.Unix(0, sv.Expiry))
		if remaining <= 0 {
			return "expired"
		}
		return fmt.Sprintf("active (%s remaining)", humanizeDuration(remaining))
	}
	return "active"
}

// humanizeDuration formats a duration into a human-readable string like "12m".
func humanizeDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
