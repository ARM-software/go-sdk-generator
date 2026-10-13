<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Configure functional options

Use `specification.options` in a [generator manifest](write-a-manifest.md) to define optional settings for an SDK operation and, where needed, expose them in the generated CLI. The [repository's pet shop manifest](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/testdata/petshop-purchase.yml) exercises the option and CLI configuration together.

## Define an option and its value field

Each generated option value has an `Enabled bool` field with `mapstructure:"enabled"`. A `default_value` prepares its value but does **not** enable the option. Calling a generated `With…(...)` helper enables it, even if the supplied value is empty. The `value_field` setting changes the name and mapstructure tag of the value held by the option; it does not rename the enabling flag.

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

Next: [choose types and defaults](types-and-defaults.md) or [generate a project](generate-a-project.md).
