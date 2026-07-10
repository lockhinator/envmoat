# Task for delegate

Create beads issues for Phase 7 (Polish). Run these commands and output all issue IDs:

bd create "asdf plugin for envmoat" \
  --description "Create asdf plugin for envmoat distribution.

- asdf plugin repository structure
- install.sh: download prebuilt macOS binary (Intel + Apple Silicon) or build from source
- Plugin URL: github.com/lockinator/asdf-envmoat (or bundled)
- Version resolution: git tags
- Installation command sets up shell hook automatically

Parent: SM-rgt (Phase 7 Polish)
Priority: P3
" \
  --parent SM-rgt --priority 3 --silent

bd create "mise plugin for envmoat" \
  --description "Create mise (rtx) plugin for envmoat distribution.

- mise plugin specification (mise.toml or .mise.toml)
- Binary download: GitHub releases with OS/arch detection
- Build-from-source fallback
- Installation command sets up shell hook automatically

Parent: SM-rgt (Phase 7 Polish)
Priority: P3
" \
  --parent SM-rgt --priority 3 --silent

bd create "Shell completions — zsh + bash" \
  --description "Generate shell completions for zsh and bash.

- zsh completion script (_envmoat)
- bash completion script (envmoat-completion.bash)
- Cover all subcommands and flags
- Dynamic completions where applicable (e.g., profile names, key names)
- 'envmoat completions <zsh|bash>' command to generate
- Install to correct system directories

Parent: SM-rgt (Phase 7 Polish)
Priority: P3
" \
  --parent SM-rgt --priority 3 --silent

bd create "Code signing + notarization for macOS Gatekeeper" \
  --description "Implement macOS code signing and notarization.

- Developer certificate signing (codesign)
- Notarization via xcrun notarytool
- Gatekeeper compliance for distributed binaries
- CI/CD integration: sign + notarize on release builds
- Entitlements: no special entitlements needed (no TCC access)

Parent: SM-rgt (Phase 7 Polish)
Priority: P3
" \
  --parent SM-rgt --priority 3 --silent

bd create "Supply chain hardening — go mod verify, reproducible builds, audit" \
  --description "Implement supply chain security measures.

- 'go mod verify' in CI pipeline
- Reproducible builds: deterministic output from same source + deps
- Dependency audit: 'govulncheck' or equivalent in CI
- Release artifacts: SHA256 checksums, signatures
- GitHub Actions: pin action versions, review third-party actions

Parent: SM-rgt (Phase 7 Polish)
Priority: P3
" \
  --parent SM-rgt --priority 3 --silent

echo "Done. Issue IDs created."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/c327abb9/progress.md

## Acceptance Contract
Acceptance level: reviewed
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope
- criterion-2: Return evidence sufficient for an independent acceptance review

Required evidence: changed-files, tests-added, commands-run, validation-output, residual-risks, no-staged-files

Review gate: required by reviewer.

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