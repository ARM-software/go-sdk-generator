<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Write a generator manifest

## When to use this guide

Write a YAML or JSON manifest to define one coherent set of user-facing Go operations. The manifest controls the generated SDK surface, optional CLI, configuration, documentation, and supporting files. Domain-specific operation logic is implemented afterwards in Go.

## Start from the contract

Use [`generator/schema/schema.yml`](https://github.com/ARM-software/go-sdk-generator/blob/main/generator/schema/schema.yml) for the complete field definitions, or print the schema as JSON with `go-sdk-generator schema`. The repository's [`generate.yml`](https://github.com/ARM-software/go-sdk-generator/blob/main/generate.yml) shows a real input. The key sections under `specification` are:

| Field | Purpose |
| --- | --- |
| `package_name`, `main_module`, `name`, `description` | Project and SDK identity. |
| `repository_url`, `public` | Repository location and whether the project is public; used across generated project metadata and hosting integrations. |
| `interface` | Public Go operations and their arguments and results. |
| `global_configuration` | Values shared across the SDK and CLI. |
| `command_line_commands` | Optional binary and commands exposing those operations. |
| `additional_types` | Optional domain types used by the contract. |
| `tag_latest`, `update_copyright_headers` | Optional release tagging and copyright-header maintenance settings. |

When operations accept functional options, use `specification.options`; see [configure functional options](configure-options.md) for option wrappers, `value_field` and CLI bindings.

A minimal starting shape is:

```yaml
specification:
  package_name: example-tool
  main_module: example
  name: Example
  description: Verify an example project.
  licence: Apache-2.0
  interface:
    notCloseable: true
    namespace: github.com/example/
    methods:
      - name: Verify
        description: Verify an example project.
        returns:
          - name: err
            lang_type: error
            description: Verification failure, if any.
  global_configuration: []
```

This describes a library surface. Add `command_line_commands` if users also need a CLI; see the [repository manifest](https://github.com/ARM-software/go-sdk-generator/blob/main/generate.yml) for commands, flags, and shared configuration. Adapt the namespace, name, operation, and licence to your project. [What the generator produces](what-it-generates.md) explains the resulting project structure.

Repository-level automation can also be controlled from the top level of `specification`:

- `tag_latest` defaults to `false`. Set it to `true` to keep a `latest` tag pointing to the most recent stable release. Consumers can follow published releases without updating their references or tracking a development branch. Because the tag moves, builds that use it are not reproducible and can pick up breaking changes; pin a version when that matters. See [Docker's guidance on tags and aliases](https://github.com/docker-library/official-images#tags-and-aliases) and [pinning image versions](https://docs.docker.com/build/building/best-practices/#pin-base-image-versions), and [GitHub's guidance on managing action versions](https://docs.github.com/en/actions/how-tos/create-and-publish-actions/manage-custom-actions).
- `update_copyright_headers` defaults to `true`. Copyright headers identify the copyright holders and ownership of source files, often with applicable years and an SPDX licence identifier. They support attribution, clear licensing terms, and automated licence checks. For GitHub-hosted projects, enabling this setting generates a workflow that updates headers on pull requests and proposes a maintenance pull request on scheduled or manual runs. Generator upgrades refresh headers too. Set it to `false` to opt out. See the [REUSE specification](https://reuse.software/spec-3.3/) and [GNU copyright notice guidance](https://www.gnu.org/prep/maintain/html_node/Copyright-Notices.html).

## Model values deliberately

- Use specific `lang_type` aliases such as `duration`, `uuid`, `filepath`, or `service` where they match the value. See [types, formats and defaults](types-and-defaults.md) for the complete built-in reference. Prefer `reference: true`, `list: true`, or `iterator: true` to encoding shapes into the type string; `list` and `iterator` cannot be combined.
- Use `enum` for closed string choices, `constraints` for validation and relationships, and `example` or `examples` for representative values. These feed generated validation, help, documentation, and action inputs.
- Keep truly shared values in `global_configuration`; put operation-specific values on that command. Mark credentials with `secret: true` and follow [secret management in generated CLI tools](secret-management.md) for storage and loading behaviour.
- Give operations verb names and write descriptions from a user's perspective. Avoid defining a transport endpoint as an entire operation if the intended action is a higher-level workflow.
- For incompatible changes to a published contract, document a replacement with `deprecation_notice` before removing the old surface. Follow [breaking changes and deprecation](evolving-a-contract.md).

The [manifest-authoring skill](https://github.com/ARM-software/go-sdk-generator/blob/main/.agents/skills/manifest-authoring/SKILL.md) contains a detailed checklist for agents and maintainers.
