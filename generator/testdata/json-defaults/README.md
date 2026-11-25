<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# JSON default regression fixtures

## Rationale

A valid JSON default can become invalid Go code or lose content during generation.
These tests exercise the real default-generation helpers for both `json` and
`validJSON`, checking Go string literals, JSON validity and round-trip preservation.
`strconv.Unquote` validates the generated literal and recovers its value.

Exact text and decoded values are checked separately: `"\u263A"` and `"☺"` mean
the same thing in JSON, but silently rewriting the escape would fail exact
preservation. Keeping cases as raw JSON avoids introducing a second layer of
Go-string escaping into the expected values.

## Fixtures

- `cases.json`: 24 values covering objects, primitives, quotes, backticks,
  backslashes, control characters and Unicode.
- `manifest.yml`: checks that JSON text survives manifest decoding in SDK
  configuration, additional types, options and CLI configuration. Defaults must
  be strings, for example `default_value: '{"name":"test"}'`.
- `minimal-object.yml` and `minimal-empty-string.yml`: small examples covering
  object quoting and preservation of the empty JSON string.

Run from the repository root:

```sh
go test ./generator -run TestJSONDefault -v
```
