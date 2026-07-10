# Architecture Review — Secrets Manager PLAN.md

**Date**: 2026-07-09  
**Scope**: Comprehensive review of PLAN.md covering Go+CGO, shell hooks, directory walk-up, CLI completeness (profiles, verify), file format, concurrency, atomic writes, HKDF bundle_id, marker parsing, debug mode, walk boundary, rotate two-phase, and remaining gaps.

---

## Prior Review Status

The previous ARCHITECTURE-REVIEW.md identified 6 blockers and 15 remaining issues. PLAN.md has been substantially updated since then. The bash `$REPLY` blocker is fixed (now uses `$BASH_COMMAND`). Several other issues are resolved. **New contradictions and gaps are identified below.**

---

## 1. Go + CGO — Solid

The plan correctly describes:
- `99designs/keychain` for basic operations, custom CGO only for `SecAccessControl`
- `kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly`
- Two-item Keychain pattern (protected + cache)

**Non-blocking notes:**
- CGO cross-compilation requires macOS host or osxcross for Intel + Apple Silicon binaries. CI must run on macOS.
- No `doctor`/`repair` command for Keychain item issues (orphaned items after password change, Time Machine restore).

---

## 2. Shell Hooks — Fixed, One Minor Gap

### Bash hook: ✅ Fixed
The `$REPLY` blocker from the prior review is resolved. The hook now uses:
```bash
[[ "$BASH_COMMAND" =~ ^cd(\ |$) ]] || return
```
This correctly matches `cd` and `cd /path` but guards against `cdpath`, `cdvar`, etc.

### zsh hook: ✅ Correct
Uses `add-zsh-hook chpwd` — fires only on directory change.

### Bundle-hash dedup: ✅ Good
Tracks `__secrets_manager_last_bundle` to avoid redundant `eval` when the bundle hasn't changed.

### 🔸 Gap: Interactive shell guard not in hook code
The plan states "Hook only installs in interactive shells (`[[ $- == *i* ]]`)" but the provided hook code snippets don't include this guard. The guard must wrap the `add-zsh-hook` / `trap` call, not the hook function itself. This is an installation concern, not a hook logic concern, but should be explicit in the plan.

---

## 3. Directory Walk-Up — Contradiction Found

### 🔴 Contradiction: Walk boundary — `$HOME` vs `/`

The plan contains **three conflicting statements**:

| Location | Statement |
|----------|-----------|
| Line 149 (walk algorithm) | `→ walk up from PWD looking for .secrets-manager marker (stops at $HOME)` |
| Line 164 (walk boundary) | `stops at \`/\` (root). Configurable via \`SECRETS_MANAGER_WALK_ROOT\` env var.` |
| Line 265 (resolved decisions) | `/ by default, configurable via SECRETS_MANAGER_WALK_ROOT` |
| Line 270 (resolved decisions) | `Stop at $HOME, validate marker content strictly` |

**Two resolved decisions contradict each other.** Line 265 says `/`, line 270 says `$HOME`. The walk algorithm (line 149) says `$HOME`. The walk boundary section (line 164) says `/`.

**Recommendation**: Pick one. `/` with `SECRETS_MANAGER_WALK_ROOT` override is more flexible and handles projects outside `$HOME` (e.g., `/opt`, `/shared`). Remove the `$HOME`-only entry.

---

## 4. CLI Completeness — Good, Minor Gaps

### Profiles: ✅ Present, Underspecified
`profiles list/create/delete/link/unlink` are all listed. However:
- `profiles link <path> <name>` — what happens if the profile doesn't exist? Does it auto-create?
- `profiles link` — does it create the bundle in `bundles/` or just the marker? The CLI comment says "ensure bundle exists in index.json" but doesn't specify if a new empty bundle is created.

### Verify: ✅ Present, Minimal Detail
`secrets-manager verify` is listed as "integrity check: verify all bundles decrypt successfully." Missing:
- Does it also validate `index.json` consistency (every index entry has a corresponding `.enc` file, every `.enc` file has an index entry)?
- Does it check file permissions?
- What exit codes? (0 = all good, non-zero = specific failures)

### Other gaps:
- **`get <KEY>`** — doesn't specify how it determines which bundle to read from. Does it walk up from PWD like `load`? Or require being in the project root?
- **No `doctor`/`repair`** — no command for diagnosing Keychain issues, permission problems, or corrupted state.
- **No `unset`/`clear`** — `set <KEY>` with no value should explicitly set to empty string, but this isn't specified.

---

## 5. File Format — Solid

```
[1B version=0x01][12B nonce][ciphertext][32B auth tag]
```

- ✅ Version byte for future migration
- ✅ Auth tag after ciphertext (RFC 8452)
- ✅ Metadata in plaintext JSON (`_meta` with timestamps)
- ✅ AES-256-GCM with 12-byte nonce

No issues.

---

## 6. Concurrency — Adequate

- ✅ Atomic writes: temp file + rename
- ✅ File locking: `flock` (BSD-compatible on macOS)
- ✅ Session cache in Keychain (not filesystem)

