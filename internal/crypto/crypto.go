// Package crypto provides encryption primitives for envmoat.
//
// Key derivation:
//   Master Password → scrypt → 32-byte LUK
//   LUK + bundle filename → HKDF-SHA256 → 32-byte per-bundle DEK
//
// Bundle encryption:
//   AES-256-GCM(DEK, nonce, plaintext)
//   File format: [1B version=0x01][12B nonce][ciphertext][16B auth tag]
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	// FileFormatVersion is the on-disk version byte for encrypted bundles.
	FileFormatVersion = 0x01

	// LUKSize is the size of the Local Unwrap Key in bytes.
	LUKSize = 32

	// DEKSize is the size of a per-bundle Data Encryption Key in bytes.
	DEKSize = 32

	// GCMNonceSize is the nonce size for AES-GCM (12 bytes).
	GCMNonceSize = 12

	// GCMAuthTagSize is the auth tag size for AES-GCM (16 bytes).
	GCMAuthTagSize = 16

	// scryptN is the CPU/memory cost parameter (2^18 = 262144).
	scryptN = 262144

	// scryptR is the block size parameter.
	scryptR = 8

	// scryptP is the parallelization parameter.
	scryptP = 1

	// hkdfInfo is the HKDF context info string.
	hkdfInfo = "envmoat/v1/dek"
)

var (
	// ErrDecryptVersion is returned when the encrypted file has an unknown version byte.
	ErrDecryptVersion = errors.New("unknown encryption version")

	// ErrDecryptShort is returned when the ciphertext is too short to contain a valid message.
	ErrDecryptShort = errors.New("ciphertext too short")
)

// DeriveLUK derives a 32-byte Local Unwrap Key from a master password and salt using scrypt.
func DeriveLUK(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, LUKSize)
}

// DeriveDEK derives a 32-byte per-bundle Data Encryption Key from the LUK using HKDF-SHA256.
// The bundleFilename is used as the HKDF salt to ensure each bundle gets a unique key.
func DeriveDEK(luk []byte, bundleFilename string) ([]byte, error) {
	return hkdf.Key(sha256.New, luk, []byte(bundleFilename), hkdfInfo, DEKSize)
}

// Encrypt encrypts plaintext with AES-256-GCM using the given DEK.
// Returns the concatenation: [nonce || ciphertext || auth tag].
func Encrypt(plaintext []byte, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, GCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the auth tag to the ciphertext.
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Return [nonce || ciphertext || tag] — Seal already appended the tag.
	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// Decrypt decrypts data produced by Encrypt using the given DEK.
// Expected input format: [nonce || ciphertext || auth tag].
func Decrypt(data []byte, dek []byte) ([]byte, error) {
	if len(data) < GCMNonceSize+GCMAuthTagSize {
		return nil, ErrDecryptShort
	}

	nonce := data[:GCMNonceSize]
	ciphertext := data[GCMNonceSize:]

	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
