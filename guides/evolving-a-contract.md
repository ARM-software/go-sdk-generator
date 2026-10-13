<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Breaking changes and deprecation

Use this guide when changing an existing [generator manifest](write-a-manifest.md) or a generated SDK, CLI or automation surface that consumers already use. The aim is to make contract evolution reviewable and migratable.

## Identify a breaking change

A change is breaking when existing manifest authors or generated-project consumers must edit a working configuration, integration or workflow. Examples include:

- renaming or removing a field, method, command, parameter or type;
- changing the meaning of an existing field incompatibly;
- tightening validation so previously valid manifests become invalid;
- changing a value shape from scalar to list, value to reference, or one incompatible type to another;
- replacing one concept with several new ones, or merging several existing concepts into one.

## Deprecate before removing

A deprecation is a transition: the old concept still works but points users to its replacement. Where practical:

1. Add the replacement first.
2. Mark the old concept with `deprecation_notice` in the manifest.
3. Name the replacement and state when deprecation begins and, when known, when removal is planned.
4. Surface the notice in generated documentation, comments and CLI help.
5. Keep the deprecated behaviour working during the transition unless safety or correctness requires otherwise.
6. Remove it only after the announced window and a migration opportunity.

## Versioning and news files

This repository follows [Semantic Versioning](https://semver.org/) conceptually: breaking changes are major in theory, features minor and fixes patch. A true `v2+` release of a Go project also changes its module and consumer import paths; it is not merely a version-number bump ([Go modules](https://go.dev/ref/mod#major-version-suffix), [Go Modules: v2 and beyond](https://go.dev/blog/v2-go-modules)).

Do not treat every breaking manifest change as an automatic major-module migration. If a change breaks consumers without a full Go major-module transition, describe the break explicitly in a `.feature` news file, for example with a `:boom:` prefix. Reserve `.major` news files for a genuine major Go module migration, including its import-path consequences. See the [contribution guide](https://github.com/ARM-software/go-sdk-generator/blob/main/CONTRIBUTING.md#news-files) for news-file conventions.

In short: **deprecate, document, migrate, then remove**. For detailed authoring checks, use the [manifest-authoring skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/manifest-authoring/SKILL.md).
