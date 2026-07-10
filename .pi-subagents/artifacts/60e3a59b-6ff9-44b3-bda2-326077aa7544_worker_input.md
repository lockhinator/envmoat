# Task for worker

You are a delegated subagent running from a fork of the parent session. Treat the inherited conversation as reference-only context, not a live thread to continue. Do not continue or answer prior messages as if they are waiting for a reply. Your sole job is to execute the task below and return a focused result for that task using your tools.

Task:
Implement SM-gtv.6: Project scaffolding for envmoat.

**Goal**: Set up the Go project foundation with Cobra CLI framework, all Phase 1 command stubs, and a basic test.

**Context**: This is a macOS+Linux secrets manager called "envmoat" that keeps encrypted env vars invisible to AI agents. The README is at `/Users/cameronlockhart/Development/secrets-manager/README.md` — read it for full context.

**Requirements**:

1. **Go module**: `go mod init github.com/lockinator/envmoat`

2. **Cobra CLI framework** (`github.com/spf13/cobra`):
   - Root command: `envmoat` with `--version` (embed a version var), `--help`
   - No-args behavior: print welcome message + usage, prompt to run `envmoat setup` if not configured
   - Subcommand stubs (just print "not implemented yet" for now):
     - `setup` — create master password + install shell hook
     - `init` — create marker + auto-named bundle
     - `set` — add/update secret
     - `get` — print decrypted value
     - `list` — list keys (values hidden)
     - `load` — emit shell-safe export lines
     - `remove` — delete a secret
     - `deinit` — remove marker and bundle
     - `verify` — integrity check
   - All subcommands support `--help`

3. **Error handling pattern**: helper function `cmdutil.Errorf(actionableHint string, format string, args ...any)` that prints to stderr with recovery hint.

4. **Debug mode**: `ENVMOAT_DEBUG=1` (not SECRETS_MANAGER_DEBUG — the plan used the old name). Package-level `debug()` helper that logs to stderr only when env var is set. **Never logs secret values.**

5. **Build tags**: Prepare for CGO vs pure Go:
   - `internal/backend/darwin_stub.go` with `//go:build darwin` — placeholder for macOS Keychain
   - `internal/backend/linux_stub.go` with `//go:build linux` — placeholder for Linux Keyring
   - `internal/backend/backend.go` — interface definitions

6. **Basic test**: `cmd/root_test.go` — verify the root command runs, `--version` prints version, subcommands exist and are reachable.

**Project structure**:
```
/cmd
  /root.go          — root command + subcommand registration
  /root_test.go     — basic tests
  /main.go          — entry point
/internal
  /backend
    /backend.go     — KeyringBackend + ClipboardBackend interfaces
    /darwin_stub.go — macOS stub (build tag)
    /linux_stub.go  — linux stub (build tag)
  /cmdutil
    /error.go       — Errorf helper
    /debug.go       — debug helper
go.mod
go.sum
README.md           — already exists, don't modify
```

**Success criteria**:
- `go build ./...` succeeds
- `go test ./...` passes
- `./envmoat --version` prints version
- `./envmoat` prints welcome message
- `./envmoat setup --help` shows help text
- All 9 subcommands are registered and reachable

**Hard constraints**:
- Do NOT modify README.md
- Use `envmoat` as the command name everywhere (not secrets-manager)
- Use `ENVMOAT_DEBUG` not `SECRETS_MANAGER_DEBUG`
- This is scaffolding only — commands print "not implemented" stubs, no real logic yet

**Validation**: Run `go build ./...` and `go test ./...` and verify both pass.

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