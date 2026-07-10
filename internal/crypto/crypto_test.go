package crypto

import (
	"crypto/rand"
	"testing"
)

func TestDeriveLUK(t *testing.T) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}

	luk, err := DeriveLUK("password", salt)
	if err != nil {
		t.Fatalf("DeriveLUK failed: %v", err)
	}
	if len(luk) != LUKSize {
		t.Fatalf("expected LUK size %d, got %d", LUKSize, len(luk))
	}

	// Deterministic: same password + salt → same LUK.
	luk2, err := DeriveLUK("password", salt)
	if err != nil {
		t.Fatal(err)
	}
	if string(luk) != string(luk2) {
		t.Fatal("DeriveLUK is not deterministic")
	}

	// Different password → different LUK.
	luk3, err := DeriveLUK("other", salt)
	if err != nil {
		t.Fatal(err)
	}
	if string(luk) == string(luk3) {
		t.Fatal("different passwords produced the same LUK")
	}
}

func TestDeriveDEK(t *testing.T) {
	luk := make([]byte, LUKSize)
	if _, err := rand.Read(luk); err != nil {
		t.Fatal(err)
	}

	dek, err := DeriveDEK(luk, "test.enc")
	if err != nil {
		t.Fatalf("DeriveDEK failed: %v", err)
	}
	if len(dek) != DEKSize {
		t.Fatalf("expected DEK size %d, got %d", DEKSize, len(dek))
	}

	// Deterministic: same LUK + filename → same DEK.
	dek2, err := DeriveDEK(luk, "test.enc")
	if err != nil {
		t.Fatal(err)
	}
	if string(dek) != string(dek2) {
		t.Fatal("DeriveDEK is not deterministic")
	}

	// Different filename → different DEK.
	dek3, err := DeriveDEK(luk, "other.enc")
	if err != nil {
		t.Fatal(err)
	}
	if string(dek) == string(dek3) {
		t.Fatal("different filenames produced the same DEK")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	dek := make([]byte, DEKSize)
	if _, err := rand.Read(dek); err != nil {
		t.Fatal(err)
	}

	plaintext := []byte(`{"_meta":{"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"},"API_KEY":"sk-123","DB_PASS":"p@ss"}`)

	ciphertext, err := Encrypt(plaintext, dek)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Ciphertext should be nonce + ciphertext + tag.
	if len(ciphertext) < GCMNonceSize+GCMAuthTagSize {
		t.Fatalf("ciphertext too short: %d bytes", len(ciphertext))
	}

	// Nonce is at the start.
	nonce := ciphertext[:GCMNonceSize]
	if allZero(nonce) {
		t.Fatal("nonce is all zeros")
	}

	// Decrypt should recover the original plaintext.
	decrypted, err := Decrypt(ciphertext, dek)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("round-trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDifferentNonces(t *testing.T) {
	dek := make([]byte, DEKSize)
	if _, err := rand.Read(dek); err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("hello")

	c1, err := Encrypt(plaintext, dek)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Encrypt(plaintext, dek)
	if err != nil {
		t.Fatal(err)
	}

	// Nonces should be different.
	if string(c1[:GCMNonceSize]) == string(c2[:GCMNonceSize]) {
		t.Fatal("two encryptions produced the same nonce")
	}

	// Full ciphertext should be different.
	if string(c1) == string(c2) {
		t.Fatal("two encryptions produced identical ciphertext")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	dek1 := make([]byte, DEKSize)
	if _, err := rand.Read(dek1); err != nil {
		t.Fatal(err)
	}
	dek2 := make([]byte, DEKSize)
	if _, err := rand.Read(dek2); err != nil {
		t.Fatal(err)
	}

	ciphertext, err := Encrypt([]byte("secret"), dek1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(ciphertext, dek2)
	if err == nil {
		t.Fatal("Decrypt with wrong key should fail")
	}
}

func TestDecryptTamperedData(t *testing.T) {
	dek := make([]byte, DEKSize)
	if _, err := rand.Read(dek); err != nil {
		t.Fatal(err)
	}

	ciphertext, err := Encrypt([]byte("secret"), dek)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with a byte in the ciphertext.
	ciphertext[20] ^= 0xff

	_, err = Decrypt(ciphertext, dek)
	if err == nil {
		t.Fatal("Decrypt tampered data should fail")
	}
}

func TestDecryptShortData(t *testing.T) {
	dek := make([]byte, DEKSize)
	if _, err := rand.Read(dek); err != nil {
		t.Fatal(err)
	}

	_, err := Decrypt([]byte("short"), dek)
	if err == nil {
		t.Fatal("Decrypt short data should fail")
	}
}

func TestFullPipeline(t *testing.T) {
	// Simulate the full key derivation + encrypt/decrypt pipeline.
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}

	password := "my-master-password"
	bundleFilename := "auto-myproject.enc"

	luk, err := DeriveLUK(password, salt)
	if err != nil {
		t.Fatalf("DeriveLUK: %v", err)
	}

	dek, err := DeriveDEK(luk, bundleFilename)
	if err != nil {
		t.Fatalf("DeriveDEK: %v", err)
	}

	plaintext := []byte(`{"_meta":{"created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"},"API_KEY":"sk-abc123"}`)

	ciphertext, err := Encrypt(plaintext, dek)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, dek)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("pipeline round-trip failed")
	}
}

func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
