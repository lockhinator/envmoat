# Plan Review v3 — Secrets Manager

**Date**: 2026-07-09
**Scope**: Comprehensive review of PLAN.md — security, architecture, UX, contradictions, edge cases, attack vectors, completeness.
**Prior reviews**: ARCHITECTURE-REVIEW.md (10 issues), PLAN-REVIEW.md (3 blockers + 5 refinements)

---

## Prior Issue Resolution Status

| Prior Issue | Status | Notes |
|---|---|---|
| Bash hook idempotency | ⚠️ Partially | Plan states "idempotent" but mechanism (grep check) not shown |
| CLI `docker-compose` vs `docker-env` | ❌ Not resolved | CLI section still lists `docker-compose` as primary |
| Auto-bundle naming unspecified | ⚠️ Partially | Naming convention in resolved decisions, not in CLI section |
| `profiles link --force` semantics | ❌ Not resolved | Still underspecified |
| Rotate rollback procedure | ❌ Not resolved | Still stated but not detailed |
| Keychain repair path | ❌ Not resolved | No repair command documented |
| `config.yaml` version field | ❌ Not resolved | `config.yaml` still has no version |

---

## Issues That Would Block Implementation or Cause Bugs

### Blocker 1: CLI Surface Contradiction — `docker-compose` vs `docker-env`

**Location**: CLI Surface section, line ~140

**Problem**: The CLI Surface section lists:
```
secrets-manager docker-compose [output-path]
```
But Resolved Design Decisions state:
> **docker-env**: primary name. `docker-compose` kept as alias.

An implementer reading the CLI Surface section would implement `docker-compose` as the primary command name, contradicting the resolved decision. The CLI section was never updated after the rename.

**Fix**: Update CLI Surface to show `docker-env [output-path]` with `docker-compose` listed as an alias.

---

### Blocker 2: Bash Hook Idempotency Mechanism Underspecified

**Location**: Shell Integration section + Setup flow

**Problem**: The plan states `setup` is "idempotent — checks for existing hook before appending." However, no mechanism is specified for how this check works. Without a dedup check (e.g., `grep -q '__secrets_manager_hook'`), running `setup` twice appends duplicate hook code to the rc file, causing `secrets-manager load` to execute multiple times per prompt.

**Fix**: Specify the dedup mechanism explicitly. E.g.:
```bash
grep -qF '__secrets_manager_hook' ~/.zshrc || { /* append hook */ }
```

---

### Blocker 3: Auto-Bundle Naming Convention Missing from CLI Section

**Location**: CLI Surface section, `init` command

**Problem**: The CLI section says `init` creates an "auto-named bundle" but does not specify the naming algorithm. The resolved decisions section has: `auto-<slugified-last-dirname>.enc` with collision handling via `-<short-hash>`. However, an implementer working from the CLI section alone would not know how to name bundles deterministically. The `auto` section of `index.json` maps paths to bundle filenames, so the naming must be reproducible.

**Fix**: Add the naming convention to the CLI section or the `init` command description.

---

## Refinement Issues (Addressable During Implementation)

### Issue 4: `profiles link --force` Behavior Underspecified

`--force` overwrites the marker file, but it is unclear whether the old profile's bundle is also deleted. If a marker points to `profile: dev` and the user runs `profiles link --force <path> staging`, should the `dev` bundle be deleted?

**Recommendation**: `--force` overwrites the marker file only. Old bundle remains (orphan cleanup via `verify`).

### Issue 5: Rotate Rollback Procedure Underspecified

"Two-phase atomic, .bak rollback" is stated but not detailed. If encryption fails on bundle 3 of 10, what is the rollback state?

**Recommendation**: Specify explicitly:
1. Decrypt all bundles → plaintext temp dir (fail-fast)
2. Encrypt all plaintext → new encrypted temp dir (fail-fast)
3. Atomic swap: `bundles/` → `bundles.bak/`, temp dir → `bundles/`
4. Delete `bundles.bak/` on success

### Issue 6: No Keychain Repair Path

If the user changes their macOS login password or restores from Time Machine, Keychain items become inaccessible. `doctor` diagnoses but cannot fix.

**Recommendation**: Add `secrets-manager repair` or `doctor --fix` that recreates Keychain items after re-auth.

### Issue 7: `config.yaml` Missing Version Field

`index.json` has `"version": 1` for schema migration. `config.yaml` has no version field. If config schema changes between versions, there is no migration path.

**Recommendation**: Add a `version` field to `config.yaml` and implement forward-compatible defaults.

### Issue 8: `edit` Command — `$EDITOR` Not Set

