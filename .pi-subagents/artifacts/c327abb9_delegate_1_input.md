# Task for delegate

Create beads issues for Phase 2b (Linux Backend). Run these commands and output all issue IDs:

bd create "GNOME Keyring via DBus/secret-tool" \
  --description "Implement GNOME Keyring backend for Linux.

- GNOME Keyring integration via DBus Secret Service or secret-tool CLI
- Store LUK in a collection (e.g., 'envmoat-protected')
- Cache item with timestamp for TTL
- Two-item pattern adapted for Linux D-Bus Secret Service
- PIN/password prompt via system dialog (no biometric equivalent)
- Implement KeyringBackend interface: StoreLUK(), GetLUK(), DeleteLUK()

Parent: SM-coy (Phase 2b Linux Backend)
Priority: P1
" \
  --parent SM-coy --priority 1 --silent

bd create "KWallet fallback via DBus" \
  --description "Implement KWallet fallback for KDE Plasma environments.

- KWallet integration via DBus
- Detect KDE Plasma environment (XDG_CURRENT_DESKTOP contains 'KDE')
- Store LUK in KWallet folder 'envmoat'
- Same two-item pattern as GNOME Keyring
- Implement KeyringBackend interface
- Auto-select: GNOME Keyring for GNOME, KWallet for KDE, error if neither available

Parent: SM-coy (Phase 2b Linux Backend)
Priority: P1
" \
  --parent SM-coy --priority 1 --silent

bd create "Clipboard backend — xclip + wl-clipboard auto-detect" \
  --description "Implement Linux clipboard backend.

- ClipboardBackend interface: Copy(string) error
- X11: xclip -selection clipboard
- Wayland: wl-clipboard -i
- Auto-detect: check $XDG_SESSION_TYPE or $WAYLAND_DISPLAY
- Fallback: if neither available, warn on stderr, skip clipboard copy
- 'envmoat get --clip' uses this backend on Linux

Parent: SM-coy (Phase 2b Linux Backend)
Priority: P1
" \
  --parent SM-coy --priority 1 --silent

echo "Done. Issue IDs created."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/c327abb9/progress.md

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