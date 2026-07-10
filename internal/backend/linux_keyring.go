//go:build linux

package backend

import (
	"encoding/base64"
	"errors"

	"github.com/zalando/go-keyring"
)

// linuxKeyringBackend stores the LUK in the Linux Secret Service via zalando/go-keyring.
type linuxKeyringBackend struct{}

// StoreLUK stores the Local Unwrap Key in the Linux Secret Service.
func (l *linuxKeyringBackend) StoreLUK(key []byte) error {
	return keyring.Set(kcService, kcKey, base64.StdEncoding.EncodeToString(key))
}

// GetLUK retrieves the Local Unwrap Key from the Linux Secret Service.
// Returns ErrNotAvailable if no key is stored or no keyring is available.
func (l *linuxKeyringBackend) GetLUK() ([]byte, error) {
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

// DeleteLUK removes the Local Unwrap Key from the Linux Secret Service.
func (l *linuxKeyringBackend) DeleteLUK() error {
	return keyring.Delete(kcService, kcKey)
}

func NewKeyringBackend() KeyringBackend {
	return &linuxKeyringBackend{}
}
