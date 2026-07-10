# Task for delegate

Create beads issues for Phase 2 (macOS Keychain Backend). Run these commands and output all issue IDs:

bd create "Keychain two-item pattern — protected + cache items" \
  --description "Implement the macOS Keychain two-item storage pattern.

- Protected item: LUK stored with SecAccessControl requiring biometry (kSecAccessControlUserPresence)
- Cache item: LUK stored without access control, with timestamp
- Fixed labels: 'envmoat-luk-protected' and 'envmoat-luk-cache' under service 'envmoat'
- SecItemAdd / SecItemCopyMatching for both items
- kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly
- Check cache item first on each access; if TTL expired, prompt via protected item
- On TTL expiry or logout: explicitly delete cache item

Parent: SM-ckp (Phase 2 macOS Keychain Backend)
Priority: P1
" \
  --parent SM-ckp --priority 1 --silent

bd create "SecAccessControl + Touch ID integration" \
  --description "Implement Touch ID biometric unlock via macOS Security Framework.

- CGO bindings for SecAccessControl creation
- kSecAccessControlUserPresence for biometric prompt
- Native Touch ID dialog (no custom UI)
- Fallback to Mac login password (macOS behavior, documented)
- Error handling: biometry unavailable, user cancel, max attempts
- Consider 99designs/keychain for basic operations, custom CGO only for SecAccessControl

Parent: SM-ckp (Phase 2 macOS Keychain Backend)
Priority: P1
" \
  --parent SM-ckp --priority 1 --silent

bd create "Session caching with sliding TTL" \
  --description "Implement session caching with sliding window TTL.

- Default TTL: 15 minutes, configurable
- Sliding window: resets on each successful access
- Keychain cache item shared across all terminals (unlock in one = unlock in all)
- Cache item includes timestamp for TTL check
- On TTL expiry: cache item deleted, next access prompts for Touch ID
- On 'envmoat logout': cache item explicitly deleted

Parent: SM-ckp (Phase 2 macOS Keychain Backend)
Priority: P1
" \
  --parent SM-ckp --priority 1 --silent

bd create "CLI: status + logout — session state, Keychain diagnostics" \
  --description "Implement status and logout commands.

envmoat status:
- Active profile/bundle name
- Session TTL remaining (human-readable)
- Keychain state: protected item exists?, cache item exists?
- Debug mode hint: 'Set ENVMOAT_DEBUG=1 for verbose logging'

envmoat logout:
- Delete Keychain cache item
- Require re-auth on next access
- Confirm: 'Session cleared. Next command will prompt for Touch ID.'

Parent: SM-ckp (Phase 2 macOS Keychain Backend)
Priority: P1
" \
  --parent SM-ckp --priority 1 --silent

bd create "FileVault check in doctor/setup" \
  --description "Implement FileVault status check.

- 'envmoat doctor': check FileVault status, warn if disabled
- 'envmoat setup': check FileVault status, warn if disabled (don't block)
- Use 'fdesetup isactive' to check status
- Warning message: 'FileVault is disabled. Full-disk encryption is recommended for storing secrets.'

Parent: SM-ckp (Phase 2 macOS Keychain Backend)
Priority: P1
" \
  --parent SM-ckp --priority 1 --silent

echo "Done. Issue IDs created."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/c327abb9/progress.md

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