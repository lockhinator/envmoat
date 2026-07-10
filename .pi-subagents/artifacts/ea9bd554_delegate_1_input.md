# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "Platform abstraction layer — KeyringBackend and ClipboardBackend interfaces" \
  --description "Cross-platform backend abstraction.

- Define KeyringBackend interface: StoreLUK(), GetLUK(), DeleteLUK()
  - macOS: Security Framework + SecAccessControl (Touch ID) — CGO
  - Linux: GNOME Keyring or KWallet via DBus
- Define ClipboardBackend interface: Copy(string) error
  - macOS: pbcopy
  - Linux: xclip (X11) or wl-clipboard (Wayland), auto-detect
- Auto-detect platform at build/runtime, wire correct backend
- Graceful degradation: if no secret store available, fall back to in-memory cache with warning
- Factory pattern: NewKeyringBackend() returns correct impl for runtime OS

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