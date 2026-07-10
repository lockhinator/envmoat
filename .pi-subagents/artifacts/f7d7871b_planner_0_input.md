# Task for planner

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md]

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Create a detailed implementation plan for Phase 1 (MVP) of envmoat. Read the plan at /Users/cameronlockhart/Development/secrets-manager/PLAN.md and the README at /Users/cameronlockhart/Development/secrets-manager/README.md for full context.

Phase 1 tasks (all under SM-gtv):
- SM-gtv.1: Platform abstraction layer — KeyringBackend and ClipboardBackend interfaces
- SM-gtv.2: CLI: set + get + list — secret CRUD, interactive prompts, stdin
- SM-gtv.3: CLI: load + remove + deinit + verify — shell injection, deletion, integrity
- SM-gtv.4: Directory walk + marker resolution — walk-up, .envmoat marker, symlink handling
- SM-gtv.5: CLI: setup + init — master password, shell hook install, project init
- SM-gtv.6: Project scaffolding — Go module, CLI framework, command structure (CLOSED)
- SM-gtv.7: Encryption + storage — scrypt, HKDF, AES-256-GCM, bundle format, index.json

For each task, add to its description via `bd update`:
1. Implementation approach: which Go packages to use, file structure, key functions
2. Dependencies on other Phase 1 tasks (what must be done first)
3. Acceptance criteria: specific testable conditions
4. Recommended implementation order

Update each issue's description using `bd update <id> --description "..."` — append the plan to the existing description, don't replace it.

Recommended implementation order: SM-gtv.7 (encryption) → SM-gtv.4 (directory walk) → SM-gtv.1 (platform abstraction, refine stubs) → SM-gtv.5 (setup+init) → SM-gtv.2 (set+get+list) → SM-gtv.3 (load+remove+deinit+verify)

Key design decisions from PLAN.md:
- scrypt N=262144, r=8, p=1
- HKDF-SHA256 for per-bundle DEK
- AES-256-GCM, file format: [1B version][12B nonce][ciphertext][32B auth tag]
- Storage: ~/.envmoat/{config.yaml, bundles/*.enc, index.json}
- Marker: .envmoat file (empty= default, "disabled", "profile: <name>")
- Shell hook: zsh chpwd, bash PROMPT_COMMAND
- Single-quote escaping with '\'' for internal single quotes
- Atomic writes: temp file + rename
- File locking: github.com/gofrs/flock
- Permissions: 0700 dirs, 0600 files
- Key validation: alphanumeric + underscore + hyphen
- Size warning at 1MB

Context: fork

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/f7d7871b/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/f7d7871b/plan-phase1.md
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