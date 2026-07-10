# Task for reviewer

Review the implementation of SM-gtv.7 (Encryption + Storage) for envmoat.

**Files to review**:
- `internal/crypto/crypto.go` — scrypt, HKDF, AES-256-GCM
- `internal/crypto/crypto_test.go` — tests
- `internal/store/store.go` — store init, config, index, bundle CRUD
- `internal/store/store_test.go` — tests
- `internal/store/config.go` — Config struct

**Review angles**: correctness, security, error handling, edge cases, Go idioms, test coverage.

**Known issues to verify**:
- Store test `TestInitStore` FAILS with nil pointer dereference in `SaveIndex` → flock is nil. Check if flock is properly initialized in `NewStore`.
- Verify scrypt params: N=262144, r=8, p=1
- Verify HKDF info string is "envmoat/v1/dek"
- Verify file format: [1B version=0x01][12B nonce][ciphertext][16B auth tag]
- Verify atomic writes use temp file + os.Rename
- Verify file permissions: dirs 0700, files 0600

**Spec reference**: Read `PLAN.md` for the full encryption and storage spec.

**Hard constraints**: Do NOT modify project/source files. Return findings only through your response.

**Output format**: List each finding with severity (blocker/bug/suggestion), file:line, description, and recommended fix. Write a complete review document — do not stop early.

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/bf7d444a-f306-4f75-b295-d0052a3f9886/review-crypto-store.md
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