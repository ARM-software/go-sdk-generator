<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Configure functional options

Use `specification.options` in a [generator manifest](write-a-manifest.md) to define optional settings for an SDK operation and, where needed, expose them in the generated CLI. The [repository's pet shop manifest](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/petshop-purchase.yml) exercises the option and CLI configuration together.

## Define an option and its value field

In the default `enable_mode: independent`, each generated option value has an
`Enabled bool` field with `mapstructure:"enabled"`. With `enable_mode: together`,
one `Enabled` field controls the whole options object. A `default_value` prepares
its value but does **not** enable the option. Calling a generated `With…(...)`
helper enables the corresponding option or set, even if the supplied value is
empty. The `value_field` setting changes an independent option's value field; it
does not rename the enabling flag.

For example, within `specification`:

```yaml
options:
  - name: Search
    package: search
    description: Functional options for product catalogue searches.
    functions:
      - parameters:
          - name: PageToken
            lang_type: string
            description: Continuation token for the next page.
            value_field:
              name: Token
              mapstructure_tag: token
```

This defines a wrapper with a separate enabling state:

```go
type SearchOptionsPageTokenOption struct {
    Token   string `mapstructure:"token"`
    Enabled bool   `mapstructure:"enabled"`
}
```

Without `value_field`, the value is named `PageToken` with a `page_token` mapstructure tag. With `name: Token`, `mapstructure_tag` is optional and also defaults to `token`.

**Migration:** Option parameters no longer accept `bindings`. Replace an option-level `bindings: [{name: Token, mapstructure_tag: token}]` with `value_field: {name: Token, mapstructure_tag: token}`. CLI configuration `bindings` still control flags and environment mappings. For other incompatible changes, see [breaking changes and deprecation](evolving-a-contract.md).

## Expose the option in a CLI command

Use the generated wrapper as a command configuration field. For example, within `specification.command_line_commands.functions[].configuration`:

```yaml
- name: PageTokenOption
  lang_type: search.SearchOptionsPageTokenOption
  mapstructure_tag: page_token
  description: Use a continuation token for the catalogue search.
  bindings:
    - name: Token
      lang_type: string
      mapstructure_tag: token
      arg: page-token-value
      description: Continuation token value.
```

The generated CLI binds `--page-token-value` to the option's `Token` value. It also creates `--page-token` to set `Enabled`, with an environment binding for `PAGE_TOKEN_ENABLED` under the command's environment prefix. The generated default configuration retains the option's default value, if any, with `Enabled: false`.

You can rename the enabling CLI flag by adding an `Enabled` binding to the **CLI configuration field**, not the option parameter:

```yaml
- name: PageSizeOption
  lang_type: search.SearchOptionsPageSizeOption
  mapstructure_tag: page_size
  description: Apply the configured page size.
  bindings:
    - name: Enabled
      lang_type: bool
      mapstructure_tag: enabled
      arg: use-paging
      description: Enable paging.
```

This changes the default `--page-size` flag to `--use-paging` while retaining the `PageSizeOption.Enabled` destination and its environment binding. Omit the binding to keep the default flag name. See the [pet shop manifest](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/petshop-purchase.yml) for the complete option and CLI definitions.

## Pass an options object to an SDK method

A locally declared options set has two generated types:

- `SearchOptions`: the editable settings object.
- `SearchOption`: a function that applies settings to a `*SearchOptions`.

An interface argument can name the object instead of spelling out the functional
option type. For example, under `specification.interface.methods`:

```yaml
- name: Search
  description: Search the product catalogue.
  args:
    - name: Settings
      lang_type: search.SearchOptions
      description: Settings to apply to the search.
  returns:
    - name: err
      lang_type: error
      description: Whether the search failed.
```

The generator recognises the declared options object and emits functional options:

```go
Search(ctx context.Context, settings ...search.SearchOption) (err error)
```

A final argument is variadic. An earlier argument uses `[]search.SearchOption`.
Setting `iterator: true` produces `iter.Seq[search.SearchOption]` instead. Object
pointer/reference markers do not make these interface arguments pointers to
collections. Return values and unrelated custom types keep their declared shapes.

### Keep CLI configuration editable

A CLI configuration field naming `search.SearchOptions` retains that object type,
even when `list` or `iterator` is specified. Its default is
`*search.DefaultSearchOptions()`, or the constructor's pointer result for a
reference field. Explicit defaults must likewise produce one settings object.
Bindings continue to edit the object's fields; collection constraints belong on
its parameters rather than on the outer CLI object.

Convert the object when implementing the call into the SDK:

```go
return manager.Search(ctx, cfg.Settings.AsList()...)
```

For an iterator-shaped interface argument, pass `cfg.Settings.AsIterator()`
without the spread operator. The operation can collect that sequence before
calling the existing variadic builder:

```go
settings := search.WithSearchOptions(slices.Collect(optionSequence)...)
```

The generated CLI implementation is still an application stub: it does not infer
which SDK operation or arguments your command should call.

### Conversion semantics

`AsList()` returns `[]SearchOption`; `AsIterator()` returns
`iter.Seq[SearchOption]`. Generated option packages use Go 1.23 or later for these
helpers. Both capture a snapshot using the object's `Clone`
semantics. The snapshot preserves enabled flags and all configured values,
including disabled values that validation may reject.

The conversion supplies one functional option that replaces the destination's
complete state. Earlier options are replaced; options applied afterwards can
override it. Reusing the list or iterator creates another clone on application.
A nil object produces an empty list or sequence. Conversion does not enable the
object or validate it: validate the settings built by the receiving operation.

`AsList` and `AsIterator` are reserved generated method names and cannot be used
as option parameter names.
