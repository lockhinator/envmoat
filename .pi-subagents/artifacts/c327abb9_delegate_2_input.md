# Task for delegate

Create beads issues for Phase 3 (Profiles). Run these commands and output all issue IDs:

bd create "profiles list + profiles create" \
  --description "Implement profiles list and profiles create commands.

envmoat profiles list:
- List all named profiles from index.json profiles section
- Show profile name -> bundle filename mapping
- Show bundle metadata (created_at, updated_at)
- Empty output if no profiles exist

envmoat profiles create <name>:
- Create new named profile in index.json profiles section
- Create empty encrypted bundle (auto-<slugified-name>.enc)
- Validate profile name: alphanumeric + underscore + hyphen
- Error if profile already exists
- Output: 'Created profile: <name>'

Parent: SM-idf (Phase 3 Profiles)
Priority: P1
" \
  --parent SM-idf --priority 1 --silent

bd create "profiles delete" \
  --description "Implement profiles delete command.

envmoat profiles delete <name>:
- Remove profile from index.json profiles section
- Delete encrypted bundle file
- Confirmation prompt: 'Delete profile <name> and its bundle? (y/N)'
- -y flag to skip confirmation
- Error if profile doesn't exist
- Atomic write for index.json (temp + rename)

Parent: SM-idf (Phase 3 Profiles)
Priority: P1
" \
  --parent SM-idf --priority 1 --silent

bd create "profiles link + profiles unlink" \
  --description "Implement profiles link and profiles unlink commands.

envmoat profiles link <path> <name>:
- Create .envmoat marker file in <path> with content 'profile: <name>'
- Create bundle if profile doesn't exist yet
- Auto-append .envmoat to project .gitignore
- Error if marker already exists (--force to overwrite)
- Validate profile name exists or create it

envmoat profiles unlink <path>:
- Remove .envmoat marker from <path>
- Confirmation prompt (-y to skip)
- Do NOT delete the profile or bundle (other projects may use it)

Parent: SM-idf (Phase 3 Profiles)
Priority: P1
" \
  --parent SM-idf --priority 1 --silent

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