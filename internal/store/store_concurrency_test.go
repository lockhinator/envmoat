package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

func TestConcurrentIndexReads(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	// Seed the index with some data.
	idx := &Index{
		Version:  1,
		Profiles: map[string]string{"dev": "bundle-dev.enc"},
		Auto:     map[string]string{"/project/a": "auto-a.enc"},
	}
	if err := s.SaveIndex(idx); err != nil {
		t.Fatalf("SaveIndex: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.LoadIndex()
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent LoadIndex failed: %v", err)
	}
}

func TestConcurrentIndexWrites(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	// Each goroutine writes a distinct profile — no overlapping keys.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			idx := &Index{
				Version:  1,
				Profiles: map[string]string{"profile": fmt.Sprintf("profile-%d.enc", n)},
				Auto:     map[string]string{"auto": fmt.Sprintf("auto-%d.enc", n)},
			}
			if err := s.SaveIndex(idx); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent SaveIndex failed: %v", err)
	}

	// Final read should succeed — index file is not corrupted.
	final, err := s.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex after concurrent writes: %v", err)
	}
	if final.Version != 1 {
		t.Errorf("version = %d, want 1", final.Version)
	}
}

func TestLockContention(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	lockPath := filepath.Join(tmpDir, ".index.lock")
	lock1 := flock.New(lockPath)
	lock2 := flock.New(lockPath)

	// lock1 acquires the file lock.
	if err := lock1.Lock(); err != nil {
		t.Fatalf("lock1.Lock: %v", err)
	}
	if !lock1.Locked() {
		t.Fatal("lock1 should be locked")
	}

	// lock2.TryLock should fail because lock1 holds the lock.
	locked, err := lock2.TryLock()
	if err != nil {
		t.Fatalf("lock2.TryLock: %v", err)
	}
	if locked {
		t.Fatal("lock2 should not have acquired the lock")
	}

	// lock1 releases; lock2 should now be able to acquire.
	if err := lock1.Unlock(); err != nil {
		t.Fatalf("lock1.Unlock: %v", err)
	}

	if err := lock2.Lock(); err != nil {
		t.Fatalf("lock2.Lock after unlock: %v", err)
	}
	if !lock2.Locked() {
		t.Fatal("lock2 should be locked after lock1 released")
	}

	lock2.Unlock()
	lock1.Close()
	lock2.Close()
}

func TestLockTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	lockPath := filepath.Join(tmpDir, ".index.lock")
	lock1 := flock.New(lockPath)
	lock2 := flock.New(lockPath)

	// lock1 holds the lock.
	if err := lock1.Lock(); err != nil {
		t.Fatalf("lock1.Lock: %v", err)
	}

	// lock2.TryLockContext with a short timeout should return false with a context error.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	locked, err := lock2.TryLockContext(ctx, 10*time.Millisecond)
	if locked {
		t.Fatal("TryLockContext should not have acquired the lock")
	}
	// TryLockContext returns context.DeadlineExceeded when it times out waiting.
	if err != context.DeadlineExceeded {
		t.Fatalf("TryLockContext error = %v, want context.DeadlineExceeded", err)
	}

	lock1.Unlock()
	lock1.Close()
	lock2.Close()
}

func TestConcurrentReadWriteContention(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	const writers = 5
	const readers = 10
	var wg sync.WaitGroup
	errs := make(chan error, writers+readers)

	// Writers each save a distinct index.
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			idx := &Index{
				Version:  1,
				Profiles: map[string]string{"writer": fmt.Sprintf("w-%d.enc", n)},
				Auto:     map[string]string{},
			}
			if err := s.SaveIndex(idx); err != nil {
				errs <- err
			}
		}(i)
	}

	// Readers all load concurrently.
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.LoadIndex()
			if err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	count := 0
	for err := range errs {
		t.Errorf("concurrent operation failed: %v", err)
		count++
	}
	if count > 0 {
		t.Fatalf("%d errors during concurrent read/write", count)
	}

	// Final consistency check.
	final, err := s.LoadIndex()
	if err != nil {
		t.Fatalf("final LoadIndex: %v", err)
	}
	if final.Version != 1 {
		t.Errorf("version = %d, want 1", final.Version)
	}
}

func TestConcurrentAutoMapping(t *testing.T) {
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	s := &Store{
		BasePath:    tmpDir,
		BundlesPath: filepath.Join(tmpDir, BundlesDirName),
		ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
		IndexPath:   filepath.Join(tmpDir, IndexFileName),
		indexLock:   flock.New(filepath.Join(tmpDir, ".index.lock")),
	}
	if err := s.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	// Each goroutine adds a unique auto mapping.
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			dirPath := fmt.Sprintf("/project/dir-%d", n)
			bundle := fmt.Sprintf("auto-%d.enc", n)
			if err := s.AddAutoMapping(dirPath, bundle); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("AddAutoMapping failed: %v", err)
	}

	// Index should still be readable and valid.
	idx, err := s.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex after concurrent AddAutoMapping: %v", err)
	}
	if idx.Version != 1 {
		t.Errorf("version = %d, want 1", idx.Version)
	}
	// At least some mappings should exist (exact count depends on race, but file must be valid).
	if len(idx.Auto) == 0 {
		t.Error("expected at least one auto mapping to survive")
	}
}

func TestFlockSurvivesMultipleStoreInstances(t *testing.T) {
	// Two separate Store instances sharing the same lock file should serialize.
	tmpDir := t.TempDir()
	os.Chmod(tmpDir, 0700)
	lockPath := filepath.Join(tmpDir, ".index.lock")

	createStore := func() *Store {
		return &Store{
			BasePath:    tmpDir,
			BundlesPath: filepath.Join(tmpDir, BundlesDirName),
			ConfigPath:  filepath.Join(tmpDir, ConfigFileName),
			IndexPath:   filepath.Join(tmpDir, IndexFileName),
			indexLock:   flock.New(lockPath),
		}
	}

	s1 := createStore()
	if err := s1.InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	s2 := createStore()

	const iterations = 20
	var wg sync.WaitGroup
	errs := make(chan error, iterations*2)

	for i := 0; i < iterations; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			idx := &Index{
				Version:  1,
				Profiles: map[string]string{"s1": fmt.Sprintf("bundle-%d.enc", n)},
				Auto:     map[string]string{},
			}
			if err := s1.SaveIndex(idx); err != nil {
				errs <- err
			}
		}(i)
		go func(n int) {
			defer wg.Done()
			idx := &Index{
				Version:  1,
				Profiles: map[string]string{"s2": fmt.Sprintf("bundle-%d.enc", n)},
				Auto:     map[string]string{},
			}
			if err := s2.SaveIndex(idx); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	count := 0
	for err := range errs {
		t.Errorf("concurrent write from two stores failed: %v", err)
		count++
	}
	if count > 0 {
		t.Fatalf("%d errors from two-store concurrent writes", count)
	}

	// Final state should be valid JSON.
	final, err := s1.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex after two-store contention: %v", err)
	}
	if final.Version != 1 {
		t.Errorf("version = %d, want 1", final.Version)
	}
}
