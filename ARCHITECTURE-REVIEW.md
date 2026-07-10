# Architecture Review — Secrets Manager (v3)

**Date**: 2026-07-09  
**Scope**: Comprehensive review of PLAN.md — Go+CGO, shell hooks, directory walk-up, CLI, file format, concurrency, atomic writes, remaining gaps.

---

## Prior Review Status

The original review (v2) identified 15 issues. PLAN.md has been substantially updated to address the majority:

| Prior Issue | Status | Notes |
|---|---|---|
| Shell `eval` of unescaped values | ✅ Resolved | `%q`-style escaping (`strconv.QuoteToASCII`) specified |
| Per-file salt + single cached key contradiction | ✅ Resolved | Single global salt + HKDF per-bundle DEK — clean model |
| No version byte | ✅ Resolved | `0x01` version byte added |
| Bash `REPLY` guard broken | ✅ Resolved | Switched to `$BASH_COMMAND` regex `^cd(\| $)` |
| No file locking | ✅ Resolved | `flock` specified |
| Symlink handling undefined | ✅ Resolved | `realpath` canonicalization specified |
| Walk boundary at `$HOME` | ✅ Resolved | Stops at `/`, configurable via `SECRETS_MANAGER_WALK_ROOT` |
| Marker parsing unspecified | ✅ Resolved | Whitespace-trimmed, case-sensitive, strict validation |
| No `verify` command | ✅ Resolved | `secrets-manager verify` added |
| No `doctor` command | ✅ Resolved | `secrets-manager doctor` added |
| `link` underspecified | ✅ Resolved | `profiles link <path> <name>` + `profiles unlink <path>` specified |
| No debug mode | ✅ Resolved | `SECRETS_MANAGER_DEBUG=1` specified |
| Shell hook installation | ✅ Resolved | `setup` (idempotent), `uninstall` (removal), `doctor` (diagnostics) |
| Per-terminal session ambiguity | ✅ Resolved | Explicitly "shared across terminals" as a feature |
| `bundle_id` for HKDF | ✅ Resolved | `bundle_filename` used as HKDF input |
| Rotate atomicity | ✅ Resolved | Two-phase with `.bak` rollback specified |

---

## 1. Shell Hook — Critical Bash Timing Bug

### 🔴 Blocker: Bash `DEBUG` trap fires **before** `cd` executes

The bash hook uses a `DEBUG` trap:

```bash
trap '__secrets_manager_hook' DEBUG
```

The `DEBUG` trap fires **before** each command executes. When the user runs `cd /new/dir`:

1. `__secrets_manager_hook` fires — `$BASH_COMMAND` is `"cd /new/dir"`
2. Hook calls `secrets-manager load` — resolves **current** PWD (the **old** directory)
3. The wrong bundle loads (or no bundle if already outside the project)
4. `cd /new/dir` executes — directory changes **after** secrets were loaded

**This means bash users never get the correct secrets on `cd`.** The zsh `chpwd` hook works correctly because it fires **after** the directory change.

**Fix options:**

1. **Use `PROMPT_COMMAND` instead** (recommended):
   ```bash
   __secrets_manager_hook() {
     local output
     output=$(secrets-manager load 2>/dev/null)
     if [ -n "$output" ]; then
       local bundle_hash
       bundle_hash=$(echo "$output" | head -1 | cut -d: -f2)
       if [ "$bundle_hash" != "$__secrets_manager_last_bundle" ]; then
         eval "$output"
         __secrets_manager_last_bundle="$bundle_hash"
       fi
     fi
   }
   PROMPT_COMMAND='__secrets_manager_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}'
   ```
   Fires after command execution, so PWD is correct. The bundle-hash dedup check makes the per-prompt cost negligible.

2. **Override `cd` function** (also works):
   ```bash
   cd() { builtin cd "$@" && __secrets_manager_hook; }
   ```
   More robust but risks conflicts with other `cd` overrides.

**Recommendation**: `PROMPT_COMMAND` — it fires after `cd` completes, PWD is correct, and the dedup check prevents redundant work.

### Other shell hook observations

- **`2>/dev/null` suppresses warnings**: The hook silences all stderr from `load`. The plan specifies useful warnings for missing/corrupted bundles ("Bundle not found. Run `secrets-manager init <path>`"). These warnings are **never shown** because `2>/dev/null` eats them. Consider removing stderr suppression or using a one-time warning mechanism.
- **`DEBUG` trap performance** (if kept): Fires before every command. With `$BASH_COMMAND` guard, short-circuit is fast, but function call overhead on every command is non-zero. `PROMPT_COMMAND` is comparable (fires every prompt) but only after user input.

---

## 2. Document Consistency — Contradiction in Open Questions

### ⚠️ Medium: Open question contradicts resolved decision

The "Open Questions" section lists:

> 1. **Marker file format**: Plain text (`disabled`, `profile: name`) vs YAML/JSON for extensibility?

But the "Resolved Design Decisions" section already resolves this:

> - **Marker parsing**: whitespace-trimmed, case-sensitive, strict validation.

