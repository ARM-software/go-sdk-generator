<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Define validation constraints on arguments and fields

Use `constraints` in a [generator manifest](write-a-manifest.md) to describe valid
values and relationships between arguments or fields. Put rules alongside the
value's `name`, `lang_type` and description so the contract carries its validation.

The generator provides a wide range of validation rules so you can make the
manifest as explicit as possible about valid inputs and their relationships.
Prefer expressing every supported rule in the manifest over scattering ad-hoc
checks through application code. This keeps validation discoverable alongside the
contract and consistent across generated validation and documentation. Reserve
handwritten checks for domain-specific requirements the manifest cannot express.

The constraint vocabulary is **inspired by JSON Schema and OpenAPI validation**,
adapted to the generator's parameter model, Go types and CLI behaviour. The
[manifest schema](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml)
defines the supported syntax; the mapping and references below explain its origins.

**Generated CLI configuration is validated on load, so it fails fast.** At the
start of command execution, the loader resolves configuration sources and validates
the effective values before invoking the command implementation or underlying
manager logic. Invalid values supplied through flags, environment variables or
configuration files are caught at that boundary, rather than during the operation.

## Choose where the rule belongs

The shared parameter model supports constraints on:

- SDK method arguments under `specification.interface.methods[].args`;
- shared fields under `specification.global_configuration`;
- CLI fields under `specification.command_line_commands.functions[].configuration`;
- domain fields under `specification.additional_types[].fields`;
- option parameters under `specification.options[].functions[].parameters`.

Define the rule on the field that owns the value. For structured CLI inputs, put
the value's constraints on the underlying generated type or option parameter and
use `bindings` to expose it. See [functional options](configure-options.md).

There are two validation stages:

1. **Manifest validation** checks the definition: compatible types, valid regular
   expressions, consistent bounds and references to existing parameters.
2. **Runtime validation** checks actual values. Generated configuration, domain
   types and options expose `Validate()`. The CLI invokes configuration validation
   through `configUtils.LoadUsing(...)` before running the operation. SDK application
   code should validate its inputs and built options at the operation boundary;
   declaring an interface argument does not itself intercept every Go call.

## Start with scalar values

For example, these entries can live under a command's `configuration`:

```yaml
- name: ProjectName
  lang_type: string
  mapstructure_tag: project_name
  description: Lowercase project identifier.
  constraints:
    required: true
    min_length: 3
    max_length: 40
    pattern: '^[a-z][a-z0-9-]*$'

- name: BatchSize
  lang_type: int
  mapstructure_tag: batch_size
  description: Number of records processed in a batch.
  default_value: "20"
  constraints:
    minimum: 10
    maximum: 100
    multiple_of: 10

- name: Mode
  lang_type: string
  mapstructure_tag: mode
  description: Validation mode.
  default_value: strict
  constraints:
    enum: [strict, relaxed]
```

Here, `ProjectName` must be present and match the length and pattern rules;
`BatchSize` must be a multiple of ten between 10 and 100 inclusive; `Mode` is a
closed string choice. Scalar enums require a `default_value` from the enum.
Manifest defaults are strings, including numeric defaults such as `"20"`.

Top-level `required`, `enum` and `pattern` are also supported. Prefer one location
for each rule. If both top-level and nested enum or pattern definitions are supplied,
they must agree. `constraints.enum` can also drive a generated scalar enum type.

| Requirement | Manifest constraint | JSON Schema counterpart |
| --- | --- | --- |
| Required value | `required: true` | Property name in the containing object's `required` array |
| Closed string choices | `enum: [strict, relaxed]` | `enum` |
| String pattern | `pattern: '^[a-z]+$'` | `pattern` |
| String length | `min_length`, `max_length` | `minLength`, `maxLength` |
| Inclusive numeric bounds | `minimum`, `maximum` | `minimum`, `maximum` |
| Exclusive numeric bounds | `exclusive_minimum`, `exclusive_maximum` | `exclusiveMinimum`, `exclusiveMaximum` |
| Numeric step | `multiple_of` (positive) | `multipleOf` |
| Collection size | `min_items`, `max_items` | `minItems`, `maxItems` |
| Unique collection items | `unique_items: true` | `uniqueItems` |
| Presence dependency | `dependent_required: [OtherField]` | `dependentRequired` |

String lengths count Unicode code points. Patterns use Go regular expressions;
anchor with `^` and `$` when the entire value must match. JSON Schema patterns use
a different regular-expression dialect, so check patterns when translating them.

Additional generator conveniences include:

- `contains`, `not_contains`, `prefix`, `suffix`: string-content rules;
- `not_empty: true`: require a non-empty value using the generator's validation
  semantics;
- `nullable: true`: allow omitted/null values through the wrapped constraint rules
  while checking supplied values; this does not cancel a separate `required` rule;
- `extension: [.json, .yaml]`: allowed filename extensions for a scalar
  string-compatible path;
- `at_least_one: true`: shorthand for `min_items: 1`;
- `exactly_one: true`: shorthand for `min_items: 1` and `max_items: 1`.

Choose a built-in `lang_type` such as `uuid`, `email`, `port` or `existingFilepath`
when it already expresses a format requirement. See [types and defaults](types-and-defaults.md).

