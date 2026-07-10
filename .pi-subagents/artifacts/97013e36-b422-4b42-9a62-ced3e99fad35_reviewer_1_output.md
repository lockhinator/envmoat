# Review: SM-gtv.1 — Platform Abstraction Layer

## Review

### Correct
- **Keychain identifiers**: `kcService = "envmoat"` and `kcKey = "envmoat-luk"` in `backend_errors.go:13-16` — match task requirements exactly.
- **Build tags**: All four platform files have correct `//go:build darwin` or `//go:build linux` tags. No stub files remain (`darwin_stub.go`, `linux_stub.go` confirmed deleted).
- **Clipboard implementations**: `darwin_clipboard.go` invokes `pbcopy` correctly via `exec.Command` with stdin pipe. `linux_clipboard.go` tries `wl-copy` first (Wayland), falls back to `xclip -selection clipboard` (X11) — correct invocation for both.
- **`zalando/go-keyring` usage**: `keyring.Set(service, key, value)`, `keyring.Get(service, key)`, `keyring.Delete(service, key)` — all three operations used correctly per the library API. Base64 encoding/decoding of the LUK is consistent between Store and Get.
- **No Touch ID / SecAccessControl**: Confirmed absent from all reviewed files — correctly deferred to Phase 2.
- **Build health**: `go build ./internal/backend/...` and `go vet ./internal/backend/...` both pass cleanly.
- **Interface contract**: `KeyringBackend` and `ClipboardBackend` interfaces in `backend.go` are clean, minimal, and correctly implemented by both darwin and linux backends.

### Fixed
_(No fixes applied — hard constraint: do not modify source files.)_

### Blocker
None.

### Note

1. **Stale comments in `backend.go:24-25`** — The doc comments for `NewKeyringBackend` and `NewClipboardBackend` reference `darwin_stub.go` and `linux_stub.go`, which were deleted. They should reference `darwin_keyring.go`/`linux_keyring.go` and `darwin_clipboard.go`/`linux_clipboard.go` respectively.

2. **Error masking in `GetLUK`** (`darwin_keyring.go:22`, `linux_keyring.go:22`) — Any error from `keyring.Get()` is swallowed and replaced with `ErrNotAvailable`. `zalando/go-keyring` returns its own `ErrNotFound` when the specific key doesn't exist, but also returns `ErrUnsupportedPlatform` or I/O errors in other failure modes. A caller cannot distinguish "key not stored" from "keyring backend failed". Consider:
   ```go
   if errors.Is(err, keyring.ErrNotFound) {
       return nil, ErrNotAvailable
   }
   return nil, err  // propagate other errors
   ```

3. **Unused `ErrNotImplemented`** (`backend_errors.go:9`) — Defined but never referenced in any `.go` file. Dead code; remove or use.

4. **Library divergence from plan**: The plan specifies `99designs/keychain` but the implementation uses `zalando/go-keyring`. This is the correct pragmatic choice since the `99designs/keychain` repository was deleted from GitHub. The plan should be updated to reflect this. Note: `zalando/go-keyring` does not support `SecAccessControl` (biometric), so Phase 2 will require custom CGO bindings to the macOS Security framework regardless.

5. **Empty value edge case**: If the keyring stores an empty string (e.g., via a race or corruption), `GetLUK` would return `([]byte{}, nil)` — a zero-length key that would silently fail later during decryption. Consider returning an error for empty values.

## Summary

Implementation is structurally sound and correctly implements the Phase 1 platform abstraction layer. The interfaces are clean, build tags are correct, clipboard backends are properly implemented, and the keyring integration with `zalando/go-keyring` is correct. The main issues are minor: stale comments, error masking in `GetLUK`, and one unused variable.

---

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Reviewed all 6 specified files against task requirements. No scope widening — review limited to SM-gtv.1 platform abstraction layer files only."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "Full file contents read for all 6 files. go build and go vet executed. Stub file absence verified via find. Build tag correctness verified. Keychain identifiers, clipboard commands, and library usage all verified from source."
    }
  ],
  "changedFiles": [],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "go build ./internal/backend/...",
      "result": "passed",
      "summary": "Clean build, no errors"
    },
    {
      "command": "go vet ./internal/backend/...",
      "result": "passed",
      "summary": "No vet issues"
    },
    {
      "command": "find internal/backend -type f -name '*.go' | sort",
      "result": "passed",
      "summary": "Confirmed 6 files: backend.go, backend_errors.go, darwin_clipboard.go, darwin_keyring.go, linux_clipboard.go, linux_keyring.go. No stub files."
    }
  ],
  "validationOutput": [
    "Build: clean",
    "Vet: clean",
    "Stub files: confirmed removed",
    "Keychain service 'envmoat', key 'envmoat-luk': correct",
    "ErrNotAvailable returned on keyring.Get error: correct (but masks all errors)",
    "pbcopy: correct",
    "wl-copy then xclip fallback: correct",
    "zalando/go-keyring used correctly: correct",
    "No SecAccessControl/Touch ID: correct (Phase 2)"
  ],
  "residualRisks": [
    "GetLUK masks all keyring errors as ErrNotAvailable — callers cannot distinguish 'key not found' from 'backend failure'",
    "Stale comments in backend.go reference deleted stub files",
    "Unused ErrNotImplemented dead code",
    "Plan still references 99designs/keychain which is gone; should be updated to zalando/go-keyring"
  ],
  "noStagedFiles": true,
  "diffSummary": "No changes made (read-only review). All 6 platform abstraction files verified.",
  "reviewFindings": [
    "no blockers",
    "suggestion: backend.go:24-25 — stale comments reference deleted stub files",
    "suggestion: darwin_keyring.go:22, linux_keyring.go:22 — GetLUK masks all errors as ErrNotAvailable",
    "suggestion: backend_errors.go:9 — ErrNotImplemented is unused dead code",
    "note: plan references 99designs/keychain but implementation correctly uses zalando/go-keyring"
  ],
  "manualNotes": "Implementation is solid for Phase 1. The error masking in GetLUK is the most actionable finding — it works correctly for the 'key not found' case but would confuse debugging if the keyring backend itself fails. Phase 2 (Touch ID) will need custom CGO since zalando/go-keyring does not support SecAccessControl."
}
```
