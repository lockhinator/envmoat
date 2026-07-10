# Review: SM-gtv.7 — Encryption + Storage (envmoat)

**Files reviewed:**
- `internal/crypto/crypto.go`
- `internal/crypto/crypto_test.go`
- `internal/store/store.go`
- `internal/store/store_test.go`
- `internal/store/config.go`

**Spec reference:** `PLAN.md`

---

## Findings

### BLOCKER: `TestInitStore` — nil pointer dereference in `SaveIndex`

- **Location:** `internal/store/store_test.go:28–35`, `internal/store/store.go:183`
- **Description:** `TestInitStore` constructs a `Store` struct directly without initializing `indexLock`. When `InitStore()` calls `SaveIndex()` (line 132 of `store.go`), which calls `s.indexLock.Lock()` (line 183), it panics with a nil pointer dereference. The `testStore()` helper (line 21) also sets `indexLock = nil` with the comment "We'll skip locking in tests by using direct methods," but `TestWriteReadBundle` and `TestDeleteBundle` both call `InitStore()` via `testStore()`, hitting the same path.
- **Evidence:** Test run output shows `panic: runtime error: invalid memory address or nil pointer dereference` at `store.go:183` in `SaveIndex`.
- **Recommended fix:** Initialize `s.indexLock = flock.New(filepath.Join(tmpDir, ".index.lock"))` in the test helper and in `TestInitStore`. Alternatively, add a nil guard in `SaveIndex`/`LoadIndex` (less ideal — hiding nil locks masks real bugs).

### BUG: HKDF info string mismatch between spec and implementation

- **Location:** `internal/crypto/crypto.go:50`, `PLAN.md:48`
- **Description:** The spec defines `info="secrets-manager/v1/dek"` but the code uses `hkdfInfo = "envmoat/v1/dek"`. This is a project naming inconsistency: the spec still uses the old "secrets-manager" name while the codebase uses "envmoat".
- **Impact:** If the spec is the source of truth and the project name is "secrets-manager", the code is wrong. If the project was renamed to "envmoat", the spec is stale. Either way, one side must be updated to match the other.
- **Recommended fix:** Decide the canonical project name. If "envmoat" is correct, update `PLAN.md` line 48. If "secrets-manager" is correct, update `crypto.go` line 50.

### BUG: Spec claims 32-byte auth tag; AES-256-GCM produces 16 bytes

- **Location:** `PLAN.md:61`, `internal/crypto/crypto.go:41`
- **Description:** The spec states `[32B auth tag]` in the file format. AES-256-GCM always produces a 16-byte authentication tag. The code correctly uses `GCMAuthTagSize = 16`. The spec is wrong.
- **Impact:** Misleading documentation. No runtime impact since the code is correct.
- **Recommended fix:** Update `PLAN.md` line 61 from `[32B auth tag]` to `[16B auth tag]`.

### BUG: `testStore()` helper creates broken Store — cascading test failures

- **Location:** `internal/store/store_test.go:12–24`
- **Description:** `testStore()` sets `indexLock = nil`. Any test using this helper that calls `InitStore()`, `LoadIndex()`, `SaveIndex()`, `AddAutoMapping()`, `RemoveAutoMapping()`, `AddProfileMapping()`, `RemoveProfileMapping()`, `GetAutoBundle()`, or `GetProfileBundle()` will panic. Currently `TestWriteReadBundle`, `TestDeleteBundle`, `TestValidatePermissions`, and `TestConfigRoundTrip` use this helper. `TestWriteReadBundle` and `TestDeleteBundle` call `InitStore()` which calls `SaveIndex()` — they would panic if `TestInitStore` didn't fail first.
- **Recommended fix:** Initialize `indexLock` in `testStore()`:
  ```go
  s.indexLock = flock.New(filepath.Join(tmpDir, ".index.lock"))
  ```

### SUGGESTION: No `fsync` before rename in `atomicWrite`