If `$EDITOR` is not set, the `edit` command has no fallback. macOS users may not have `$EDITOR` configured.

**Recommendation**: Fall back to `vi` or `nano`, or error with a helpful message suggesting `export EDITOR=...`.

### Issue 9: `set --file` Parsing Rules Unspecified

The plan lists `set --file <.env>` for bulk import but does not specify the parsing format. Does it support `.env` conventions (comments with `#`, quoted values, `KEY=value`)?

**Recommendation**: Document parsing rules: `KEY=VALUE` format, `#` comments, single/double-quote stripping, ignore blank lines.

### Issue 10: Secure Erase Ineffective on SSDs

`uninstall` offers "secure erase" but multi-pass overwriting is ineffective on SSDs with TRIM/wear-leveling. The plan mentions FileVault encryption at rest as the real protection but doesn't note that secure erase is misleading on modern storage.

**Recommendation**: Rename to "delete store" and note that encryption at rest (FileVault) provides the actual security guarantee.

---

## Security Assessment

| Area | Status | Notes |
|---|---|---|
| Encryption (scrypt + AES-256-GCM + HKDF) | ✅ Solid | OWASP-recommended parameters, clean key hierarchy |
| Keychain two-item pattern | ✅ Solid | Touch ID + sliding TTL, shared across terminals |
| Shell injection prevention | ✅ Solid | Single-quote escaping, `'\''` for internal quotes |
| File permissions (0700/0600) | ✅ Solid | Validated on startup |
| Atomic writes + flock | ✅ Solid | Prevents corruption |
| Walk boundary at `/` | ✅ Solid | Configurable via env var |
| Bundle rename breaks DEK | ✅ Acceptable | Bundles are tool-managed, not user-renamed |
| Shell history leakage | ⚠️ Known | `get` output to stdout; mitigated by `get --clip` |
| Keychain fallback to login password | ⚠️ Documented | macOS behavior, cannot be avoided |
| Secure erase on SSDs | ⚠️ Misleading | See Issue 10 |

---

## Architecture Assessment

| Area | Status | Notes |
|---|---|---|
| Central store outside projects | ✅ Solid | Prevents accidental commit |
| index.json schema | ✅ Solid | Versioned, clean profiles/links/auto separation |
| Directory inheritance (walk-up) | ✅ Solid | Marker files, symlink handling, configurable boundary |
| Shell integration (zsh + bash) | ✅ Solid | chpwd + PROMPT_COMMAND, interactive guard, dedup hash |
| Concurrency (flock + atomic writes) | ✅ Solid | Both bundle and index.json locked |
| CLI surface | ⚠️ Minor issues | docker-compose/docker-env contradiction (Blocker 1) |
| Phased implementation | ✅ Solid | Logical progression, Touch ID in Phase 2 after MVP |

---

## UX Assessment

| Area | Status | Notes |
|---|---|---|
| Setup flow | ✅ Good | Single command, idempotent, --reset for password change |
| Error messages | ✅ Good | Actionable recovery hints on all errors |
| Destructive ops confirmation | ✅ Good | Prompts with `-y`/`--yes` bypass |
| Debug mode | ✅ Good | `SECRETS_MANAGER_DEBUG=1`, stderr only, no secret values |
| Load error paths | ✅ Good | Silent for no-marker, warnings for missing/corrupt bundles |
| Session expiry in hook | ✅ Acceptable | Stderr hint, no injection, re-auth on next interactive command |
| `edit` fallback | ⚠️ Gap | No `$EDITOR` fallback (Issue 8) |
| Secret size warning | ✅ Good | 1MB warning, practical shell limit noted |

---

## Completeness Check

| Feature Area | Status | Notes |
|---|---|---|
| Setup / uninstall | ✅ Complete | Idempotent setup, full uninstall |
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

## Verdict

**Three blockers remain that would cause bugs or ambiguity during implementation:**

1. **CLI surface lists `docker-compose` as primary** but resolved decisions say `docker-env` is primary — implementer contradiction
2. **Bash hook idempotency mechanism unspecified** — `setup` claims idempotent but no dedup check documented, causes duplicate hooks
3. **Auto-bundle naming convention not in CLI section** — `init` cannot be implemented deterministically without the naming algorithm

**Ten refinement issues** are addressable during implementation (rotate rollback details, Keychain repair, config versioning, profiles link --force semantics, edit $EDITOR fallback, set --file parsing, secure erase on SSDs, docker secrets: directive, doctor FileVault admin handling, supply chain details).

The encryption model, Keychain authentication, shell integration, directory inheritance, file format, and CLI surface design are solid. The plan is well-structured and comprehensive with these exceptions.
