<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Run the generator in GitHub Actions

## Choose an action

- `actions/generate` installs the tool and runs `go-sdk-generator generate` with structured inputs.
- `actions/install-tool` installs the tool for later shell steps and exposes a `cmd` output.

Pin a released action ref: `@v0` selects the latest v0 release; an exact tag such as `@v0.47.0` selects that release. Branch names and commit SHAs build the tool from source at that ref. Actions at `./actions/...` use a local checkout when developing this repository.

## Generate from a manifest

The manifest must exist in the checked-out workspace. For example:

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

Use your target project's Go version. The action uses `summary=github`, `logger=rich`, and `annotation=github` by default. Check [`actions/generate/action.yml`](https://github.com/ARM-software/go-sdk-generator/blob/main/actions/generate/action.yml) for all inputs and their spelling; action inputs use `go-version`, whereas the CLI uses `--go_version`.

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

Use this pattern for loops or custom shell orchestration. The installer verifies the downloaded tool with `--version`; when no matching release exists, installation and execution are skipped. See the [action integration skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/using-go-sdk-generator-in-a-github-action/SKILL.md) for platform and version resolution details. For a container-based workflow, see [install and inspect](install-and-inspect.md).
