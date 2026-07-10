# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "Project scaffolding — Go module, CLI framework, command structure" \
  --description "Set up the Go project foundation.

- Initialize Go module (go.mod) with module path github.com/lockinator/envmoat
- CLI framework: Cobra (github.com/spf13/cobra) for command structure
- Root command with --version, --help, no-args welcome message
- Subcommand stubs for all Phase 1 commands: setup, init, set, get, list, load, remove, deinit, verify
- Error handling pattern: all errors include actionable recovery hint
- Debug mode: SECRETS_MANAGER_DEBUG=1 enables verbose stderr logging (never logs secret values)
- Build tags for CGO (macOS Security framework) vs pure Go (Linux)

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