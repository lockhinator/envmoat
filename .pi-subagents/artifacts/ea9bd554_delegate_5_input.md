# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "CLI: set + get + list — secret CRUD, interactive prompts, stdin" \
  --description "Core secret management commands.

secrets-manager set <KEY> [VALUE]:
- Omitting VALUE prompts interactively (readline, hidden input)
- --stdin flag: read value from stdin
- --file flag: bulk import from .env file
- Key validation: alphanumeric + underscore + hyphen only
- Secret size warning at 1MB (shell export practical limit ~2MB)
- Re-auths if session expired
- Context resolution: walk up from PWD for marker, look up in index.json
- Error if not in tracked directory

secrets-manager get <KEY>:
- Print decrypted value to stdout
- Context resolution same as set

secrets-manager list:
- List keys only (values hidden)
- Show active profile/bundle name for context
- Context resolution same as set

All commands resolve active bundle via walk-up + index.json lookup.

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