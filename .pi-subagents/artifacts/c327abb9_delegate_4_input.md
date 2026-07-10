# Task for delegate

Create beads issues for Phase 6 (Backup/Rotate). Run these commands and output all issue IDs:

bd create "backup + restore — encrypted export/import" \
  --description "Implement backup and restore commands.

envmoat backup --to-file <path>:
- Export all bundles to single encrypted file
- Encrypted with current LUK (requires auth)
- Include metadata: version, timestamp, bundle count
- 'envmoat export' alias

envmoat restore --from-file <path>:
- Import bundles from encrypted backup file
- Requires auth (decrypt with current LUK)
- Merge with existing bundles
- Prompt on key conflict: skip, overwrite, or cancel
- 'envmoat import' alias

Parent: SM-1i3 (Phase 6 Backup/Rotate)
Priority: P2
" \
  --parent SM-1i3 --priority 2 --silent

bd create "rotate — two-phase atomic re-encryption" \
  --description "Implement password rotation with two-phase atomic re-encryption.

envmoat rotate:
- Phase 1: decrypt all bundles with current LUK
- Phase 2: encrypt all bundles with new LUK (new master password)
- Atomic rename with .bak rollback on failure
- Progress messages: 'Rotating bundle 1/3...'
- After re-encryption: delete old Keychain items, create new
- User re-auths on next access
- On success: 'Rotated N bundles. Run any command to authenticate with the new password.'

Parent: SM-1i3 (Phase 6 Backup/Rotate)
Priority: P2
" \
  --parent SM-1i3 --priority 2 --silent

bd create "edit command — \$EDITOR with temp file" \
  --description "Implement edit command for in-place secret editing.

envmoat edit <KEY>:
- Open \$EDITOR with current secret value
- Temp file in ~/.envmoat/ (not /tmp), mode 0600
- Temp file deleted on exit (success or cancel)
- On save: re-auth if session expired, update bundle with new value
- Atomic write for bundle (temp + rename)
- Cancel (non-zero editor exit): no changes, no error

Parent: SM-1i3 (Phase 6 Backup/Rotate)
Priority: P2
" \
  --parent SM-1i3 --priority 2 --silent

bd create "get --clip — clipboard copy via platform backend" \
  --description "Implement clipboard copy for secret values.

envmoat get <KEY> --clip:
- Copy decrypted value to clipboard via ClipboardBackend
- macOS: pbcopy
- Linux: xclip or wl-clipboard
- Print 'Copied to clipboard' to stderr (not stdout)
- Never print secret value to stdout when --clip is used
- Re-auth if session expired

Parent: SM-1i3 (Phase 6 Backup/Rotate)
Priority: P2
" \
  --parent SM-1i3 --priority 2 --silent

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