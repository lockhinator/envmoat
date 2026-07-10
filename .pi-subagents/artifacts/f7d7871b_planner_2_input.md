# Task for planner

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md]

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Create detailed implementation plans for Phase 2b (Linux Backend) tasks. Read the plan at /Users/cameronlockhart/Development/secrets-manager/PLAN.md and README at /Users/cameronlockhart/Development/secrets-manager/README.md for context.

Phase 2b tasks (all under SM-coy):
- SM-coy.1: Clipboard backend — xclip + wl-clipboard auto-detect
- SM-coy.2: GNOME Keyring via DBus/secret-tool
- SM-coy.3: KWallet fallback via DBus

For each task, update its description via `bd update <id> --description "..."` — append implementation plan to existing description.

Include:
1. Go packages to use (e.g., github.com/godbus/dbus for DBus)
2. DBus API details (org.freedesktop.secrets for Secret Service)
3. Auto-detection logic (XDG_CURRENT_DESKTOP, $WAYLAND_DISPLAY, $XDG_SESSION_TYPE)
4. Acceptance criteria
5. Dependencies

Recommended order: SM-coy.2 (GNOME Keyring) → SM-coy.3 (KWallet fallback) → SM-coy.1 (clipboard)

Key details:
- Secret Service API (org.freedesktop.secrets) is the standard — works with GNOME Keyring and KWallet
- secret-tool CLI as fallback
- Clipboard: xclip -selection clipboard (X11), wl-clipboard -i (Wayland)
- Auto-detect: check $XDG_SESSION_TYPE or $WAYLAND_DISPLAY
- Implement KeyringBackend + ClipboardBackend interfaces from internal/backend

Context: fork

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/f7d7871b/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/f7d7871b/plan-phase2b.md
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