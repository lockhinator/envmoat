# Review: SM-gtv.4 — Directory Walk + Marker Resolution (envmoat resolver)

## Review
- Correct: what is already good (with evidence)
- Fixed: issue, location, and resolution (if you applied a fix)
- Blocker: critical issue that must be resolved before proceeding
- Note: observation, risk, or follow-up item

---

## Correct

| # | Claim | Evidence |
|---|-------|----------|
| 1 | Marker file is `.envmoat` (not `.secrets-manager`) | `resolver.go:12` — `const markerName = ".envmoat"` |
| 2 | Walk stops at `/` or `ENVMOAT_WALK_ROOT` | `resolver.go:52-55` — checks `dir == walkRoot \|\| dir == "/"`; `FindWalkRoot()` reads `ENVMOAT_WALK_ROOT` |
| 3 | Marker parsing: empty → `MarkerDefault` | `resolver.go:72-73` — `content == ""` returns `MarkerDefault` |
| 4 | Marker parsing: "disabled" → `MarkerDisabled` | `resolver.go:76-78` — `content == "disabled"` returns `MarkerDisabled` |
| 5 | Marker parsing: "profile: \<name\>" → `MarkerProfile` | `resolver.go:80-86` — `strings.HasPrefix(content, "profile: ")` extracts profile name |
| 6 | Marker parsing: other → error | `resolver.go:88-89` — unrecognized content returns error |
| 7 | `ResolveFromPWD` uses `os.Getwd()` | `resolver.go:61-66` — calls `os.Getwd()` then delegates to `Resolve` |
| 8 | `ENVMOAT_DEBUG=1` logs walk steps to stderr | `resolver.go:100-103` — checks `ENVMOAT_DEBUG`, writes to `os.Stderr` |
| 9 | No symlink following (Phase 4 deferral) | `resolver.go:34` — uses `os.Stat` (not `os.Lstat` or `filepath.EvalSymlinks`); symlinks to directories are not resolved |
| 10 | Infinite-walk guard | `resolver.go:57-60` — `parent == dir` break prevents infinite loop |
| 11 | Marker parsing is case-sensitive | `resolver.go:76` — literal `"disabled"` comparison; test `TestParseMarkerCaseSensitive` confirms `"DISABLED"` errors |
| 12 | Walk-up inheritance works (child finds parent marker) | `TestResolveWalksUp` — marker in root found when resolving from `root/child/grandchild` |
| 13 | Disabled marker stops walk at closest match | `TestResolveStopsAtDisabled` — child's `.envmoat` with "disabled" takes precedence over parent's profile marker |
| 14 | All 18 tests pass | `go test ./internal/resolver/ -v -count=1` — PASS, 0.373s |

---

## Note — Test Coverage Gap: `ResolveFromPWD` untested

**Severity:** suggestion
**File:** `resolver_test.go`
**Description:** `ResolveFromPWD()` is a public function but has no dedicated test. It delegates to `Resolve(os.Getwd())`, which is indirectly covered by the `Resolve` tests, but there is no test that exercises `os.Getwd()` failure paths or confirms the function's contract in isolation.
**Recommended fix:** Add a test that `os.Chdir`s into a temp directory with a marker and calls `ResolveFromPWD()`.

---

## Note — Test Resource Leak: `tmpDir()` called without `*testing.T`

**Severity:** suggestion
**File:** `resolver_test.go`, lines 109, 120, 131, 142, 153, 164, 175
**Description:** Several tests call `tmpDir()` without passing `t`, so the temp directory is never cleaned up. Affected tests: `TestParseMarkerDisabled`, `TestParseMarkerDisabledWithWhitespace`, `TestParseMarkerProfile`, `TestParseMarkerProfileWithExtraWhitespace`, `TestParseMarkerEmptyProfileName`, `TestParseMarkerUnrecognized`, `TestParseMarkerCaseSensitive`. (Tests using `tmpDir(t)` on lines 82, 93, 214 correctly register cleanup.)
**Recommended fix:** Pass `t` to `tmpDir(t)` in all affected tests.

---

## Bug — `ParseMarker` returns `MarkerDefault` for error cases

**Severity:** bug (low severity — API hygiene)
**File:** `resolver.go:88`
**Description:** When marker content is unrecognized, `ParseMarker` returns `MarkerDefault, "", fmt.Errorf(...)`. The `MarkerDefault` return value is semantically wrong for an error path — a caller who (incorrectly) ignores the error would get `MarkerDefault` instead of a clearly invalid value.
**Recommended fix:** Return a distinct sentinel (e.g., `MarkerDefault` is fine if documented as "always check error first", but a zero-value or dedicated error constant would be clearer). Alternatively, change the signature to return `(*ResolveResult, error)` or add a `MarkerError` sentinel.

