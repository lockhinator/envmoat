# Task for worker

Fix review findings for SM-gtv.7 (Encryption + Storage).\n\n**Blocker — Fix `testStore()` helper** (`store_test.go:12-24`):\n- `indexLock = nil` causes panics in any test that calls `InitStore()`, `LoadIndex()`, `SaveIndex()`, or any index method.\n- Fix: Initialize `s.indexLock = flock.New(filepath.Join(tmpDir, ".index.lock"))` in `testStore()`.\n\n**Suggestion — Add `f.Sync()` before `f.Close()`** (`store.go:345-371`):\n- In `atomicWrite`, call `f.Sync()` before `f.Close()` for crash durability.\n\n**Suggestion — Use `os.Lstat` instead of `os.Stat`** (`store.go:315-334`):\n- In `checkDirPerm` and `checkFilePerm`, use `os.Lstat` to avoid following symlinks in permission checks.\n\n**Suggestion — Wrap flock errors** (`store.go:158`):\n- In `LoadIndex` and `SaveIndex`, use `fmt.Errorf("acquire index lock: %w", err)` instead of `errors.New(...)`.\n\n**Suggestion — Validate bundle filename** (`store.go:202`):\n- In `WriteBundle`, `ReadBundle`, `DeleteBundle`, validate that `filename` is a simple basename (no `..`, no `/`). Return an error if path traversal detected.\n\n**Validation**: `go build ./...`, `go vet ./...`, and `go test ./internal/crypto/ ./internal/store/ -v -count=1` must ALL pass.\n\n**Hard constraints**: Only fix these issues. Do not change anything else. Do not close any tickets.

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/2c0615df-7308-407e-a9fc-d3c55c3234fb/fix-crypto-store.md
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