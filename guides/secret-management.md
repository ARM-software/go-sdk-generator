<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Secret management in generated CLI tools

Use this guide when a [manifest](write-a-manifest.md) defines CLI configuration containing tokens, passwords or other secret strings. The generator can provide a separate `credentials` command for storing those values in the system keyring where supported.

## Mark secret inputs in the manifest

Set `secret: true` on a string-compatible CLI configuration field. For example, within a command's `configuration` list:

```yaml
- name: Token
  lang_type: string
  mapstructure_tag: token
  description: Token used by the command.
  secret: true
```

For a structured configuration value, you can instead mark an individual `bindings` entry as secret. The generator validates the type: a secret field or binding must resolve to a string-compatible value. Consult the [manifest schema](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml) for both `secret` fields and the [manifest-authoring skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/manifest-authoring/SKILL.md) for authoring guidance.

## Store and load credentials

If the CLI has secret-marked inputs but no `Credentials` command, the generator creates one from those inputs. If a `Credentials` command already exists, it merges the secret-marked inputs into that command. The generated command stores its configuration in the system keyring instead of invoking SDK operation logic, and no GitHub Action wrapper is generated for it.

In the **generated tool**, inspect `<generated-binary> credentials --help` for its inputs. Run the command with the required values to store them when the platform supports a system keyring. An unsupported keyring reports an error rather than silently storing the values elsewhere. `go-sdk-generator` itself has no `credentials` command unless its own manifest defines secret inputs.

Commands configured to load from the keyring use non-empty stored values ahead of CLI flags, environment variables, configuration files and defaults. Such commands expose `--no-keyring` to disable keyring loading for that invocation. The [generated command reference](https://github.com/ARM-software/go-sdk-generator/blob/main/DOCUMENTATION.md#cli-tool) describes configuration precedence; check the generated project's own command documentation for which commands enable keyring loading. A command cannot be configured to store to and load from the keyring at the same time.

## Security boundaries

- Keep actual secret values out of manifests, `default_value`, committed `.env` files and generated source: secret marking describes how an input is handled, but does not remove a literal value supplied in the contract.
- Prefer the system keyring for local credential storage where supported. For CI, supply credentials through the CI system's secret mechanism; see [GitHub Actions integration](use-github-actions.md).
- Generated `secretutil.ProtectedString` is currently a Go `string` alias. It does not itself encrypt values in memory or prevent a caller from logging them; handle values in handwritten operation code accordingly.

For the broader project workflow, return to [what the generator produces](what-it-generates.md) or [generate a project](generate-a-project.md).
