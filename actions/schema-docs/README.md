<!--
Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
SPDX-License-Identifier: Apache-2.0
-->

# Generate JSON Schema guides

This local composite action uses
[jsonschema-markdown](https://github.com/elisiariocouto/jsonschema-markdown) to turn
a JSON Schema into a Markdown reference guide. It accepts `.json`, `.yaml` and
`.yml` sources, including YAML anchors and merge keys. References to definitions
within the schema remain links to sections in the generated document.

## Inputs

| Input | Required | Description |
| --- | --- | --- |
| `schema` | Yes | Source schema path, relative to the checked-out workspace. |
| `output` | Yes | Destination Markdown path; missing parent directories are created. |
| `title` | No | Override the document heading; otherwise use the schema's title. |

The action runs on Linux runners, sets up Python 3.11 and installs pinned renderer
dependencies in a temporary virtual environment. This keeps them separate from
the release tooling's Python dependencies. It generates one guide per invocation;
call it again with different paths to document more schemas.

## Use in this repository

```yaml
- uses: actions/checkout@v7
- name: Generate manifest schema reference
  uses: $/actions/schema-docs
  with:
    schema: generator/schema/schema.yml
    output: guides/schema-reference.md
    title: Generator manifest schema reference
- name: Stage generated schema guide
  run: git add -- guides/schema-reference.md
```

`$/actions/schema-docs` selects the action from the running workflow's repository
and revision. The checkout supplies the source schema and destination workspace.

The action writes the requested Markdown file but does not commit it. The calling
workflow is responsible for committing the generated documentation.

In this repository, `generator/schema/schema.yml` is the source of the manifest
schema reference. The action reads its field definitions and descriptions and
writes the reference to `guides/schema-reference.md`. That Markdown file is stored
in Git, so readers can browse it directly from the repository's guides index.
The documentation build also converts it into a page on the documentation site,
alongside the other guides.

To change the reference, edit the definitions or descriptions in
`generator/schema/schema.yml`, then run the command below to regenerate
`guides/schema-reference.md`. Commit both files when updating the reference in a
pull request. Avoid editing the Markdown reference directly: the next generation
run replaces its contents. The release workflow also regenerates it automatically
before tagging, using the schema in the release revision.

## Run locally

With Python 3.11 or later and the packages in `requirements.txt` installed in a
virtual environment, run from the repository root:

```bash
python actions/schema-docs/generate.py \
  --schema generator/schema/schema.yml \
  --output guides/schema-reference.md \
  --title "Generator manifest schema reference"
python -m unittest discover -s actions/schema-docs/tests -v
```

The renderer documents the JSON Schema features supported by jsonschema-markdown;
the source schema remains authoritative. This action generates reference material,
not task-oriented examples or a replacement for runtime validation.
