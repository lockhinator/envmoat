# Task for worker

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md, /Users/cameronlockhart/Development/secrets-manager/plan.md]

Implement SM-gtv.1: Platform Abstraction Layer for envmoat.

**Goal**: Replace stub implementations in `internal/backend/` with real implementations.

**Context**: envmoat is a Go CLI secrets manager. Read `PLAN.md` and existing files in `internal/backend/`.

Interfaces in `internal/backend/backend.go`:
```go
type KeyringBackend interface {
    StoreLUK(key []byte) error
    GetLUK() ([]byte, error)  // returns ErrNotAvailable if no key
    DeleteLUK() error
}
type ClipboardBackend interface {
    Copy(text string) error
}
```

**macOS** (`internal/backend/darwin_keyring.go`, build tag `darwin`):
- Use `github.com/99designs/keychain` for Keychain operations
- Service: `envmoat`, Key: `envmoat-luk`
- Touch ID / SecAccessControl is Phase 2 — basic Keychain storage for now

**macOS Clipboard** (`internal/backend/darwin_clipboard.go`, build tag `darwin`):
- Use `pbcopy` via exec.Command

**Linux** (`internal/backend/linux_keyring.go`, build tag `linux`):
- Use `github.com/99designs/keychain` (supports Linux Secret Service)
- Same service/key names. Return `ErrNotAvailable` if no keyring.

**Linux Clipboard** (`internal/backend/linux_clipboard.go`, build tag `linux`):
- Try `wl-copy` first, fallback to `xclip`

**Shared** (`internal/backend/backend_errors.go`, no build tag):
- `ErrNotAvailable`, `ErrNotImplemented`

**Remove** old stub files after creating new ones.

**Validation**: `go build ./...` must succeed.
**Dependencies**: `github.com/99designs/keychain`.
**Hard constraints**: Do NOT implement Touch ID yet. Do NOT implement CLI commands.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/wave1-platform.md
This path is authoritative for this run.
Ignore any other output filename or output path mentioned elsewhere, including output destinations in the base agent prompt, system prompt, or task instructions.

## Acceptance Contract
Acceptance level: reviewed
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope
- criterion-2: Return evidence sufficient for an independent acceptance review

Required evidence: changed-files, tests-added, commands-run, validation-output, residual-risks, no-staged-files

Review gate: required by reviewer.

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```