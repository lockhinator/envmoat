## Architecture Review Results

**Not CLEAN — 10 issues found (1 critical, 2 medium, 7 low).**

### Critical Blocker
- **Bash `DEBUG` trap fires before `cd`**: The hook resolves PWD *before* the directory changes, so secrets load for the wrong directory. The zsh `chpwd` hook is correct; the bash hook is fundamentally broken. **Fix**: replace `DEBUG` trap with `PROMPT_COMMAND` (or override `cd`).

### Medium Issues
- **`2>/dev/null` swallows warnings**: The plan specifies actionable warnings for missing/corrupted bundles, but the hook suppresses all stderr — users never see them.
- **No secret size limits**: Large values (TLS certs, SSH keys) could exceed shell export limits (~2MB on macOS) and break the hook.

### Low Issues
- Document contradiction: "Open Questions" lists marker format as undecided, but "Resolved Design Decisions" already chose plain text.
- No `config.yaml` version/migration strategy (only `index.json` is versioned).
- Docker `secrets:` directive not considered (`.env` only).
- No Go `flock` package specified (`github.com/gofrs/flock` recommended).
- No Keychain repair path for orphaned items after password change/restore.
- No `flock` for `index.json` writes (only bundle files are locked).
- `rotate` rollback procedure underspecified (needs explicit step-by-step).

### Prior Issues Resolved (6 of 15)
The updated PLAN.md correctly addresses: bash `REPLY` guard (now `$BASH_COMMAND`), walk boundary (now `/`), marker parsing (strict validation), `verify`/`doctor` commands, `profiles link` specification, debug mode, session sharing, `bundle_filename` for HKDF, and rotate atomicity.

Full review written to `ARCHITECTURE-REVIEW.md`.