**Minor note:** `flock` on macOS uses BSD semantics. The Go `github.com/gofrs/flock` package handles this portably.

---

## 7. Atomic Writes — Solid

- ✅ Temp file + rename for bundle mutations
- ✅ `.bak` for `rotate` rollback
- ✅ File permissions (0600/0700) validated on startup

---

## 8. HKDF bundle_id — Resolved, One Concern

The plan now explicitly defines `bundle_filename` as the HKDF salt:
```
DEK = HKDF-SHA256(LUK, bundle_filename, info="secrets-manager/v1/dek")
```

**Concern**: The bundle filename is a hash of the project root path. If `index.json` is ever rebuilt or the bundle filename changes (e.g., path normalization difference), the old encrypted data becomes permanently unrecoverable because the derived DEK would differ. The plan should clarify that bundle filenames are **immutable** once created, or use a stable UUID stored in the index rather than a path-derived hash.

---

## 9. Marker Parsing — Specified

The plan now specifies:
- Content trimmed of whitespace and trailing newline
- Case-sensitive
- Accepts exactly: empty file (default), `disabled`, or `profile: <name>`
- All other content produces an error with the marker path

**Minor note**: Should a multi-line marker file be rejected (only first line parsed, or full content must match)? The plan says "content trimmed" which implies the entire file content is trimmed and checked — a multi-line file would fail validation, which is reasonable.

---

## 10. Debug Mode — ✅ Specified

`SECRETS_MANAGER_DEBUG=1` enables verbose stderr logging (directory walk, bundle resolution, Keychain access). Never logs secret values. Good.

---

## 11. Walk Boundary — See §3 (Contradiction)

The walk boundary is specified in multiple places with conflicting values. See the contradiction table in Section 3.

---

## 12. Rotate Two-Phase — Detailed, One Gap

The plan specifies:
> two-phase — (1) decrypt all bundles to temp dir, (2) encrypt all with new key to new bundles dir, (3) atomic rename `bundles/` → `bundles.bak/`, new → `bundles/`. On failure, restore from `.bak`

**Gap**: Step (3) is two renames. If the first rename succeeds (`bundles/` → `bundles.bak/`) but the second fails (new → `bundles/`), the `bundles/` directory doesn't exist and the system is in a broken state. The plan says "on failure, restore from `.bak`" but doesn't specify:
- How failure is detected (panic? error return?)
- How rollback is triggered (defer? explicit error handling?)
- Whether `bundles.bak/` is cleaned up on success

**Recommendation**: Add explicit rollback logic: wrap step (3) in a transaction where failure of either rename triggers `mv bundles.bak/ bundles/` and cleanup of the temp dir.

---

## 13. Session Independence Contradiction

The plan states:
- Line 79: "The Keychain cache item is **shared across all terminals**"
- Line 252: "**Session tokens**: Keychain two-item pattern (protected + cache), **independent per-terminal**"
- Line 262: "**Session TTL**: sliding window (resets on each access), **shared across terminals**"

Lines 79 and 262 agree (shared). Line 252 contradicts (independent). With a single cache item in Keychain, sessions **are shared** — this is the correct behavior for the two-item pattern. **Remove "independent per-terminal" from line 252.**

---

## 14. Remaining Gaps

| Area | Issue | Severity |
|------|-------|----------|
| Walk boundary | `$HOME` vs `/` contradiction (3 locations) | **Medium** |
| Session independence | "independent per-terminal" contradicts "shared across terminals" | **Low** |
| Interactive guard | Mentioned but not shown in hook code | **Low** |
| `profiles link` | Underspecified — auto-create profile? Create bundle? | **Low** |
| `verify` | Minimal detail — no index consistency check, no exit codes | **Low** |
| `get <KEY>` | Doesn't specify bundle resolution (walk-up like `load`?) | **Low** |
| Rotate rollback | Two renames in step (3) — no explicit failure/rollback handling | **Low** |
| HKDF bundle_id | Filename-as-salt means path change = data loss; no immutability guarantee | **Low** |
| No `doctor`/`repair` | No command for Keychain diagnostics or recovery | **Low** |
| No migration strategy | No config/bundle format migration between versions | **Low** |
| No secret size limit | Large secrets could break shell exports (~2MB macOS limit) | **Low** |
| No uninstall | No command to remove shell hook from rc file | **Low** |

---

## Summary

The plan is substantially improved from the prior review. The critical bash hook blocker is fixed. The encryption model, file format, concurrency strategy, and CLI surface are well-designed.

**Two contradictions need resolution before implementation:**
1. **Walk boundary**: `$HOME` vs `/` — pick `/` with `SECRETS_MANAGER_WALK_ROOT` override for flexibility
2. **Session independence**: "independent per-terminal" on line 252 contradicts the shared Keychain cache — remove "independent per-terminal"

**One operational gap to clarify:**
3. **Rotate rollback**: Step (3) has two renames; specify explicit error handling and rollback for partial failure

The remaining items are low-severity refinements addressable during implementation.
