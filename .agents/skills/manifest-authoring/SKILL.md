---
name: manifest-authoring
description: Define generator manifests clearly, choose built-in types where possible, and model SDK or CLI interfaces with stable, idiomatic names.
license: Proprietary
compatibility: opencode
---

# Skill: Manifest Authoring

Use this skill when creating or updating a generator manifest such as `generate.yml`, `verify.yml`, `agent.yml`, or another YAML/JSON file that follows `generator/schema/schema.yml`.

## Goal

Write manifests that are:

- precise
- easy to regenerate safely
- aligned with the generator's built-in capabilities
- consistent across SDKs, CLIs, actions, and generated documentation

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
- Is the manifest modelling one concern coherently?
