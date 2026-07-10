## Review Results

**2 blockers found** — both are text issues in PLAN.md that would cause bugs or implementation confusion if left as-is.

---

### Blocker 1: zsh hook `2>/dev/null` suppresses all warnings (line 108)

The zsh hook silences stderr:
```bash
output=$(secrets-manager load 2>/dev/null)
```
But the plan specifies meaningful warnings for missing/corrupted bundles ("Bundle not found. Run `secrets-manager init`", "Bundle corrupted. Run `secrets-manager verify`"). These are **never shown** because `2>/dev/null` eats them. The bash hook (line 115) correctly omits stderr suppression.

**Fix**: Remove `2>/dev/null` from the zsh hook, matching the bash hook.

---

### Blocker 2: "Resolved Design Decisions" contradicts actual bash hook code (lines 295, 319)

Two entries describe bash as using a `DEBUG` trap:
- Line 295: "`$BASH_COMMAND` regex `^cd(\ |$)` on DEBUG trap"
- Line 319: "zsh (`chpwd`) and bash (DEBUG trap)"

But the actual shell hook code (line 112–128) uses `PROMPT_COMMAND`, not a `DEBUG` trap. The code is correct (the prior architecture review identified the DEBUG trap as a critical bug); the text is stale. An implementer reading the resolved decisions would implement the wrong approach.

**Fix**: Update lines 295 and 319 to reference `PROMPT_COMMAND` instead of `DEBUG` trap.

---

### Non-blocking issues (5)

| # | Issue | Line | Impact |
|---|-------|------|--------|
| 3 | `config.yaml` has no `version` field; migration strategy still listed as "Open Question" | 287 | Future schema changes lack migration path. Add `version` to config.yaml. |
| 4 | Marker file format "open question" (line 287) contradicts resolved decision that chose plain text | 287 | Document inconsistency, not a bug. Remove from open questions. |
| 5 | `set` doesn't specify re-auth behavior on session expiry; `edit` explicitly says "re-auths if session expired" | 222 | Ambiguous behavior — implementer may not prompt for Touch ID. Add re-auth to `set` description. |
| 6 | `profiles link` doesn't mention `.gitignore` auto-add; `init` does ("Auto-adds `.secrets-manager` to `.gitignore`") | 234 | Marker file could be committed. Add `.gitignore` auto-add to `profiles link`. |
| 7 | No-args behavior for already-configured users underspecified — shows "welcome + usage" but no status | 207, 292 | Minor UX gap. Consider showing active profile/session status. |

---