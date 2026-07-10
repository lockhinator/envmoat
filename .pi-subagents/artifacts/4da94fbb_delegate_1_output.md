Here is the comprehensive review of PLAN.md, incorporating findings from the two prior reviews and fresh analysis.

---

## Review Summary

The plan is substantially solid. The encryption model (scrypt → LUK → HKDF per-bundle DEK → AES-256-GCM), Keychain two-item authentication pattern, shell-safe output format, directory walk-up inheritance, and CLI surface are well-designed. The prior critical blocker (bash `DEBUG` trap) has been resolved with `PROMPT_COMMAND`.

**Prior review status**: ARCHITECTURE-REVIEW.md identified 10 issues; PLAN-REVIEW.md identified 3 blockers + 5 refinements. The current PLAN.md has addressed most prior findings. Three items from the second review (bash hook dedup, CLI naming inconsistency, auto-bundle naming) are now resolved in the current text.

---

## Issues That Would Block Implementation or Cause Bugs

### Issue 1: `set --file` Import Format Underspecified — Blocker

**Location**: CLI Surface → `secrets-manager set --file <.env>`

**Problem**: The plan specifies `set --file <.env>` for bulk import but does not define the input format. Is it `.env`-style (`KEY=VALUE`, one per line, `#` comments, quoted values)? JSON? YAML? Without this, the parser cannot be implemented. The `.env` filename hint suggests `.env` format, but that is not explicit.

**Impact**: Implementer must guess the format, risking incompatibility with common `.env` files users actually have.

**Recommendation**: Specify explicitly: "Parses standard `.env` format: `KEY=VALUE` or `KEY="VALUE"` per line, `#` comments ignored, leading/trailing whitespace trimmed, no multi-line values."

---

### Issue 2: `docker-env` Output Path Relative to PWD, Not Project Root — Bug

**Location**: Docker Compose Integration section

**Problem**: `docker-env` generates `.env.secrets` in the current working directory (`./.env.secrets`). If the user is in a subdirectory that inherits secrets from a parent marker (e.g., `myapp/src/` inherits from `myapp/.secrets-manager`), the `.env.secrets` file is created in `myapp/src/` instead of `myapp/`. Docker Compose typically runs from the project root, so the file ends up in the wrong place.

**Impact**: User runs `docker-compose up` from project root but the `.env.secrets` file was generated in a subdirectory.

**Recommendation**: Write the file next to the resolved marker (project root), not PWD. Or accept PWD behavior but document it clearly.

---

### Issue 3: `setup` Idempotency Mechanism Underspecified — Bug

**Location**: Shell Integration section / `setup` command

**Problem**: The plan states `setup` is "idempotent; checks for existing hook before appending." However, the mechanism is not specified. Running `setup` on a system where the user manually deleted the hook from their rc file should re-install it. The check must be specific enough to detect the hook's presence (e.g., `grep -q '__secrets_manager_hook'`) but not so fragile that it fails on minor formatting differences.

**Impact**: Without a dedup check, running `setup` twice appends duplicate hook entries, causing `secrets-manager load` to run multiple times per prompt.

**Recommendation**: Specify: "Before appending hook code to rc file, check if `__secrets_manager_hook` is already present via `grep`. If found, skip hook installation (but still validate store setup)."

---

## Issues That Are Addressable During Implementation (Not Blockers)

### Issue 4: `init` Idempotency Not Specified

Running `init` on an already-initialized directory (marker + bundle exist) is not specified. Should it be a no-op, error, or offer to reinitialize? **Recommendation**: Make it a no-op with a message like "Directory already initialized with profile `<name>`."

### Issue 5: `profiles link` Path Validation Not Specified

Does `profiles link <path> <name>` require `<path>` to exist? Should it resolve symlinks? **Recommendation**: Require path exists, resolve to canonical path.

### Issue 6: Auto-Bundle Slugification Not Defined

The plan says `auto-<slugified-last-dirname>.enc` but "slugified" is not defined (lowercase? hyphens for spaces? strip special chars?). **Recommendation**: Define as: lowercase, alphanumeric + hyphens only, non-alphanumeric replaced with hyphens, consecutive hyphens collapsed, leading/trailing hyphens trimmed.

### Issue 7: `backup`/`restore` Auth Flow Not Specified

`backup` encrypts with the current LUK. If the session is expired, does it prompt for Touch ID? `restore` "requires master password" — is this a fresh prompt or does it use the cached session? **Recommendation**: Both should prompt for authentication if the session is expired, consistent with other commands.

### Issue 8: Non-UTF-8 Secret Values

The plan references `strconv.QuoteToASCII` for shell escaping, which rejects non-ASCII bytes. Binary secrets (certificates, keys in DER format) would fail. **Recommendation**: Document that secrets must be UTF-8 text. For binary data, suggest base64 encoding before storing.

### Issue 9: No Way to Set Secrets in a Profile Without Being in a Linked Directory

Context resolution requires being in a tracked directory. There is no `--profile` flag on `set` to target a specific profile by name. **Recommendation**: Add `--profile <name>` flag to `set`/`get`/`list`/`remove`/`edit` to override context resolution. Alternatively, document `profiles link` as the required workflow.

---

## Completeness Verification

| Feature Area | Status |
|---|---|
| Setup / uninstall | ✅ |
| Master password + scrypt | ✅ |
| Touch ID + Keychain two-item | ✅ |
| Secret CRUD (set/get/list/remove/edit) | ✅ |
| Shell integration (zsh + bash) | ✅ |
| Directory inheritance + markers | ✅ |
| Profiles (named bundles) | ✅ |
| Docker .env generation | ✅ |
| Import/export/rotate | ✅ |
| Diagnostics (verify/doctor/status) | ✅ |
| Concurrency (flock + atomic writes) | ✅ |
| Distribution (asdf/mise, code signing) | ✅ |