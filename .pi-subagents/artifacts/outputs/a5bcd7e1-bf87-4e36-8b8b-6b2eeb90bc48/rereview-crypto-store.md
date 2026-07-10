## SM-gtv.7 Re-review: Encryption + Storage

### Previous Findings — Verification

| # | Finding | Status | Evidence |
|---|---------|--------|----------|
| 1 | BLOCKER: testStore() nil indexLock | **FIXED** | `store_test.go:15` — `s.indexLock = flock.New(...)` now set in helper |
| 2 | f.Sync() before f.Close() | **FIXED** | `store.go:322` — `f.Sync()` called before `f.Close()` in `atomicWrite` |
| 3 | os.Lstat instead of os.Stat | **FIXED** | `store.go:304,315` — both `checkDirPerm` and `checkFilePerm` use `os.Lstat` |
| 4 | Wrap flock errors with fmt.Errorf | **FIXED** | `store.go:139,157` — `LoadIndex`/`SaveIndex` wrap with `fmt.Errorf("acquire index lock: %w", err)` |
| 5 | Validate bundle filename path traversal | **FIXED** | `store.go:32-40` — `validateBundleFilename` rejects `/`, `\`, `..` |

### Test Results

```
go test ./internal/crypto/ ./internal/store/ -v -count=1

crypto:  8 tests — all PASS (1.797s)
store:   8 tests — all PASS (0.411s)
```

### New Issues Introduced by Fixes

**None.** Reviewed each fix for regressions:

- `testStore()` lock init: lock file path uses temp dir; `flock.New` on a non-existent path is fine (flock creates on first lock). No side effects.
- `f.Sync()` placement: correctly between `Chmod` and `Close`; error path cleans up temp file via defer.
- `os.Lstat`: semantically correct for permission checks (avoids symlink follow). No behavior change for regular files/dirs.
- `fmt.Errorf` wrapping: preserves error chain with `%w`; no functional change.
- `validateBundleFilename`: applied in all three bundle entry points (`WriteBundle`, `ReadBundle`, `DeleteBundle`). Rejects empty, `/`, `\`, `..`. No false positives on legitimate names.

### Additional Observations

- **File format comment** (`crypto.go:10`): says `[16B auth tag]` which matches `GCMAuthTagSize = 16` and actual AES-GCM behavior. The plan.md mentions `[32B auth tag]` which is incorrect — the code is right, the plan has the typo. Not a code issue.
- **hkdfInfo constant** (`crypto.go:42`): `"envmoat/v1/dek"` — consistent with project naming (`envmoat`), not `secrets-manager`. This is fine as long as it's consistent across the codebase (it is).
- **atomicWrite cleanup**: defer correctly removes temp file on any error path before `written = true`.

## Review

- **Correct**: All 5 previous findings properly fixed. All 16 tests pass. No regressions introduced.
- **Fixed**: N/A (review-only, no modifications made)
- **Blocker**: None
- **Note**: Plan.md has a typo — says `[32B auth tag]` in file format; actual code correctly uses 16-byte AES-GCM tag. Consider updating plan.md for consistency.

PASS — no issues