- **Location:** `internal/store/store.go:345–371`
- **Description:** `atomicWrite()` calls `f.Close()` then `os.Rename()`. On a crash between `Write` and `Rename`, the data may be in the page cache but not on disk. The rename itself is atomic on the same filesystem, but the data may not be durable yet.
- **Impact:** In the event of a power loss or kernel crash during a write, the target file could be left with stale data or the temp file could be orphaned. For a secrets manager this is a low-probability but real risk.
- **Recommended fix:** Call `f.Sync()` before `f.Close()` to ensure data hits disk before the rename.

### SUGGESTION: `ValidatePermissions` — `os.Stat` vs `os.Lstat`

- **Location:** `internal/store/store.go:315–334`
- **Description:** `checkDirPerm` and `checkFilePerm` use `os.Stat`, which follows symlinks. If a symlink points outside the store directory with permissive permissions, the check would pass on the symlink target rather than the symlink itself. For a security-critical permission check, `os.Lstat` is more appropriate.
- **Impact:** Low. Unlikely in practice since the store directory shouldn't contain symlinks, but the check is meant to be a security guard.
- **Recommended fix:** Use `os.Lstat` instead of `os.Stat` in `checkDirPerm` and `checkFilePerm`.

### SUGGESTION: Missing test for `ValidatePermissions` with bad permissions

- **Location:** `internal/store/store_test.go:147–154`
- **Description:** `TestValidatePermissions` only tests the happy path (permissions are correct after init). There is no test that verifies the function actually catches too-open permissions.
- **Recommended fix:** Add a test that makes a file 0644 and verifies `ValidatePermissions` returns `ErrPermissionTooOpen`.

### SUGGESTION: Missing test for `atomicWrite` directly

- **Location:** `internal/store/store_test.go`
- **Description:** `atomicWrite` is tested indirectly through `WriteBundle` and `InitStore`, but there is no direct test verifying the atomic write behavior (temp file cleanup on failure, permissions, rename).
- **Recommended fix:** Add a unit test for `atomicWrite` that verifies: (1) file is written with correct permissions, (2) temp file is cleaned up on write failure, (3) original file is untouched on rename failure.

### SUGGESTION: `slugify` edge case — all-special-characters directory name

- **Location:** `internal/store/store.go:293–309`
- **Description:** `slugify` falls back to `"bundle"` if the result is empty. There is no test for this case (e.g., a directory named `"---"` or `"___"`).
- **Recommended fix:** Add a test case for `AutoBundleName` with a directory like `"/tmp/---"` to verify the fallback.

### SUGGESTION: `WriteConfig` in `config.go` uses `atomicWrite` but no test verifies atomicity

- **Location:** `internal/store/config.go:45`
- **Description:** `WriteConfig` delegates to `atomicWrite`, but `TestConfigRoundTrip` only verifies the round-trip, not atomicity or permissions.
- **Recommended fix:** Add a test that verifies `WriteConfig` creates files with 0600 permissions.

### SUGGESTION: Bundle filename not validated against path traversal

- **Location:** `internal/store/store.go:202`
- **Description:** `WriteBundle`, `ReadBundle`, and `DeleteBundle` accept a `filename` parameter and join it with `BundlesPath` via `filepath.Join`. If `filename` contains `..` components, `filepath.Join` resolves them, but the resulting path could escape the bundles directory. For example, `filename = "../index.json"` would resolve to `~/.envmoat/index.json`.
- **Impact:** Moderate. If an attacker controls the bundle filename, they could overwrite the index file or config. In practice, bundle filenames are generated internally, but the API is public.
- **Recommended fix:** Validate `filename` contains no `..` components and is a simple basename, or use `filepath.Clean` and verify the resolved path is still under `BundlesPath`.

### SUGGESTION: `LoadIndex` swallows `flock.Lock` errors

