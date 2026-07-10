# Task for delegate

Read /Users/cameronlockhart/Development/secrets-manager/PLAN.md lines 100-200 for CLI/shell hook details.

Then run these commands to add implementation plans to Phase 1 CLI tasks.

For SM-gtv.5 (setup + init):
bd update SM-gtv.5 --description "$(bd show SM-gtv.5 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: cmd/ (add to root.go or new setup.go, init.go)\n- setup: golang.org/x/term for hidden password prompt, confirm password match\n- Generate global salt: crypto/rand.Read(32 bytes), hex encode\n- Store in ~/.envmoat/config.yaml: {version: 1, salt: \"...\"}\n- Shell hook install: detect rc file (~/.zshrc or ~/.bashrc), grep for __envmoat_hook, append if missing\n- Hook: zsh chpwd + bash PROMPT_COMMAND (see PLAN.md for exact hook scripts)\n- --reset flag: re-run password + hook setup\n- init: create .envmoat marker (empty), create bundle, add to index.json auto section\n- Auto-bundle name: auto-<slugified-last-dirname>.enc, collision: append -<short-hash>\n- Auto-append .envmoat to .gitignore\n- Dependencies: SM-gtv.7 (encryption), SM-gtv.1 (backend)\n- Acceptance: setup creates config.yaml + installs hook, init creates marker + bundle + index entry, idempotent hook install, --reset works" 2>&1

For SM-gtv.2 (set + get + list):
bd update SM-gtv.2 --description "$(bd show SM-gtv.2 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: cmd/ (set.go, get.go, list.go) or add to root.go\n- set: resolve bundle via resolver, decrypt, update JSON, re-encrypt, atomic write\n- Interactive prompt: golang.org/x/term.ReadPassword(0, os.Stdin, prompt)\n- --stdin: io.ReadAll(os.Stdin)\n- --file: parse .env format (KEY=VALUE, skip # comments, handle quotes)\n- Key validation: regexp `^[a-zA-Z0-9_-]+$`\n- Size warning: len(value) > 1MB -> stderr warning\n- get: resolve bundle, decrypt, fmt.Print(value) to stdout\n- list: resolve bundle, decrypt, print keys only + active profile name\n- Context resolution: resolver.Resolve() -> bundle path, error if not tracked\n- Dependencies: SM-gtv.7 (encryption), SM-gtv.4 (resolver)\n- Acceptance: set/get roundtrip, list shows keys no values, --stdin works, --file bulk import, key validation rejects invalid chars, size warning at 1MB" 2>&1

For SM-gtv.3 (load + remove + deinit + verify):
bd update SM-gtv.3 --description "$(bd show SM-gtv.3 --json | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['description'])")\n\n## Implementation Plan\n- Package: cmd/ (load.go, remove.go, deinit.go, verify.go) or add to root.go\n- load: resolve bundle, decrypt, emit shell-safe output\n- Shell-safe: single-quote all values, escape internal ' as '\''\n- First line: #bundle_hash:sha256:<hash> for change detection\n- Errors to stderr, exit 0 no output when no bundle\n- remove: resolve bundle, decrypt, delete key, re-encrypt, atomic write, confirmation prompt\n- deinit: remove .envmoat marker, remove bundle file, remove from index.json, confirmation\n- verify: iterate all bundles in index.json, attempt decrypt, report orphans\n- Confirmation: fmt.Print(\"Are you sure? [y/N] "), read stdin, check 'y' or 'Y'\n- -y/--yes flag: skip confirmation\n- Dependencies: SM-gtv.7 (encryption), SM-gtv.4 (resolver)\n- Acceptance: load emits valid shell exports, single quotes escaped, hash line present, remove deletes key, deinit cleans up, verify detects orphans" 2>&1

echo "Done updating Phase 1 CLI tasks."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/222a0ee2/progress.md

## Acceptance Contract
Acceptance level: checked
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope

Required evidence: changed-files, tests-added, commands-run, residual-risks, no-staged-files

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```