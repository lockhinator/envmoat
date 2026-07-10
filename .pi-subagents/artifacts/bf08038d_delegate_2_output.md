## Review Findings

The plan is substantially solid. The encryption model, Keychain two-item pattern, shell-safe output, directory walk-up, atomic writes, and CLI surface are well-designed and internally consistent. The ARCHITECTURE-REVIEW.md's critical bash `DEBUG` trap issue is resolved (now uses `PROMPT_COMMAND`), and most other prior findings are addressed.

**Three blockers remain that would cause bugs or prevent correct implementation:**

### Blocker 1: `auto` section resolution flow underspecified
`init` creates an empty `.secrets-manager` marker + an `auto` entry in `index.json`. The `load` flow describes walking up for markers, parsing content, and looking up in `index.json` — but doesn't specify how empty markers resolve to the `auto` section vs. `links`. An implementer can't wire `init` → `load` without knowing: when marker is empty, do we look up PWD in `auto`? What about `links`? Which takes priority?

### Blocker 2: `setup` idempotency — bash `PROMPT_COMMAND` duplication
`setup` is documented as idempotent, but the bash hook installation appends to `PROMPT_COMMAND` without checking if `__secrets_manager_hook` is already present. Running `setup` twice duplicates the hook, causing `secrets-manager load` to fire twice per prompt. Needs a guard (e.g., grep before append).

### Blocker 3: Silent session expiry in shell hook
When the Keychain cache TTL expires, `load` exits 0 with no output. The hook sees empty output and does nothing — secrets silently disappear. The plan says "Next interactive command prompts for Touch ID" but doesn't specify how the hook triggers re-auth. Either `load` should emit a stderr message, or the hook should detect expiry and prompt.

**Non-blocking issues** (should be addressed, won't cause bugs): `rotate` plaintext exposure on crash, `edit` "saves on exit" ambiguity (Ctrl+C), `get --clip` clipboard persistence warning, `config.yaml` lacks version field, `init` has no guard for existing markers (inconsistent with `profiles link --force`).