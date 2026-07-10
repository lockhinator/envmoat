After a thorough review of the plan, I found several UX issues across onboarding, discoverability, error handling, and general usability:

---

**Onboarding Issues**

1. **No guided post-setup flow** — After `setup` completes (password + hook installed), the user is on their own to discover `init`. The plan says the no-args welcome message shows usage, but there's no interactive wizard or "next step" prompt guiding the user through `init <project-root>`.

2. **`init <project-root>` argument is opaque** — New users won't know to pass an absolute path vs. relative path, or whether it should be the current directory. No `.` shorthand or "current directory" default is mentioned.

3. **`init` doesn't explain what it creates** — The user won't know a `.secrets-manager` marker file and an encrypted bundle are being created until they look. No confirmation message is specified.

4. **Shell hook install is silent on re-run** — The plan says `setup` is idempotent and "re-running re-installs hook only." No message is printed to confirm what happened, leaving users unsure if anything occurred.

**CLI Discoverability Issues**

5. **`setup --reset` is hidden** — The `--reset` flag is documented in the plan but has no discoverable path from `--help` or interactive prompts. Users who want to change their password won't find it.

6. **`profiles link` vs. `init` distinction is unclear** — The plan notes "`init` = auto-named bundle; `profiles link` = shared named profile," but this distinction isn't self-evident from the command names. A user wanting to link an existing profile to a directory won't intuitively try `profiles link`.

7. **`docker-compose` is a misleading command name** — It generates a `.env` file; it doesn't interact with Docker Compose. Users expecting Docker Compose integration will be confused. `docker-env` or `env-file` would be clearer.

8. **`export`/`import` naming conflict with common meaning** — `export` produces an *encrypted* file, not a plaintext dump. Users expecting a plaintext export for backup/migration will be surprised. Consider `backup`/`restore` instead.

9. **No `--all` flag on `get`** — To dump all secrets for the current bundle (e.g., for scripting), there's no `get --all` equivalent. Users have to `list` keys and `get` each one individually.

**Error Handling Issues**

10. **`load` silently swallows all errors** — Missing bundle, corrupted bundle, and session expiry all exit 0 with stderr-only warnings. If stderr is suppressed (common in scripts), the user gets no feedback at all. There's no way to get a non-zero exit for debugging.

11. **`SECRETS_MANAGER_DEBUG=1` is not discoverable** — The debug mode is only mentioned in `status` output. Users troubleshooting why secrets aren't loading won't find it via `--help` or error messages.

12. **Marker file parse errors are unrecoverable** — If a user accidentally writes invalid content to `.secrets-manager`, they get an error with the marker path but no guidance on valid values. The error message format isn't specified.

13. **`profiles link` with non-existent profile is undefined** — What happens if you `profiles link /path nonexistent-profile`? The plan doesn't specify error behavior.

**Usability Issues**

14. **No "which bundle am I in?" feedback** — The `status` command shows active profile, but there's no visual indicator in the shell prompt (e.g., a `PROMPT_COMMAND` or `precmd` hook that sets a visual marker). Users switching between projects won't know which secrets are active.

15. **`set` overwrites without warning** — `secrets-manager set KEY VALUE` updates an existing key silently. No confirmation prompt is specified for overwrites, unlike `remove`/`deinit`/`profiles delete` which all prompt.

16. **`edit` behavior on non-existent key is undefined** — Does `edit NEW_KEY` create the key, or error? The plan only says "open $EDITOR with current value."

17. **Key naming restriction (alphanumeric + underscore + hyphen) excludes dots** — Many real-world env vars use dots (e.g., `DATABASE_URL`, `AWS_SECRET_ACCESS_KEY` are fine, but `COMPOSE_PROJECT_NAME` patterns with dots in values aside, some conventions like `APP.VERSION` use dots in keys). This is a minor friction point.

18. **`get --clip` prints confirmation to stderr** — "Copied to clipboard" on stderr is correct (to avoid polluting stdout), but if the user pipes `get --clip` into another command, the stderr message still appears. Consider making it a `--quiet`-suppressible message.

19. **No fish shell support with no migration path** — Fish is a significant macOS shell user base. The plan documents it's unsupported but offers no `secrets-manager shell-hook fish` generation or guidance for fish users.

20. **`verify` orphan cleanup is silent** — The plan says `verify` does "orphan cleanup" but doesn't specify whether it deletes orphans automatically, prompts, or just lists them. Automatic deletion of encrypted bundles is a destructive action that needs confirmation.

21. **`rotate` success confirmation is unspecified** — After re-encrypting all bundles, there's no defined success message or verification step confirming the rotation completed correctly.

22. **`uninstall` secure erase is a single prompt** — "Offers secure erase" is vague. This is a high-stakes operation; it should require explicit confirmation (e.g., typing "DELETE" or similar).

23. **`list` shows only keys, no bundle context** — When multiple profiles exist, `list` doesn't indicate which bundle/profile is being listed. Easy to lose context.

24. **No `rename` for profiles** — Profiles can be created and deleted but not renamed. Users must create a new profile, migrate secrets, and delete the old one.

---

**Summary**: 24 issues found across 4 categories. The plan is well-structured architecturally but has meaningful UX gaps in onboarding guidance, command naming clarity, error feedback visibility, and destructive-operation safeguards.