- **Location:** `internal/store/store.go:158`
- **Description:** `LoadIndex` and `SaveIndex` return `errors.New("failed to acquire index lock")` which loses the original error. Use `fmt.Errorf("acquire index lock: %w", err)` to preserve the error chain.
- **Recommended fix:** Wrap the flock error with `fmt.Errorf`.

### NOTE: `InitStore` — config written without lock

- **Location:** `internal/store/store.go:122–128`
- **Description:** `InitStore` writes `config.yaml` using `WriteConfig` (which calls `atomicWrite`) but does not hold the index lock. If two processes call `InitStore` concurrently, both may try to write the config simultaneously. The `os.Stat` check is a TOCTOU race.
- **Impact:** Low. `atomicWrite` prevents data corruption, and the config content (random salt) would differ between the two writes, meaning one salt would silently win. This could cause key derivation failures for bundles encrypted under the other salt.
- **Recommended fix:** Add a file lock around the config write in `InitStore`, or use `os.OpenFile` with `O_CREATE|O_EXCL` semantics.

### NOTE: `DeleteBundle` — no lock, no atomic delete

- **Location:** `internal/store/store.go:227`
- **Description:** `DeleteBundle` uses `os.Remove` directly without locking or atomicity. If another process is reading the bundle concurrently, this could cause issues.
- **Impact:** Low for a CLI tool (concurrent deletes are unlikely), but inconsistent with the spec's concurrency guarantees.
- **Recommended fix:** Consider adding a per-bundle flock before deletion, consistent with the spec's concurrency model.

---

## Spec Verification Checklist

| Spec Requirement | Status | Notes |
|---|---|---|
| scrypt N=262144, r=8, p=1 | ✅ PASS | `crypto.go:38-44` |
| HKDF info = "envmoat/v1/dek" | ⚠️ MISMATCH | Code says "envmoat/v1/dek", spec says "secrets-manager/v1/dek" |
| AES-256-GCM | ✅ PASS | `crypto.go:68-78` |
| File format: [1B ver][12B nonce][ciphertext][16B tag] | ✅ PASS | Code correct; spec has typo (says 32B tag) |
| Atomic writes (temp + rename) | ✅ PASS | `store.go:345-371` |
| Dir perms 0700, file perms 0600 | ✅ PASS | `store.go:31-32`, enforced in `atomicWrite` and `MkdirAll` |
| flock on index.json | ✅ PASS | `store.go:80`, `LoadIndex`/`SaveIndex` |

---

## Test Coverage Summary

| Area | Coverage | Notes |
|---|---|---|
| scrypt derivation | ✅ | Determinism, different passwords |
| HKDF derivation | ✅ | Determinism, different filenames |
| Encrypt/Decrypt round-trip | ✅ | Including nonce uniqueness |
| Wrong key rejection | ✅ | |
| Tampered data rejection | ✅ | |
| Short data rejection | ✅ | |
| Full pipeline | ✅ | |
| Store init | ❌ FAILS | Nil pointer on `indexLock` |
| Bundle CRUD | ⚠️ BROKEN | Uses broken `testStore()` helper |
| Index CRUD | ❌ NOT TESTED | No tests for `AddAutoMapping`, `AddProfileMapping`, etc. |
| Permissions validation | ⚠️ PARTIAL | Only happy path tested |
| Config round-trip | ✅ | |
| AutoBundleName | ✅ | Including collision |

---

## Summary

**1 blocker** (test panic), **3 bugs** (spec/code mismatches, broken test helper), **7 suggestions**, **2 notes**.

