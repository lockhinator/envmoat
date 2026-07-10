# SM-gtv.7 Fix: Encryption + Storage

## Changes Made

### Blocker: Fix `testStore()` helper (`store_test.go`)
- Replaced `s.indexLock = nil` with `s.indexLock = flock.New(filepath.Join(tmpDir, ".index.lock"))`
- Also fixed `TestInitStore` and `TestIsInitialized` which construct `Store` directly without `indexLock`
- Added `github.com/gofrs/flock` import to test file
- Changed `os.MkdirAll(tmpDir, 0755)` to `os.Chmod(tmpDir, 0700)` to fix latent test failure (Go 1.25 `t.TempDir()` creates 0755 dirs)

### Suggestion: `f.Sync()` before `f.Close()` (`store.go:atomicWrite`)
- Added `f.Sync()` call before `f.Close()` in `atomicWrite` for crash durability

### Suggestion: `os.Lstat` instead of `os.Stat` (`store.go`)
- Changed `checkDirPerm` and `checkFilePerm` to use `os.Lstat` to avoid following symlinks in permission checks

### Suggestion: Wrap flock errors (`store.go`)
- Changed `LoadIndex` and `SaveIndex` from `errors.New("failed to acquire index lock")` to `fmt.Errorf("acquire index lock: %w", err)`

### Suggestion: Validate bundle filename (`store.go`)
- Added `validateBundleFilename()` helper that rejects empty names, path separators (`/`, `\`), and `..` components
- Added validation call at the start of `WriteBundle`, `ReadBundle`, and `DeleteBundle`
- Added `TestValidateBundleFilename` test with 7 test cases

## Changed Files
- `internal/store/store.go` — 5 fixes (sync, lstat, flock wrapping, filename validation)
- `internal/store/store_test.go` — blocker fix + new test + import

## Validation
- `go build ./...` — PASS
- `go vet ./...` — PASS
- `go test ./internal/crypto/ ./internal/store/ -v -count=1` — ALL 16 tests PASS

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Only the 5 specified review findings were addressed. No new features, refactoring, or scope changes. Added 1 test (TestValidateBundleFilename) to cover the new filename validation, and 1 latent test environment fix (os.Chmod 0700) required because Go 1.25 t.TempDir() creates 0755 dirs."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "All changed files listed, all commands run with output captured, all 16 tests passing, no staged files."
    }
  ],
  "changedFiles": [
    "internal/store/store.go",
    "internal/store/store_test.go"
  ],
  "testsAddedOrUpdated": [
    "internal/store/store_test.go (TestValidateBundleFilename added; testStore/TestInitStore/TestIsInitialized updated)"
  ],
  "commandsRun": [
    {
      "command": "go build ./...",
      "result": "passed",
      "summary": "Clean build, no errors"
    },
    {
      "command": "go vet ./...",
      "result": "passed",
      "summary": "No vet issues"
    },
    {
      "command": "go test ./internal/crypto/ ./internal/store/ -v -count=1",
      "result": "passed",
      "summary": "16/16 tests pass (8 crypto, 8 store)"
    }
  ],
  "validationOutput": [
    "go build: OK",
    "go vet: OK",
    "go test: 16 PASS, 0 FAIL"
  ],
  "residualRisks": [
    "None. All fixes are narrow and verified."
  ],
  "noStagedFiles": true,
  "diffSummary": "5 review findings fixed: (1) nil indexLock panic in tests, (2) f.Sync() for crash durability, (3) os.Lstat for symlink safety, (4) wrapped flock errors, (5) bundle filename path traversal validation. Added 1 new test function.",
  "reviewFindings": [
    "blocker: store_test.go:12 — FIXED: indexLock = nil replaced with flock.New()",
    "suggestion: store.go:345 — FIXED: f.Sync() added before f.Close()",
    "suggestion: store.go:315 — FIXED: os.Stat replaced with os.Lstat",
    "suggestion: store.go:158 — FIXED: flock errors wrapped with fmt.Errorf",
    "suggestion: store.go:202 — FIXED: validateBundleFilename added to Write/Read/DeleteBundle"
  ],
  "manualNotes": "TestValidatePermissions had a latent failure (Go 1.25 t.TempDir creates 0755 dirs, test expects 0700). Fixed with os.Chmod in testStore. This was pre-existing and not caused by the os.Lstat change (os.Stat would have returned the same result for a regular directory)."
}
```
