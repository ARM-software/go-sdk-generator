<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# Embedded Agent Skills

This guide explains how Go SDK Generator includes its product skills inside its own binary. For installing and using skills, see the [Agent Skills guide](agent-skills.md) and [product reference](../DOCUMENTATION.md#agent-skills).

## Source, mirror and installed copies

```text
generator/templates/*.skill.md.tmpl
    | generate from generate.yml
    v
.agents/skills/<product-skill>/
    | copy selected complete directories after rendering
    v
generator/skills/<product-skill>/
    | go:embed at build time
    v
go-sdk-generator binary
    | agent --init in a target repository
    v
<target>/.agents/skills/<product-skill>/
```

The templates are the source of the generated instructions. `.agents/skills/` is the repository's discovery location and also contains handwritten contributor skills. `generator/skills/` is a generated build input, not a second independently maintained implementation.

The mirror exists because Go's `go:embed` paths are relative to the package directory and cannot reach a parent directory using `..`. Keeping the selected skills inside the `generator` package lets the binary provide them without downloading the source repository.

## Generation responsibilities

[`Generate`](../generator/generate.go) coordinates the stages. The skills-specific helpers live in [`skills_generation.go`](../generator/skills_generation.go):

- `addAgentFiles` selects agent instructions and product skills whenever a CLI is defined.
- `addSkillQualityFiles` selects the guide and validation tooling. GitHub workflows and the local validation action remain applicable when command actions are disabled.
- `copyGoSDKGeneratorSkillsForEmbedding` coordinates the build-time mirror, after the product skill templates have been rendered.
- `discoverEmbeddedSkillFiles` walks the selected directories and returns their regular files, including nested references, examples and scripts.
- `copyEmbeddedSkillFile` creates parent directories and copies one resource with its source permission bits.

The embedding step runs only when the manifest's `package_name` is `go-sdk-generator`. Ordinary generated CLI projects receive skills, documentation and validation tooling, but are not automatically given an embedded-skills API or an `agent --init` command.

`embeddedAgentSkillPaths` in [`skills.go`](../generator/skills.go) selects the generator's three product directories: run, install and GitHub Action integration. Each listed `SKILL.md` identifies a directory; the mirror includes the whole directory, not just that file. Unregistered contributor skills are not copied. Missing source directories are skipped.

## Overwrite rules and resource changes

Discovered mirror files are registered with the `agent` overwrite group before copying. Consequently, excluding that group also protects supporting resources. Exact file selectors still apply; when selecting individual outputs, include both the source and mirror paths if both should be refreshed.

Copying is an overlay: it updates selected files and does not prune obsolete mirror files. When removing or renaming a resource, review and remove its obsolete mirror explicitly. Keep `generator/skills/` free of unrelated content: `//go:embed all:skills` includes the entire tree, including hidden files.

When adding a new embedded product skill, update its template/generation wiring and the directory selection in `embeddedAgentSkillPaths`. Review the product catalogue and distribution selection too. Adding a file inside an already selected directory does not require another registry entry.

## Building and installing

1. Change the appropriate skill template, or supporting resources in the selected source skill directory.
2. Build the generator CLI and regenerate this repository from `generate.yml`, following [generate and regenerate a project](generate-a-project.md). Ensure the agent outputs and their mirror are included by any generation selectors.
3. Review `.agents/skills/` and `generator/skills/` together.
4. **Rebuild the CLI after refreshing the mirror.** The binary used to perform regeneration contains the old embedded files; editing files on disk does not change an already-built binary.
5. In a temporary target repository, run the rebuilt binary's `agent --init` command and inspect the installed skills.

`EmbeddedAgentSkills()` reads the compiled tree. `InitialiseAgentSkills()` writes those files under the target repository's `.agents/skills/`, creating nested directories as needed. New files are written with mode `0644`; scripts under a `scripts/` directory with a shebang use `0755` (subject to the operating system's permission handling). Existing matching target files are overwritten; unrelated files are not removed. This is an explicit command, not a package-install hook.

Linux packages follow a separate delivery path: GoReleaser packages the selected product directories directly from `.agents/skills/` into `/etc/<package-name>/.agents/skills/`. Both delivery paths originate from the same rendered product skills.

## Verification

Run the [local validation commands](agent-skills.md#validation-and-maintenance). Validation checks that the repository and embedded copies have matching complete file inventories and content, normalising text line endings. Those intentional mirrors are the only duplicate-name exception; other duplicate names fail validation.

The Go tests in [`skills_test.go`](../generator/skills_test.go) cover embedded loading, explicit installation, recursive resource copying and overwrite protection. The generation tests verify that CLI projects receive skill outputs and library-only projects do not. Regeneration changes should follow the [integration/regression workflow](../.agents/skills/integration-regression-tests/SKILL.md).
