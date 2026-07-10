package backend

// KeyringBackend defines the interface for platform-specific secret storage.
type KeyringBackend interface {
	// StoreLUK stores the Local Unwrap Key in the platform's secure store.
	StoreLUK(key []byte) error

	// GetLUK retrieves the Local Unwrap Key from the platform's secure store.
	// Returns ErrNotAvailable if no key is stored.
	GetLUK() ([]byte, error)

	// DeleteLUK removes the Local Unwrap Key from the platform's secure store.
	DeleteLUK() error
}

// ClipboardBackend defines the interface for platform-specific clipboard operations.
type ClipboardBackend interface {
	// Copy writes the given text to the system clipboard.
	Copy(text string) error
}

// NewKeyringBackend returns the appropriate KeyringBackend for the current platform.
// Implementation is in darwin_stub.go or linux_stub.go based on build tags.

// NewClipboardBackend returns the appropriate ClipboardBackend for the current platform.
// Implementation is in darwin_stub.go or linux_stub.go based on build tags.