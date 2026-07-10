# Task for delegate

Read /Users/cameronlockhart/Development/secrets-manager/PLAN.md and review it as a **plan document**, not an implementation spec. This plan defines the feature set and high-level approach — it is NOT meant to specify every edge case, parsing rule, or error path. Those will be decided during implementation.

Flag ONLY:
- Contradictions within the document (e.g., section A says X, section B says Y)
- Missing high-level features from the original requirements
- Security issues that change the architecture

Do NOT flag:
- Unspecified parsing rules, error messages, or edge cases
- Missing implementation details (slugification, temp file handling, etc.)
- Suggestions for additional features

If the plan is internally consistent and covers the feature set, respond with exactly "CLEAN — plan is ready for implementation."

---
Update progress at: /Users/cameronlockhart/Development/secrets-manager/.pi-subagents/artifacts/progress/95532639/progress.md

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