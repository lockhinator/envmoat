# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "CLI: load + remove + deinit + verify — shell injection, deletion, integrity" \
  --description "Shell injection, deletion, and integrity commands.

secrets-manager load:
- Shell hook command: emit shell-safe \"export K=V\" lines
- Output format:
  #bundle_hash:sha256:abc123...
  export API_KEY='sk-1234567890abcdef'
  export DB_PASS='p@\$\$w\"rd; with \$pecial chars'
- Single-quote escaping with '\'' for internal single quotes
- Errors to stderr, exit 0 with no output when no bundle found
- Bundle hash on first line for change detection in hook

secrets-manager remove <KEY>:
- Delete a secret from current bundle
- Confirmation prompt (-y to skip)
- Atomic write (temp + rename)

secrets-manager deinit <project-root>:
- Remove .envmoat marker and bundle
- Remove from index.json
- Confirmation prompt (-y to skip)

secrets-manager verify:
- Integrity check: all bundles decrypt successfully
- Index consistency: all referenced bundles exist
- Orphan cleanup: list orphaned bundles, prompt for deletion (not automatic)

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