---

## Note — `FindWalkRoot` does not enforce absolute path

**Severity:** suggestion
**File:** `resolver.go:95-100`
**Description:** `FindWalkRoot` applies `filepath.Clean` but does not convert to an absolute path. If `ENVMOAT_WALK_ROOT` is set to a relative path (e.g., `ENVMOAT_WALK_ROOT=~/projects`), the comparison `dir == walkRoot` may never match because `dir` is absolute (from `filepath.Abs` or `os.Getwd`) but `walkRoot` is relative.
**Recommended fix:** Call `filepath.Abs(root)` or document that `ENVMOAT_WALK_ROOT` must be an absolute path.

---

## Note — `debug()` always evaluates `os.Getenv`

**Severity:** suggestion
**File:** `resolver.go:101-104`
**Description:** `debug()` calls `os.Getenv("ENVMOAT_DEBUG")` on every invocation. For a hot path (walk loop), this is a minor inefficiency. The env var is read once per directory check.
**Recommended fix:** Cache the debug flag in a package-level var at init time, or use a `sync.Once` / `atomic.Bool`.

---

## Note — `go.mod` module path vs. naming

**Severity:** note
**File:** inferred from test output
**Description:** Test output shows module `github.com/lockinator/envmoat/internal/resolver`. The repo is named `secrets-manager` but the Go module is `envmoat`. This is consistent with the task framing (envmoat is the tool name), but may cause confusion if the repo name and module name are expected to match.

---

## Summary

The implementation is solid and correctly implements the SM-gtv.4 spec for directory walk + marker resolution. All core behaviors (marker name, walk boundaries, parsing rules, debug logging, no symlink following) are correct and tested. The 18 tests all pass.

The main gaps are:
1. **No test for `ResolveFromPWD`** — the primary public entry point is untested directly.
2. **7 tests leak temp directories** — `tmpDir()` called without `t` in parse tests.
3. **`ParseMarker` error returns `MarkerDefault`** — cosmetic API issue.
4. **`FindWalkRoot` doesn't enforce absolute paths** — edge case if env var is relative.

No blockers found.

---

## Acceptance Report

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Reviewed implementation against SM-gtv.4 spec without widening scope; all 7 check items verified from source code and tests"
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "14 correctness claims cited with file:line evidence; 5 findings with severity, location, description, and recommended fix; test run output attached"
    }
  ],
  "changedFiles": [],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "go test ./internal/resolver/ -v -count=1",
      "result": "passed",
      "summary": "18 tests passed, 0 failures, 0.373s"
    }
  ],
  "validationOutput": [
    "All 18 resolver tests pass (TestParseMarkerEmpty through TestMarkerContentValues)",
    "Marker file is .envmoat (resolver.go:12)",
    "Walk stops at / or ENVMOAT_WALK_ROOT (resolver.go:52-55)",
    "Marker parsing handles all 4 cases correctly (resolver.go:72-89)",
    "ResolveFromPWD uses os.Getwd() (resolver.go:61)",
    "ENVMOAT_DEBUG=1 logs to stderr (resolver.go:101-103)",
    "No symlink following — Phase 4 deferral confirmed (resolver.go:34 uses os.Stat)"
  ],
  "residualRisks": [
    "ResolveFromPWD has no direct test coverage",
    "7 parse tests leak temp directories (tmpDir() without *testing.T)",
    "ParseMarker returns MarkerDefault for error cases (API hygiene)",
    "FindWalkRoot does not enforce absolute path for ENVMOAT_WALK_ROOT"
  ],
  "noStagedFiles": true,
  "diffSummary": "No changes made — review-only task per hard constraints",
  "reviewFindings": [
    "no blockers",
    "bug (low): resolver.go:88 - ParseMarker returns MarkerDefault for unrecognized content error path",
    "suggestion: resolver_test.go - ResolveFromPWD has no dedicated test",
    "suggestion: resolver_test.go - 7 tests call tmpDir() without *testing.T, leaking temp dirs",
    "suggestion: resolver.go:95 - FindWalkRoot does not enforce absolute path for ENVMOAT_WALK_ROOT",
    "suggestion: resolver.go:101 - debug() reads env var on every call; consider caching"
  ],
  "manualNotes": "Implementation is correct and well-structured. The 4 findings are all non-blocking improvements. The test suite is comprehensive for ParseMarker and Resolve but missing direct coverage of ResolveFromPWD."
}
```
