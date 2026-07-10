// Package store provides encrypted bundle storage and index management for envmoat.
//
// Storage layout:
//
//	~/.envmoat/
//	├── config.yaml    # global settings (TTL, global salt)
//	├── bundles/
//	│   └── <id>.enc   # encrypted secret bundles
//	└── index.json     # path → bundle mapping
//
// All files are created with 0600 permissions; directories with 0700.
// Writes are atomic (temp file + rename). Index operations use flock.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
	"github.com/lockinator/envmoat/internal/crypto"
)

// validateBundleFilename ensures filename is a simple basename with no path traversal.
func validateBundleFilename(filename string) error {
	if filename == "" {
		return errors.New("bundle filename is empty")
	}
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return errors.New("bundle filename must not contain path separators")
	}
	if strings.Contains(filename, "..") {
		return errors.New("bundle filename must not contain '..' components")
	}
	return nil
}

const (
	// StoreDirName is the name of the envmoat home directory.
	StoreDirName = ".envmoat"

	// BundlesDirName is the subdirectory for encrypted bundles.
	BundlesDirName = "bundles"

	// ConfigFileName is the name of the global config file.
	ConfigFileName = "config.yaml"

	// IndexFileName is the name of the index file.
	IndexFileName = "index.json"

	// DirPerm is the permission mode for directories.
	DirPerm = 0700

	// FilePerm is the permission mode for files.
	FilePerm = 0600
)

var (
	// ErrStoreNotInitialized is returned when the store has not been set up.
	ErrStoreNotInitialized = errors.New("envmoat store not initialized; run 'envmoat setup'")

	// ErrBundleNotFound is returned when a bundle file does not exist.
	ErrBundleNotFound = errors.New("bundle not found")

	// ErrPermissionTooOpen is returned when file/dir permissions are too permissive.
	ErrPermissionTooOpen = errors.New("permissions too open; refusing to operate for security")
)

// Index represents the index.json file structure.
type Index struct {
	Version  int               `json:"version"`
	Profiles map[string]string `json:"profiles"`
	Auto     map[string]string `json:"auto"`
}

// Store provides access to the envmoat encrypted storage backend.
type Store struct {
	// BasePath is the absolute path to ~/.envmoat/.
	BasePath string

	// BundlesPath is the absolute path to ~/.envmoat/bundles/.
	BundlesPath string

	// ConfigPath is the absolute path to ~/.envmoat/config.yaml.
	ConfigPath string

	// IndexPath is the absolute path to ~/.envmoat/index.json.
	IndexPath string

	// indexLock is the file lock for index.json operations.
	indexLock *flock.Flock
}

// NewStore creates a Store pointing to the user's ~/.envmoat/ directory.
// It does not create the directory structure; use InitStore() for that.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	basePath := filepath.Join(home, StoreDirName)
	return &Store{
		BasePath:    basePath,
		BundlesPath: filepath.Join(basePath, BundlesDirName),
		ConfigPath:  filepath.Join(basePath, ConfigFileName),
		IndexPath:   filepath.Join(basePath, IndexFileName),
		indexLock:   flock.New(filepath.Join(basePath, ".index.lock")),
	}, nil
}

