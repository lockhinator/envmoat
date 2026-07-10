# Architecture Review — Secrets Manager PLAN.md

**Date**: 2026-07-09  
**Scope**: Comprehensive review of PLAN.md covering Go+CGO, shell hooks, directory walk-up, CLI, file format, concurrency, atomic writes, profiles subcommand, and remaining gaps.

---

## Prior Review Status

The previous ARCHITECTURE-REVIEW.md identified 6 blockers. PLAN.md has been substantially updated. All prior blockers are resolved:

| Prior Issue | Status |
|---|---|
| Shell `eval` of unescaped values | ✅ Resolved — `%q`-style escaping specified |
| Per-file salt + single cached key contradiction | ✅ Resolved — single global salt + HKDF per-bundle DEK |
| No version byte | ✅ Resolved — `0x01` version byte added |
| `REPLY` guard broken in bash hook | ✅ Resolved — switched to `$BASH_COMMAND` |
| No file locking | ✅ Resolved — `flock` specified |
| Symlink handling undefined | ✅ Resolved — `realpath` canonicalization |

---

## 1. Go + CGO Approach — ✅ Solid

**Assessment**: Well-designed. Using `99designs/keychain` for basic Keychain operations with custom CGO only for `SecAccessControl` is the right balance — avoids reinventing basic Keychain plumbing while retaining full biometric control.

**Notes (non-blocking):**
- **CGO cross-compilation**: Building Intel + Apple Silicon requires macOS host or `osxcross`. CI must run on macOS runners. Manageable but plan for it.
- **Keychain orphan on login password change / Time Machine restore**: No `doctor`/`repair` subcommand to recover orphaned Keychain items. Consider adding in Phase 7.
- **Per-terminal session independence**: Plan says "independent per-terminal" but the two-item Keychain cache is shared across all terminals using the same Keychain. One terminal's Touch ID unlock benefits all terminals. This is arguably a feature (convenience), but if truly independent sessions are desired, the cache item needs a terminal-specific key (e.g., keyed by `$TTY`).

---

## 2. Shell Hook — ⚠️ Minor Pattern Bug in Bash

### Bash `$BASH_COMMAND` guard too broad

The plan uses `[[ "$BASH_COMMAND" == cd* ]]` which matches any command starting with `cd`:

```
cd /tmp        → MATCHED ✅ (correct)
cd             → MATCHED ✅ (correct)
cdpath_test    → MATCHED ❌ (false positive)
cdsomevar=1    → MATCHED ❌ (false positive)
```

While unlikely in practice, this is a correctness issue. The hook would invoke `secrets-manager load` on every `cdpath_*` or `cd*var` command.

**Fix**: Use a regex or explicit pattern:
```bash
[[ "$BASH_COMMAND" =~ ^cd(\ |$) ]] || return
```
This matches only `cd`, `cd /path`, `cd -`, etc.

### zsh `chpwd` hook — ✅ Correct

Fires only on directory change. No issues.

### Other observations
- **Bundle-hash dedup**: ✅ Tracking `bundle_hash` avoids redundant reloads.
- **Interactive shell guard**: ✅ `[[ $- == *i* ]]` check mentioned.
- **`DEBUG` trap performance**: Fires before every command in bash, but the short-circuit is fast. Acceptable; document if needed.
- **Silent failure when binary missing**: `2>/dev/null` suppresses all stderr. If a `.secrets-manager` marker exists but `load` fails (e.g., binary uninstalled), the user gets no feedback. Consider a one-time warning.

---

## 3. Directory Walk-Up — ✅ Solid, Minor Boundary Concern

**Assessment**: Well-designed. `realpath` canonicalization, walk-up from PWD, marker file detection, and `disabled`/`profile:` override support are all correct.

**Notes:**
- **Walk boundary at `$HOME`**: Projects outside `$HOME` (e.g., `/opt/projects`, `/shared`, `/Volumes/external`) won't work. Consider stopping at filesystem root (`/`) or first mount point instead, or make the boundary configurable.
- **Marker parsing edge cases**: Plan says "validate marker content strictly" but doesn't specify:
  - Does `Profile: name` (capital P) match?
  - Does `disabled ` (trailing whitespace) match?
  - Does `profile: name\n` (trailing newline from `echo`) parse correctly?
  - **Recommendation**: Trim whitespace, case-insensitive keyword matching, single-line parsing.
- **Walk performance**: For deeply nested paths (10+ levels), each `cd` triggers a walk. Bundle-hash dedup mitigates repeated prompts in the same directory. Consider caching the last resolved marker path in a hook variable for subdirectory navigation.

---

## 4. CLI Completeness — ✅ Good, Minor Gaps

**Assessment**: Comprehensive surface covering setup, project management, secret operations, profiles, shell integration, Docker, and migration.

**Remaining gaps:**

| Gap | Severity | Recommendation |
|---|---|---|
| No `verify`/`integrity-check` | Medium | Add `secrets-manager verify` to audit encrypted store for corruption |
| No `doctor`/`repair` | Low | For Keychain orphan recovery, permission fixes |
| `profiles link` underspecified | Low | Clarify: creates marker file? Updates `index.json`? Creates new bundle or references existing? |
| No secret size limit enforcement | Low | Shell exports have ~2MB limit on macOS; large secrets (TLS certs) could break the hook |
| No `set --clear` for empty values | Low | `remove` deletes keys, but no way to explicitly set a key to empty string |
| No fish shell support | Low | Fish is popular on macOS; worth noting for Phase 6 |
| No debug/verbose mode | Low | `load` failures are silent by design; `SECRETS_MANAGER_DEBUG=1` or `--verbose` aids troubleshooting |

