# Wave 1: Resolver — Directory Walk + Marker Resolution

## Summary

Implemented `internal/resolver/` package for envmoat's directory walk and `.envmoat` marker resolution.

## API

```go
package resolver

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

## Implementation Details

### Walk behavior
1. `filepath.Clean(dir)` — canonicalize input (no symlink resolution yet, Phase 4)
2. Walk up from `dir` checking for `.envmoat` at each level
3. Stop at `/` or `ENVMOAT_WALK_ROOT` env var
4. `Resolve` returns `nil, nil` when no marker found (not an error)

### Marker parsing
- Content trimmed with `strings.TrimSpace` (whitespace + trailing newline)
- Case-sensitive matching
- Empty → `MarkerDefault` (auto bundle)
- `disabled` → `MarkerDisabled` (stop, no bundle)
- `profile: <name>` → `MarkerProfile` with extracted profile name
- Anything else → error with descriptive message

### Debug mode
- `ENVMOAT_DEBUG=1` enables verbose stderr logging of walk steps
- Never logs secret values

## Files Created

| File | Lines | Description |
|------|-------|-------------|
| `internal/resolver/resolver.go` | ~140 | Core implementation |
| `internal/resolver/resolver_test.go` | ~200 | 17 unit tests |

## Test Coverage

| Test | What it covers |
|------|----------------|
| `TestParseMarkerEmpty` | Empty file → MarkerDefault |
| `TestParseMarkerWhitespaceOnly` | Whitespace-only → MarkerDefault |
| `TestParseMarkerDisabled` | "disabled\n" → MarkerDisabled |
| `TestParseMarkerDisabledWithWhitespace` | "  disabled  \n" → MarkerDisabled |
| `TestParseMarkerProfile` | "profile: myapp-dev\n" → MarkerProfile |
| `TestParseMarkerProfileWithExtraWhitespace` | Whitespace around profile name trimmed |
| `TestParseMarkerEmptyProfileName` | "profile: " → error |
| `TestParseMarkerUnrecognized` | Random text → error |
| `TestParseMarkerCaseSensitive` | "DISABLED" → error (case-sensitive) |
| `TestParseMarkerNonexistent` | Missing file → error |
| `TestResolveFindsMarker` | Marker in starting dir |
| `TestResolveWalksUp` | Marker in ancestor directory |
| `TestResolveStopsAtDisabled` | Disabled marker stops walk |
| `TestResolveNoMarker` | No marker → nil result |
| `TestFindWalkRootDefault` | Default "/" when env var unset |
| `TestFindWalkRootEnvVar` | ENVMOAT_WALK_ROOT honored |
| `TestMarkerContentValues` | iota values correct |

## Validation

```
$ go build ./internal/resolver/
# success — no output

$ go test -v ./internal/resolver/
# 17 tests, all PASS, 0.388s
```

## Constraints Honored

- ✅ No bundle decryption implemented
- ✅ No symlink canonicalization (Phase 4)
- ✅ Package only finds and parses markers
- ✅ Uses `.envmoat` marker file (not `.secrets-manager`)
- ✅ Uses `ENVMOAT_WALK_ROOT` and `ENVMOAT_DEBUG` env vars

## Known Issues

- `go build ./...` fails due to pre-existing missing dependencies in other packages (`github.com/99designs/keychain`, `golang.org/x/crypto/scrypt`, `github.com/gofrs/flock`, `gopkg.in/yaml.v3`). This is not caused by the resolver package.
