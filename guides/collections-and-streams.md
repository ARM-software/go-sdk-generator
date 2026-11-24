<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Collections, lists and iterators

Use a collection when a value represents several items: tags on an asset,
supported platforms, or files returned by a search. Choose the **item type**, the
**collection shape**, and the **validation rules** separately.

This guide builds on [writing a manifest](write-a-manifest.md) and
[types and defaults](types-and-defaults.md). The [schema][manifest-schema] is the
authoritative field reference.

## Choose the item type and shape

Prefer a base `lang_type` with a shape flag. For `lang_type: string`:

| Shape flags | Usual Go shape | Choose it when |
| --- | --- | --- |
| Neither flag set | `string` | There is one value. |
| `list: true` | `[]string` | Stored values, with indexing. |
| `iterator: true` | `iter.Seq[string]` | Iterated SDK values. |

The same shape rules apply to other item types, such as `int` or a declared
additional type. `list` and `iterator` both default to `false` and cannot both be
`true`.

Shape flags are available on method arguments and returns, configuration fields,
additional-type fields, and option parameters. The field examples below are
fragments to place in the appropriate manifest section; a complete SDK example
appears at the end.

Use `lang_type` for the item, rather than combining `list: true` with
`lang_type: "[]string"`: that would add another collection layer. String lists also
support repeated CLI flags; see [CLI input](#cli-input).

## Lists: values held together

For example, put this field under `specification.global_configuration`:

```yaml
- name: Tags
  lang_type: string
  list: true
  mapstructure_tag: tags
  description: Tags identifying the assets to process.
  examples: [alpha, beta]
  constraints:
    unique_items: true
```

The generated configuration field is `Tags []string`. Callers can provide a slice
and inspect it repeatedly. `unique_items` rejects duplicates; it does not silently
remove them.

### The final method argument becomes variadic

In an SDK interface method, a **final, non-reference** argument with `list: true`
becomes `...T`, rather than `[]T` in the signature. Earlier list arguments, list
return values, configuration fields and option parameters keep their slice shape.

For example, an entry under `specification.interface.methods`:

```yaml
- name: Collect
  description: Collect assets with the supplied tags.
  args:
    - name: Tags
      lang_type: string
      list: true
      description: Tags to collect.
  returns:
    - name: err
      lang_type: error
      description: Whether collection failed.
```

The generated method has this shape:

```go
Collect(ctx context.Context, tags ...string) (err error)
```

Pass an existing slice using Go's spread syntax:

```go
tags := []string{"alpha", "beta"}
err := manager.Collect(ctx, tags...)
```

This variadic rule is about SDK method signatures, not how CLI flags are parsed.

## Iterators: values supplied through iteration

Use `iterator: true` for a stream-like SDK surface. For example, a method return
value:

```yaml
- name: Paths
  lang_type: string
  iterator: true
  description: Asset paths produced by the search.
```

The generated type is `iter.Seq[string]`. The implementation supplies the sequence,
and a caller can consume it with a Go range loop:

```go
for path := range paths {
    // Process this path.
}
```

An iterator can produce values lazily, but the flag itself does not implement a
producer, network stream, channel or file reader. Use Go 1.23 or later for
range-over-function iteration; see the [Go iter package](https://pkg.go.dev/iter).

**Validation can consume the sequence.** Collection validation helpers may
enumerate and materialise iterator items, and different rules may traverse the
sequence separately. For validation followed by processing, use a finite,
replayable sequence or supply a list. Do not assume that adding `iterator: true`
guarantees constant-memory validation of an unbounded stream.

## Arrays, lengths and references

`list: true` represents a variable-length Go slice, not a fixed-size Go array such
as `[3]string`. To require three items, use `min_items: 3` and `max_items: 3`; the
generated type is still a slice or iterator. If positions have different meanings,
named fields in an additional type can express the contract more clearly.

`reference: true` applies to the **whole shape**: combined with `list: true`, it
produces `*[]T`, not `[]*T`; combined with `iterator: true`, it produces
`*iter.Seq[T]`. Leave it unset unless a pointer to the collection is intentional.
A reference list argument is not converted to a variadic argument.

## Collection rules versus item rules

Where generated validation is emitted, such as a configuration's `Validate()`
method, distinguish the number of items from the validity of each item.

| Collection rule under `constraints` | Meaning |
| --- | --- |
| `min_items: N` | Minimum item count. |
| `max_items: N` | Maximum item count. |
| `unique_items: true` | Reject duplicate items. |
| `at_least_one: true` | Shorthand for `min_items: 1`. |
| `exactly_one: true` | Shorthand for `min_items: 1` and `max_items: 1`. |

These rules require `list: true` or `iterator: true`. Keep bounds consistent:
`min_items` cannot exceed `max_items`, and `exactly_one` cannot be combined with
bounds excluding one item. `exactly_one` means one item in this collection, not
one populated field among several fields.

Use `required: true` when omitting the configuration value must be rejected. Item
rules alone do not require a non-empty collection; express its cardinality
separately.

The generator also supports item-by-item validation for:

- Top-level `enum` and `pattern`.
- String constraints such as `not_empty`, `contains`, `not_contains`, `prefix`,
  `suffix`, `min_length` and `max_length`.
- Numeric constraints such as `minimum`, `maximum`, `exclusive_minimum`,
  `exclusive_maximum` and `multiple_of`.

For example, this configuration requires one to five distinct tags, each
containing only lowercase letters and having at least two characters:

```yaml
- name: Tags
  lang_type: string
  list: true
  required: true
  mapstructure_tag: tags
  description: Lowercase asset tags.
  pattern: '^[a-z]+$'
  constraints:
    min_items: 1
    max_items: 5
    unique_items: true
    min_length: 2
```

Here `min_items` counts tags; `min_length` measures each tag. Likewise, `contains`
tests a substring in each string item, not membership in the collection.
`constraints.extension` is currently restricted to scalar path values. For format
aliases such as `uuid` or `filepath`, check the emitted validation rather than
assuming every built-in scalar validator is collection-aware.

Manifest validation checks the definition. Runtime validation checks actual values
where the generated API provides it; declaring method arguments or return types
does not by itself implement validation of your operation's inputs or outputs.

## Enums for collection items are supported

Use `lang_type: string` with `enum` to constrain each item of a list or iterator:

```yaml
- name: Platforms
  lang_type: string
  list: true
  required: true
  mapstructure_tag: platforms
  description: Platforms to include in the search.
  enum: [linux, windows, macos]
  examples: [linux, windows]
  constraints:
    at_least_one: true
    unique_items: true
```

Each supplied platform must be an allowed value. Replace `list: true` with
`iterator: true` for the iterator form. The generated collection retains its base
string shape (`[]string` or `iter.Seq[string]`); collection enums use item validation
rather than a separate generated enum type for every item. Do not add handwritten
enum validation just because the value is a collection.

## Examples are items; defaults are different

For a list or iterator, `example` describes one item, and each entry in `examples`
describes one item. Use one property or the other, not both. Example values are
strings in the manifest, including numeric examples:

```yaml
- name: Ports
  lang_type: int
  list: true
  mapstructure_tag: ports
  description: Ports to inspect.
  examples: ["8080", "8443"]
  constraints:
    minimum: 1
    maximum: 65535
```

These examples are two representative ports, not a default two-item collection.
They describe input and help generated documentation; they do not initialise the
configuration.

`default_value` remains a single manifest string. It is not a YAML sequence and is
not automatically expanded into a list or iterator. The collection examples in
this guide deliberately omit it. Initialise non-empty collection defaults in
application code, and inspect and compile any custom default expression before
relying on it. See [types and defaults](types-and-defaults.md).

## CLI input

For string-valued CLI lists, `lang_type: string` with `list: true` registers a
Cobra StringSlice flag as well as generating a `[]string` configuration field.
This also applies to string-compatible aliases and list-valued bindings, including
root-level flags.

| CLI configuration declaration | Go field | Registered flag |
| --- | --- | --- |
| `lang_type: string` and `list: true` | `[]string` | `StringSlice` |
| `lang_type: "[]string"`, without `list` | `[]string` | `StringSlice` |

Prefer the shape-flag form, which also supports the collection constraints
described above. Put this field under a command's `configuration`:

```yaml
- name: Tags
  lang_type: string
  list: true
  mapstructure_tag: tags
  description: Tags supplied on the command line.
  constraints:
    unique_items: true
```

For a command named `check`, [StringSlice flags][stringslice] accept both forms:

```sh
tool check --tags alpha,beta
tool check --tags alpha --tags beta
```

The raw `[]string` spelling remains supported for compatibility, but does not opt
into shape-aware collection constraints. Prefer `string` plus `list: true` for new
manifests. The repeated-flag behaviour above is for string-valued lists; check flag
support separately for other collection types. Iterators need an
application-supplied producer; a CLI flag does not create a lazy sequence. Prefer
stored list values at CLI/configuration boundaries.

## Complete SDK example

Save this as `collections.yml`. It models a catalogue that searches by tags and
returns paths through an iterator:

```yaml
specification:
  package_name: asset-catalog-sdk
  main_module: catalog
  name: Catalog
  description: Search an asset catalogue by tags and platform.
  licence: Apache-2.0
  interface:
    notCloseable: true
    namespace: github.com/example/
    methods:
      - name: Find
        description: Find matching asset paths.
        args:
          - name: Tags
            lang_type: string
            list: true
            description: Tags to search for.
        returns:
          - name: Paths
            lang_type: string
            iterator: true
            description: Matching asset paths.
          - name: err
            lang_type: error
            description: Whether starting the search failed.
  global_configuration:
    - name: Platforms
      lang_type: string
      list: true
      required: true
      mapstructure_tag: platforms
      description: Platforms to include in the search.
      enum: [linux, windows, macos]
      constraints:
        at_least_one: true
        unique_items: true
```

Generate it with your project's target Go version, for example:

```sh
go-sdk-generator generate -i collections.yml -o ./catalog-generated \
  --go-version 1.27.3
```

The SDK method uses `tags ...string` and returns `paths iter.Seq[string]`. Callers
supply `Platforms` before validating the SDK configuration. Implement the search
and its iterator in the generated implementation stub; see
[generate and regenerate a project](generate-a-project.md) for the remaining module
setup steps.

Existing regression manifests provide further examples:

- [List and iterator shapes][shape-examples].
- [Collection enums][enum-examples].
- [Item constraints][constraint-examples].

[manifest-schema]: https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml
[stringslice]: https://pkg.go.dev/github.com/spf13/pflag#FlagSet.StringSlice
[shape-examples]: https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/list-iterator.yml
[enum-examples]: https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/enum-collections.yml
[constraint-examples]: https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/collection-item-constraints.yml
