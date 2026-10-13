<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Types, formats and default values

Use this reference when choosing `lang_type`, value shapes and `default_value` in a [generator manifest](write-a-manifest.md). The [schema](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml) defines the contract; this page explains what the built-in types produce.

## Built-in types and formats

The generator accepts native Go types and convenience aliases in `lang_type`.

| `lang_type` | Generated Go type | Notes |
| --- | --- | --- |
| `string` | `string` | Plain string values. |
| `bool` | `bool` | Boolean values. |
| `int`, `int8`, `int16`, `int32`, `int64` | Same as input | Signed integers. |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | Same as input | Unsigned integers. |
| `float32`, `float64` | Same as input | Floating-point values. |
| `[]string` | `[]string` | String list values. |
| `[]byte` | `[]byte` | Byte slice values. |
| `any` | `any` | Arbitrary values. |
| `time.Duration`, `duration` | `time.Duration` | Durations such as `5m` or `5*time.Minute`. |
| `decimal` | `float64` | Convenience alias for decimal numbers. |
| `datetime` | `string` | RFC3339-style date/time input with special default handling. |
| `uuid` | `string` | UUID-formatted string. |
| `json`, `validJSON` | `string` | JSON payload stored as a string; `validJSON` enables strict JSON validation. |
| `url`, `uri`, `host`, `pathParameter` | `string` | URL-like strings; `uri` and `pathParameter` use specialised validators. |
| `connection`, `onlineTarget`, `service`, `server` | `httpUtils.RequestConfiguration` | HTTP request target configuration. |
| `protectedConnection`, `protectedService`, `protectedServer`, `behindAuthConnection`, `behindAuthService`, `behindAuthServer`, `serviceBehindAuth`, `serverBehindAuth` | `httpUtils.RequestConfiguration` | Request target configuration with authorisation-oriented defaults. |
| `email` | `string` | Email address string. |
| `ip` | `string` | IP address string. |
| `version`, `semver` | `string` | Semantic version strings. |
| `port` | `uint32` | Network port number. |
| `filepath`, `existingFilepath` | `string` | Filesystem paths; `existingFilepath` also requires the path to exist. |
| `hex` | `[]byte` | Hex-encoded bytes decoded at runtime. |
| `base64`, `byte`, `binary`, `regex` | `string` | Validated string-based formats. |

Declare project-specific types with `specification.additional_types`. For a pointer, list or iterator, prefer the manifest's `reference`, `list` and `iterator` flags to embedding Go syntax in `lang_type`. `list` and `iterator` cannot be combined. A final method argument marked `list: true` becomes variadic (`...T`). Use `enum` for closed string-like choices.

## Built-in `default_value` shortcuts

Most values are emitted as literals. The following shortcuts produce computed defaults; use them only when their behaviour suits the contract.

| Type | `default_value` | Generated behaviour |
| --- | --- | --- |
| `string` and string-like aliases such as `uuid`, `host`, `url`, `uri`, `email`, `ip`, `version`, `semver`, `base64`, `byte`, `binary`, `regex` | Any other value | Emit a quoted string literal. |
| String-like aliases | `hostname` | Use `os.Hostname()`. |
| String-like aliases | `uuid` | Generate a UUID with `idgen.GenerateUUID4()`. |
| String-like aliases | `temp`, `tempdir`, `temporary` | Use `filesystem.TempDirectory()`. |
| String-like aliases | `bin`, `bindir` | Use `xdg.BinHome`. |
| String-like aliases | `cache`, `cachedir` | Use `xdg.CacheHome`. |
| String-like aliases | `config`, `configdir`, `conf`, `configuration` | Use `xdg.ConfigHome`. |
| String-like aliases | `here` | Use the working directory from `filesystem.CurrentDirectory()`. |
| String-like aliases | `home` | Use `os.UserHomeDir()`. |
| String-like aliases | `<username>'s home` | Use `platform.GetHomeDirectory("<username>")`, for example `maddie's home`. |
| String-like aliases | `random` | Use `rand.Text()`. |
| String-like aliases | `now` | Use `time.Now().UTC().Format(time.RFC3339)`. |
| String-like aliases | `func()...` | Emit the custom function default as-is. |
| `datetime` | `now` | Use `time.Now().UTC()`. |
| `datetime` | RFC3339 timestamp | Parse with `time.Parse(time.RFC3339, ...)`. |
| `duration` | `5m`, `1h30m`, etc. | Parse with `time.ParseDuration(...)`. |
| `duration` | Expression containing `time.` | Emit as-is, for example `5 * time.Minute`. |
| `connection`, `onlineTarget`, `service`, `server` | URL such as `https://some-mcp.org/mcp` | Build a default `httpUtils.RequestConfiguration`, set `Host` and use a generated user agent. |
| Protected and `behindAuth` connection/service/server aliases | URL such as `https://some-mcp.org/mcp` | Build an authorised `httpUtils.RequestConfiguration`, set the bearer scheme and `Host`, and use a generated user agent. |
| `hex` | Hex text such as `aaffbb77` | Decode with `hex.DecodeString(...)`. |
| `[]byte` | Any string | Emit as `[]byte("...")`. |
| `json` | JSON object text such as `{"key":1}` | Emit a quoted JSON string. |
| Numeric and boolean types | Literal such as `42`, `3.14`, `false` | Emit the corresponding Go literal. |

Next: [configure functional options](configure-options.md), [write a manifest](write-a-manifest.md) or [generate a project](generate-a-project.md).
