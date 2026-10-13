<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Generate and regenerate a Go project

## Inputs and first run

You need a manifest that follows the [schema](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml), an output directory, and a target Go version matching the project. From the project directory:

```sh
go-sdk-generator generate -i generate.yml -o . --go-version 1.26.4
```

Replace the Go version with the version your project uses. `-i`/`--input` selects the manifest, `-o`/`--output` selects the destination, and `-g`/`--go-version` selects the target Go version. Use `go-sdk-generator generate --help` for current options. The generator validates the manifest before producing files.

After generation, run these commands from each directory containing a generated `go.mod`:

```sh
go mod tidy
go generate ./...
go test ./...
```

`go mod tidy` resolves dependencies; `go generate ./...` runs generated directives, such as mock generation. Implement the generated operation stubs in Go and test the resulting behaviour. The manifest is the source of truth for the *surface*, not an implementation of the operations.

For a multi-module workspace, repeat the commands in each generated module rather than running `go mod tidy` from the workspace root. See [what the generator produces](what-it-generates.md) for the distinction between generated scaffolding and handwritten logic.

## Change the contract later

1. Edit the manifest, checking [manifest authoring](write-a-manifest.md) for type and compatibility guidance.
2. Re-run `go-sdk-generator generate` with the same input, output, and Go version.
3. Review the file diff before accepting it. Generation preserves stateful files by default; `--overwrite` also regenerates them, including files such as `go.mod`. Use it only when you intend to replace and review that state.
4. Run `go mod tidy`, `go generate ./...`, and the project's tests again in each affected Go module. Keep handwritten operation logic and project-specific wiring when resolving changes.

This repository generates parts of itself from [`generate.yml`](https://github.com/ARM-software/go-sdk-generator/blob/main/generate.yml). Its [development guide](https://github.com/ARM-software/go-sdk-generator/blob/main/DEVELOPMENT.md#code-generation) covers the repository-specific workflow. For CLI configuration, logging, and the full generated interface, consult [`DOCUMENTATION.md`](https://github.com/ARM-software/go-sdk-generator/blob/main/DOCUMENTATION.md).
