# Fix SM-gtv.1 — Platform Abstraction Layer Review Findings

## Changes Applied

### 1. Stale comments in `backend.go` (lines 24-25)
**File:** `internal/backend/backend.go`
- Updated doc comment for `NewKeyringBackend` to reference `darwin_keyring.go` / `linux_keyring.go` instead of deleted `darwin_stub.go` / `linux_stub.go`.
- Updated doc comment for `NewClipboardBackend` to reference `darwin_clipboard.go` / `linux_clipboard.go` instead of deleted stub files.

### 2. Error masking in `GetLUK`
**Files:** `internal/backend/darwin_keyring.go`, `internal/backend/linux_keyring.go`
- Changed `GetLUK` to use `errors.Is(err, keyring.ErrNotFound)` to distinguish "key not found" from other keyring errors.
- `ErrNotFound` → returns `ErrNotAvailable` (correct semantic: no key stored).
- All other errors → propagated unchanged to the caller.
- Added `"errors"` import to both files.

### 3. Unused `ErrNotImplemented`
**File:** `internal/backend/backend_errors.go`
- Removed `ErrNotImplemented` variable and its doc comment (unused in the codebase).

## Validation

| Command | Result |
|---------|--------|
| `go build ./...` | Pass (no output) |
| `go vet ./...` | Pass (no output) |
