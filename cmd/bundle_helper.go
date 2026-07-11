// Package cmd provides the CLI commands for envmoat.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/lockinator/envmoat/internal/backend"
	"github.com/lockinator/envmoat/internal/cmdutil"
	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/resolver"
	"github.com/lockinator/envmoat/internal/store"
	"golang.org/x/term"
)

// keyringBackend is the platform-specific keyring backend.
// Can be overridden in tests with a mock.
var keyringBackend backend.KeyringBackend = backend.NewKeyringBackend()

// BundleContext holds the resolved bundle and its decryption key.
type BundleContext struct {
	Store        *store.Store
	BundleFile   string
	DEK          []byte
	ProfileName  string
	MarkerDir    string
	Marker       resolver.MarkerContent
}

// resolveBundle resolves the active bundle from the current working directory.
// It walks up from PWD for a marker, looks up the bundle in index.json,
// authenticates via keyring, and returns the bundle context with DEK.
func resolveBundle() (*BundleContext, error) {
	s, err := store.NewStore()
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}

	if !s.IsInitialized() {
		return nil, fmt.Errorf("store not initialized; run 'envmoat setup'")
	}

	result, err := resolver.ResolveFromPWD()
	if err != nil {
		return nil, fmt.Errorf("resolve marker: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("not in a tracked directory; run 'envmoat init' or cd into a tracked project")
	}

	if result.Marker == resolver.MarkerDisabled {
		return nil, fmt.Errorf("directory is disabled; remove .envmoat or set it to empty/profile")
	}

	var bundleFile string
	var profileName string
	var ok bool

	switch result.Marker {
	case resolver.MarkerDefault:
		bundleFile, ok = s.GetAutoBundle(result.MarkerDir)
		if !ok {
			return nil, fmt.Errorf("no bundle found for directory %s; run 'envmoat init'", result.MarkerDir)
		}
	case resolver.MarkerProfile:
		profileName = result.ProfileName
		bundleFile, ok = s.GetProfileBundle(profileName)
		if !ok {
			return nil, fmt.Errorf("profile %q not found; run 'envmoat profiles create %s'", profileName, profileName)
		}
	default:
		return nil, fmt.Errorf("unknown marker content")
	}

	// Get LUK from keyring or prompt.
	luk, err := keyringBackend.GetLUK()
	if err != nil {
		if err == backend.ErrNotAvailable {
			// Prompt for master password.
			cfg, cfgErr := store.ReadConfig(s.ConfigPath)
			if cfgErr != nil {
				return nil, fmt.Errorf("read config: %w", cfgErr)
			}
			fmt.Fprint(os.Stderr, "Enter master password: ")
			password, readErr := readPassword()
			if readErr != nil {
				return nil, fmt.Errorf("read password: %w", readErr)
			}
			fmt.Fprintln(os.Stderr)
			luk, err = crypto.DeriveLUK(password, cfg.GlobalSalt)
			if err != nil {
				return nil, fmt.Errorf("derive LUK: %w", err)
			}
			// Store in keyring for session caching.
			if storeErr := keyringBackend.StoreLUK(luk); storeErr != nil {
				cmdutil.Debug("failed to cache LUK in keyring: %v", storeErr)
			}
		} else {
			return nil, fmt.Errorf("get LUK from keyring: %w", err)
		}
	}

	dek, err := crypto.DeriveDEK(luk, bundleFile)
	if err != nil {
		return nil, fmt.Errorf("derive DEK: %w", err)
	}

	return &BundleContext{
		Store:       s,
		BundleFile:  bundleFile,
		DEK:         dek,
		ProfileName: profileName,
		MarkerDir:   result.MarkerDir,
		Marker:      result.Marker,
	}, nil
}

// readPassword reads a password from stdin without echoing.
func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(bytePassword), "\n"), nil
}
