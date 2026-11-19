<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
---
name: manifest-collections
description: Use when modeling repeated manifest values such as lists, iterators, or stream-like inputs.
license: Apache-2.0
compatibility: opencode
---

# Skill: Manifest Collections

Use this skill when a manifest field should represent many values rather than one scalar value.

Start with [Collections, lists and iterators](../../../guides/collections-and-streams.md). Use the [schema](../../../generator/schema/schema.yml) as the authoritative reference.

## Core rule

Prefer:

- a base `lang_type`
- plus `list: true` or `iterator: true`

Avoid encoding the collection shape directly into `lang_type` such as `[]string` unless you are working with already-generated Go code rather than authoring the manifest itself.

## Choose the right shape

- Use `list: true` when callers or CLI configuration should provide all values up front.
- Use `iterator: true` when the generated Go surface should model a lazy or stream-like sequence.
- Treat "stream" guidance as iterator guidance in this repository: the manifest shape for that is `iterator: true`.

## Collection checklist

- Is the base item type still precise, for example `string`, `filepath`, `host`, or `uuid`?
- Should examples describe one item value rather than the whole collection?
- Should the collection enforce `unique_items`, `min_items`, `max_items`, `at_least_one`, or `exactly_one`?
- If the value is exposed through CLI configuration, should it be a list rather than an iterator so flag parsing remains natural?

## Current limitation

Scalar string enums are supported directly, but item-level enums for `list` and `iterator` values are not currently first-class schema features.

When a repeated field should accept only a closed set of item values:

- keep the manifest collection shape correct
- describe the allowed items explicitly
- add only the minimum targeted runtime validation needed outside the schema
