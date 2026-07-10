# Task for delegate

Add implementation plans to Phase 2 (macOS Keychain) tasks. Run these commands:

bd update SM-ckp.3 --description "$(bd show SM-ckp.3 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/backend/darwin_keychain.go (//go:build darwin)\n- CGO: #import <Security/Security.h>\n- Functions: SecItemAdd, SecItemCopyMatching, SecItemDelete\n- Protected item: kSecAttrService=\"envmoat\", kSecAttrAccount=\"envmoat-luk-protected\", with SecAccessControlRef\n- Cache item: kSecAttrAccount=\"envmoat-luk-cache\", no SecAccessControl, include timestamp in value\n- Value format: JSON {\"luk\": \"<hex>\", \"ts\": \"<unix>\"}\n- kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly\n- Dependencies: SM-gtv.1 (backend interface)\n- Acceptance: StoreLUK creates both items, GetLUK returns from cache first then protected, DeleteLUK removes both" 2>&1

bd update SM-ckp.5 --description "$(bd show SM-ckp.5 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/backend/darwin_keychain.go\n- CGO: SecAccessControlCreateWithFlags(NULL, kSecAccessControlUserPresence, kSecAccessControlAllowPasswordChange, &err)\n- Attach SecAccessControlRef to SecItemAdd via kSecAttrAccessControl\n- Error handling: errSecUserCanceled (-25300), errSecAuthFailed (-25293)\n- Fallback: macOS shows login password prompt automatically when biometry unavailable\n- Consider 99designs/keychain for basic SecItemAdd/CopyMatching, custom CGO only for SecAccessControl\n- Dependencies: SM-ckp.3 (two-item pattern)\n- Acceptance: Touch ID prompt appears on GetLUK, cancel returns error, login password fallback works" 2>&1

bd update SM-ckp.1 --description "$(bd show SM-ckp.1 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/session/\n- Files: session.go (TTL management, cache check)\n- Default TTL: 15 minutes, configurable via config.yaml\n- Sliding window: on each GetLUK success, update cache timestamp\n- Cache check: GetLUK reads cache item, check timestamp, if expired delete cache + prompt via protected\n- Cross-terminal: Keychain cache item is shared (fixed account name)\n- Dependencies: SM-ckp.3 (two-item pattern)\n- Acceptance: session stays unlocked for 15min, resets on access, expiry deletes cache, shared across terminals" 2>&1

bd update SM-ckp.2 --description "$(bd show SM-ckp.2 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: cmd/ (status.go, logout.go) or add to root.go\n- status: check Keychain items exist (SecItemCopyMatching with kSecMatchLimitOne), show TTL remaining, show active profile\n- logout: SecItemDelete cache item only, print confirmation to stderr\n- Debug hint: \"Set ENVMOAT_DEBUG=1 for verbose logging\"\n- Dependencies: SM-ckp.1 (session), SM-gtv.4 (resolver for active profile)\n- Acceptance: status shows TTL + keychain state, logout clears cache, next command prompts for auth" 2>&1

bd update SM-ckp.4 --description "$(bd show SM-ckp.4 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: cmd/ (doctor.go) or add to root.go\n- exec.Command(\"fdesetup\", \"isactive\") -> check exit code\n- Warn: \"FileVault is disabled. Full-disk encryption is recommended.\"\n- Integrate into setup (warn, don't block) and doctor (report status)\n- Dependencies: none\n- Acceptance: setup warns if FileVault off, doctor reports FileVault status" 2>&1

echo "Done updating Phase 2 tasks."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/222a0ee2/progress.md

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