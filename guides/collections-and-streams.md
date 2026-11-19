<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Collections, Lists and Iterators

Use collection shapes when a manifest value naturally represents many items rather than one scalar value.

## Prefer shape flags over raw slice spellings

In manifests, prefer a base `lang_type` together with the shape flags:

- `list: true` for generated slice-style values
- `iterator: true` for generated stream-style iterator values

Prefer this over spelling collection shapes directly in `lang_type` such as `[]string`.

## Lists

Use `list: true` when the generated SDK or CLI configuration should hold all values at once.

Example:

```yaml
- name: tags
  lang_type: string
  list: true
  description: Tags to include in the operation.
  constraints:
    unique_items: true
```

This generates a Go slice shape such as `[]string`.

For CLI configuration fields, string lists are exposed as Cobra string-slice flags, so users can pass either repeated flags or comma-separated values.

Example:

```sh
tool verify --tags alpha,beta
tool verify --tags alpha --tags beta
```

## Iterators and Streams

Use `iterator: true` when the generated surface should represent a stream-like sequence instead of a stored slice.

Example:

```yaml
- name: files
  lang_type: filepath
  iterator: true
  description: Files to process lazily.
```

This generates an `iter.Seq[T]` shape in Go.

Use iterators when:

- values are naturally streamed or produced lazily
- consumers should not need to materialise the whole collection first
- the generated API should emphasise iteration rather than random access

Use lists when:

- callers already have all values in memory
- CLI/config loading should capture the full set directly
- ordering and repeated inspection matter more than streaming

## Arrays versus lists

The manifest models repeated values through `list` and `iterator` rather than fixed-length Go arrays.

In practice:

- choose `list: true` for ordinary repeated values
- choose `iterator: true` for stream-style values
- avoid designing manifests around fixed-length array semantics unless a scalar structure or dedicated type models the contract more clearly

## Collection constraints

Collection-aware constraints apply to `list` and `iterator` shapes:

- `min_items`
- `max_items`
- `unique_items`
- `at_least_one`
- `exactly_one`

Examples and defaults also follow item semantics:

- `example` represents one item value
- each entry in `examples` represents one item value

## Current limitation

The schema currently supports scalar string enums directly, but not item-level enums for `list` or `iterator` values.

When a collection should accept only a closed set of item values:

- still model the field as a real collection with `list: true` or `iterator: true`
- document the allowed items in the description and examples
- add targeted runtime validation only where necessary

That keeps the manifest as the source of truth for the field shape even when item-level enum validation needs a small amount of supporting code.
