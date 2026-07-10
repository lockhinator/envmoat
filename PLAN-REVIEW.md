# Plan Review — Secrets Manager

**Date**: 2026-07-09  
**Scope**: Comprehensive review of PLAN.md — security, architecture, UX, contradictions, edge cases, attack vectors, completeness.

---

## Prior Review Status

ARCHITECTURE-REVIEW.md identified 10 issues (1 critical, 3 medium, 6 low). PLAN.md has been updated to address all of them:

| Prior Issue | Status | How Resolved |
|---|---|---|
| Bash DEBUG trap fires before cd | ✅ Resolved | Switched to PROMPT_COMMAND |
| `2>/dev/null` suppresses warnings | ✅ Resolved | Hook code has no stderr suppression |
| Open question contradicts resolved decision | ✅ Resolved | Open questions now only list clipboard backend |
| Secret size limits | ✅ Resolved | 1MB warning added |
| Config migration strategy | ⚠️ Partial | `index.json` has version; `config.yaml` does not |
| Docker `secrets:` directive | ✅ Resolved | Documented as future enhancement |
| `flock` Go package | ✅ Resolved | `github.com/gofrs/flock` specified |
| Keychain orphan repair | ❌ Not addressed | No repair path documented |
| `index.json` flock | ✅ Resolved | "index.json writes are also locked" |
| Rotate rollback specifics | ⚠️ Partial | "two-phase atomic, .bak rollback" stated but not detailed |

---

## 1. Security Review

### Encryption Model ✅ Solid
- **scrypt** (N=262144, r=8, p=1) — OWASP 2026 recommended minimum
- **Single global salt + HKDF per-bundle DEK** — clean key hierarchy, compromise containment
- **AES-256-GCM** with explicit nonce and auth tag — correct authenticated encryption
- **Version byte** (0x01) — enables future format migration
- **File format**: `[1B version][12B nonce][ciphertext][32B auth tag]` — RFC 8452 convention

**Note**: Global salt in `config.yaml` is not secret (standard for KDF salts). This is correct but should be documented as such.

### Authentication ✅ Solid
- **Keychain two-item pattern** — protected (Touch ID) + cache (TTL) — well-designed
- **Sliding TTL** — resets on each access, configurable
- **Shared across terminals** — documented trade-off (one terminal compromise exposes all)
- **Fallback to Mac login password** — documented as known limitation (macOS behavior)
- **Fixed Keychain labels** — enables reliable find/delete by `rotate`, `logout`, `doctor`

**Edge case**: If user changes macOS login password, Keychain items may become inaccessible. No `repair` command documented. Low risk but worth noting.

### Shell Injection ✅ Solid
- **Single-quote escaping** with `'\''` for internal quotes — correct, handles all shell metacharacters
- **`$()`, backticks, newlines** inert inside single quotes — correct
- **Errors to stderr**, `load` exits 0 with no output when no bundle — correct
- **Interactive shell guard** (`[[ $- == *i* ]]`) — prevents injection in scripts
- **No `2>/dev/null`** — warnings for missing/corrupted bundles will display

### File Permissions ✅ Solid
- 0700 directories, 0600 files — correct
- Validated on startup — refuses to operate if too open
- Docker `.env.secrets` also 0600 — correct

### Attack Vectors Considered
- **Bundle rename breaks DEK**: DEK derived from `bundle_filename`. Renaming a `.enc` file would break decryption. This is acceptable — bundles are managed by the tool, not renamed manually.
- **Keychain item collision**: Fixed labels under service `secrets-manager`. macOS Keychain is per-user, so no cross-user collision. Single-user assumption is reasonable.
- **Shell history leakage**: `get <KEY>` output to stdout could appear in history. Mitigated by `get --clip` (pbcopy, stderr-only confirmation). General shell issue, not tool-specific.
- **Secure erase on SSDs**: `uninstall` multi-pass overwrite is ineffective on SSDs with TRIM/wear-leveling. Modern best practice is relying on encryption at rest (FileVault) rather than overwriting. Worth noting as a limitation.

---

## 2. Architecture Review

### Storage Layout ✅ Solid
- Central store (`~/.secrets-manager/`) outside project directories — prevents accidental commit
- `index.json` with `profiles`, `links`, `auto` sections — clean separation of concerns
- Atomic writes (temp file + rename) — prevents corruption
- `flock` for both bundle and index.json writes — prevents concurrent mutation

