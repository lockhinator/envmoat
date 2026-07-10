# Task for worker

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md, /Users/cameronlockhart/Development/secrets-manager/plan.md]

Fix review findings for SM-gtv.1 (Platform Abstraction Layer).

**Findings to fix**:

1. **Stale comments in `backend.go:24-25`** — Doc comments reference `darwin_stub.go` and `linux_stub.go` which were deleted. Update to reference `darwin_keyring.go`/`linux_keyring.go` and `darwin_clipboard.go`/`linux_clipboard.go`.

2. **Error masking in `GetLUK`** (`darwin_keyring.go:22`, `linux_keyring.go:22`) — All errors from `keyring.Get()` are replaced with `ErrNotAvailable`. Fix to distinguish `keyring.ErrNotFound` from other errors:
   ```go
   if errors.Is(err, keyring.ErrNotFound) {
       return nil, ErrNotAvailable
   }
   return nil, err  // propagate other errors
   ```

3. **Unused `ErrNotImplemented`** (`backend_errors.go:9`) — Remove this unused variable.

**Validation**: `go build ./...` and `go vet ./...` must pass.

**Hard constraints**: Only fix these 3 issues. Do not change anything else. Do not close any tickets.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/10dbe2ad-a7f4-4f22-bd44-e94750e8eb9c/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/10dbe2ad-a7f4-4f22-bd44-e94750e8eb9c/fix-platform.md
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