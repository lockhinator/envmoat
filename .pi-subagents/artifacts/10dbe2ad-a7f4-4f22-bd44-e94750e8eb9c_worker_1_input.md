# Task for worker

[Read from: /Users/cameronlockhart/Development/secrets-manager/context.md, /Users/cameronlockhart/Development/secrets-manager/plan.md]

Fix review findings for SM-gtv.4 (Directory Walk + Marker Resolution).

**Findings to fix**:

1. **`ResolveFromPWD` untested** — Add a test that creates a temp dir with a marker, `os.Chdir`s into it, and calls `ResolveFromPWD()`.

2. **7 tests leak temp dirs** — In `resolver_test.go`, these tests call `tmpDir()` without `*testing.T` so cleanup never runs. Fix by passing `t`: `TestParseMarkerDisabled`, `TestParseMarkerDisabledWithWhitespace`, `TestParseMarkerProfile`, `TestParseMarkerProfileWithExtraWhitespace`, `TestParseMarkerEmptyProfileName`, `TestParseMarkerUnrecognized`, `TestParseMarkerCaseSensitive`.

3. **`ParseMarker` returns `MarkerDefault` for error path** (`resolver.go:88`) — Return `0` (zero value) or add a `MarkerUnknown` sentinel instead of `MarkerDefault` for unrecognized content.

4. **`FindWalkRoot` doesn't enforce absolute path** (`resolver.go:95-100`) — Call `filepath.Abs(root)` after `filepath.Clean` so relative paths and `~` are properly resolved. If `filepath.Abs` fails, return "/" as fallback.

**Validation**: `go build ./...`, `go vet ./...`, and `go test ./internal/resolver/ -v -count=1` must all pass.

**Hard constraints**: Only fix these 4 issues. Do not change anything else. Do not close any tickets.

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/10dbe2ad-a7f4-4f22-bd44-e94750e8eb9c/progress.md

---
**Output:**
Write your findings to exactly this path: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/outputs/10dbe2ad-a7f4-4f22-bd44-e94750e8eb9c/fix-resolver.md
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