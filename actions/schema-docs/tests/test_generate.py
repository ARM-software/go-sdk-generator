#
# Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
"""Exercise schema documentation through the action's command-line entry point."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

import yaml


ACTION = Path(__file__).resolve().parents[1]
ROOT = ACTION.parents[1]
SCHEMA = """\
title: Example schema
description: A schema with shared fields and references.
type: object
required: [name]
$defs:
  label: &label
    type: string
    description: A readable label containing a literal `|`.
    minLength: 2
properties:
  name:
    <<: *label
    default: demo
  mode:
    type: string
    enum: [fast, safe]
  label:
    $ref: '#/$defs/label'
"""


class GenerateGuideTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)

    def render(self, source, output, *extra):
        return subprocess.run(
            [sys.executable, str(ACTION / "generate.py"), "--schema", str(source),
             "--output", str(output), *extra],
            capture_output=True, text=True, encoding="utf-8", check=False,
        )

    def test_yaml_and_json_generate_identical_guides(self):
        outputs = []
        for suffix in (".yaml", ".yml", ".json"):
            with self.subTest(suffix=suffix):
                source = self.root / f"schema with spaces{suffix}"
                text = json.dumps(yaml.safe_load(SCHEMA)) if suffix == ".json" else SCHEMA
                source.write_text(text, encoding="utf-8")
                output = self.root / "new guides" / f"schema{suffix}.md"
                result = self.render(source, output)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(source.read_text(encoding="utf-8"), text)
                outputs.append(output.read_bytes())
        self.assertEqual(outputs[0], outputs[1])
        self.assertEqual(outputs[0], outputs[2])
        markdown = outputs[0].decode("utf-8")
        for expected in ("# Example schema", "✅", '`"demo"`', "`fast`", "[label](#label)",
                         r"A readable label containing a literal `\|`."):
            self.assertIn(expected, markdown)
        self.assertNotIn(b"\r\n", outputs[0])

    def test_title_override_and_repeatable_output(self):
        source = self.root / "schema.yml"
        source.write_text(SCHEMA, encoding="utf-8")
        output = self.root / "guide.md"
        first = self.render(source, output, "--title", "Custom reference")
        self.assertEqual(first.returncode, 0, first.stderr)
        expected = output.read_bytes()
        self.assertIn(b"# Custom reference", expected)
        second = self.render(source, output, "--title", "Custom reference")
        self.assertEqual(second.returncode, 0, second.stderr)
        self.assertEqual(output.read_bytes(), expected)

    def test_invalid_input_preserves_existing_guide(self):
        output = self.root / "guide.md"
        output.write_text("Existing guide\n", encoding="utf-8")
        for suffix, text in ((".yml", "properties: ["), (".json", "{broken"),
                             (".yaml", ""), (".yaml", "[]"), (".yaml", "true")):
            with self.subTest(suffix=suffix, text=text):
                source = self.root / f"schema{suffix}"
                source.write_text(text, encoding="utf-8")
                result = self.render(source, output)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(output.read_text(encoding="utf-8"), "Existing guide\n")

    def test_output_cannot_overwrite_schema(self):
        source = self.root / "schema.yml"
        source.write_text(SCHEMA, encoding="utf-8")
        result = self.render(source, source)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("must not overwrite", result.stderr)
        self.assertEqual(source.read_text(encoding="utf-8"), SCHEMA)

    def test_repository_schema_includes_nested_definitions(self):
        output = self.root / "reference.md"
        result = self.render(ROOT / "generator/schema/schema.yml", output)
        self.assertEqual(result.returncode, 0, result.stderr)
        markdown = output.read_text(encoding="utf-8")
        for expected in ("specification.package_name", "## parameter_constraints",
                         "dependent_required", "require_together_as_cli_flags",
                         "[parameter_constraints](#parameter_constraints)"):
            self.assertIn(expected, markdown)


if __name__ == "__main__":
    unittest.main()
