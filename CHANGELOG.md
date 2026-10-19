<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# 0.50.1 (2026-10-19)

## Bugfixes

- :green_heart: Run TPIP compliance in generated public CI, declare the root LICENSE for CDS, and include the licence file in released tool packages under `/etc/<tool>/LICENSE` on Linux. (#202610151230)
- Dependency upgrade: golang-1.27.1 (#20261014060409694259825)
- Dependency upgrade: tools-0.52.0 (#20261014060412873157088)


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# 0.50.0 (2026-10-13)

## Features

- :boom: Generated option values now default to the option name and snake_case tag; replace option-level bindings with value_field to customise them (old option-level bindings are no longer accepted). (#20261007170813587457600)
- :sparkles: Add an Enabled field to generated functional options so callers and CLI users can explicitly enable or disable individual options without relying on empty values. (#20261008112438055755700)
- Validate manifest `condition` definitions so invalid condition shapes are rejected during manifest validation. (#20261008165000000000000)

## Bugfixes

- :bug: Correct generated GitHub Action skills to report missing releases as failures and recognise source builds. (#20261013)
- :bug: Preserve the correct licence and repository links when regenerating this project's agent skills. (#2026101301)

## Improved Documentation

- :book: Add linked generator guides for installation, outputs, types, defaults, functional options, secret management and contract evolution, with task-specific agent skills. (#20261012)


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# 0.49.1 (2026-10-06)

## Bugfixes

- :bug: Enable CDS release-time licence summary generation into `docs/` and use the latest CDS install in the generated public release workflow. (#202610061145)


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
# 0.49.0 (2026-10-06)

## Features

- :building_construction: Add generated public-project SBOM publishing and third-party licence compliance CI workflows, and expose the latest SPDX SBOM run from the generated README. (#202610021840)

## Bugfixes

- :bug: Prefix generated Towncrier release titles with `# ` so markdown changelogs keep a proper top-level heading after the [Towncrier formatting change](https://github.com/twisted/towncrier/issues/758). (#202610021955)
- :bug: Record manual CDS licence approvals for `github.com/stretchr/testify/internal/spew`, `sigs.k8s.io/yaml`, and `github.com/klauspost/compress` so the project's compliance check explains and accepts those permissive dependencies. (#202610022015)
- Dependency upgrade: testify-1.12.1 (#20260820214537, #20260820214543)
- Dependency upgrade: mod-0.41.0 (#20260908214459)
- Dependency upgrade: text-0.42.0 (#20260911214446)
- Dependency upgrade: v4-4.12.0 (#20260916214526)
- Dependency upgrade: alpine-3.24.2 (#20260921214543)
- Dependency upgrade: golang-1.27rc3 (#20261002160300251099622)
- Dependency upgrade: utils-1.170.3 (#20261006121313596500320)
- Dependency upgrade: tools-0.50.0 (#20261006121328249371977)

## Misc

- #202610022035


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.48.1" (2026-08-20)
=====================

Bugfixes
--------

- :bug: Fix show-config so pointer-backed values such as `SDK_GENERATOR_TOOLVERSION` are printed with their dereferenced string value instead of only a raw pointer address. (#202608201132)
- :bug: Run `cd-license-files` in the generated upgrade-generator workflow after regeneration so licence headers stay in sync with automated generator updates. (#202608201220)
- :broom: Rename repository documentation and skill references to `ghcr.io/arm-software/go-sdk-generator`. (#202608201330)


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.48.0" (2026-08-20)
=====================

Features
--------

- :seedling: Open-source the project so generating boilerplate code for Go tools can be made easier. (#20260819135603)


Bugfixes
--------

- :bug: Fix generated CI workflows to avoid unsupported YAML anchors in GitHub Actions. (#202608191355)
- Dependency upgrade: go-sdk-generator-0.46.0 (#20260818144845, #20260818145007, #20260818150923)
- :camel: Regenerate project output after upgrading go-sdk-generator version 0.46.0. (#20260818144921)
- :camel: Regenerate project output after upgrading go-sdk-generator version 0.47.0. (#20260818150908)
- Dependency upgrade: testify-1.12.0 (#20260818152324, #20260818152408)
- Dependency upgrade: v4-4.11.0 (#20260818152342)
- Dependency upgrade: tools-0.49.0 (#20260818152404)
- Dependency upgrade: create-pull-request-8 (#20260818154523)
- Dependency upgrade: mod-0.39.0 (#20260818154626)


<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.47.0" (2026-08-18)
=====================

Features
--------

- :sparkles: Add manifest support for generated functional options packages. (#202608171215)
- :sparkles: Add configurable generated CLI and action logging and annotation support, including custom and rich logger support. (#20260812022604)
- :sparkles: Reject manifest-defined CLI configuration fields that reuse protected built-in argument names. (#20260813153433)
- :sparkles: Add manifest-driven result output support for CLI commands with human, JSON, and YAML destinations. (#20260814095502)
- :recycle: Set generated enum zero values to match the default option where applicable. (#20260814101526)
- :sparkles: Add parameter_value.when to control always, empty, and non-empty comparisons. (#20260817112135)


Bugfixes
--------

- :builder: Add a workflow that reports Go dependency licences, runs `go-licenses check`, and publishes an SPDX SBOM artifact. (#202608130135)


Deprecations and Removals
-------------------------

- :broom: Remove unused template helper registrations and dead helper code from the generator. (#202608171620)


Misc
----

- #202608171735, #202608171905


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.46.0" (2026-08-12)
=====================

Features
--------

- :sparkles: Generated CLIs can now opt commands into summary-aware wrappers and GitHub Actions default those commands to `--summary=github`. (#202608061300)
- :sparkles: Generated CLIs can now emit human-readable summaries, including GitHub step summaries and file summaries, with corresponding generated validation and GitHub Action defaults. (#202608071430)
- :sparkles: The generator schema now defines built-in summary destinations, built-in `lang_type` values, and semantic `default_value` tokens so generator logic can use generated schema constants instead of hard-coded strings. (#202608071431)
- :sparkles: Generated validation now uses the newer golang-utils field-oriented helpers and state-style `Required` semantics so required zero values such as `false` and `0` remain valid in generated SDK and CLI configuration. (#202608071432)
- :recycle: Require explicit default values for enum-defined parameters to avoid ambiguity or accidental misinterpretation when generated configuration values are left unset. (#20260810123000)
- :sparkles: Add a generated `--show-config` CLI flag that logs the effective per-command configuration before execution, with generated wrapper tests and redaction support for sensitive values. (#20260812000500)
- :recycle: Move generated show-config CLI options to the shared root command path. (#20260812100355)


Bugfixes
--------

- :bug: Fix generated delegate forwarding for interface arguments declared with explicit variadic `lang_type` values such as `...string`. (#202608111019)
- Dependency upgrade: generator-0.45.0 (#20260810043555)
- :camel: Regenerate project output after upgrading go-sdk-generator version 0.45.0. (#20260810043810)
- CLI enum-backed configuration fields now generate as strings while still emitting enum constants and `enumer` helpers for validation and runtime parsing. This avoids relying on Viper to decode string inputs like `"file"` into generated enum types, which it does not do automatically through the default mapstructure decode hooks. (#20260810101500)
- :recycle: Improved the generated constraint-validation and CLI enum implementation so the resulting code is easier to read and follow, while keeping enum-backed CLI values parsed through generated helpers instead of relying on implicit Viper conversion. (#20260810114500)
- :green_heart: Regression fixture tests now build every generated Go module with `go build ./...` before comparing output against fixtures or rewriting fixture archives, so non-compiling generated output is caught before snapshot updates can hide it. (#20260810143000)


Improved Documentation
----------------------

- :book: Clarify the dogfooding skill guidance so contributors pass the repository's current Go version explicitly during self-generation instead of changing the manifest default value. (#20260812121500)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.45.0" (2026-08-05)
=====================

Features
--------

- :sparkles: Add explicit manifest parameter constraints, including CLI-only required-together flag relationships, and generate clearer constraint validation output. (#20260805114500)
- :sparkles: Add many more manifest parameter constraint rules so manifests can describe parameters as precisely as possible, with validation that roughly aligns with JSON Schema and OpenAPI-style constraints. (#20260805161604)


Bugfixes
--------

- :camel: Regenerate project output after upgrading go-sdk-generator version 0.44.1. (#20260805072119, #20260805072226, #20260805072329)
- Dependency upgrade: setup-go-7 (#20260805072235)
- Dependency upgrade: alpine-3.24.1 (#20260805072247)
- Dependency upgrade: generator-0.43.4 (#20260805074655)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.44.1" (2026-08-05)
=====================

Bugfixes
--------

- :bug: Fix Dependabot generation so GitHub Actions updates are included and private registries are only used for non-public projects. (#20260804223000)
- :camel: Regenerate project output after upgrading go-sdk-generator version 0.43.4. (#20260804230724, #20260804232538, #20260804232554, #20260804232718)
- Dependency upgrade: go-sdk-generator-0.43.4 (#20260804230731)
- Dependency upgrade: setup-go-7 (#20260804232548)
- Dependency upgrade: checkout-7 (#20260804232607)
- Dependency upgrade: create-pull-request-8 (#20260804232610)
- Dependency upgrade: ubuntu-26.04 (#20260804232624)
- Dependency upgrade: alpine-3.24.1 (#20260804232635)
- :bug: Preserve generated workflow files unless overwrite is requested, and only regenerate generated Dockerfiles automatically when the generator default Go version is newer than the targeted version. (#20260805001000)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.44.0" (2026-08-04)
=====================

Features
--------

- :sparkles: Add built-in filepath, existingFilepath, pathParameter, and validJSON parameter aliases with stronger generated validation. (#20260723230916)
- :sparkles: Add generated `upgrade-generator` workflows to update the SDKs file using the generator. (#20260724010118)
- :sparkles: Add structured deprecation notices to generated parameters, methods, types, and CLI commands. (#20260724183610)


Bugfixes
--------

- :book: Document GHCR authentication troubleshooting for OCI image access in generated README, documentation, and tool-related skills. (#20260723115352)
- :camel: Update dependencies (#20260723225345)
- :recycle: use validation rules from `golang-utils` (#20260723225414)
- Dependency upgrade: utils-1.167.0 (#20260724042401, #20260724042406)
- Dependency upgrade: v4-4.10.1 (#20260727043213)
- :book: Add a README badge showing the go-sdk-generator version used for generated output. (#20260803094421)
- Dependency upgrade: utils-1.168.0 (#20260804042403)
- :green_heart: Fix the upgrade-generator workflow so Dependabot updates run correctly, workflow-file changes can be pushed, and regenerated output creates a versioned news file. (#20260804170655)
- :bug: fix the generator upgrade CI flow (#20260804224411)


Improved Documentation
----------------------

- :book: Indicate the generator version used to generate the project structure in the README (#20260803095047)
- :book: Improve the description of some of the skills of this project (#20260804143640)
- :broom: Remove out of date generated skills (#20260804143701)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.43.4" (2026-07-22)
=====================

Bugfixes
--------

- :bug: Use the working GitHub Container Registry package page URL in OCI README badges and links. (#20260722174737)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.43.3" (2026-07-22)
=====================

Bugfixes
--------

- :bug: Disable Alpine OCI image publishing until goreleaser/nfpm APK installation failures are resolved upstream. (#20260722170732)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.43.2" (2026-07-22)
=====================

Bugfixes
--------

- :bug: Assign explicit GoReleaser package IDs when splitting Alpine and Debian-style Linux package definitions. (#20260722161651)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.43.1" (2026-07-22)
=====================

Bugfixes
--------

- :bug: Work around nfpm APK tree packaging issues in OCI image builds while keeping Debian-style package trees for tool-related skills. (#20260722153840)
- :lock: Run generated OCI images as a non-root user by default. (#20260722154523)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.43.0" (2026-07-22)
=====================

Features
--------

- :sparkles: Add OCI distribution support with multi-architecture images, GHCR publishing, and packaged skills related to the tool. (#20260722115521)


Bugfixes
--------

- Dependency upgrade: v4-4.10.0 (#20260720043208)
- :recycle: Change documentation templates and content to be more accurate (#20260722002912)
- :recycle: Update `generator` definition so the different commands it has are correct (#20260722053720)
- Dependency upgrade: golang-1.27rc2 (#20260722084756)
- :book: Include skills related to the tool in generated Linux packages. (#20260722115551)


Improved Documentation
----------------------

- :book: improve the generated documentation to better describe what is generated. (#20260721223951)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.42.1" (2026-07-16)
=====================

Bugfixes
--------

- :bug: Generated `pre.sh` now creates the `.manpages` directory with `mkdir -p`, so re-running the script (or running it when `.manpages` already exists) no longer fails. (#202607161431)
- :bug: validate generated CLI argument names and short flags for uniqueness after preprocessing (#20260715185003)
- :bug: Fix import issues in `[cli]` (#20260716135505)
- :camel: Update dependencies (#20260716140359)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.42.0" (2026-07-15)
=====================

Features
--------

- :sparkles: support example and examples metadata on manifest field values, validate them against compatible types, and surface them in generated documentation, skills, CLI help, and action input descriptions. (#20260710010415)


Bugfixes
--------

- Dependency upgrade: generator-0.41.0 (#20260715042344)
- :bug: fix some language spelling mistakes (#20260715132455)
- Dependency upgrade: utils-1.165.2 (#20260715170450)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.41.0" (2026-07-14)
=====================

Features
--------

- :sparkles: add a dedicated install-tool GitHub Action for generated CLIs (#20260711192747)


Bugfixes
--------

- Dependency upgrade: generator-0.40.1 (#20260710042403)
- :memo: improve README structure, badges, and agent guidance links (#20260714145102)
- :bug: fix global configuration field settings to use the correct casing as dogfooding the generator revealed a bug was present in generated configuration (#20260714164956)


Improved Documentation
----------------------

- :memo: improve generated documentation for actions, manifests, and exported configuration fields (#20260714145102)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.40.2" (2026-07-11)
=====================

Bugfixes
--------

- :bug: fix goreleaser template selection for CLI manifests (#20260711190829)


Improved Documentation
----------------------

- :book: Improve generator documentation and how it compares to close but different types of tools/generators (#20260710103141)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.40.1" (2026-07-09)
=====================

Bugfixes
--------

- :bug: default generated GitHub Action inputs such as no-keyring to safer values when the generated CLI expects those flags to be enabled by default in automation. (#20260709172846)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.40.0" (2026-07-09)
=====================

Features
--------

- :sparkles: allow generated CLI projects to declare their top-level package directory through `command_line_commands.package`. This should help use of tools directly as [go tools](https://go.dev/ref/mod#go-mod-file-tool) without having them named `cli`. See https://go.dev/doc/modules/gomod-ref#tool (#20260709100915)


Bugfixes
--------

- Dependency upgrade: generator-0.39.0 (#20260709042355)
- Dependency upgrade: tools-0.48.0 (#20260709042505)
- :bug: fix generated enum validation and enum constant emission so enum-backed fields do not emit incorrect required validation and generated enum const blocks work correctly with go generate. (#20260709142002)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.39.0" (2026-07-08)
=====================

Features
--------

- :sparkles: allow additional generated types to declare their package and emit grouped type code outside the default types package. (#20260708135916)


Bugfixes
--------

- Dependency upgrade: generator-0.38.0 (#20260708042352)
- :bug: fix generated defaults for locally defined additional types so blank defaults resolve to the appropriate Default<Type>() constructor across generated packages. (#20260708135916)
- :bug: fix generated additional type and documentation text so acronym-heavy type names and descriptions preserve expected casing, including custom types such as HybridAESRSAEncryptedPayload and acronym-led phrases. (#20260708143036)
- :bug: fix generated comments and documentation so acronym-heavy identifiers and descriptions preserve expected casing instead of producing corrupted words like aI or hTTP. (#20260708143614)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.38.0" (2026-07-07)
=====================

Features
--------

- :sparkles: Add manifest support for list and iterator parameter shapes in generated types and method signatures. (#20260701113626)
- :recycle: improve the validation of manifest using the embedded json schema (#20260701132128)
- :sparkles: Add string enum manifest support with generated Go enums, validation, and CLI help text. (#20260702124212)


Bugfixes
--------

- Dependency upgrade: generator-0.37.1 (#20260626042351)
- Dependency upgrade: v4-4.9.0 (#20260629042716)
- Dependency upgrade: tools-0.47.0 (#20260630135426)
- Dependency upgrade: utils-1.164.1 (#20260702042343, #20260702042351)
- :bug: Generate StringSlice CLI flags for string list parameters. (#20260702101210)
- Dependency upgrade: text-0.39.0 (#20260707042357)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.37.1" (2026-06-25)
=====================

Bugfixes
--------

- :bug: Fix generated custom lang_type references so acronym-normalised types are used consistently. (#20260624163640)
- :bug: Fix generated method and field references for custom lang_type values so acronym-normalised type names are used consistently. (#20260624191214)
- Dependency upgrade: generator-0.37.0 (#20260625042354)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.37.0" (2026-06-24)
=====================

Features
--------

- :sparkles: Add built-in `connection` manifest types for generating HTTP request configuration defaults. (#20260624134227)


Bugfixes
--------

- :bug: Fix the licence value in some frontmatter (#20260624141852)


Improved Documentation
----------------------

- :book: Added a skill for authoring a manifest (#20260624144410)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.36.0" (2026-06-24)
=====================

Features
--------

- :sparkles: Update the list of acronyms to cover most of the ones we are likely to encounter when creating SDKs or CLIs (#20260624112411)


Bugfixes
--------

- Dependency upgrade: generator-0.35.0 (#20260619042337)
- :sparkles: Updated the acronyms list to handle HAL properly (#20260623182545)
- :bug: Fix the support for multiline descriptions (#20260624110436)


Improved Documentation
----------------------

- :book: Improve the README and generator skill documentation to better explain the generator rationale and usage. (#20260623141229)
- :book: Added a skill to document how to add to the acronym list (#20260623222141)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.35.0" (2026-06-18)
=====================

Features
--------

- :recycle: Better handle what is generated when a project is `public` (#20260617151706)
- :lock: Introduced a `secret` flag to better handle CLI parameters so they can be stored in secret stores such as keyring (#20260617183859)


Bugfixes
--------

- Dependency upgrade: generator-0.34.2 (#20260605042353)
- Dependency upgrade: alpine-3.24.0 (#20260610042332)
- Dependency upgrade: utils-1.161.2 (#20260610042347)
- Dependency upgrade: tools-0.46.0 (#20260612042351)
- Dependency upgrade: v4-4.8.0 (#20260615043142)
- Dependency upgrade: alpine-3.24.1 (#20260617042332)
- Dependency upgrade: utils-1.162.0 (#20260617042405, #20260617042406)
- :recycle: Improve how acronyms are handled and maintained to avoid situations like having functions called `IDentify` (#20260618133942)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.34.2" (2026-06-04)
=====================

Bugfixes
--------

- Dependency upgrade: alpine-3.23.4 (#20260604133914)
- Dependency upgrade: text-0.37.0 (#20260604133936)
- Dependency upgrade: v2-2.9.0 (#20260604133947)
- :dog: Dog fooding (#20260604143048)
- Dependency upgrade: utils-1.161.1 (#20260604151644, #20260604154601)
- Dependency upgrade: tools-0.45.0 (#20260604160426)


Misc
----

- #20260604145518


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.34.1" (2026-06-04)
=====================

Bugfixes
--------

- :bug: Ensure that install fails if it cannot find a release that has been requested (#20260604131824)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.34.0" (2026-06-04)
=====================

Features
--------

- :sparkles: Improve generated Dockerfiles and workflow templates for security and lint compliance (#202606012206)
- :recycle: Moving to [go workspace](https://go.dev/doc/tutorial/workspaces) (#20260603192548)


Bugfixes
--------

- Dependency upgrade: alpine-3.23.4 (#20260603220741)
- Dependency upgrade: golang-1.26.4 (#20260603220747)
- :bug: Ensure that `inputs.github-token` is preferred over `github.token` (#20260604112519)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.33.1" (2026-06-01)
=====================

Bugfixes
--------

- :lock: Update generated Dockerfiles to run CLI container images as a non-root app user. (#20260528175202)
- :bug: Update ci.yml.tmpl to always run go generate in library CI (#20260529140325)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.33.0" (2026-05-26)
=====================

Features
--------

- :sparkles: Add generated support for a follow-up public release upload job in release workflows (#202605211125)
- :sparkles: Add sub-commands for use when working with agents (#20260417115900)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.32.3" (2026-05-20)
=====================

Bugfixes
--------

- Dependency upgrade: alpine-3.23.4 (#20260519165215)
- :sparkles: Ensure that package name is used when downloading assets for the GitHub actions (#20260520160031)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.32.2" (2026-05-20)
=====================

Bugfixes
--------

- :bug: Ensure that the binary name is derived from the binary name in the schema, not the package name (#20260519135312)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.32.1" (2026-05-19)
=====================

Bugfixes
--------

- :bug: Require a github token with the correct scopes for fetching released artefacts (#20260519094237)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.32.0" (2026-05-14)
=====================

Features
--------

- :sparkles: Add ADR documentation for generator architecture decisions (#20260511150637)
- :sparkles: Add a `grill-me` skill, inspired by Matt Pocock's equivalent skill, to challenge plans against generator documentation and decisions. It should help with supporting AI-assisted generator change. (#20260511160059)
- :book: Update the contribution guide with clearer news file and pull request guidance (#20260511164237)
- :sparkles: Add a skill (`change-manifest-schema`) for handling changes to the manifest schema. (#20260512115806)
- :sparkles: Support generating GitHub-hosted tools outside the `Arm-Debug` organisation (#20260513134817)
- :sparkles: Support generating tools that are not hosted on GitHub (#20260513134837)
- :sparkles: Support custom SDK namespace (#20260513134925)


Bugfixes
--------

- Dependency upgrade: golang-1.26.3 (#20260508043915)
- Dependency upgrade: alpine-3.23.4 (#20260511105834)
- :bug: Fix some small rendering issues in markdown files (#20260513153357)


Improved Documentation
----------------------

- :book: Document the rationale behind the separation of SDK namespace and repository URL (#20260512132538)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.31.0" (2026-04-27)
=====================

Features
--------

- :sparkles: Automatically create an action per sub-command (#20260416112141)


Bugfixes
--------

- :gear: Ensure that unreleased tool fetching fetches from the correct repository (#20260427144453)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.30.2" (2026-04-21)
=====================

Bugfixes
--------

- :bug: Add missing 'Configure Git with secret' step in dogfood-action job (#20260421145831)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.30.1" (2026-04-21)
=====================

Bugfixes
--------

- :bug: Fix archive installs for packages which have a binary name that differs from the package name (#20260421123350)
- :bug: Add missing 'Set up Go' step in dogfood-action job (#20260421123628)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.30.0" (2026-04-09)
=====================

Features
--------

- :sparkles: The generator now also creates a GitHub Action that wraps the CLI, making it straightforward to integrate and use in workflows. (#20260409110310)
- :sparkles: Add option not to generate an action if the CLI does not seem to be used in this context (#20260409163139)
- :sparkles: Add some documentation about the generated actions (#20260409163351)


Bugfixes
--------

- Dependency upgrade: golang-1.26.2 (#20260408043246)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.29.0" (2026-04-02)
=====================

Features
--------

- :sparkles: Added built-in `default_value` shortcuts to ease the definition of some common defaults. (#202604011231)


Improved Documentation
----------------------

- :book: Documented supported `lang_type` values and built-in `default_value` shortcuts in the [README formats section](README.md#formats). (#202604011230)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.28.0" (2026-03-26)
=====================

Features
--------

- :robot: Improve integration with Github and `copilot` (#20260326142407)


Bugfixes
--------

- :bug: Fix documentation link for the CLI usage (#20260326142230)
- :robot: Improve installation skill and documentation (#20260326142313)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.27.0" (2026-03-26)
=====================

Features
--------

- :bug: Reject mixed callCloud and explicit root auth config (#20260320114452)
- :sparkles: Generate `SKILL.md` so the generated tool is :robot: compatible (#20260325001405)


Bugfixes
--------

- :dog: Dogfooding (#20260324231703)
- :camel: Upgrade dependencies (#20260325001239)
- :bug: Make sure files which need to be regenerated on manifest change are (#20260325001327)
- :bug: Fix gitattributes content on Windows (#20260325001534)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.26.2" (2026-03-19)
=====================

Bugfixes
--------

- :bug: Ensure that the release workflow is valid and uses the target go version in the inputs for the release workflow (#20260317130805)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.26.1" (2026-02-25)
=====================

Bugfixes
--------

- :bug: Fix comment syntax on `go.mod.cli.tmpl` (#20260225160803)
- :gear: Include generation comment in `go.mod.sdk.tmpl` (#20260225160955)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.26.0" (2026-02-20)
=====================

Features
--------

- :sparkles: Generate .gitattributes file to ensure that generated files are actually recognised by GitHub. See [this page](https://docs.github.com/en/repositories/working-with-files/managing-files/customizing-how-changed-files-appear-on-github) for more information. (#20260219154802)


Bugfixes
--------

- :gear: Update release.yml template (#20260218155956)
- :bug: Add back accidentally deleted tests (#20260219175758)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.25.0" (2026-02-16)
=====================

Features
--------

- :sparkles: Added an [EditorConfig](https://editorconfig.org/) to help with files formatting (#20260212174129)


Bugfixes
--------

- Dependency upgrade: alpine-3.23.3 (#20260216100308)
- Dependency upgrade: golang-1.26rc2 (#20260216105014)


Misc
----

- #20260212165336, #20260212174040


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.24.0" (2026-01-29)
=====================

Features
--------

- :sparkles: Allow callCloud to be defined at the top level of the CLI manifest with conditional root-level handling of APIKey and Host arguments (#20260123174444)


Bugfixes
--------

- :bug: Add the generator name and version to the header comment of all generated files. Add a badge to Documentation.md indicating the go-sdk-generator version used. (#20260127011427)
- Dependency upgrade: alpine-3.23.3 (#20260128053123)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.23.0" (2026-01-23)
=====================

Features
--------

- :sparkles: Support mapping []string to StringArray argument in cobra (#20260122131137)


Bugfixes
--------

- :bug: Trim trailing newlines when generating multiline YAML default values (| / >) (#20260120142508)
- :bug: Fix docstrings not followinng standard Go documentation conventions (#20260121143611)
- :bug: Fix Qlty template to disable start line. Rename unused argument in root template. Add targeted Go version flag in development template (#20260122091519)
- :bug: Ensure that the use of cobra.Flags().MarkHidden() is assigned to _ = so the the linter is happy (#20260122165416)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.22.3" (2026-01-16)
=====================

Bugfixes
--------

- :bug: Add build cli job in ci workflow (#20260116123201)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.22.2" (2026-01-15)
=====================

Bugfixes
--------

- 🐛 : Generate correct command in development template, adding missing keyword , using schema file name instead of module name as input path name (#20260115115730)


Improved Documentation
----------------------

- :sparkles: Update development template, adding github secret note (#20260114090753)


<!--
Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.22.1" (2026-01-02)
=====================

Bugfixes
--------

- :bug: fix API key regex which is not supported by go (#20260102144124)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.22.0" (2025-12-31)
=====================

Features
--------

- :sparkles: add ability to specify the licence (#20251231113050)


Bugfixes
--------

- :bug: fix argument/parameter name casing (#20251231111234)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.21.0" (2025-12-30)
=====================

Features
--------

- :sparkles: Better load configuration from system/environment (#20251230141827)


Bugfixes
--------

- :camel: Upgrade dependencies (#20251230141756)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.20.0" (2025-12-30)
=====================

Features
--------

- :sparkles: Add ability to mark a CLI function as Hidden or SilenceUsage (#20251229165247)


Bugfixes
--------

- :bug: Fix typo in template (#20251229165334)
- :gear: Handle automatically the different extensions for yaml (#20251229171340)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.19.0" (2025-12-29)
=====================

Features
--------

- :sparkles: Add a way to specify a command calls Arm online services and so, required arguments will be added if missing (#20251222112700)


Bugfixes
--------

- :camel: Upgrade dependencies (#20251222112614)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.18.0" (2025-12-19)
=====================

Features
--------

- :sparkles: Generate a valid module name with invalid recipe input (#20251218170225)
- :bug: Fix dependabot template file (#20251218170251)


Bugfixes
--------

- Dependency upgrade: alpine-3.23.2 (#20251218051818)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.17.0" (2025-12-16)
=====================

Features
--------

- :sparkles: Add support for controlling the exit code when the command receives an interrupt signal (#20251215161106)


Bugfixes
--------

- Dependency upgrade: golang-1.25.5 (#20251203051707)
- Dependency upgrade: alpine-3.23.0 (#20251204051747)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.16.1" (2025-11-19)
=====================

Bugfixes
--------

- Dependency upgrade: alpine-3.22.2 (#20251009041838)
- :bug: Ensure that pointers are shown correctly in the generated documentation (#20251020144032)
- Dependency upgrade: golang-1.25.4 (#20251110051731)
- :bug: Ensure that `decimal` type uses floats as the default value (#20251118153521)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.16.0" (2025-10-07)
=====================

Features
--------

- :sparkles: Support go-releaser version 2 (#20251007104310)
- :sparkles: Support detect-secrets version 1.5.0 (#20251007104324)


Bugfixes
--------

- Dependency upgrade: golang-1.25.1 (#20250925041555)
- :bug: Ensure that root flags specified as root across multiple commands are not duplicated (#20251007102340)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.15.0" (2025-09-24)
=====================

Features
--------

- :sparkles: Add support for hidden flags. See the `Hidden` field in cobra https://pkg.go.dev/github.com/spf13/cobra. `MarkHidden()` can be used to hide a flag. (#20250828120011)


Bugfixes
--------

- Dependency upgrade: alpine-3.22.1 (#20250812130420)
- Dependency upgrade: golang-1.25.0 (#20250814043649)
- :bug: fix close stack overflow (#20250924164325)
- :camel: Upgrade dependencies (#20250924165037)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.14.0" (2025-08-11)
=====================

Features
--------

- :sparkles: define the manager as Closeable (#20250811133109)
- :sparkles: define short arguments for the generator (#20250811151440)


Bugfixes
--------

- Dependency upgrade: golang-1.24.5 (#20250709050051)
- Dependency upgrade: alpine-3.22.1 (#20250716045300)
- Dependency upgrade: golang-1.24.6 (#20250807050811)
- :camel: Upgrade dependencies (#20250811133148)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.13.1" (2025-07-08)
=====================

Bugfixes
--------

- :bug: Make sure that port validation rule expects a uint32 instead of a string (#20250707161852)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.13.0" (2025-07-07)
=====================

Features
--------

- :sparkles: Add support for binding multiple flags to a single argument (#20250704173259)
- :sparkles: Only regenerate .goreleaser.yml if overwrite is enabled (#20250707150640)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.12.3" (2025-07-04)
=====================

Bugfixes
--------

- :bug: Make sure that gosec tags don't remove newlines in CLI template (#20250703155657)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.12.2" (2025-07-01)
=====================

Bugfixes
--------

- :bug: Make sure that gosec tags don't remove newlines in root template (#20250701093318)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.12.1" (2025-06-25)
=====================

Bugfixes
--------

- :bug: Make sure that you can't specify -h or -v short args (#20250625140230)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.12.0" (2025-06-24)
=====================

Features
--------

- :sparkles: Add support for specifying short args on the CLI functions (#20250623152933)


Bugfixes
--------

- :bug: Make sure that any extra LangTypes are included in the ExtraValidTypes (#20250623152922)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.11.2" (2025-06-20)
=====================

Bugfixes
--------

- :bug: Explicitly call `chmod scripts/pre.sh` so that SDKs generated on windows will work (#20250620143449)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.11.1" (2025-06-20)
=====================

Bugfixes
--------

- Dependency upgrade: golang-1.24.4 (#20250620043232)
- Dependency upgrade: alpine-3.22.0 (#20250620043235)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.11.0" (2025-06-19)
=====================

Features
--------

- :camel: change the schema to use yaml so that yaml aliases can be used (#20250618111945)
- :sparkles: Automatic management of bindings (#20250619104617)
- :sparkles: Generation of [QLTY](https://docs.qlty.sh/qlty-toml) configuration file (#20250619113601)
- :sparkles: Automatically generate and install manpages during installation (#20250619140044)
- :sparkles: Make the output of help and usage commands pretty using fang (#20250619140046)
- :sparkles: Run go mod tidy and generate before releasing (#20250619144539)
- :recycling: dogfooding! (#20250619144625)


Bugfixes
--------

- :bug: Make sure that the arguments used in bindings take precedent over everything else when naming the arguments (#20250618084049)
- :bug: If no argument is specified in a binding, make sure to use the full name of the field (#20250618084123)
- :bug: Make sure that generated types use concrete types instead of interface{} (#20250618112020)
- :camel: Upgrade dependencies (#20250619104756)
- :bug: fix project definition template (#20250619111827)
- :gear: make more files not overwritten if manually modified (#20250619112239)
- :bug: Make sure that gitignore file is generated (#20250619144523)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.10.0" (2025-06-16)
=====================

Features
--------

- :sparkles: Make sure `TTL` is considered as an acronym (#202506161424)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.9.0" (2025-06-16)
====================

Features
--------

- :sparkles: Add support for POSIX style operands (#20250616100028)


Bugfixes
--------

- Dependency upgrade: alpine-3.22.0 (#20250602050743)
- :bug: Make sure that bindings work when using custom types (#20250603105205)
- Dependency upgrade: golang-1.24.4 (#20250606041601)


Misc
----

- #20250609111531


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.8.3" (2025-05-22)
====================

Bugfixes
--------

- :bug: fix problem with names containing identifier (#20250522155216)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.8.2" (2025-05-20)
====================

Bugfixes
--------

- :bug: update acronyms list to include `HTTP` (#20250520153505)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.8.1" (2025-05-20)
====================

Bugfixes
--------

- :bug: ensure correct casing and default value (#20250518001822)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.8.0" (2025-05-15)
====================

Features
--------

- :sparkles: Add a way to pass tool version information to each CLI commands (#20250430153155)


Bugfixes
--------

- :book: Add doc strings to better explain the purpose of the different files and the next steps (#20250430153313)
- :bug: Fix links and doc strings issues (#20250430153347)
- Dependency upgrade: golang-1.24.3 (#20250507044720)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.7.0" (2025-04-25)
====================

Features
--------

- :sparkles: Make go version configurable (#20250422163652)
- :sparkles: `[boom]` change interface name to be more explicit (#20250422181756)
- :sparkles: Improve the argument validations (#20250423172034)
- :sparkles: extend the possible types for lang_type (#20250423172050)


Bugfixes
--------

- :gear: Specify required go tools (#20250422160754)
- :recycle: Split CLI code from SDK code (#20250422160826)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.6.2" (2025-04-22)
====================

Bugfixes
--------

- :camel: Upgrade dependencies (#20250422112455)
- :gear: Use `go 1.24 tool` command for generators (#20250422112541)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.6.1" (2025-04-11)
====================

Bugfixes
--------

- :bug: Make sure that when a default value is a function it doesn't wrap it in quotes (#20250411142251)
- :bug: Make sure that the type of an argument in the CLI is respected (#20250411142323)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.6.0" (2025-04-10)
====================

Features
--------

- :sparkles: Add support for using yml directly (#20250409162959)
- :sparkles: Add information on the installation of the tool to DEVELOPMENT.md template (#20250409163025)


Bugfixes
--------

- Dependency upgrade: golang-1.24.2 (#20250402045325)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.5.0" (2025-03-19)
====================

Features
--------

- :sparkles: Add support for conditional validation of fields (#20250319103210)


Bugfixes
--------

- :bug: Make sure that extra imports are included in config.go and types.go (#20250319115417)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.4.1" (2025-03-13)
====================

Bugfixes
--------

- :bug: Fix variadic argument forwarding in function definition (#20250312175334)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.4.0" (2025-03-11)
====================

Features
--------

- :sparkles: Update `mockgen` import to use `go 24` [tool command](https://go.dev/doc/go1.24#tools) (#20250311104426)


<!--
Copyright (C) 2020-2025 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.3.0" (2025-03-10)
====================

Features
--------

- :sparkles: Add support for not regenerating gomod files (#20240918125706)
- :sparkles: Add support for config files (#20240918125718)
- :sparkles: Add `version` command for all CLIs (#20250310161007)


Bugfixes
--------

- :bug: Make sure that the go.mod does not have superfluous replace directives (#20240617115401)
- Dependency upgrade: golang-1.22.6 (#20240813150941)
- :arrow_up: Upgrade dependencies (#20250310161018)
- Dependency upgrade: alpine-3.21.3 (#20250310162303)
- Dependency upgrade: golang-1.24.1 (#20250310162310)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.2.0" (2024-08-13)
====================

Features
--------

- :sparkles: Add support for specifying time.Duration as CLI arguments (#20240813154502)
- :sparkles: Remove inactive linters from golangci-lint template (#20240813154532)


Bugfixes
--------

- Dependency upgrade: alpine-3.20.1 (#20240621043617)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.1.2" (2024-06-06)
====================

Bugfixes
--------

- :book: correct the spelling in the generated documentation (#202406060937)
- Dependency upgrade: golang-1.22.3 (#20240528145224)
- Dependency upgrade: golang-1.22.4 (#20240606043443)
- :bug: Update generator to use correct [goreleaser](https://goreleaser.com/) configuration (#20240606084014)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.1.1" (2024-05-28)
====================

Bugfixes
--------

- Dependency upgrade: alpine-3.20.0 (#20240528145151)


<!--
Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
"0.1.0" (2024-05-28)
====================

Features
--------

- :sparkles: Create initial SDK generator (#20240422092734)
