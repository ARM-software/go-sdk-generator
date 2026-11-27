<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->
---
name: manifest-validation
description: Use when defining or reviewing validation constraints on manifest arguments, configuration fields, domain fields or option parameters, including fail-fast CLI validation.
license: Apache-2.0
compatibility: opencode
---

<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Skill: Manifest Validation

Use this skill to express allowed values and cross-field relationships in a
generator manifest. Start with
[Define validation constraints](../../../guides/validation-constraints.md) for
examples and the keyword reference, and check
[the schema](../../../generator/schema/schema.yml) for supported definitions.
Use [Manifest Authoring](../manifest-authoring/SKILL.md) for the wider contract.

The generator offers a wide range of validation rules. Use them to make the
manifest as explicit as possible about valid inputs and relationships, rather
than adding ad-hoc checks throughout application code. Before writing a custom
check, look for a supported built-in type or constraint that expresses it. Keep
supported rules in the manifest so generated validation and documentation share
the same contract; reserve handwritten checks for requirements it cannot express.

## Author constraints in this order

1. Identify the owning argument or field and its parameter collection. Choose the
   most specific built-in `lang_type` and intentional `reference`, `list` or
   `iterator` shape.
2. Decide whether the value is required, optional or nullable. Consider defaults
   and the type's zero/empty/nil semantics. Presence is based on runtime values,
   not just whether a CLI flag was supplied.
3. Add local rules under `constraints`: enum/pattern, string length/content,
   numeric bounds/steps or collection size/uniqueness.
4. Add relationships using exact sibling manifest names, not flag names or
   mapstructure tags. Prefer structured rules over legacy `condition` expressions.
5. Check the definition with the generator, inspect emitted validation, and verify
   representative valid, boundary and invalid inputs at the relevant entry point.

## Preserve the intended semantics

- Top-level `enum`, `pattern` and `required` remain supported. Prefer one location
  per rule; duplicated enum/pattern definitions must agree.
- Scalar string enums need a default from the allowed values. Collection enums
  validate items and do not require a scalar default.
- Keep bounds consistent and `multiple_of` positive. String lengths count Unicode
  code points, and patterns must compile as Go regular expressions.
- Separate per-item rules from `min_items`, `max_items` and `unique_items`.
  `not_empty` on a collection checks items; use a cardinality rule to require items.
  See [Manifest Collections](../manifest-collections/SKILL.md) for iterator
  consumption and CLI parsing.
- `nullable` wraps constraint rules; it does not override a separate requirement
  that the field be present.
- `dependent_required` is directional. `require_together_as_cli_flags` is an
  additional opt-in on all participating CLI fields; runtime dependencies still
  operate on effective values from all sources.
- Distinguish `required_when`/`forbidden_when` (presence) from
  `required_values_when`/`forbidden_values_when` (allowed values).
- Use `parameter_value` for comparisons between field values. Its `when` selector
  refers to the current field's emptiness.
- Put nested option rules on option parameters. Field validation is gated by the
  option's enabling state; defaults alone normally do not enable it. Option-object
  conversions do not validate the resulting settings.

## Explain when validation runs

Keep manifest-definition validation distinct from runtime input validation.
**The generated CLI validates configuration on load and fails fast**, before the
command implementation or manager operation runs. The loader validates the
effective configuration assembled from supported sources, including defaults,
files, environment variables, flags and keyring values when enabled.

Generated configuration, domain types and options expose `Validate()`. For direct
SDK use, check that application code validates inputs and built options at the
operation boundary. An interface declaration alone does not validate every call.
When verifying CLI behaviour, include an invalid non-flag configuration source
and confirm the operation is not invoked.

## Standards inspiration

Describe these constraints as inspired by **JSON Schema and OpenAPI validation**,
with generator-specific spelling and semantics. Reference
[JSON Schema validation](https://json-schema.org/understanding-json-schema/reference),
[conditional schemas](https://json-schema.org/understanding-json-schema/reference/conditionals)
and the [OpenAPI 3.1 Schema Object](https://spec.openapis.org/oas/v3.1.1.html#schema-object).
Use the guide's mapping when translating concepts: `minLength` becomes
`min_length`, property requiredness becomes a per-field boolean, and string
`contains` is not JSON Schema's array `contains`. Only use supported manifest
keywords, rather than assuming full JSON Schema/OpenAPI compatibility.
