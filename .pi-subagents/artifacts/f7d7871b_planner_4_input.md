# Task for planner

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md]

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Create detailed implementation plans for Phase 4 (Inheritance) and Phase 6 (Backup/Rotate) tasks. Read the plan at /Users/cameronlockhart/Development/secrets-manager/PLAN.md and README at /Users/cameronlockhart/Development/secrets-manager/README.md for context.

Phase 4 tasks (all under SM-1pz):
- SM-1pz.1: Symlink canonicalization for all path operations
- SM-1pz.2: Marker disabled support
- SM-1pz.3: Configurable walk root via ENVMOAT_WALK_ROOT
- SM-1pz.4: Marker profile override support

Phase 6 tasks (all under SM-1i3):
- SM-1i3.1: backup + restore — encrypted export/import
- SM-1i3.2: edit command — $EDITOR with temp file
- SM-1i3.3: get --clip — clipboard copy via platform backend
- SM-1i3.4: rotate — two-phase atomic re-encryption

For each task, update its description via `bd update <id> --description "..."` — append implementation plan to existing description.

Include:
1. Implementation approach, Go packages
2. Dependencies on other tasks
3. Acceptance criteria
4. Edge cases to handle

Recommended order Phase 4: SM-1pz.2 (disabled) → SM-1pz.4 (profile override) → SM-1pz.1 (symlinks) → SM-1pz.3 (walk root)
Recommended order Phase 6: SM-1i3.3 (clipboard) → SM-1i3.2 (edit) → SM-1i3.1 (backup+restore) → SM-1i3.4 (rotate)

Key details from PLAN.md:
Phase 4:
- Marker parsing: whitespace-trimmed, case-sensitive, strict validation
- Symlinks: canonical paths via filepath.EvalSymlinks
- Walk root: ENVMOAT_WALK_ROOT env var, must be absolute

Phase 6:
- backup: encrypt with current LUK, include metadata (version, timestamp, count)
- restore: merge with existing, prompt on key conflict
- rotate: two-phase atomic, .bak rollback, progress messages
- edit: temp file in ~/.envmoat/, mode 0600, deleted on exit
- get --clip: use ClipboardBackend, stderr "Copied to clipboard", never print to stdout

Context: fork

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/f7d7871b/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/f7d7871b/plan-phase4-6.md
This path is authoritative for this run.
Ignore any other output filename or output path mentioned elsewhere, including output destinations in the base agent prompt, system prompt, or task instructions.

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