### Directory Inheritance ✅ Solid
- Walk up from PWD looking for `.secrets-manager` marker — clean, intuitive
- Marker content: empty (default), `disabled`, `profile: <name>` — simple, well-specified
- Symlink handling via `realpath` — correct
- Walk boundary at `/`, configurable via `SECRETS_MANAGER_WALK_ROOT` — good
- Auto `.gitignore` entry for marker — prevents accidental commit

### CLI Surface ✅ Comprehensive
- Well-organized command hierarchy (setup, project, secrets, profiles, shell, docker, password)
- `--help`/`--version` on every subcommand — good
- Destructive ops require confirmation; `-y`/`--yes` to bypass — good
- Context resolution consistent across `set`/`get`/`list`/`remove`/`edit` — walks up from PWD
- `status` command shows active profile, TTL, Keychain state — good for debugging
- `doctor` command for diagnostics — comprehensive

### Concurrency ✅ Solid
- `flock` (BSD-compatible via `github.com/gofrs/flock`) for bundle writes
- `flock` for index.json writes
- Atomic temp+rename pattern

### Shell Integration ✅ Solid
- **zsh**: `chpwd` hook — fires after directory change, correct timing
- **bash**: `PROMPT_COMMAND` — fires after command execution, PWD is correct
- **Bash 5.1+ array form** with string fallback for older versions — correct
- **Bundle hash dedup** prevents redundant `eval` — efficient
- **No output** when no bundle found — zero overhead outside tracked dirs

---

## 3. UX Review

### Setup Flow ✅ Good
- `setup` creates master password + installs shell hook — single command
- Idempotent — re-running re-installs hook only
- `--reset` flag to change password — clear
- No-args shows welcome + usage, prompts to run `setup` if not configured — friendly

### Error Messages ✅ Good
- All errors include actionable recovery hint
- Debug mode via `SECRETS_MANAGER_DEBUG=1` — stderr only, never logs values
- Load error paths well-specified (no marker = silent, missing bundle = warning, corrupt = warning)

### Edge Cases ✅ Mostly Covered
- Secret size warning at 1MB (practical shell limit ~2MB)
- `edit` temp file in `~/.secrets-manager/` (not `/tmp`), mode 0600, deleted on exit
- `get --clip` uses pbcopy, confirmation on stderr only
- `rotate` two-phase with `.bak` rollback
- `verify` for integrity checks and orphan cleanup
- `profiles link --force` to overwrite existing marker

---

## 4. Issues That Would Block Implementation or Cause Bugs

### Issue 1: Bash Hook Duplicate Installation ⚠️ Bug

**Location**: Shell Integration section / `setup` idempotency claim

**Problem**: The `setup` command claims to be idempotent ("re-running re-installs hook only"). However, the hook installation code does not check whether `__secrets_manager_hook` is already defined in the rc file. Running `setup` twice would append duplicate hook entries to `PROMPT_COMMAND`, causing `secrets-manager load` to run multiple times per prompt.

**Fix**: Before appending the hook to the rc file, check if `__secrets_manager_hook` is already defined:
```bash
grep -q '__secrets_manager_hook' ~/.bashrc || { /* append hook */ }
```

### Issue 2: CLI Surface Inconsistency — `docker-compose` vs `docker-env` ⚠️ Bug

**Location**: CLI Surface section vs Resolved Design Decisions

**Problem**: The CLI Surface section lists:
```
secrets-manager docker-compose [output-path]
```
But the Resolved Design Decisions state:
> - **docker-env command**: renamed from `docker-compose` for clarity. `docker-compose` kept as alias.

The CLI surface section was not updated to reflect this rename. An implementer reading only the CLI section would implement `docker-compose` as the primary command name.

**Fix**: Update CLI surface to show `docker-env [output-path]` with `docker-compose` as alias.

### Issue 3: Auto-Bundle Naming Convention Unspecified ⚠️ Ambiguity

**Location**: `init` command specification / index.json `auto` section

**Problem**: `init` creates an "auto-named bundle" but the naming convention is not specified. The index.json schema shows `auto-simple.enc` as an example, but the derivation logic is missing. Is it:
- `auto-<hash-of-path>.enc`?
- `auto-<basename-of-path>.enc`?
- Random UUID?