The plan has chosen plain text, but the open question remains listed. Either remove the question from "Open Questions" or acknowledge the trade-off (plain text is simpler but less extensible).

---

## 3. Secret Size Limits — Not Addressed

### ⚠️ Medium: No enforcement or warning for large secrets

Shell `export` has practical limits (~2MB on macOS). Large secrets (TLS certificates, SSH keys, large JSON configs) could:
- Break the `eval` in the hook
- Cause shell instability
- Slow down every prompt (if using `PROMPT_COMMAND`)

**Recommendation**: Add a configurable size limit (e.g., 64KB default) with a warning on `set`. For larger secrets, suggest using `get --clip` or file-based access instead.

---

## 4. Config/Index Migration — Not Addressed

### ⚠️ Low-Medium: No `config.yaml` or `index.json` migration strategy

The plan has:
- ✅ Version byte in bundle file format (enables detection)
- ✅ `rotate` for password changes
- ❌ No `config.yaml` schema migration
- ❌ No `index.json` schema migration

If `config.yaml` gains new fields or `index.json` changes structure between versions, the tool needs either:
- Backward-compatible defaults for missing fields
- An explicit migration step on startup
- A version field in `config.yaml` (currently only `index.json` has `"version": 1`)

**Recommendation**: Add a `version` field to `config.yaml` and implement forward-compatible defaults for missing config fields.

---

## 5. Docker Integration — Limited Scope

### ⚠️ Low: Only `.env` file generation, no `secrets:` directive support

The plan generates `.env.secrets` files. Docker Compose also supports the `secrets:` directive, which mounts secrets as files (more secure — no env var leakage into container processes).

**Recommendation**: Document the `.env` approach as the primary path (matches the tool's env-var focus) but note `secrets:` as a future enhancement. Not a blocker.

---

## 6. `flock` Implementation Detail

### ⚠️ Low: No Go package specified

The plan mentions `flock` (BSD-compatible on macOS) but doesn't specify the Go package. The standard choice is `github.com/gofrs/flock`, which handles BSD vs. POSIX semantics portably.

**Recommendation**: Add `github.com/gofrs/flock` to the dependency list.

---

## 7. Keychain Orphan Handling

### ⚠️ Low: No repair path for Keychain issues

The plan mentions `doctor` for diagnostics and `rotate` for password changes. But if the user changes their login password or restores from Time Machine, Keychain items can become orphaned (the old Keychain is no longer accessible).

**Recommendation**: Add a `repair` or `doctor --fix` subcommand that recreates Keychain items after re-auth. Mention this in the `doctor` output when orphaned items are detected.

---

## 8. `index.json` Concurrency

### ⚠️ Low: No file locking for `index.json`

The plan specifies `flock` for bundle writes but doesn't mention locking for `index.json` mutations. If two processes simultaneously run `init` or `profiles link`, the index could be corrupted.

**Recommendation**: Apply the same `flock` pattern to `index.json` writes.

---

## 9. Rotate Rollback Specifics

### ⚠️ Low: Rollback procedure underspecified

The plan says "two-phase — decrypt all → encrypt all → atomic rename with `.bak` rollback." But the exact procedure isn't detailed:

- Does it decrypt all bundles first, then encrypt all with the new key?
- Where are the new bundles written before the rename?
- What happens if encryption fails halfway through?

**Recommendation**: Specify the procedure explicitly:
1. Decrypt all bundles → plaintext temp dir
2. Encrypt all plaintext → new encrypted temp dir
3. Atomic swap: rename `bundles/` → `bundles.bak/`, rename temp dir → `bundles/`
4. If any step fails, restore from `bundles.bak/`

---

## Summary of Remaining Issues

| # | Area | Issue | Severity |
|---|---|---|---|
| 1 | Shell hook (bash) | `DEBUG` trap fires before `cd` — secrets load for wrong directory | **Critical** |
| 2 | Shell hook (bash) | `2>/dev/null` suppresses bundle warnings the plan specifies | **Medium** |
| 3 | Document consistency | Open question contradicts resolved decision on marker format | **Low** |
| 4 | Secret size | No limit enforcement or warning for large values | **Medium** |
| 5 | Migration | No `config.yaml` version/migration strategy | **Low-Medium** |
| 6 | Docker | `secrets:` directive not considered (`.env` only) | **Low** |
| 7 | `flock` | No Go package specified | **Low** |
| 8 | Keychain | No repair path for orphaned items | **Low** |
| 9 | Concurrency | No `flock` for `index.json` writes | **Low** |
| 10 | Rotate | Rollback procedure underspecified | **Low** |

### Verdict

**One critical blocker remains**: The bash `DEBUG` trap fires **before** `cd` executes, so `secrets-manager load` resolves the old PWD. Use `PROMPT_COMMAND` instead (or override `cd`) so the hook fires after the directory change.

The rest of the architecture is solid. The encryption model, Keychain two-item pattern, file format, CLI surface, and directory walk-up logic are well-designed. The remaining issues are refinements addressable during implementation.
