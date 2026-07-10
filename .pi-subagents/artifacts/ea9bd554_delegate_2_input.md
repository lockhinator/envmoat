# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "Encryption + storage — scrypt, HKDF, AES-256-GCM, bundle format, index.json" \
  --description "Core encryption and storage layer.

Key Derivation:
- Master Password -> scrypt(password, global_salt, N=262144, r=8, p=1) -> 32-byte LUK
- DEK = HKDF-SHA256(LUK, bundle_filename, info=\"envmoat/v1/dek\") -> 32-byte per-bundle DEK
- Global salt stored in config.yaml

Bundle Encryption:
- AES-256-GCM(DEK, nonce, plaintext)
- File format: [1B version=0x01][12B nonce][ciphertext][32B auth tag]
- Bundle JSON: { \"_meta\": {\"created_at\": \"...\", \"updated_at\": \"...\"}, \"API_KEY\": \"sk-...\", ... }

Storage Layout:
- ~/.envmoat/config.yaml (global settings)
- ~/.envmoat/bundles/<bundle-id>.enc (encrypted bundles)
- ~/.envmoat/index.json (path -> bundle mapping, versioned schema)

File Operations:
- Atomic writes: temp file + rename pattern
- File locking: flock (BSD-compatible, github.com/gofrs/flock)
- Permissions: 0700 directories, 0600 files, validated on startup
- Concurrency: flock on bundle and index.json writes

Parent: SM-gtv (Phase 1 MVP)
Priority: P0
" \
  --parent SM-gtv \
  --priority 0 \
  --silent

Output only the issue ID returned.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/ea9bd554/progress.md

## Acceptance Contract
Acceptance level: checked
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope

Required evidence: changed-files, tests-added, commands-run, residual-risks, no-staged-files

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