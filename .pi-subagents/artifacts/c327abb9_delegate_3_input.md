# Task for delegate

Create beads issues for Phase 4 (Inheritance). Run these commands and output all issue IDs:

bd create "Marker disabled support" \
  --description "Implement marker disabled content support.

- Marker file content 'disabled' (trimmed of whitespace/newline): stop walk, emit nothing
- Use case: place .envmoat with 'disabled' in subfolder to block parent inheritance
- Load resolution: on finding 'disabled', exit 0, no output
- Debug mode: log 'Marker disabled at <path>' to stderr

Parent: SM-1pz (Phase 4 Inheritance)
Priority: P2
" \
  --parent SM-1pz --priority 2 --silent

bd create "Marker profile override support" \
  --description "Implement marker profile override content support.

- Marker file content 'profile: <name>' (trimmed, case-sensitive): load named profile instead of default
- Use case: place .envmoat with 'profile: staging' in subfolder to override parent bundle
- Load resolution: parse 'profile: <name>', look up name in index.json profiles section
- Error if profile name not found in index.json
- Debug mode: log 'Loading profile <name> from marker at <path>' to stderr

Parent: SM-1pz (Phase 4 Inheritance)
Priority: P2
" \
  --parent SM-1pz --priority 2 --silent

bd create "Symlink canonicalization for all path operations" \
  --description "Implement symlink-aware path canonicalization.

- All path operations use canonical paths (realpath, follows symlinks)
- PWD resolution: resolve to canonical path before walk-up
- Index.json path storage: store canonical paths
- Prevent duplicate bundles from symlink aliases of same directory
- Test: symlink to project dir should resolve to same bundle as real dir

Parent: SM-1pz (Phase 4 Inheritance)
Priority: P2
" \
  --parent SM-1pz --priority 2 --silent

bd create "Configurable walk root via ENVMOAT_WALK_ROOT" \
  --description "Implement configurable walk boundary.

- Walk stops at '/' by default
- ENVMOAT_WALK_ROOT env var overrides walk boundary
- Use case: container environments, chroot, restricted filesystem access
- Validation: path must be absolute, must exist
- Debug mode: log 'Walk boundary: <path>' to stderr

Parent: SM-1pz (Phase 4 Inheritance)
Priority: P2
" \
  --parent SM-1pz --priority 2 --silent

echo "Done. Issue IDs created."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/c327abb9/progress.md

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