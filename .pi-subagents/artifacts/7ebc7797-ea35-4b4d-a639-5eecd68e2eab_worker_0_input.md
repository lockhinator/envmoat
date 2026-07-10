# Task for worker

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md, /Users/cameronlockhart/Development/secrets-manager/plan.md]

Implement SM-gtv.7: Encryption + Storage for envmoat.

**Goal**: Create `internal/crypto/` and `internal/store/` packages with full encryption, key derivation, and bundle storage.

**Context**: envmoat is a Go CLI secrets manager. Binary name is `envmoat`. CLI skeleton is in `cmd/root.go`. Backend interfaces are in `internal/backend/backend.go`. Read `PLAN.md` for full spec.

**Encryption Model** (from PLAN.md):
- Master Password → scrypt(password, global_salt, N=262144, r=8, p=1) → 32-byte LUK
- DEK = HKDF-SHA256(LUK, bundle_filename, info="envmoat/v1/dek") → 32-byte per-bundle DEK
- Bundle = JSON {"_meta": {"created_at": "...", "updated_at": "..."}, "API_KEY": "sk-...", ...}
- Encrypted = AES-256-GCM(DEK, nonce, plaintext)
- File format = [1B version=0x01][12B nonce][ciphertext][16B auth tag]

**Storage Layout**:
```
~/.envmoat/
├── config.yaml              # global settings (TTL, global salt)
├── bundles/
│   ├── <bundle-id>.enc
└── index.json               # path → bundle mapping
```

**index.json schema**:
```json
{"version": 1, "profiles": {}, "auto": {}}
```

**File Permissions**: dirs 0700, files 0600. Validate on startup.
**Atomic writes**: temp file + os.Rename.
**Concurrency**: `github.com/gofrs/flock` for flock.
**Auto-bundle naming**: `auto-<slugified-last-dirname>.enc`. Collision: append `-<short-hash>`.

**Packages to create**:

1. `internal/crypto/crypto.go` — scrypt, HKDF, AES-256-GCM
   - `DeriveLUK(password string, salt []byte) ([]byte, error)`
   - `DeriveDEK(luk []byte, bundleFilename string) ([]byte, error)`
   - `Encrypt(plaintext []byte, dek []byte) ([]byte, error)` — returns [nonce || ciphertext || tag]
   - `Decrypt(ciphertext []byte, dek []byte) ([]byte, error)`

2. `internal/store/store.go` — store init, config, index, bundle CRUD
   - `Store` struct, `NewStore() (*Store, error)` — locate ~/.envmoat/, create if needed
   - `InitStore()` — create config.yaml with random global salt
   - `LoadIndex()` / `SaveIndex()` — atomic with flock
   - `WriteBundle(filename, plaintextJSON, dek) error` — encrypt + atomic write
   - `ReadBundle(filename, dek) ([]byte, error)` — read + decrypt
   - `DeleteBundle(filename) error`
   - `AddAutoMapping(dirPath, bundleFilename)`, `RemoveAutoMapping(dirPath)`
   - `AddProfileMapping(profileName, bundleFilename)`, `RemoveProfileMapping(profileName)`
   - `GetAutoBundle(dirPath) (string, bool)`, `GetProfileBundle(profileName) (string, bool)`
   - `ValidatePermissions()`

3. `internal/store/config.go` — Config struct + YAML
   - `Config`: Version int, GlobalSalt []byte, SessionTTLMinutes int (default 15)

**Validation**: `go build ./...` must succeed. Add unit tests for crypto encrypt/decrypt roundtrip.
**Dependencies**: `github.com/gofrs/flock`, `gopkg.in/yaml.v3`.
**Hard constraints**: Do NOT implement CLI commands. Do NOT implement Keychain backend. Use `os.UserHomeDir()`.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/wave1-crypto-store.md
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