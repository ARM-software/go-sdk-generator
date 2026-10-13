<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
---
name: manifest-authoring
description: Use when creating or editing manifest files such as `generate.yml`, `verify.yml`, or `agent.yml`, especially to choose types, names, and configuration shapes clearly.
license: Apache-2.0
compatibility: opencode
---

# Skill: Manifest Authoring

Use this skill when creating or updating a generator manifest such as `generate.yml`, `verify.yml`, `agent.yml`, or another YAML/JSON file that follows `generator/schema/schema.yml`.

Start with [write a manifest](../../../guides/write-a-manifest.md) for the contract shape and minimal example; use [generate and regenerate a project](../../../guides/generate-a-project.md) for the CLI workflow. This skill adds authoring and review detail. Check the [schema](../../../generator/schema/schema.yml) for the definitive field definitions.

## Goal

Write manifests that are:

- precise
- easy to regenerate safely
- aligned with the generator's built-in capabilities
- consistent across SDKs, CLIs, actions, and generated documentation
- detailed enough to define a real project and generate a useful starting point
- understandable by a human and easy to read

## Primary Mindset

The goal of a manifest is to define a real project in enough detail that its SDK, CLI, and surrounding project structure can be bootstrapped quickly.

In general, the more precise and detailed the project definition is in the manifest, the less boilerplate and bespoke code should need to be written by hand afterwards.

A manifest should also be understandable by a human. Prefer names, descriptions, and structure that make the project easy to read and reason about without having to mentally simulate the generator.

This means manifests should:

- be real project definitions rather than synthetic schema-only exercises
- model one coherent concern clearly enough that a team could start implementation from the generated output
- use generator features such as `enum`, `constraints`, `options`, result output, and structured built-in types when they naturally fit the application
- avoid adding strange or artificial parameters purely to hit edge cases unless those parameters still make sense for a real project

## Quick Checklist

Before finalising a manifest, check:

- Does it look like a real project or tool a team could plausibly generate and continue developing?
- Are built-in `lang_type` values used wherever they already match the intent?
- Are closed choices expressed with `enum` rather than unconstrained strings?
- Are relationships and validation semantics expressed with `constraints` where appropriate?
- If the API or CLI would otherwise require repeated boilerplate, would `options` make the interface cleaner?
- If the project has several commands, do some naturally produce structured results while others do not?
- Does the manifest use the generator features that help keep the generated project structure consistent?
- Does the manifest express validation as thoroughly as the real project contract allows?
- Does it validate against the current schema, and does a generation run produce the intended SDK and CLI surface?

When a manifest definition needs to be more thorough, prefer using richer manifest features where they naturally fit, such as:

- several commands with different operational modes
- a mix of commands with and without result output
- functional options for repetitive or mode-specific behaviour
- cross-parameter constraints
- representative defaults, enums, and examples

## Core Guidance

### Prefer built-in types where they already match the intent

Use the generator's built-in `lang_type` aliases when they express the intent cleanly instead of falling back to raw `string` or ad hoc custom types.

Common examples:

- `host`, `url`, `uri`
- `duration`
- `uuid`
- `email`
- `json`
- `ip`
- `port`
- `hex`
- `binary`, `base64`, `regex`
- `connection`, `onlineTarget`, `service`, `server`
- protected request-target aliases such as `protectedConnection`, `protectedService`, `protectedServer`, `behindAuthConnection`, `behindAuthService`, `behindAuthServer`, `serviceBehindAuth`, `serverBehindAuth`

Use these built-ins because the generator can then:

- emit a better Go type
- apply matching validation logic
- handle default values more intelligently
- generate more consistent CLI/configuration behaviour

If the meaning is already covered by a built-in type, prefer the built-in type.

### Use `enum` for closed string choices

If a field or parameter is really a closed set of string-like choices, prefer `enum` in the manifest rather than leaving it as an unconstrained string.

This is intended for:

- `string`
- `reference: true` on a string-like value

When present, `enum` should drive:

- generated Go enum definitions
- generated validation
- clearer CLI help text listing allowed values

Use this when the value set is truly closed and stable. Do not use it for open-ended free text.

### Add `example` or `examples` when representative values help users

If a field or parameter benefits from showing what a valid value looks like, add:

- `example` for one representative value
- `examples` for several representative values

These values are intended to improve generated documentation, CLI help, action input descriptions, and agent-facing skills.

For `list` or `iterator` shapes, treat each example as an item value rather than trying to serialise the whole collection.

### Prefer `constraints` for parameter relationships and conditional requirements

