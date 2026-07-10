# Task for delegate

Create a beads issue using the `bd` CLI. Run this exact command:

bd create "CLI: setup + init — master password, shell hook install, project init" \
  --description "Setup and project initialization commands.

secrets-manager setup:
- Create master password (interactive prompt, never CLI args)
- Generate global salt, store in config.yaml
- Install shell hook into detected rc file (~/.zshrc or ~/.bashrc)
- Idempotent: grep for __envmoat_hook before appending
- --reset flag to change password
- FileVault check: warn if disabled

secrets-manager init [project-root]:
- Create .envmoat marker file in project root (defaults to PWD)
- Create auto-named bundle (auto-<slugified-dirname>.enc)
- Add entry to index.json auto section
- Auto-append .envmoat to project .gitignore
- File permissions: 0600 for marker, 0700 for store dirs

Shell Hook (zsh):
- chpwd hook, fires only on directory change
- __envmoat_last_bundle change detection
- eval shell-safe output

Shell Hook (bash):
- PROMPT_COMMAND (array form for bash 5.1+, string fallback)
- Same change detection pattern

Interactive guard: hook only installs in interactive shells ([[ \$- == *i* ]])

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