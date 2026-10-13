<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# What the generator produces

The [manifest](write-a-manifest.md) defines one concern and the user-facing operations around it. Depending on that contract, `go-sdk-generator` can create:

- a Go module layout for the SDK, with exported configuration and domain types;
- interfaces and implementation scaffolding for SDK operations;
- an optional [Cobra](https://cobra.dev/)-style CLI backed by the same configuration model and [Viper](https://github.com/spf13/viper)-style configuration sources;
- a reusable Go-based entrypoint for step-level [GitHub Actions](use-github-actions.md);
- generated documentation and supporting project files.

The [types and defaults reference](types-and-defaults.md) describes built-in value formats. The [generated command reference](https://github.com/ARM-software/go-sdk-generator/blob/main/DOCUMENTATION.md) covers this repository's Go SDK and CLI interfaces.

## What you still implement

Generation provides structure and boilerplate, not the domain-specific logic behind each operation. Implement the generated stubs in Go, then [regenerate and review the project](generate-a-project.md) when its contract changes.

## One operation, several ways to use it

- Import the generated SDK in another Go project and call its methods directly.
- Run the generated CLI locally, in scripts or in a container.
- Use a generated action in CI to invoke the same Go tool rather than duplicating its logic in shell or JavaScript.

The same implementation can serve library users, command-line users and automation while remaining testable as Go code. A generated action represents an operation or a step; a [GitHub Actions workflow](https://docs.github.com/actions/using-workflows) or task runner still defines how multiple steps fit together. See [use GitHub Actions](use-github-actions.md) for workflow examples.

As the generator gains improvements to documentation, support files, defaults and repository conventions, regenerated projects can adopt them. Review regenerated files, preserve handwritten implementation logic and [manage incompatible contract changes](evolving-a-contract.md) deliberately.
