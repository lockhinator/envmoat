Review complete. Two internal contradictions identified:

**1. `docker-compose` vs `docker-env`**
- CLI Surface: `secrets-manager docker-compose [output-path]` (listed as the command)
- Resolved Decisions: `docker-env` is primary, `docker-compose` is alias
- `docker-env` does not appear in the CLI Surface section at all

**2. `export`/`import` vs `backup`/`restore`**
- CLI Surface: `secrets-manager export --to-file` / `secrets-manager import --from-file`
- Resolved Decisions: `backup`/`restore` are primary, `export`/`import` are aliases
- `backup` and `restore` do not appear in the CLI Surface section at all

Both contradictions would cause an implementer reading the CLI Surface section to implement the wrong primary command names. The CLI Surface section should be updated to list `docker-env`, `backup`, and `restore` as primary commands, with aliases noted.

No missing high-level features. No architectural security issues.