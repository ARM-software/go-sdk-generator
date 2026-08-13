---
name: change-manifest-schema
description: Use when editing `generator/schema/schema.yml` or changing how manifests are validated or generated. This skill covers the follow-up work: regenerate generated files and dogfood the generator in the correct order.
license: Proprietary
compatibility: opencode
---

# Skill: Change Manifest Schema

Use this skill when the user wants to add, remove, or change fields or behaviour in the generator manifest schema in `schema.yml`.

## Required Workflow

Follow this workflow in order:

1. Load and use the local `grill-me` skill first to challenge the design before implementation.
2. Treat `schema.yml` as the source of truth for schema changes.
3. If the schema change introduces or tightens validation, follow the repository validation pattern: `Validate() error`, `config.ValidateEmbedded(...)`, `validation.ValidateStruct(...)`, and collection helpers for nested slices.
4. If the change is breaking, design the deprecation path first: define the replacement, use `deprecation_notice` where the schema allows it, document the migration, and give consumers a deprecation window rather than making an abrupt deletion.
5. Make the implementation changes needed in the generator and templates.
6. Create a news file with the correct change type and a message written as described in `CONTRIBUTING.md`. Use `cd-create-news-file`. This is required for generator feature work, not optional.
7. Regenerate generated files using `go generate ./...`.
8. Never edit generated files matching `*.gen.*` manually.
9. Build the updated generator using `go build` in the `cli` project.
10. Dogfood the updated generator against this repository's own manifest `generate.yml`. The `generator` skill may be useful for that process, although the newly built CLI tool should be used rather than a previous release.

## Implementation Rules

- Keep schema descriptions in `schema.yml` clear and authoritative.
- Do not consider a schema or generator feature complete until the corresponding news file has been created.
- If generated files change because of the schema update, regenerate them instead of patching them.
- When integration or regression fixtures may need updating, follow the local `integration-regression-tests` skill.
- Prefer the smallest coherent schema change that supports the requested behaviour.
- Make sure naming is clear, self-explanatory, and precise.
- Treat renames, removals, semantic repurposing, incompatible type or shape changes, and validation tightening that breaks existing manifests as breaking changes.
- Prefer deprecate, document, migrate, then remove.
- Do not abruptly delete or silently repurpose a manifest concept when a deprecation window is practical.
- Remember that the generator output is a Go project: a true major version bump also implies module-path and import-path work, so a breaking change that is not doing that full migration should usually be communicated with a `.feature` news file and an explicit breakage message rather than treated as a routine `.major` automation case.
- Relevant references: [Semantic Versioning](https://semver.org/), [Go modules major version suffix](https://go.dev/ref/mod#major-version-suffix), and [Go Modules: v2 and beyond](https://go.dev/blog/v2-go-modules).

## Dogfooding

After implementing the change:

1. Build the generator CLI from this repository using `go build` in the `cli` project.
2. Run the built generator against `generate.yml`. The `generator` skill may be useful for that process, although the newly built CLI tool should be used rather than a previous release. When doing this in this repository, pass the repository's current Go version explicitly via `--go_version` instead of changing the manifest default value.
3. Inspect the generated output for regressions in the generator's own SDK and CLI.
4. Treat dogfooding as a validation step, not as permission to remove repository-local development wiring or handwritten logic that this repository still needs in order to build, test, or develop the generator itself.

In particular, do not blindly accept dogfooded changes if they remove or break things such as:

- `go.work` entries required for local multi-module development
- `go.mod` requirements needed for the repository's current development and test workflow
- `go tool` entries declared in `go.mod`, unless the work explicitly changes the repository's toolchain setup
- targeted Go versions or dependency versions that were not meant to change as part of the work
- the repository's own manifest content in `generate.yml` when the feature is about generator behaviour rather than a deliberate manifest edit
- checked-in implementation logic in this repository that is intentionally not meant to be replaced by generated placeholders
- any other repository-local bootstrap or development behaviour that the generated output does not yet preserve correctly

If dogfooding removes required local logic, keep the dogfooding insight, but preserve the working repository state and fix the generator/templates rather than committing a broken self-generated result.

## Verification

Minimum verification expectations:

- `go generate ./...`
- relevant `go test` coverage for the changed generator behaviour
- build the generator successfully
- dogfood the built generator against `generate.yml`

## Notes

- Use `CONTRIBUTING.md` for news-file expectations and wording.
- Use `grill-me` to resolve terminology and design trade-offs before editing code.
- Dogfooding is required for this skill, not optional.
- Dogfooding should validate generator changes, but it must not leave the repository in a state where the generator can no longer be built, tested, or evolved locally.
- Unless the user explicitly asked for a version or manifest-content change, treat Go version bumps, dependency version churn, and generated edits to `generate.yml` as suspect and review them critically before keeping them.
