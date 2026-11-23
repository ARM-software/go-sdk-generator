<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Generate selected parts of a project

Use granular generation to refresh selected parts of a generated SDK while preserving customised files. The **manifest** is the YAML or JSON input; the **generated SDK** is the output project written from it.

The examples below use `generate.yml`, the current directory as the output, and Go `1.27.3`. Adjust the manifest, destination and target Go version for your project. For the initial generation workflow, see [generate and regenerate a project](generate-a-project.md).

## Why use granular generation?

A generated SDK often develops project-specific changes: command handlers gain business logic, `go.mod` gains chosen dependency versions, and CI or lint configuration is adapted to the repository. A small manifest change may only require a few generated files to be refreshed.

The selection options let you match the scope of regeneration to the task. Exclusions preserve known customisations during a broad refresh; `only` selections limit a refresh to the outputs you actually need. This reduces unrelated changes and makes the resulting diff easier to review.

## Understand the default modes

### Without `--overwrite`

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3
```

The generator creates missing files and refreshes contract-driven outputs, including SDK source, CLI declarations and wrappers, `DOCUMENTATION.md`, generated actions, `agent.md`, Copilot instructions, run/action skills and `.gitattributes`.

Existing stateful scaffolding is generally preserved, including `README.md`, Go module/workspace files, CI and lint configuration, and CLI `implementations.go`. Version-sensitive Dockerfiles also follow the generator's target-Go-version update rule.

This is the ordinary workflow for keeping the generated contract in sync with the manifest while retaining stateful project customisations. It is **not** a create-missing-files-only mode: contract-driven files can still be rewritten.

### With `--overwrite` on its own

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite
```

All normal generated outputs are eligible for regeneration, including existing stateful scaffolding such as Go module files, workflows and CLI implementations. Built-in protection still applies to existing `module.properties` files.

This mode is useful when adopting updated generator templates across the whole project. Use the selection flags below when you want a narrower refresh or want to retain specific customisations.

## Choose what to generate or preserve

All four selection flags require `--overwrite`:

| Flag | Effect |
| --- | --- |
| `--only-groups` | Generate files belonging to the selected groups. |
| `--only-files` | Generate files matching the selected basenames or output-relative paths. |
| `--exclude-groups` | Preserve files belonging to the selected groups. |
| `--exclude-files` | Preserve files matching the selected basenames or output-relative paths. |

The rules are:

1. If neither `only` option contains values, all normal outputs are eligible for generation.
2. Otherwise, a file is eligible if it matches **either** `--only-groups` **or** `--only-files`.
3. A matching exclusion always wins, even when an `only` option selected the file.

Files outside the final selection are neither written nor deleted. Empty lists behave as unset options. Unknown group names are rejected; valid selectors that match no output are ignored. If nothing is selected, or exclusions remove the whole selection, the command succeeds without writing generated files.

Selection applies to supporting files too. For example, `--only-groups sdk` does not refresh `.gitattributes`, qlty configuration, documentation or action files unless you also select them.

## Available groups

Group membership uses exact generated output paths, including paths derived from the manifest's module and package names.

| Group | Generated outputs |
| --- | --- |
| `sdk` | SDK implementation, interfaces, configuration, additional-type packages, option packages and enum definitions. |
| `cli` | CLI entrypoints, command definitions, configuration, wrappers and supporting CLI source, excluding `implementations.go`. |
| `cli-implementation` | `<cli_package>/cmd/<main_module>/implementations.go`. |
| `gomod` | SDK and CLI `go.mod` files and the generated `go.work`. |
| `ci` | Workflows, GitHub Actions, release configuration, container build files and installation scripts. |
| `linters` | Lint, editor, pre-commit, Checkov, YAML lint and qlty configuration. |
| `readme` | The root `README.md`. |
| `project` | Repository support files such as `pyproject.toml`, `CODEOWNERS`, Dependabot files, contribution/security documentation, `.gitignore`, `.gitattributes` and `module.properties`. |
| `agent` | `agent.md`, Copilot instructions, generated skills and applicable embedded skill copies. |

A file can belong to more than one group. For example, the Dependabot workflow belongs to both `ci` and `project`; excluding either group preserves it.

## Refresh only the SDK source

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite --only-groups sdk
```

This updates SDK source and enum definitions while leaving CLI and repository support files alone.

Use it for an SDK-only contract change when the CLI and repository automation do not need to be refreshed.

## Preserve command implementations during a wider refresh

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite --exclude-groups cli-implementation,gomod
```

This preserves CLI command implementations and Go module/workspace definitions while regenerating the other eligible outputs.

This is useful when command handlers contain your own logic and module files contain deliberate dependency versions or `replace` directives.

## Combine group and file selections

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite --only-groups sdk --only-files README.md,.gitattributes
```

The result is the union: SDK source, the README and `.gitattributes`. Other supporting files remain untouched.

Use this when a contract update also needs refreshed project documentation and generated-file tracking, without updating CI or lint configuration.

## Exclude a file from a selected group

For the default CLI package name `cli`:

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite --only-groups cli --exclude-files cli/cmd/root.go
```

This refreshes the generated CLI scaffolding except `cli/cmd/root.go`. The separate `cli-implementation` group is not selected, so command implementations are also preserved.

This is useful when command definitions should follow the manifest but the root command contains project-specific initialisation or shared flags.

## Match files precisely

`--only-files` and `--exclude-files` share the same matching rules:

- `enums.go` matches that exact basename at any depth.
- `catalogue/options/enums.go` matches only that exact output-relative path. Replace `catalogue` with the manifest's `main_module`.
- `./README.md` explicitly targets the root README rather than every generated file with that basename.
- Paths are relative to `--output`, not to the manifest or current working directory.
- Both `/` and `\` separators are accepted and normalised.
- Matching is exact and case-sensitive. There are no substring, glob or recursive-directory matches.
- Absolute paths and paths escaping the output directory, such as `../README.md`, are rejected.

You can provide comma-separated values or repeat a flag:

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.27.3 --overwrite --only-groups sdk --only-groups readme
```

## After generation

The entire manifest is still validated, even when only a subset of outputs is selected. Existing protected files, such as `module.properties`, retain their normal protection when selected.

Review the resulting diff and run the appropriate checks in the affected Go modules. Commands such as `go generate ./...` and `go mod tidy` are separate operations: these selection flags do not control their writes, including generated enum parsers, mocks or checksum files.

For the current flag descriptions, run `go-sdk-generator generate --help` or consult the [command reference](../DOCUMENTATION.md).