### What counts as a supplied value?

`required` and presence-based relationships check the **value after configuration
loading**, once defaults and configured sources have been resolved. They do not
simply check whether a key exists in a configuration file or whether the user
typed a flag.

For example, for a required string field named `ProjectName`:

- A configuration entry `project_name: ""` still fails the required check: the
  key exists, but its value is an empty string.
- Passing `--project-name=""` also fails: explicitly passing the flag does not
  make an empty string valid.
- Omitting the flag can still pass if an environment variable, configuration file
  or `default_value` supplies a non-empty project name. `required: true` means
  the resulting value must satisfy the required check, not that the user must
  type the flag on every invocation.

The same distinction matters for dependencies: if `ProjectName` declares
`dependent_required: [Region]`, a non-empty default project name activates that
dependency even when the user does not pass `--project-name`. `Region` must then
have a value that satisfies its required check too.

Other types have their own presence/emptiness rules. When `0`, `false`, `nil` or an
empty collection is a valid input for your contract, verify that case against the
generated validator rather than assuming that explicitly supplying it makes it
"present". To require flags to be passed together, use the separate
[`require_together_as_cli_flags` setting](#requiring-flags-together).

## Separate collection rules from item rules

```yaml
- name: Labels
  lang_type: string
  list: true
  mapstructure_tag: labels
  description: Labels used to select records.
  constraints:
    required: true
    min_items: 1
    max_items: 8
    unique_items: true
    min_length: 2
    pattern: '^[a-z][a-z0-9-]*$'
```

`min_items`, `max_items` and `unique_items` apply to the collection. String enum,
pattern, length and substring rules, numeric bounds and `not_empty` apply to each
item when `list: true` or `iterator: true` is used. Thus `min_length` measures each
label, not the number of labels. `contains` tests string contents, unlike JSON
Schema's array `contains` keyword. `extension` is scalar-only.

Use a collection-size rule to require items; an item rule alone does not require
a populated collection. Iterator validation can consume/materialise the sequence.
See [collections, lists and iterators](collections-and-streams.md) for replayability
and CLI parsing details.

## Express relationships between fields

Reference another field by its manifest `name`, with the same spelling and case,
in the **same parameter collection**. Use `ReportPath`, for example, rather than
its `mapstructure_tag` or CLI flag name.

| Constraint | Meaning |
| --- | --- |
| `dependent_required: [OtherField]` | When this value is present, the named values must also be present. This is directional. |
| `conflicts_with: [OtherField]` | This value and the named values cannot be present together. |
| `required_when` | Require this value when a condition on another parameter matches. |
| `forbidden_when` | Forbid this value when a condition on another parameter matches. |
| `required_values_when` | When a condition matches, restrict this value to `required_values`. |
| `forbidden_values_when` | When a condition matches, reject this value if it is in `forbidden_values`. |
| `parameter_value` | Compare this value with another parameter's value, for example using `equals` or `not_equals`. |

For `required_when`, `forbidden_when` and the conditions in the `*_values_when`
lists, provide `parameter` and an `operator`:

- `present` (the default) or `absent`: test presence;
- `equals` or `not_equals`: compare with one `value`;
- `in` or `not_in`: compare with a `values` list.

Use typed YAML values for comparisons, such as `true` for a boolean and `10` for
a number. `required_when` and `forbidden_when` each take one condition object;
`required_values_when` and `forbidden_values_when` take lists.

For example, require a destination only when file reporting is selected:

```yaml
- name: ReportMode
  lang_type: string
  mapstructure_tag: report_mode
  description: Where to send the report.
  default_value: console
  enum: [console, file]

- name: ReportPath
  lang_type: string
  mapstructure_tag: report_path
  description: Destination used for file reports.
  constraints:
    required_when:
      parameter: ReportMode
      operator: equals
      value: file
    forbidden_when:
      parameter: ReportMode
      operator: not_equals
      value: file
```

Loading `ReportMode: file` without a non-empty `ReportPath` fails. Supplying a
`ReportPath` with `ReportMode: console` also fails. These checks apply to the
resolved configuration regardless of the source of either value.

To restrict a value conditionally, add this to a string `Compression` field's
`constraints` when a sibling `ReportMode` exists:

```yaml
required_values_when:
  - parameter: ReportMode
    operator: equals
    value: console
    required_values: [none]
forbidden_values_when:
  - parameter: ReportMode
    operator: equals
    value: file
    forbidden_values: [legacy]
```

To keep a destination different from a sibling `SourcePath`, use:

```yaml
parameter_value:
  parameter: SourcePath
  operator: not_equals
  when: not_empty
```

Here, `when` controls evaluation against the **current** value: `always` (default),
`empty` or `not_empty`. Value comparisons and conditional value restrictions are
different from requiring or forbidding the field's presence.

Prefer structured constraints for new contracts. The older `condition` field is
a legacy conditional-required expression referencing one other parameter.

### Requiring flags together

There are two separate requirements you can express:

1. **The values must be available together.** With
   `ProjectName.constraints.dependent_required: [Region]`, a non-empty project name
   requires a region. The region can come from any supported configuration source.
   For example, `--project-name=demo` passes this dependency check if a configuration
   file already supplies `region: eu-west`, even without a `--region` flag.
2. **The flags must be typed together.** Add
   `require_together_as_cli_flags: true` to both fields when supplying either flag
   should require the user to supply the other flag in the same invocation.
   A value from a configuration file, environment variable or default does not
   count as typing that flag.

For example, under a command's `configuration`:

```yaml
- name: ProjectName
  lang_type: string
  mapstructure_tag: project_name
  arg: project-name
  description: Project to deploy.
  constraints:
    dependent_required: [Region]
    require_together_as_cli_flags: true

- name: Region
  lang_type: string
  mapstructure_tag: region
  arg: region
  description: Deployment region.
  constraints:
    require_together_as_cli_flags: true
```

With this opt-in on **both fields**:

- `--project-name=demo --region=eu-west` satisfies the flag-together requirement.
- `--project-name=demo` fails that requirement, even if the region is configured
  elsewhere.
- `--region=eu-west` also fails that requirement: the flag grouping applies in
  both directions, although `dependent_required` is directional.
- Passing neither flag does not trigger the flag-together requirement. The loaded
  values must still pass runtime validation: a project name supplied by a default
  or configuration file still requires a region.

The flag grouping checks which flags were supplied; it does not validate their
contents. Supplying both flags with empty values does not bypass any required or
other value constraints. Set `require_together_as_cli_flags: true` on every field
in the dependency group to enable this additional CLI check.

### Functional options

An option stores both a value and an enabling state. These answer different
questions: the value says **what setting to use**, whilst the enabling state says
**whether to apply that setting**. Having a value stored in an option does not,
by itself, mean the option is enabled.

The `enable_mode` setting determines which options are enabled together:

- **`independent` (the default):** each option has its own wrapper with an
  `Enabled` field. Enabling a page-size option does not enable a separate
  page-token option. Each enabled option's value is validated separately.
- **`together`:** the whole options object has one shared `Enabled` field.
  Enabling it activates the set, and validation checks the fields in that set.

For example, suppose a page-size option has `minimum: 1` and `maximum: 100`.
If it stores `200` but is disabled, its maximum-value check is skipped. Once the
option is enabled, validating it rejects `200`. Disabling an option skips its
field-level checks; relationships between options can still impose requirements
on which options may be enabled together.

Setting `default_value: "20"` initialises the stored page size to 20, but normally
leaves the option disabled. Calling its generated `With…(...)` helper enables the
option (or the whole set in together mode). The helper enables it even if the
supplied value is empty; validation then determines whether that value is allowed.

There is an explicit exception for together mode:
`infer_enabled_from_non_empty_values: true`. With this setting, `Validate()`,
`AsList()` and `AsIterator()` automatically set the shared `Enabled` field to
`true` if any generated field has a non-empty value. A non-empty default counts
too. Enable this behaviour only when a stored non-empty value should activate the
whole set, including when that value came from a default.

In SDK code, validate the final options object after applying the caller's
options and before using its settings. `AsList()` and `AsIterator()` only convert
an editable options object into functional options that can be passed to an SDK
method; they do not check whether its values satisfy the constraints. For example,
converting an enabled page-size option containing `200` does not reject it during
conversion. The receiving operation must validate the options object it builds
from those functional options before processing any records.

See [configure functional options](configure-options.md) for manifest examples,
enabling flags and SDK conversion details.

## Verify the contract

1. Check the manifest against the current schema and generator's semantic validation.
2. Generate into a scratch project and inspect the emitted `Validate()` methods
   and CLI flags. See [generate a project](generate-a-project.md).
3. Check a valid value, each relevant boundary and an invalid combination. Include
   empty/omitted values and defaults where presence matters.
4. Exercise invalid CLI configuration through both flags and another source, such
   as environment variables, and confirm it fails during loading before the
   operation runs.

The [Manifest Validation skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/manifest-validation/SKILL.md)
provides an agent-oriented authoring checklist.

## Inspiration and references

The generator uses its own snake_case constraint syntax and runtime semantics.
Use these standards as design references rather than assuming arbitrary JSON
Schema or OpenAPI keywords can be copied into `constraints`:

- [JSON Schema validation overview](https://json-schema.org/understanding-json-schema/reference): value types and validation keywords.
- [String validation](https://json-schema.org/understanding-json-schema/reference/string), [numeric validation](https://json-schema.org/understanding-json-schema/reference/numeric), and [array validation](https://json-schema.org/understanding-json-schema/reference/array): patterns, lengths, bounds and uniqueness.
- [Conditional validation](https://json-schema.org/understanding-json-schema/reference/conditionals): `dependentRequired` and `if`/`then`/`else`, which inspire the explicit parameter-relationship rules.
- [OpenAPI 3.1 Schema Object](https://spec.openapis.org/oas/v3.1.1.html#schema-object): JSON Schema-based validation of API data.
- [OpenAPI 3.1 Parameter Object](https://spec.openapis.org/oas/v3.1.1.html#parameter-object): describing API parameters, including requiredness and their value schemas.
