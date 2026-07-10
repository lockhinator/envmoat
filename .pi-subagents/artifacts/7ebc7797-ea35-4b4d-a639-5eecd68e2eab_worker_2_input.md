# Task for worker

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md, /Users/cameronlockhart/Development/secrets-manager/plan.md]

Implement SM-gtv.4: Directory Walk + Marker Resolution for envmoat.

**Goal**: Create `internal/resolver/` package.

**Context**: envmoat uses `.envmoat` marker file (NOT `.secrets-manager`). Read `PLAN.md` for full spec.

**Walk behavior**:
1. Resolve PWD to canonical path (filepath.Clean, no symlink following yet — that's Phase 4)
2. Walk up from PWD looking for `.envmoat`
3. Stop at `/` or `ENVMOAT_WALK_ROOT` env var
4. Return directory containing marker

**Marker parsing** (trimmed whitespace + trailing newline, case-sensitive):
- Empty file → default auto bundle
- `disabled` → stop, no bundle
- `profile: <name>` → named profile bundle
- Other → error

**API** (`internal/resolver/resolver.go`):
```go
type MarkerContent int
const (
    MarkerDefault MarkerContent = iota
    MarkerDisabled
    MarkerProfile
)

type ResolveResult struct {
    MarkerDir   string
    Marker      MarkerContent
    ProfileName string
}

func Resolve(dir string) (*ResolveResult, error)
func ResolveFromPWD() (*ResolveResult, error)
func ParseMarker(path string) (MarkerContent, string, error)
func FindWalkRoot() string
```

**Debug mode**: `ENVMOAT_DEBUG=1` — verbose stderr of walk steps. Never logs secrets.

**Validation**: `go build ./...` must succeed. Add unit tests for marker parsing.
**Hard constraints**: Do NOT implement bundle decryption. Do NOT implement symlink canonicalization yet. This package only finds and parses markers.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/7ebc7797-ea35-4b4d-a639-5eecd68e2eab/wave1-resolver.md
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