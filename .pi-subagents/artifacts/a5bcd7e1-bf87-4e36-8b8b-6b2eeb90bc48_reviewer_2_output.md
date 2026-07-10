## Review

### Previous Findings — All Fixed

| # | Previous Finding | Status | Evidence |
|---|-----------------|--------|----------|
| 1 | ResolveFromPWD untested | **Fixed** | `TestResolveFromPWD` added (resolver_test.go:207). Creates temp dir with marker, Chdir's into it, asserts result. |
| 2 | 7 tests leaking temp dirs | **Fixed** | `tmpDir()` (resolver_test.go:243) now accepts variadic `*testing.T`, registers `t.Cleanup(func() { os.RemoveAll(dir) })`. All 18 test functions pass `t`. |
| 3 | ParseMarker returns MarkerDefault for error path | **Fixed** | Error paths return `MarkerUnknown`: read error (resolver.go:103), empty profile name (resolver.go:117), unrecognized content (resolver.go:123). |
| 4 | FindWalkRoot doesn't enforce absolute path | **Fixed** | `FindWalkRoot` (resolver.go:131-135) calls `filepath.Abs(root)`, falls back to `"/"` on error. |

### Test Results

```
go test ./internal/resolver/ -v -count=1
18 tests — all PASS (0.363s)
```

### New Issues

- **Note** — Tests that mutate shared state (`ENVMOAT_WALK_ROOT` in `TestResolveNoMarker`, `TestFindWalkRootDefault`, `TestFindWalkRootEnvVar`; `os.Chdir` in `TestResolveFromPWD`) are not safe for `t.Parallel()`. Not a blocker since tests pass sequentially, but worth noting if parallelism is added later.

### Verdict

**PASS — no blockers.** All four previous findings are correctly fixed. No new issues beyond a minor parallelism note.
