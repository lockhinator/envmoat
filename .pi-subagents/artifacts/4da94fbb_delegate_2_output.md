**Review complete.** One blocker and several underspecified behaviors found.

---

## Findings

### 🔴 Blocker (1)

**CLI Surface contradicts Resolved Decisions** — The CLI section lists `secrets-manager docker-compose [output-path]` as the primary command, but Resolved Design Decisions state `docker-env` is primary with `docker-compose` as alias. An implementer reading only the CLI section would build the wrong command name.

### ⚠️ Bug-Risk Issues (7)

| # | Issue | Impact |
|---|---|---|
| 2 | `set --file` parsing rules unspecified | `.env` files have comments, quotes, `export` prefix — without rules, implementation is guesswork |
| 3 | `profiles link` bundle creation underspecified | "Create bundle if needed" — empty bundle structure and pre-requisites unclear |
| 4 | Key case sensitivity not stated | `API_KEY` vs `api_key` — ambiguous behavior |
| 5 | `edit` non-zero exit code behavior missing | Editor crash/Ctrl+C could corrupt secrets without explicit discard logic |
| 6 | `backup` auth flow on expired session unstated | Unlike `set`, `backup` doesn't mention re-auth on expiry |
| 7 | `restore` conflict resolution options unspecified | "Prompts on conflict" but options (keep/overwrite/skip) not defined |
| 8 | `init` re-run behavior unspecified | Running `init` twice in same directory — error, no-op, or re-create? |

### Refinements (5)
Rotate rollback steps, Keychain repair path, `config.yaml` versioning, `profiles link --force` cleanup, `doctor` FileVault admin handling — all addressable during implementation.