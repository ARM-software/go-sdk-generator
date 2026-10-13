<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Go SDK Generator: task guides

`go-sdk-generator` turns a validated YAML or JSON operation contract into a Go SDK, an optional CLI, and supporting project files. Use these guides to move from a manifest to a working project. The [generated command reference](https://github.com/ARM-software/go-sdk-generator/blob/main/DOCUMENTATION.md) lists every CLI option; the guides explain when and how to use them.

## Choose a task

- [Install and inspect the tool](install-and-inspect.md): choose a package or OCI image, check the CLI, and inspect the schema.
- [Write a manifest](write-a-manifest.md): define a focused project, choose types and constraints, and validate the contract.
- [What the generator produces](what-it-generates.md): SDK, CLI, action and documentation outputs, and what remains to implement.
- [Types, formats and defaults](types-and-defaults.md): choose built-in `lang_type` values and generated default behaviour.
- [Configure functional options](configure-options.md): enable option values and expose them through CLI flags.
- [Generate and regenerate a project](generate-a-project.md): run generation, implement operations, and preserve handwritten work.
- [Use the GitHub Actions](use-github-actions.md): run generation in a workflow or install the binary for later steps.
- [Manage secrets in a generated CLI](secret-management.md): mark secret inputs, use keyring-backed commands and understand security boundaries.
- [Evolve a contract safely](evolving-a-contract.md): deprecate incompatible manifest and generated-surface changes.

## Further reference

- [Schema source](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml): authoritative manifest field definitions.
- [This repository's manifest](https://github.com/ARM-software/go-sdk-generator/blob/main/generate.yml): a real generator input.
- [Generated SDK and CLI reference](https://github.com/ARM-software/go-sdk-generator/blob/main/DOCUMENTATION.md): configuration, commands, and Go interfaces.
- [Repository README](https://github.com/ARM-software/go-sdk-generator/blob/main/README.md): purpose, quick start and design decisions.
- [Maintained agent skills](https://github.com/ARM-software/go-sdk-generator/tree/main/.agents/skills): task-specific instructions for agents working in this repository. The tool can also install its embedded skills with `go-sdk-generator agent --init`.

The Markdown in `guides/` is the source. `cd-generate-docs` renders it alongside the API documentation in `docs/guides/` when documentation is rebuilt; edit the Markdown rather than the rendered output.