// InitStore creates the store directory structure and writes a default config if needed.
// It is idempotent: existing directories and config are left untouched.
func (s *Store) InitStore() error {
	// Create base directory.
	if err := os.MkdirAll(s.BasePath, DirPerm); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}

	// Create bundles directory.
	if err := os.MkdirAll(s.BundlesPath, DirPerm); err != nil {
		return fmt.Errorf("create bundles directory: %w", err)
	}

	// Write config only if it doesn't exist.
	if _, err := os.Stat(s.ConfigPath); os.IsNotExist(err) {
		cfg, err := DefaultConfig()
		if err != nil {
			return fmt.Errorf("generate default config: %w", err)
		}
		if err := WriteConfig(s.ConfigPath, cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	// Write index only if it doesn't exist.
	if _, err := os.Stat(s.IndexPath); os.IsNotExist(err) {
		idx := &Index{
			Version:  1,
			Profiles: make(map[string]string),
			Auto:     make(map[string]string),
		}
		if err := s.SaveIndex(idx); err != nil {
			return fmt.Errorf("write initial index: %w", err)
		}
	}

	return nil
}

// IsInitialized checks whether the store directory and config exist.
func (s *Store) IsInitialized() bool {
	if _, err := os.Stat(s.ConfigPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(s.IndexPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// LoadIndex reads the index.json file under a lock.
func (s *Store) LoadIndex() (*Index, error) {
	if err := s.indexLock.Lock(); err != nil {
		return nil, fmt.Errorf("acquire index lock: %w", err)
	}
	defer s.indexLock.Unlock()

	data, err := os.ReadFile(s.IndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrStoreNotInitialized
		}
		return nil, fmt.Errorf("read index: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	if idx.Profiles == nil {
		idx.Profiles = make(map[string]string)
	}
	if idx.Auto == nil {
		idx.Auto = make(map[string]string)
	}

	return &idx, nil
}

// SaveIndex writes the index.json file atomically under a lock.
func (s *Store) SaveIndex(idx *Index) error {
	if err := s.indexLock.Lock(); err != nil {
		return fmt.Errorf("acquire index lock: %w", err)
	}
	defer s.indexLock.Unlock()

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	return atomicWrite(s.IndexPath, data, FilePerm)
}

// WriteBundle encrypts plaintext JSON and writes it as an encrypted bundle file.
// File format: [1B version][12B nonce][ciphertext][16B auth tag].
// The write is atomic (temp file + rename).
func (s *Store) WriteBundle(filename string, plaintextJSON []byte, dek []byte) error {
	if err := validateBundleFilename(filename); err != nil {
		return err
	}
	encrypted, err := crypto.Encrypt(plaintextJSON, dek)
	if err != nil {
		return fmt.Errorf("encrypt bundle: %w", err)
	}

	// Prepend version byte.
	data := make([]byte, 1, 1+len(encrypted))
	data[0] = crypto.FileFormatVersion
	data = append(data, encrypted...)

	path := filepath.Join(s.BundlesPath, filename)
	return atomicWrite(path, data, FilePerm)
}

// ReadBundle reads an encrypted bundle file and decrypts it.
func (s *Store) ReadBundle(filename string, dek []byte) ([]byte, error) {
	if err := validateBundleFilename(filename); err != nil {
		return nil, err
	}
	path := filepath.Join(s.BundlesPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrBundleNotFound
		}
		return nil, fmt.Errorf("read bundle: %w", err)
	}

	if len(data) < 1 {
		return nil, errors.New("bundle file is empty")
	}

	// Check version byte.
	if data[0] != crypto.FileFormatVersion {
		return nil, fmt.Errorf("unsupported bundle version: 0x%02x", data[0])
	}

	// Decrypt the rest (nonce || ciphertext || tag).
	plaintext, err := crypto.Decrypt(data[1:], dek)
	if err != nil {
		return nil, fmt.Errorf("decrypt bundle: %w", err)
	}

	return plaintext, nil
}

// DeleteBundle removes a bundle file from the store.
func (s *Store) DeleteBundle(filename string) error {
	if err := validateBundleFilename(filename); err != nil {
		return err
	}
	path := filepath.Join(s.BundlesPath, filename)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrBundleNotFound
		}
		return fmt.Errorf("delete bundle: %w", err)
	}
	return nil
}

// AddAutoMapping adds a directory → bundle mapping to the index (auto-bundle).
func (s *Store) AddAutoMapping(dirPath, bundleFilename string) error {
	idx, err := s.LoadIndex()
	if err != nil {
		return err
	}
	idx.Auto[dirPath] = bundleFilename
	return s.SaveIndex(idx)
}

// RemoveAutoMapping removes a directory → bundle mapping from the index.
func (s *Store) RemoveAutoMapping(dirPath string) error {
	idx, err := s.LoadIndex()
	if err != nil {
		return err
	}
	delete(idx.Auto, dirPath)
	return s.SaveIndex(idx)
}

// AddProfileMapping adds a profile name → bundle mapping to the index.
func (s *Store) AddProfileMapping(profileName, bundleFilename string) error {
	idx, err := s.LoadIndex()
	if err != nil {
		return err
	}
	idx.Profiles[profileName] = bundleFilename
	return s.SaveIndex(idx)
}

// RemoveProfileMapping removes a profile name → bundle mapping from the index.
func (s *Store) RemoveProfileMapping(profileName string) error {
	idx, err := s.LoadIndex()
	if err != nil {
		return err
	}
	delete(idx.Profiles, profileName)
	return s.SaveIndex(idx)
}

// GetAutoBundle looks up a directory in the auto mapping.
func (s *Store) GetAutoBundle(dirPath string) (string, bool) {
	idx, err := s.LoadIndex()
	if err != nil {
		return "", false
	}
	bundle, ok := idx.Auto[dirPath]
	return bundle, ok
}

// GetProfileBundle looks up a profile name in the profiles mapping.
func (s *Store) GetProfileBundle(profileName string) (string, bool) {
	idx, err := s.LoadIndex()
	if err != nil {
		return "", false
	}
	bundle, ok := idx.Profiles[profileName]
	return bundle, ok
}

// ValidatePermissions checks that store files and directories have correct permissions.
func (s *Store) ValidatePermissions() error {
	// Check base directory.
	if err := checkDirPerm(s.BasePath, DirPerm); err != nil {
		return fmt.Errorf("base directory: %w", err)
	}

	// Check bundles directory.
	if err := checkDirPerm(s.BundlesPath, DirPerm); err != nil {
		return fmt.Errorf("bundles directory: %w", err)
	}

	// Check config file.
	if err := checkFilePerm(s.ConfigPath, FilePerm); err != nil {
		return fmt.Errorf("config file: %w", err)
	}

	// Check index file.
	if err := checkFilePerm(s.IndexPath, FilePerm); err != nil {
		return fmt.Errorf("index file: %w", err)
	}

	// Check all bundle files.
	entries, err := os.ReadDir(s.BundlesPath)
	if err != nil {
		return fmt.Errorf("read bundles directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := checkFilePerm(filepath.Join(s.BundlesPath, entry.Name()), FilePerm); err != nil {
			return fmt.Errorf("bundle %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// AutoBundleName generates an auto-bundle filename from a directory path.
// Format: auto-<slugified-last-dirname>.enc
// On collision, appends -<short-hash> to make it unique.
func AutoBundleName(dirPath string, existing map[string]bool) string {
	lastDir := filepath.Base(filepath.Clean(dirPath))
	slug := slugify(lastDir)
	base := fmt.Sprintf("auto-%s.enc", slug)

	if !existing[base] {
		return base
	}

	// Collision: append short hash.
	hash := shortHash(dirPath)
	return fmt.Sprintf("auto-%s-%s.enc", slug, hash)
}

func slugify(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '_' || r == '-' {
			b.WriteRune('-')
		}
	}
	s := b.String()
	// Collapse multiple dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	// Trim leading/trailing dashes.
	s = strings.Trim(s, "-")
	if s == "" {
		s = "bundle"
	}
	return s
}

// shortHash returns a short hex hash of the input string.
func shortHash(s string) string {
	h := simpleHash(s)
	return fmt.Sprintf("%08x", h)
}

// simpleHash is a fast non-cryptographic hash for collision avoidance.
func simpleHash(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return h
}

// checkDirPerm verifies a directory exists and has at least the required permissions.
func checkDirPerm(path string, required os.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}
	// Check that owner has rwx and group/others have nothing.
	if info.Mode().Perm()&^required != 0 {
		return ErrPermissionTooOpen
	}
	return nil
}

// checkFilePerm verifies a file exists and has at least the required permissions.
func checkFilePerm(path string, required os.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("not a file")
	}
	// Check that owner has rw and group/others have nothing.
	if info.Mode().Perm()&^required != 0 {
		return ErrPermissionTooOpen
	}
	return nil
}

// atomicWrite writes data to a file atomically using temp file + rename.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := f.Name()

	// Ensure cleanup on failure.
	written := false
	defer func() {
		if !written {
			os.Remove(tmpPath)
		}
	}()

	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := f.Chmod(perm); err != nil {
		f.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	written = true
	return nil
}
