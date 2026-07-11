// Package auth manages the Local Unwrap Key (LUK) for envmoat.
package auth

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lockinator/envmoat/internal/crypto"
	"github.com/lockinator/envmoat/internal/store"
)

const lukFileName = ".luk"

// EncodeLUK returns the hex-encoded representation of a LUK.
func EncodeLUK(luk []byte) string {
	return hex.EncodeToString(luk)
}

// GetLUK retrieves the Local Unwrap Key.
// Priority: ENVMOAT_LUK env var > ~/.envmoat/.luk file
// Returns nil, nil if no LUK is available (session expired / not set up).
func GetLUK(storePath string) ([]byte, error) {
	// Check env var first (testing / scripting).
	if envLUK := os.Getenv("ENVMOAT_LUK"); envLUK != "" {
		return hex.DecodeString(envLUK)
	}

	// Read from file.
	lukPath := filepath.Join(storePath, lukFileName)
	data, err := os.ReadFile(lukPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read LUK: %w", err)
	}

	return hex.DecodeString(string(data))
}

// SetLUK stores the LUK hex-encoded in the store directory.
// Creates the directory if it doesn't exist.
func SetLUK(storePath string, luk []byte) error {
	if err := os.MkdirAll(storePath, store.DirPerm); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}

	lukPath := filepath.Join(storePath, lukFileName)
	return os.WriteFile(lukPath, []byte(hex.EncodeToString(luk)), store.FilePerm)
}

// DeriveDEK derives a per-bundle Data Encryption Key from the LUK.
func DeriveDEK(luk []byte, bundleFilename string) ([]byte, error) {
	return crypto.DeriveDEK(luk, bundleFilename)
}

// HasLUK checks whether an LUK is available without reading it.
func HasLUK(storePath string) bool {
	if os.Getenv("ENVMOAT_LUK") != "" {
		return true
	}
	lukPath := filepath.Join(storePath, lukFileName)
	_, err := os.Stat(lukPath)
	return err == nil
}
