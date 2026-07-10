# Fix Report: SM-gtv.4 Resolver Review Findings

## Findings Fixed

### 1. `ResolveFromPWD` untested
- Added `TestResolveFromPWD` that creates a temp dir with a `profile: pwd-test` marker, `os.Chdir`s into it, calls `ResolveFromPWD()`, and verifies the result.

### 2. 7 tests leak temp dirs
- Changed `tmpDir()` → `tmpDir(t)` in all 7 tests:
  - `TestParseMarkerDisabled`
  - `TestParseMarkerDisabledWithWhitespace`
  - `TestParseMarkerProfile`
  - `TestParseMarkerProfileWithExtraWhitespace`
  - `TestParseMarkerEmptyProfileName`
  - `TestParseMarkerUnrecognized`
  - `TestParseMarkerCaseSensitive`

### 3. `ParseMarker` returns `MarkerDefault` for error path
- Added `MarkerUnknown` sentinel (iota value 3) to `MarkerContent` constants.
- Changed 3 error-return sites in `ParseMarker` from `MarkerDefault` → `MarkerUnknown`:
  - `os.ReadFile` error (line 88)
  - Empty profile name error (line 94)
  - Unrecognized content error (line 99)
- Added `MarkerUnknown` case to `describeContent()` returning `"unknown"`.

### 4. `FindWalkRoot` doesn't enforce absolute path
- After `filepath.Clean(root)`, now calls `filepath.Abs(root)`.
- If `filepath.Abs` fails, returns `"/"` as fallback.

## Changed Files
- `internal/resolver/resolver.go` — fixes 3 and 4
- `internal/resolver/resolver_test.go` — fixes 1 and 2

## Validation
- `go build ./...` — passed
- `go vet ./...` — passed
- `go test ./internal/resolver/ -v -count=1` — 18 tests passed
