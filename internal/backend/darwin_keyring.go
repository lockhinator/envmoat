//go:build darwin

package backend

import (
	"encoding/base64"
	"errors"

	"github.com/zalando/go-keyring"
)

// darwinKeyringBackend stores the LUK in the macOS Keychain via zalando/go-keyring.
type darwinKeyringBackend struct{}

// StoreLUK stores the Local Unwrap Key in the macOS Keychain.
func (d *darwinKeyringBackend) StoreLUK(key []byte) error {
	return keyring.Set(kcService, kcKey, base64.StdEncoding.EncodeToString(key))
}

// GetLUK retrieves the Local Unwrap Key from the macOS Keychain.
// Returns ErrNotAvailable if no key is stored.
func (d *darwinKeyringBackend) GetLUK() ([]byte, error) {
	val, err := keyring.Get(kcService, kcKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNotAvailable
		}
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// DeleteLUK removes the Local Unwrap Key from the macOS Keychain.
func (d *darwinKeyringBackend) DeleteLUK() error {
	return keyring.Delete(kcService, kcKey)
}

func NewKeyringBackend() KeyringBackend {
	return &darwinKeyringBackend{}
}
