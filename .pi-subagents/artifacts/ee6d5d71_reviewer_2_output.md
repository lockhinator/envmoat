## Review

### Correct

- **Central store outside projects** — `~/.secrets-manager/bundles/` is the right default. Users never need to worry about secret files in repos, and migration between machines is a single directory copy.
- **Walk-up inheritance model** — walking from `PWD` upward for `.secrets-manager` markers is intuitive. It mirrors how `.gitignore`, `.env`, and Node.js `package.json` resolution work. Developers will already have a mental model for it.
- **Touch ID via Keychain** — leveraging macOS's native `SecAccessControl` is the most frictionless path. The native dialog is familiar, and automatic fallback to login password is handled by the OS.
- **Session TTL (15 min default)** — reasonable for a developer workflow. Long enough to avoid constant re-auth during a coding session, short enough to feel secure when you step away.
- **Shell hook via `precmd` / `PROMPT_COMMAND`** — correct mechanisms for zsh and bash. The `last_dir` guard prevents redundant calls.

### Note

- **First-run experience is underspecified.** The plan has no `secrets-manager install` or `secrets-manager setup` command. The shell hook snippet lives in the docs but there's no automated way to inject it into `~/.zshrc`. A user who `asdf install secrets-manager` will have a binary that does nothing until they manually edit their shell config. **Recommendation:** Add an `install` subcommand that appends the hook to the detected shell rc file, or at minimum prints a one-liner to copy-paste. This is the single biggest UX gap.

- **`secrets-manager set` has no interactive mode.** For secrets with special characters (e.g., passwords with `!`, `$`, spaces), passing values as CLI arguments is fragile — the shell interprets them before the tool sees them. **Recommendation:** When `<VALUE>` is omitted, prompt with a masked stdin read (like `git credential fill` or `pass`). Example: `secrets-manager set API_KEY` → prompts for value.

- **Open question 2 (auto-detect vs `--profile`) affects discoverability.** The plan leaves this open. **Recommendation:** Default to auto-detect from the nearest `.secrets-manager` marker. Require `--profile` only when no marker is found or when explicitly requested. This is the path of least resistance for the 90% case.

- **Error messages are not designed.** The plan doesn't specify what users see on common failure modes:
  - Keychain access denied (Touch ID failure after retries)
  - Bundle corrupted or tampered (GCM auth tag mismatch)
  - Multiple markers found at different levels (ambiguous inheritance)
  - `load` runs in a non-interactive shell (e.g., CI, cron)
  **Recommendation:** Define at least 3-5 error message templates in the design doc. Every error should include the command to run to fix it (e.g., "Touch ID failed. Run `secrets-manager unlock` to authenticate with password.").

- **`docker-compose` subcommand naming inconsistency.** The subcommand is `docker-compose` (hyphen) but it's a single command, not two tools composing. Consider `secrets-manager docker` or `secrets-manager compose-env` for consistency with the rest of the CLI which uses single words (`init`, `set`, `get`, `list`, `load`, `rotate`).

### Fixed

None — this is a plan review, no code changes applied.

### Blocker

None. The plan is architecturally sound and the UX model is solid. The first-run gap is a significant omission but doesn't invalidate the design.

### Additional Observations

- **Marker file content format** — using plain text content (`disabled`, `profile: <name>`) is simple but fragile. A YAML or JSON marker would be more extensible (e.g., adding `exclude: [KEY1, KEY2]` later). However, plain text wins on simplicity for the MVP. Keep it and add a migration path.
- **Open question 3 (shared vs independent session tokens)** — independent per-terminal is the safer default. If Terminal A is left unlocked for 15 minutes, Terminal B shouldn't inherit that session without its own Touch ID. The session token should be scoped to the terminal process group, not just PID.
- **`rotate` command UX** — re-encrypting all bundles with a new password is a destructive-sounding operation. The command should confirm, show how many bundles will be affected, and ideally create a backup before proceeding.

---