When one parameter depends on another, or parameters have to be provided together or kept apart, prefer the `constraints` block instead of encoding that relationship in the older `condition` field.

Use `constraints` for cases such as:

- `dependent_required`
- `conflicts_with`
- `required_when`
- `forbidden_when`
- `require_together_as_cli_flags` when `dependent_required` should also require the related values to be passed together as CLI arguments

Prefer `condition` only for the older single-expression required logic when there is a concrete reason to stay with that shape.

When a value has additional validation semantics beyond its base `lang_type`, prefer expressing them directly in the manifest with `constraints` rather than leaving them implicit in prose or expecting downstream runtime code to catch them later.

Recommended mindset:

- use as many clear constraint rules as the contract genuinely supports
- prefer early manifest validation over deferring invalid combinations to generated runtime behaviour
- use `enum`, `pattern`, length, numeric, collection, and string-shape constraints whenever they make the allowed input space clearer
- prefer precise validation at authoring time so generated SDKs, CLIs, docs, and actions all inherit the same contract

These rules are intended to stay roughly aligned with JSON Schema and OpenAPI-style constraint vocabulary where practical, while still fitting the generator's parameter model and CLI-oriented behaviour.

Example:

```yaml
- name: Profile
  lang_type: string
  mapstructure_tag: profile
  description: Validation profile.
  constraints:
    dependent_required:
      - Output
    require_together_as_cli_flags: true

- name: Output
  lang_type: string
  mapstructure_tag: output
  description: Output file.
  constraints:
    require_together_as_cli_flags: true
```

### Prefer structured request-target built-ins over separate host strings when appropriate

If the manifest is modelling an HTTP target or online service configuration, prefer a request-target built-in such as:

- `connection`
- `service`
- `server`
- or one of the protected variants

instead of splitting the concept into several loosely related string fields unless there is a strong reason to keep them separate.

This is especially useful when the generator can produce a `httpUtils.RequestConfiguration` default and bindings for nested settings such as host or authorisation fields.

### Use verbs for function names

Function and command names should usually be verbs or verb phrases.

Good examples:

- `Verify`
- `Compile`
- `DownloadArtifacts`
- `StoreCredentials`
- `Initialise`
- `Decommission`

Avoid vague or noun-only names for operations when the manifest is describing something the tool does.

The generated SDK and CLI are easier to understand when operation names read like actions.

### Pass struct-like domain values as references when appropriate

When a manifest field, method argument, or method return value represents a struct-like domain object, prefer `reference: true`.

Good cases:

- generated or imported struct types such as `types.Project`
- request-target configuration objects such as `httpUtils.RequestConfiguration` when pointer semantics make sense
- larger domain values that should not be copied unnecessarily

Do not apply `reference: true` blindly to:

- raw scalar types such as `string`, `bool`, `int`, `uint32`, `time.Duration`
- slice types unless there is a concrete reason
- interface types such as `error` or other already-reference-like abstractions

In short:

- struct-like concrete data: usually use `reference: true`
- raw value types: usually keep as values
- interfaces: usually keep as values in the manifest type declaration

### Prefer shape flags over composing shape into `lang_type`

When the manifest wants to express pointer, list, or iterator semantics, prefer the dedicated flags:

- `reference: true`
- `list: true`
- `iterator: true`

instead of encoding those shapes directly into the `lang_type` string.

Prefer:

```yaml
lang_type: string
list: true
```

instead of:

```yaml
lang_type: "[]string"
```

Prefer:

```yaml
lang_type: types.Project
reference: true
```

instead of:

```yaml
lang_type: "*types.Project"
```

Avoid composed forms such as:

- `*[]string`
- `[]*types.Project`
- `*httpUtils.RequestConfiguration`
- `iter.Seq[string]`

unless there is a very specific repository constraint that the dedicated manifest flags cannot express.

This keeps the manifest more declarative and lets the generator own the final Go type construction consistently.

It also keeps the manifest more Go-agnostic. Using separate shape flags such as `reference`, `list`, and `iterator` preserves the intent of the contract without baking Go syntax directly into the manifest. That makes the manifest easier to reason about as a language-neutral description and keeps the door open for other generators to interpret the same contract in other ways later.

### Keep naming intent obvious

Prefer names that explain the role of the field directly.

Examples:

- `McpTarget` is better than `Mcp`
- `DocumentationURL` is better than `URL`
- `DefaultJobTimeout` is better than `Timeout`

If a field describes a target service, connection, credentials, or payload, say that explicitly in the name.

### Model one concern clearly

The manifest should define one coherent concern fully.