Without this, `init` cannot be implemented deterministically. The `auto` section maps paths to bundle filenames, so the naming must be reproducible.

**Fix**: Specify the naming convention explicitly. E.g., `auto-<sha256(path)[:8]>.enc` or `auto-<sanitized-dirname>.enc`.

---

## 5. Refinement Issues (Not Blockers, Addressable During Implementation)

### Issue 4: `profiles link --force` Behavior Underspecified

The plan says `profiles link` "errors if marker already exists (use --force to overwrite)." It does not specify whether `--force` overwrites only the marker file, or also the bundle if the target profile differs. If a marker points to `profile: dev` and the user runs `profiles link --force <path> staging`, should the `dev` bundle be deleted?

**Recommendation**: `--force` overwrites the marker file only. The old profile's bundle remains (orphan cleanup via `verify`).

### Issue 5: `rotate` Rollback Procedure Underspecified

The plan states "two-phase atomic, .bak rollback" but does not detail the exact procedure for partial failure. If encryption of bundle 3 of 10 fails, what is the rollback state?

**Recommendation**: Specify explicitly:
1. Decrypt all bundles → plaintext temp dir (fail-fast: if any decrypt fails, abort)
2. Encrypt all plaintext → new encrypted temp dir (fail-fast: if any encrypt fails, delete temp dirs and abort)
3. Atomic swap: rename `bundles/` → `bundles.bak/`, rename temp dir → `bundles/`
4. Delete `bundles.bak/` on success

### Issue 6: No Keychain Repair Path

If the user changes their macOS login password or restores from Time Machine, Keychain items can become inaccessible. The plan has `doctor` for diagnostics but no repair command.

**Recommendation**: Add `secrets-manager repair` or `doctor --fix` that recreates Keychain items after re-auth.

### Issue 7: `config.yaml` Missing Version Field

`index.json` has `"version": 1` for schema migration. `config.yaml` has no version field. If config schema changes between versions, there is no migration path.

**Recommendation**: Add `"version"` to `config.yaml` and implement forward-compatible defaults for missing fields.

### Issue 8: `doctor` FileVault Check May Require Admin

`fdesetup status` may require admin privileges or may hang if the user lacks permissions. The `doctor` command should handle this gracefully.

**Recommendation**: Wrap `fdesetup status` in a timeout and handle permission errors gracefully with a non-blocking warning.

---

## 6. Completeness Check

| Feature Area | Status | Notes |
|---|---|---|
| Setup / uninstall | ✅ Complete | Idempotent setup, full uninstall with secure erase |
| Master password | ✅ Complete | scrypt, Keychain storage, rotate support |
| Touch ID | ✅ Complete | Two-item pattern, sliding TTL, shared sessions |
| Secret CRUD | ✅ Complete | set, get, list, remove, edit, --stdin, --file, --clip |
| Shell integration | ✅ Complete | zsh chpwd, bash PROMPT_COMMAND, interactive guard |
| Directory inheritance | ✅ Complete | Walk-up, marker files, disabled/override, symlinks |
| Profiles | ✅ Complete | Named bundles, link/unlink, list/create/delete |
| Docker | ✅ Complete | .env.secrets generation, 0600, auto .gitignore |
| Import/Export | ✅ Complete | backup/restore with encryption, merge on conflict |
| Diagnostics | ✅ Complete | verify, doctor, status, debug mode |
| Concurrency | ✅ Complete | flock, atomic writes |
| Distribution | ✅ Complete | asdf/mise, code signing, supply chain |
| Error handling | ✅ Complete | Actionable hints, load error paths, size warnings |

---

## 7. Verdict

**The plan is substantially solid.** The encryption model, Keychain authentication pattern, shell integration, directory inheritance, and CLI surface are well-designed. The prior critical blocker (bash DEBUG trap) has been resolved.

**Three issues remain that would cause bugs or ambiguity during implementation:**

1. **Bash hook duplicate installation** — `setup` idempotency claim is broken without dedup check
2. **CLI surface inconsistency** — `docker-compose` vs `docker-env` rename not reflected in CLI section
3. **Auto-bundle naming unspecified** — `init` cannot be implemented without a naming convention

The remaining refinement issues (rotate rollback details, Keychain repair, config versioning, profiles link --force semantics, doctor FileVault handling) are addressable during implementation and do not block starting work.
