<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Run the generator in GitHub Actions

Choose what the workflow step should do, then choose whether to use a published release or build an unreleased revision. Both actions verify the tool with `--version` before using it.

## Choose an action

| Action | Operating mode |
| --- | --- |
| `actions/generate` | Install the tool and run `go-sdk-generator generate` with structured inputs. |
| `actions/install-tool` | Install only; return a `cmd` output for later commands, loops or scripts. |

`workdir` determines where `generate` runs and how its relative paths are resolved. For install-only use, call the installed tool through the `cmd` output in subsequent steps.

## Choose a version mode

The action's `@ref` selects one of these modes. `action-ref` overrides that ref; `action-repo` selects a different repository, and `github-token` supplies access when the default workflow token is insufficient.

| Mode | Choose | Outcome |
| --- | --- | --- |
| **Published release** | An exact tag such as `@v0.47.0`, a major/minor ref such as `@v0` or `@v0.47`, or `@latest`. | Install the matching release. Major/minor refs select the latest matching stable release. |
| **Unreleased source** | A branch such as `@main` or a commit SHA. | Build and install the CLI from that revision, whether or not a release exists. |

With no ref (for example, a local action path without `action-ref`), the action tries the latest release. Using `./actions/generate` or `./actions/install-tool` does **not** by itself select a source build; set `action-ref` to a branch or SHA for that mode.

**Failure behaviour:** A release ref that cannot be resolved or installed fails the action; it neither skips execution nor switches to a source build. An unavailable branch/SHA or unsuccessful source build likewise fails. Check the ref and, if needed, the `github-token` permissions.

The [shared installer](https://github.com/ARM-software/go-sdk-generator/blob/main/actions/_shared/install-and-run.sh) and [action metadata](https://github.com/ARM-software/go-sdk-generator/blob/main/actions/generate/action.yml) define the exact ref resolution and inputs.

## Generate from a manifest

The manifest must exist in the checked-out workspace. This example installs a released binary and runs `generate`:

```yaml
steps:
  - uses: actions/checkout@v4
  - name: Generate Go project
    uses: ARM-software/go-sdk-generator/actions/generate@v0
    with:
      input: generate.yml
      output: .
      go-version: '1.26.4'
```

To build the tool from the current `main` branch instead of installing a release, replace `@v0` with `@main`.

Set `go-version` to the **generated project's** target Go version, not the version used to build the action. The action defaults to `summary=github`, `logger=rich` and `annotation=github`; `go-version` corresponds to the CLI flag `--go-version` (or `-g`).

## Install for later steps

```yaml
steps:
  - uses: actions/checkout@v4
  - name: Install generator
    id: generator
    uses: ARM-software/go-sdk-generator/actions/install-tool@v0
  - name: Inspect schema
    shell: bash
    run: '"${{ steps.generator.outputs.cmd }}" schema'
```

Use the install-only mode when later steps need custom shell processing. For container-based use, see [install and inspect](install-and-inspect.md). For agent-specific instructions, see the [GitHub Action integration skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/using-go-sdk-generator-in-a-github-action/SKILL.md).
