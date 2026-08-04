---
name: change-manifest-schema
description: Add or change fields and features in the generator manifest schema, regenerate generated files correctly, and dogfood the updated generator against its own manifest.
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
2. Run the built generator against `generate.yml`. The `generator` skill may be useful for that process, although the newly built CLI tool should be used rather than a previous release.
3. Inspect the generated output for regressions in the generator's own SDK and CLI.

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