The crypto implementation is solid — correct algorithms, proper parameters, good test coverage. The main issues are in the store layer: the test helper creates a broken `Store` with a nil `indexLock`, the spec has stale naming and an incorrect auth tag size, and several defensive measures (path traversal validation, fsync, permission test cases) are missing.

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Review is scoped to the 5 specified files and 6 known issues. No source files were modified. All findings are from direct code inspection and test execution."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "Each finding cites exact file paths and line numbers. Test output captured showing the nil pointer panic. Spec cross-referenced for all known issues. Coverage gaps documented."
    }
  ],
  "changedFiles": [],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "go test ./internal/crypto/ -v",
      "result": "passed",
      "summary": "All 8 crypto tests pass"
    },
    {
      "command": "go test ./internal/store/ -v",
      "result": "failed",
      "summary": "TestInitStore panics with nil pointer dereference at store.go:183 (SaveIndex calls s.indexLock.Lock on nil)"
    },
    {
      "command": "grep -n flock store.go",
      "result": "passed",
      "summary": "Confirmed indexLock is set in NewStore (line 97) but not in test helpers"
    },
    {
      "command": "grep HKDF info string in spec vs code",
      "result": "passed",
      "summary": "Spec says secrets-manager/v1/dek, code says envmoat/v1/dek"
    },
    {
      "command": "grep auth tag size in spec",
      "result": "passed",
      "summary": "Spec says 32B auth tag, code correctly uses 16B (AES-GCM standard)"
    }
  ],
  "validationOutput": [
    "Crypto tests: 8/8 PASS",
    "Store tests: FAIL — TestInitStore panics (nil indexLock), remaining tests not reached",
    "scrypt params verified: N=262144, r=8, p=1 — MATCH spec",
    "HKDF info: code='envmoat/v1/dek', spec='secrets-manager/v1/dek' — MISMATCH",
    "File format: [1B version][12B nonce][ciphertext][16B tag] — CODE CORRECT, SPEC TYPO (says 32B)",
    "Atomic writes: temp file + os.Rename — MATCH spec",
    "Permissions: dirs 0700, files 0600 — MATCH spec"
  ],
  "residualRisks": [
    "Bundle filename path traversal: no validation on filename parameter in WriteBundle/ReadBundle/DeleteBundle",
    "No fsync before rename in atomicWrite: data durability risk on crash",
    "TOCTOU race in InitStore config write: concurrent setup could produce mismatched salts",
    "ValidatePermissions uses os.Stat (follows symlinks) instead of os.Lstat"
  ],
  "noStagedFiles": true,
  "diffSummary": "No changes made. Read-only review of 5 files.",
  "reviewFindings": [
    "blocker: store_test.go:28-35 — TestInitStore panics: indexLock is nil, SaveIndex calls .Lock() on nil pointer",
    "bug: crypto.go:50 — HKDF info 'envmoat/v1/dek' mismatches spec 'secrets-manager/v1/dek'",
    "bug: PLAN.md:61 — Spec says 32B auth tag; AES-GCM produces 16B (code is correct, spec is wrong)",
    "bug: store_test.go:21 — testStore() sets indexLock=nil; breaks all tests using it that call InitStore",
    "suggestion: store.go:345-371 — atomicWrite missing f.Sync() before rename for durability",
    "suggestion: store.go:315-334 — ValidatePermissions uses os.Stat instead of os.Lstat",
    "suggestion: store_test.go — no test for ValidatePermissions catching bad permissions",
    "suggestion: store_test.go — no direct test for atomicWrite behavior",
    "suggestion: store.go:202 — bundle filename not validated against path traversal",
    "suggestion: store.go:158 — LoadIndex/SaveIndex swallow flock.Lock errors with errors.New instead of fmt.Errorf",
    "note: store.go:122 — InitStore config write has TOCTOU race, no lock",
    "note: store.go:227 — DeleteBundle uses os.Remove without flock"
  ],
  "manualNotes": "The crypto layer is well-implemented and thoroughly tested. The primary issue is the store test infrastructure: the test helper creates Store structs with nil indexLock, causing panics. This blocks all store tests beyond the first failure. The spec/code naming mismatch (envmoat vs secrets-manager) needs a decision from the project owner. No source files were modified per hard constraints."
}
```
