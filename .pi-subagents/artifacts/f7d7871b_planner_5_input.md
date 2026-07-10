# Task for planner

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md]

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Create detailed implementation plans for Phase 7 (Polish) tasks. Read the plan at /Users/cameronlockhart/Development/secrets-manager/PLAN.md and README at /Users/cameronlockhart/Development/secrets-manager/README.md for context.

Phase 7 tasks (all under SM-rgt):
- SM-rgt.1: Supply chain hardening — go mod verify, reproducible builds, audit
- SM-rgt.2: Shell completions — zsh + bash
- SM-rgt.3: asdf plugin for envmoat
- SM-rgt.4: mise plugin for envmoat
- SM-rgt.5: Code signing + notarization for macOS Gatekeeper

For each task, update its description via `bd update <id> --description "..."` — append implementation plan to existing description.

Include:
1. Implementation approach
2. Files to create
3. CI/CD integration details
4. Acceptance criteria

Recommended order: SM-rgt.3 (asdf) → SM-rgt.4 (mise) → SM-rgt.2 (completions) → SM-rgt.5 (code signing) → SM-rgt.1 (supply chain)

Key details from PLAN.md:
- asdf: plugin repo, install.sh, version resolution via git tags
- mise: mise.toml plugin spec, binary download with OS/arch detection
- Completions: cobra-cli generate, zsh _envmoat, bash envmoat-completion.bash
- Code signing: codesign + xcrun notarytool, CI integration
- Supply chain: go mod verify, govulncheck, reproducible builds, SHA256 checksums

Context: fork

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/f7d7871b/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/f7d7871b/plan-phase7.md
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