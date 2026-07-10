# Progress: SM-gtv.4 Resolver Review Fixes

## Status: COMPLETE

## Completed
- [x] Fix 1: Added `TestResolveFromPWD` test
- [x] Fix 2: Fixed 7 temp dir leaks by passing `t` to `tmpDir()`
- [x] Fix 3: Added `MarkerUnknown` sentinel; error paths return `MarkerUnknown` instead of `MarkerDefault`
- [x] Fix 4: `FindWalkRoot` now calls `filepath.Abs` after `filepath.Clean`, with `"/"` fallback

## Changed Files
- `internal/resolver/resolver.go`
- `internal/resolver/resolver_test.go`

## Validation
All passing: `go build ./...`, `go vet ./...`, `go test ./internal/resolver/ -v -count=1` (18 tests)
