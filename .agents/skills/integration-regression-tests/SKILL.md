---
name: integration-regression-tests
description: Update and verify generator integration or regression tests, especially snapshot and fixture-based generation tests, in the correct sequence.
license: Proprietary
compatibility: opencode
---

# Skill: Integration And Regression Tests

Use this skill when changing the generator, templates, preprocessing, schema, or generated project shape and the repository's integration or regression tests may need updating.

This skill is especially relevant for fixture-backed tests such as `generator/generate_test.go`, where generated output is compared against archived snapshots like:

- `generator/testdata/workspace.zip`
- `generator/testdata/verify.zip`
- `generator/testdata/agent.zip`

## Goal

Keep integration tests trustworthy.

The purpose of these tests is to detect real output regressions. Fixture updates should reflect intentional final output changes, not transient intermediate states while the implementation is still in flux.

## Required Sequence

Follow this order:

1. Finish the implementation change first.
2. Run focused tests against the changed behaviour.
3. Generate the regression output and compare its archive hash to the baseline fixture hash.
4. Only inspect unpacked contents if the archive hash differs.
5. Only update fixtures or snapshot archives if the final, intended output truly changed.
6. Re-run the regression tests after refreshing fixtures.
7. Analyse the changed generated files and explicitly call out anything that smells like a regression or bug before accepting the fixture update.

Do not refresh fixtures pre-emptively while the implementation is still unstable.

## Why This Sequence Matters

If fixtures are refreshed too early, you can end up:

- hiding real regressions
- losing signal about what actually changed
- rewriting archives unnecessarily
- making review harder because snapshot churn no longer corresponds to the final behaviour

The fixture change should be the result of the implementation, not a tool used to paper over an incomplete implementation.

## Typical Workflow In This Repository

For generator output changes, the usual flow is:

1. update code, templates, preprocessing, or schema
2. run focused tests such as:

```bash
go test -tags integration ./generator -run TestGenerate -count=1
```

When this repository refers to `TestGenerate` or the fixture-backed generator regression tests, include the `integration` build tag as part of the command.

3. zip the generated output and compare its hash to the baseline regression archive
4. if the hash differs, inspect which generated files actually differ
5. decide whether the differences are intentional
6. refresh fixture archives only for the intentional final differences
7. run the focused regression tests again
8. review the file-level diff and highlight suspicious changes, not just accepted ones

## Fixture Updates

When a fixture-backed test is correct but the generated output intentionally changed:

- regenerate the fixture from the final code
- keep the tool version used by the test stable where the test expects one
- compare the generated archive hash to the baseline archive hash first
- verify the unpacked contents, not just the ZIP file timestamp or binary blob
- prefer understanding file-level differences before accepting a refreshed archive
- explicitly note suspicious changes such as broken casing, unexpected type changes, incorrect imports, malformed comments, or generated text that looks less correct than before
- if a fixture refresh only changes unstable metadata such as regenerated UUIDs in files like `pyproject.toml`, revert the archive refresh instead of keeping meaningless churn
- if unpacked inspection shows no meaningful generated-output changes, revert the ZIP file from the worktree instead of leaving it in the commit

If the generated archive hash is unchanged, do not rewrite or commit a new ZIP fixture. If the archive hash changed, inspect the unpacked contents before deciding whether the fixture update is legitimate.

## Review Checklist

Before finishing, check:

- Was the implementation completed before fixtures were updated?
- Were focused tests run before and after the fixture refresh?
- Do the fixture changes correspond to intentional generated-output changes?
- Was archive-hash comparison used before falling back to unpacked-content comparison?
- Are unchanged archives avoided?
- Can the final diff be explained file by file?
- Were any suspicious changes explicitly highlighted for review instead of being silently accepted into refreshed fixtures?

## Anti-patterns

Avoid these unless there is a strong reason:

- refreshing snapshots before the implementation is stable
- updating archives just to make failing tests pass temporarily
- treating fixture churn as harmless
- rewriting or committing a ZIP fixture when the generated archive hash is unchanged
- skipping the inspection step for generated-output differences
- keeping rewritten archives when the final unpacked content is unchanged
- refreshing fixtures without calling out suspicious file-level changes that may indicate a regression
- keeping a refreshed fixture when the only diffs are unstable metadata churn and the meaningful generated content is unchanged
- leaving a refreshed ZIP in the worktree when unpacked inspection shows there is no meaningful generated-output change to keep
