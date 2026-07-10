Read the full plan. Here's the UX review:

## UX Review Findings

### Blockers / Gaps

1. **No `secrets-manager` (no args) behavior defined.** Running the binary with no subcommand is the most common first interaction. The plan doesn't specify what happens — should it print a welcome message, usage summary, or point to `setup`?

2. **`--help` / `--version` not mentioned.** Every CLI needs these. No reference to help text, man pages, or per-subcommand help anywhere in the plan.

3. **Fish shell not supported.** macOS users frequently use fish. The shell hook section covers only zsh and bash. Either document fish support or explicitly state it's unsupported with a clear error.

4. **`init` vs `profiles create` + `profiles link` relationship unclear.** `init <project-root>` creates a marker + empty bundle. `profiles link <path> <name>` does the same but with a named profile. Users won't know which to pick. The plan needs a clear distinction: `init` = "quick start for this directory with its own bundle" vs `profiles link` = "associate this directory with a shared named profile."

5. **First-run / onboarding experience absent.** No detection that this is the user's first interaction. No "welcome" flow. A user who installs via asdf has no guidance beyond `secrets-manager --help` (which isn't even documented).

6. **`edit` command UX unspecified.** Listed in Phase 6 but no detail on how it works — does it open `$EDITOR` with the current value? Does it require auth each time? What happens on save?

7. **`load` failure UX under-specified.** If setup hasn't been run, or Keychain is corrupted, or the binary is missing from PATH — the hook runs `2>/dev/null` so the user sees nothing. There should be a one-time warning mechanism or a `secrets-manager doctor`-style diagnostic.

8. **`SECRETS_MANAGER_DEBUG=1` discoverability.** Environment variable flags are invisible to users who don't read docs. Should `status` or `--help` mention it, or should there be a `--debug` flag on commands?

### Minor Issues

9. **`docker-compose` output file naming.** Default is `.env.secrets` — but the hint says `docker compose --env-file=.env.secrets up`. Users familiar with Docker expect `.env`. Consider documenting why the non-standard name is used (to avoid accidentally committing secrets if `.env` is already gitignored).

10. **`profiles delete` deletes the bundle too?** Unclear if deleting a profile also removes the encrypted bundle file, or just the name mapping. If bundles are orphaned, the store grows silently.

11. **Marker file in `$HOME` edge case.** The walk stops at `$HOME`, but what if the user places `.secrets-manager` in their home directory? The plan says "stops at `$HOME`" but doesn't clarify whether the `$HOME` marker itself is checked before stopping.

12. **`rotate` UX during two-phase process.** No progress indicator mentioned. If the user has many bundles, the command could take seconds with no feedback. A progress message or at least "Decrypting N bundles…" / "Re-encrypting…" would help.

13. **`get --clip` provides no confirmation.** Copying to clipboard is silent — user won't know if it succeeded. A brief "Copied to clipboard" on stderr would improve confidence.

14. **`setup` re-run behavior not specified.** What happens if the user runs `setup` again? Does it reset the master password? Re-install the hook? The plan should define idempotent or destructive behavior.

### Already Resolved / Not Issues

- Marker file format (plain text vs YAML) is listed as an open question — acceptable for a plan document.
- Clipboard backend (`pbcopy`) is resolved in design decisions.
- Error messages with actionable hints are addressed.
- Confirmation prompts for destructive ops are specified.
- Debug mode is defined (though discoverability is weak — see #8).

---

**Summary:** 8 substantive gaps, 6 minor issues. The plan is architecturally solid but has noticeable blind spots in CLI onboarding (`no-args` behavior, `--help`, first-run), shell coverage (fish), and command relationship clarity (`init` vs profiles). These should be addressed before implementation begins.