That may surface as:

- a library only
- a CLI only
- multiple CLI commands
- actions
- or a combination

But the manifest should still describe one concern cleanly rather than mixing unrelated responsibilities into one project.

## Practical Rules

### Handle breaking changes through deprecation, not abrupt deletion

If you are changing an existing manifest contract, treat the following as breaking changes unless proven otherwise:

- renaming a field, parameter, method, command, or type
- removing an existing field, parameter, method, command, or type
- changing the meaning of an existing concept incompatibly
- tightening validation so previously valid manifests stop validating
- changing a shape or type incompatibly, such as scalar to list, value to reference, or one domain type to another incompatible one

In this repo, those changes should normally follow a deprecation process:

1. Add the replacement or new shape first where possible.
2. Mark the old concept with `deprecation_notice`.
3. Provide a deprecation window, including an issue date and, when known, a removal date.
4. Describe the replacement clearly so generated docs, comments, and CLI help can guide users.
5. Keep the deprecated concept working during that window unless there is a strong correctness or safety reason not to.
6. Remove it only after consumers have had time to migrate.

Versioning and news-file note for this Go project:

- SemVer still applies in principle, so breaking changes are major in theory.
- But the output of the generator is a Go project, so a true `v2+` transition also implies Go module path and import-path changes.
- Because that is more than a routine version bump, do not assume every breaking manifest change should immediately become a normal `.major` automation path change.
- If the change is breaking but the project is not doing a full Go major-module migration, prefer a `.feature` news file with an explicit breakage message such as a `:boom:` prefix.
- Reserve `.major` news files for genuine major Go module transitions where the module-path consequences are being handled deliberately.

References:

- [Semantic Versioning](https://semver.org/)
- [Go modules major version suffix](https://go.dev/ref/mod#major-version-suffix)
- [Go Modules: v2 and beyond](https://go.dev/blog/v2-go-modules)

Do not perform abrupt deletion or silent incompatible repurposing when a managed deprecation path is practical.

### Global configuration vs command configuration

Put configuration in `global_configuration` when it is genuinely shared across the SDK or several commands.

Put configuration in command-specific `configuration` when it belongs to one operation only.

Avoid promoting command-specific options to global scope just because several commands happen to use similar names.

### Use bindings for structured subfields, not for everything

Bindings are most useful when:

- a structured type should be configured through CLI flags or env vars
- nested values such as `Host`, `Port`, or `Token` need user-facing bindings

Do not add bindings mechanically for every field if the resulting CLI becomes noisy or confusing.

### Mark secrets explicitly

If a field or binding represents secret material, mark it with `secret: true` when the schema supports it.

That allows the generator to handle credentials storage and secret-aware code paths more consistently.

### Prefer authoritative descriptions

Descriptions should explain what the field or operation means to the user, not just restate the name.

Good descriptions help generated docs, CLI help text, action metadata, and code comments all at once.

Multiline descriptions are supported when the text needs more room.

## Good Patterns

### Better built-in type choice

Prefer:

```yaml
- name: AIPortalMcpTarget
  lang_type: serviceBehindAuth
  reference: true
  default_value: https://mcp.api.example.com/mcp
```

instead of splitting the same concern into several loosely related raw strings unless you explicitly need that flexibility.

### Better function naming

Prefer:

```yaml
- name: Verify
  description: Verify that a project is valid.
```

instead of:

```yaml
- name: Verification
```

### Better reference use

Prefer:

```yaml
- name: project
  lang_type: types.Project
  reference: true
```

but usually keep plain scalars as values.

## Anti-patterns

Avoid these unless there is a strong reason:

- using raw `string` for a value already covered by a built-in type
- introducing a custom type when an existing built-in alias already models the intent
- naming operations as nouns when they represent actions
- using `reference: true` for basic scalar types by default
- using `reference: true` for interface types such as `error`
- creating a large unstructured bag of global configuration
- adding bindings that make the CLI noisier without improving usability

## Review Checklist

Before finishing a manifest change, check:

- Does each field use the most specific built-in `lang_type` available?
- Are HTTP or service targets represented as connection/service/server built-ins where appropriate?
- Are operation names verbs?
- Are struct-like concrete types passed as references where appropriate?
- Are raw scalar types kept as values unless there is a clear reason otherwise?
- Are interfaces left as values rather than wrapped in pointer semantics?
- Are descriptions clear enough to power generated docs and CLI help text well?
- If an existing concept is being replaced or retired, has it been deprecated first with a clear migration path and deprecation window?
- Is the manifest modelling one concern coherently?
