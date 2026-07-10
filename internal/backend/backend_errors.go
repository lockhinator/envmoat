package backend

import "errors"

// ErrNotAvailable is returned when no keyring backend is available on the platform.
var ErrNotAvailable = errors.New("keyring backend not available")

// Keychain item identifiers shared across platforms.
const (
	kcService = "envmoat"
	kcKey     = "envmoat-luk"
)
