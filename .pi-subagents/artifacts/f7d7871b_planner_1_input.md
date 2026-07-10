# Task for planner

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md]

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Create detailed implementation plans for Phase 2 (macOS Keychain) tasks. Read the plan at /Users/cameronlockhart/Development/secrets-manager/PLAN.md and README at /Users/cameronlockhart/Development/secrets-manager/README.md for context.

Phase 2 tasks (all under SM-ckp):
- SM-ckp.1: Session caching with sliding TTL
- SM-ckp.2: CLI: status + logout — session state, Keychain diagnostics
- SM-ckp.3: Keychain two-item pattern — protected + cache items
- SM-ckp.4: FileVault check in doctor/setup
- SM-ckp.5: SecAccessControl + Touch ID integration

For each task, update its description via `bd update <id> --description "..."` — append implementation plan to existing description.

Include:
1. CGO bindings needed (exact Security.framework functions)
2. Keychain item attributes (kSecAttrService, kSecAttrAccount, etc.)
3. Error handling for biometry unavailable, user cancel, max attempts
4. Acceptance criteria
5. Dependencies on other tasks

Recommended order: SM-ckp.3 (two-item pattern) → SM-ckp.5 (SecAccessControl) → SM-ckp.1 (session caching) → SM-ckp.2 (status+logout) → SM-ckp.4 (FileVault)

Key details from PLAN.md:
- Two Keychain items: "envmoat-luk-protected" and "envmoat-luk-cache" under service "envmoat"
- Protected: SecAccessControl with kSecAccessControlUserPresence
- Cache: no access control, includes timestamp
- Sliding TTL: default 15 min, resets on each access
- kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly
- Consider 99designs/keychain for basic ops, custom CGO for SecAccessControl
- FileVault: fdesetup isactive

Context: fork

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/f7d7871b/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/f7d7871b/plan-phase2.md
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