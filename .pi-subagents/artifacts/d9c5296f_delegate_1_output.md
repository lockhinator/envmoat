## Architecture Review Findings

The plan is substantially improved from the prior review — the critical bash `$REPLY` blocker is fixed (now uses `$BASH_COMMAND` with `^cd(\ |$)` regex). The encryption model, file format, concurrency strategy, and CLI surface are well-designed.

**Two contradictions must be resolved before implementation:**

| # | Contradiction | Locations | Fix |
|---|---|---|---|
| 1 | **Walk boundary**: `$HOME` vs `/` | Line 149 (`$HOME`), 164 (`/`), 265 (`/`), 270 (`$HOME`) | Pick `/` with `SECRETS_MANAGER_WALK_ROOT` override — handles projects outside home |
| 2 | **Session independence**: "independent per-terminal" vs "shared across terminals" | Line 252 (`independent`), 79/262 (`shared`) | Remove "independent per-terminal" — the two-item Keychain pattern is inherently shared |

**One operational gap:**

| # | Gap | Detail |
|---|---|---|
| 3 | **Rotate rollback** | Step (3) has two renames; if the first succeeds but the second fails, `bundles/` doesn't exist. Need explicit error handling + `mv bundles.bak/ bundles/` rollback |

**Low-severity items (addressable during implementation):**
- `profiles link` underspecified (auto-create profile? create bundle?)
- `verify` minimal detail (no index consistency check, no exit codes)
- `get <KEY>` doesn't specify bundle resolution
- No `doctor`/`repair` command for Keychain diagnostics
- No config/bundle format migration strategy
- No uninstall command for shell hook
- No secret size limit warning
- Interactive shell guard (`[[ $- == *i* ]]`) mentioned but not shown in hook code
- HKDF uses filename-as-salt — path change = unrecoverable data loss; no immutability guarantee

Full review written to `.pi-subagents/artifacts/progress/d9c5296f/progress.md`.