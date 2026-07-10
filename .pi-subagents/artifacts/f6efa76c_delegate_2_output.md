## UX Review Results

**Not clean — 7 issues found.** The plan is well-architected overall, but has meaningful UX gaps.

### Blockers (2)

| # | Issue | Impact |
|---|-------|--------|
| 1 | **No `setup` command** — shell hook injection is described as automatic but no CLI command exists to do it. Users must manually edit rc files. | First-run breaks without manual intervention |
| 2 | **Master password creation is undocumented** — encryption model references it but no command creates it. | User cannot complete first-run |

### High (2)

| # | Issue | Recommendation |
|---|-------|----------------|
| 3 | **Error messages lack actionable recovery** — no guidance for wrong password, corrupted bundles, Touch ID unavailable, or permission failures | Define error patterns with recovery hints |
| 4 | **`get <KEY>` stdout risks shell history exposure** — piping secrets to commands leaves them in history | Add `--clip` flag as recommended pattern, document the risk |

### Medium (3)

| # | Issue | Recommendation |
|---|-------|----------------|
| 5 | **`link` command purpose is unclear** — "profile" vs "bundle" relationship is ambiguous | Rename to `profile link` or add `profiles` subcommand |
| 6 | **Docker workflow lacks integration guidance** — no `.gitignore` auto-entry, no `COMPOSE_ENV_FILES` hint | Auto-add `.gitignore`, print usage hint after generation |
| 7 | **No confirmation for destructive ops** — `remove`, `deinit`, `rotate` have no `--yes`/confirmation | Add confirmation prompts with `-y` bypass |

### Positive findings

Shell hook design, shell-safe output escaping, hierarchical inheritance, session TTL pattern, atomic writes, and phase breakdown are all solid. No changes needed there.

Full review written to `.pi-subagents/artifacts/progress/f6efa76c/progress.md`.