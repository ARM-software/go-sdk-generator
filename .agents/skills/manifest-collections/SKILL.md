---
name: manifest-collections
description: Use when authoring or reviewing manifest collections, list or iterator shapes, item enums, collection constraints, or repeated CLI inputs.
license: Apache-2.0
metadata:
  scope: "tool"
---

<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Skill: Manifest Collections

Use this skill when a manifest value represents several items, such as asset tags,
supported platforms or streamed search results.

Optional background: [Collections, lists and iterators](https://github.com/ARM-software/go-sdk-generator/blob/main/guides/collections-and-streams.md),
including its complete SDK example. Use the
[schema](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml) as the authoritative field reference
and [Manifest Authoring](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/manifest-authoring/SKILL.md) for broader design guidance. For offline, version-matched definitions run `go-sdk-generator schema`.

## Author the contract in this order

1. Identify one item's meaning and choose its base `lang_type`.
2. Choose `list: true` for stored values or `iterator: true` for an SDK sequence.
3. Separate collection rules from item rules.
4. Check how the value enters the application: a Go caller, configuration loading,
   or CLI flags.
5. Validate the manifest and inspect the generated types, flags and validation
   before accepting the design.

## Choose the shape deliberately

- Prefer a base item type with a shape flag over raw Go slice or iterator spellings.
- `list: true` normally produces `[]T`; choose it for values that need indexing,
  repeated inspection or configuration loading.
- `iterator: true` produces `iter.Seq[T]`; the application supplies the producer.
  It does not automatically create a channel, network stream or lazy file reader.
- Never set both `list` and `iterator`.
- A final non-reference list argument in an SDK method becomes `...T`. Existing
  slices are passed using `values...`. List returns, earlier arguments and option
  parameters remain slices.
- `reference: true` wraps the whole collection: `*[]T` or `*iter.Seq[T]`, not a
  collection of pointers. Use it only when that distinction is intentional.
- Avoid combining `list: true` with `lang_type: "[]string"`; that adds another
  collection layer.
- For declared options objects, retain the editable object in CLI configuration
  and convert it with `AsList()` or `AsIterator()` when calling the SDK. Interface
  arguments use functional-option collections. See
  [options-object conversions](https://github.com/ARM-software/go-sdk-generator/blob/main/guides/configure-options.md#pass-an-options-object-to-an-sdk-method).

## Keep collection and item validation separate

Collection constraints under `constraints`:

| Requirement | Use |
| --- | --- |
| Minimum or maximum number of items | `min_items`, `max_items` |
| No duplicates | `unique_items: true` |
| At least one item | `at_least_one: true` |
| Exactly one item | `exactly_one: true` |

These require a list or iterator shape. Keep bounds consistent. Use
`required: true` when a configuration value must be provided; checking individual
items alone does not require a populated collection.

Item rules include `enum`, `pattern`, string constraints such as `min_length`, and
numeric constraints such as `minimum`. For a list of strings, `min_items` counts
strings while `min_length` measures each string. `contains` tests each string's
contents rather than collection membership. `constraints.extension` is scalar-only.

Check the emitted rules for format aliases such as `uuid` or `filepath`; do not
assume every scalar validator is automatically adapted to collections. Distinguish
manifest-definition validation from runtime validation of actual inputs and outputs.

### Collection enums are supported

Use `lang_type: string` with an `enum` for allowed item values. Both `list: true`
and `iterator: true` support item enum validation:

```yaml
- name: Platforms
  lang_type: string
  list: true
  required: true
  mapstructure_tag: platforms
  description: Platforms to inspect.
  enum: [linux, windows, macos]
  constraints:
    at_least_one: true
    unique_items: true
```

The collection retains its string item type; do not expect the scalar enum's
generated named type for each element. Prefer the manifest enum to handwritten
membership checks.

## Do not confuse examples and defaults

- `example` represents one item; each entry in `examples` represents one item.
- Choose `example` or `examples`, not both.
- Examples are strings in the manifest, including numeric examples such as `"8080"`.
- Examples describe values; they do not initialise a collection.
- `default_value` is a single string, not a collection-default syntax. Omit it for
  ordinary collection examples and initialise non-empty defaults in application
  code. Verify any custom default against generated Go code.

## Check iterator consumption

Collection validation may traverse and materialise an iterator, potentially once
per rule. Use finite, replayable sequences when validation is followed by
processing, or materialise values into a list. Do not describe an iterator as
guaranteeing constant-memory validation or safe reuse of a one-shot stream.

## Check CLI parsing separately

For string-valued CLI lists:

- Prefer `lang_type: string` plus `list: true`.
  This creates a `[]string` field and a StringSlice flag, supporting repeated flags
  and comma-separated values.
- String-compatible aliases and list-valued bindings use the same flag shape,
  including root-level flags. Verify the item validation separately for format aliases.
- `lang_type: "[]string"` without `list` remains supported, but does not enable
  shape-aware collection constraints. Prefer the shape-flag form for new manifests.
- Check flag and decoding support separately for numeric or custom collection types.
- A CLI flag does not construct an iterator producer. Prefer stored values at
  CLI/configuration boundaries.

See the guide's
[CLI input section](https://github.com/ARM-software/go-sdk-generator/blob/main/guides/collections-and-streams.md#cli-input)
for an example.

## Review checklist

- Is the item type separate from the collection shape?
- Are `list`, `iterator` and `reference` being used intentionally?
- Does the final-argument variadic rule produce the intended SDK API?
- Are cardinality, uniqueness and per-item rules distinguished?
- Are string item enums expressed in the manifest?
- Are examples item values rather than encoded whole collections?
- Is iterator validation compatible with how the producer is consumed?
- Does the actual generated CLI flag match the documented command-line usage?
- Have the relevant manifest snippets been validated and generated output inspected?
