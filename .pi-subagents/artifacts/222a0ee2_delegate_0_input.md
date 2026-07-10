# Task for delegate

Read /Users/cameronlockhart/Development/secrets-manager/PLAN.md lines 1-100 for encryption details.

Then run these commands to add implementation plans to Phase 1 tasks. Append the plan section to each existing description — read current desc first with `bd show <id> --json`, then update.

For SM-gtv.7 (encryption + storage):
bd update SM-gtv.7 --description "$(bd show SM-gtv.7 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/storage/\n- Files: storage.go (bundle CRUD), encryption.go (scrypt+HKDF+AES-GCM), index.go (index.json), config.go (config.yaml)\n- Go packages: golang.org/x/crypto/scrypt, golang.org/x/crypto/hkdf, github.com/gofrs/flock, gopkg.in/yaml.v3\n- scrypt: N=262144, r=8, p=1 -> 32-byte LUK\n- HKDF-SHA256: LUK + bundle_filename -> 32-byte DEK per bundle\n- AES-256-GCM: crypto/aes, crypto/cipher\n- File format: [0x01][12B nonce][ciphertext][32B auth tag]\n- Bundle JSON: {\"_meta\":{\"created_at\":\"...\",\"updated_at\":\"...\"}, \"KEY\": \"value\", ...}\n- Atomic writes: ioutil.TempFile + os.Rename\n- Permissions: os.MkdirAll 0700, os.OpenFile 0600\n- Dependencies: SM-gtv.6 (scaffolding, CLOSED)\n- Acceptance: encrypt/decrypt roundtrip, file format parse, concurrent write safety, permission validation" 2>&1

For SM-gtv.4 (directory walk):
bd update SM-gtv.4 --description "$(bd show SM-gtv.4 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/resolver/\n- Files: resolver.go (walk-up + marker parse), marker.go (marker file format)\n- Walk: filepath.EvalSymlinks(PWD), then walk up checking for .envmoat\n- Marker parse: read file, trim whitespace/newline, case-sensitive match\n  - Empty string -> default bundle (check index.json auto section)\n  - \"disabled\" -> return nil, no error\n  - \"profile: <name>\" -> look up in index.json profiles section\n  - Other -> error with marker path\n- Walk boundary: / by default, ENVMOAT_WALK_ROOT override\n- Dependencies: SM-gtv.7 (encryption, for bundle loading)\n- Acceptance: walk-up finds marker, disabled stops walk, profile override loads correct bundle, symlink resolution, walk root respected" 2>&1

For SM-gtv.1 (platform abstraction):
bd update SM-gtv.1 --description "$(bd show SM-gtv.1 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: internal/backend/ (already has stubs from SM-gtv.6)\n- Refine backend.go: add ErrNotAvailable sentinel error\n- darwin_stub.go: implement NewKeyringBackend/ NewClipboardBackend returning real types (Phase 2)\n- linux_stub.go: implement NewKeyringBackend/NewClipboardBackend returning real types (Phase 2b)\n- Add runtime detection: runtime.GOOS switch in factory functions\n- Dependencies: SM-gtv.6 (scaffolding, CLOSED)\n- Acceptance: NewKeyringBackend() returns non-nil on darwin/linux, NewClipboardBackend() returns non-nil, interfaces compile" 2>&1

echo "Done updating Phase 1 infra tasks."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/222a0ee2/progress.md

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