---

## 5. File Format Design — ✅ Solid

**Assessment**: Clean, versioned, well-structured format:
```
[1B version=0x01][12B nonce][ciphertext][32B auth tag]
```

- ✅ Version byte for future migration
- ✅ Auth tag after ciphertext (RFC 8452 convention)
- ✅ Metadata in plaintext JSON (`_meta` with `created_at`, `updated_at`)
- ✅ Single global salt + HKDF per-bundle DEK — clean compromise model

**Note:**
- **`bundle_id` for HKDF not explicitly defined**: Plan says `LUK + bundle_id → HKDF-SHA256 → DEK` but doesn't specify what `bundle_id` is — the path hash? The bundle filename? This must be explicitly defined to ensure deterministic key derivation. **Recommendation**: Define `bundle_id` as the path hash (same as the filename prefix).

---

## 6. Concurrency — ✅ Adequate

- ✅ Atomic writes: temp file + rename pattern
- ✅ File locking: `flock` for concurrent `set`/`remove`
- ✅ Session cache in Keychain (not filesystem) — avoids PID reuse issues

**Note:**
- **`flock` on macOS**: BSD `flock` semantics differ slightly from POSIX. The Go `github.com/gofrs/flock` package handles this portably. Test specifically on macOS.

---

## 7. Atomic Writes — ✅ Solid

- ✅ Temp file + rename for bundle mutations
- ✅ `.bak` for `rotate` rollback
- ✅ File permissions (0600/0700) validated on startup

**Note:**
- **`rotate` multi-step atomicity**: Rotating all bundles is inherently multi-step. If rotation fails halfway, `.bak` files should allow full rollback. **Recommendation**: Decrypt all → encrypt all with new key → write new bundles to temp dir → atomic rename of entire `bundles/` directory (rename old to `.bak`, rename temp to `bundles/`).

---

## 8. Profiles Subcommand — ✅ Defined, Needs Specification Detail

The `profiles` subcommand surface is complete:
- `profiles list` — list all named profiles
- `profiles create <name>` — create a new named profile
- `profiles delete <name>` — delete a profile
- `profiles link <path> <name>` — link a directory to a profile
- `profiles unlink <path>` — remove a profile link

**Underspecified**: What `profiles link` does exactly — does it create the `.secrets-manager` marker file with `profile: <name>`? Does it update `index.json`? Does it create a new bundle or reference an existing named profile's bundle? This needs clarification for implementation.

---

## 9. Remaining Gaps Not Previously Noted

### Distribution & Installation
- Shell hook installation into `~/.zshrc`/`~/.bashrc` needs idempotency check (don't install duplicate hooks)
- Uninstallation that removes the hook cleanly
- Consider oh-my-zsh plugin structure for zsh

### Migration & Versioning
- No config migration strategy if `config.yaml` format changes
- No bundle format migration mechanism (version byte enables detection, but no migration specified)
- No `index.json` schema versioning

### Observability
- No structured logging or debug mode
- No audit log (mentioned as "nice to have" in Phase 7)

### Docker Compose
- Only `.env` file generation supported; Docker Compose `secrets:` directive (mounts secrets as files, more secure) not considered

---

## Summary

| # | Area | Issue | Severity |
|---|---|---|---|
| 1 | Bash hook | `cd*` pattern too broad — matches `cdpath_test`, `cdsomevar`, etc. | **Medium** |
| 2 | Walk-up | `$HOME` boundary excludes projects outside home directory | Low |
| 3 | Marker parsing | Whitespace, case sensitivity, trailing newline handling unspecified | Low |
| 4 | CLI | No `verify`/`integrity-check` command | Medium |
| 5 | CLI | No debug/verbose mode for troubleshooting | Low |
| 6 | File format | `bundle_id` for HKDF derivation not explicitly defined | Low |
| 7 | Profiles | `profiles link` underspecified (marker creation? index update?) | Low |
| 8 | Keychain | Per-terminal session independence contradicts shared Keychain cache | Low |
| 9 | `rotate` | Multi-step atomicity needs explicit rollback procedure | Low |
| 10 | Concurrency | `flock` BSD semantics on macOS — test specifically | Low |
| 11 | Distribution | Shell hook installation idempotency/uninstallation undefined | Low |
| 12 | Migration | No config/bundle/index migration strategy | Low |
| 13 | CLI | No fish shell support | Low |
| 14 | CLI | No secret size limit enforcement/warning | Low |

### Verdict

**No critical blockers.** The prior `REPLY` guard blocker is fixed. The plan is architecturally sound:

- **Go+CGO**: ✅ Clean approach, `99designs/keychain` + custom CGO for biometrics
- **Shell hooks**: ⚠️ Bash `cd*` pattern needs tightening (`^cd(\ |$)` regex)
- **Directory walk-up**: ✅ Correct, boundary at `$HOME` is the only concern
- **CLI**: ✅ Comprehensive, minor gaps (`verify`, debug mode)
- **File format**: ✅ Well-designed, `bundle_id` needs explicit definition
- **Concurrency**: ✅ Atomic writes + `flock` — adequate
- **Atomic writes**: ✅ Solid, `rotate` needs explicit rollback procedure
- **Profiles**: ✅ Surface defined, `link` behavior needs specification

**Highest priority fix**: Tighten the bash hook pattern from `cd*` to `^cd(\ |$)` to prevent false positives on commands starting with "cd".

**All other issues are low-severity refinements** addressable during implementation.
