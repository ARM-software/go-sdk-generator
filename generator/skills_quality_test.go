/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestPortableSkillNames(t *testing.T) {
	for _, binary := range []string{"tool", "My_Tool.EXE", "a--b", strings.Repeat("long-", 25)} {
		for _, name := range []string{runSkillName(binary), installSkillName(binary), githubActionSkillName(binary)} {
			assert.LessOrEqual(t, len(name), 64)
			assert.Regexp(t, `^[a-z0-9]+(-[a-z0-9]+)*$`, name)
		}
	}
	assert.NotEqual(t, runSkillName(strings.Repeat("a", 80)), runSkillName(strings.Repeat("a", 79)+"b"))
}

func TestAdditionalSkillCatalogue(t *testing.T) {
	d := &DataModel{OutputDirectory: t.TempDir(), SDK: SDK{Specification: SDKSpecification{
		PackageName: "example", MainModule: "example", Licence: "MIT",
		CommandLineCommands: &CommandLineCommands{Binary: field.ToOptionalString("example")},
	}}}
	root := filepath.Join(d.OutputDirectory, ".agents", "skills", "custom")
	require.NoError(t, os.MkdirAll(root, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("---\nname: custom\ndescription: Use when maintaining examples.\n---\nInstructions.\n"), 0600))
	catalogue, err := d.AdditionalSkillCatalogue()
	require.NoError(t, err)
	require.Len(t, catalogue, 1)
	assert.Equal(t, SkillInfo{Name: "custom", Description: "Use when maintaining examples.", Path: ".agents/skills/custom"}, catalogue[0])
	product, err := d.SkillCatalogue()
	require.NoError(t, err)
	assert.Len(t, product, 2, "repository-maintained skills do not become product distribution entries")
}

// These projects are generated through the public entrypoint, rather than by
// rendering isolated templates. CI enables the same local QA script they emit.
func TestGeneratedSkillsQuality(t *testing.T) {
	for _, tc := range []struct {
		name         string
		cli, actions bool
	}{
		{"with-actions", true, true},
		{"without-actions", true, false},
		{"library", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("testdata", "cli-package.yml"))
			require.NoError(t, err)
			var sdk SDK
			require.NoError(t, unmarshalManifestContent(".yml", content, &sdk))
			sdk.Specification.PackageName = "different-package"
			sdk.Specification.RepositoryUrl = field.ToOptionalString("https://github.com/example/skills-product")
			if tc.cli {
				sdk.Specification.CommandLineCommands.Binary = field.ToOptionalString("sample-tool")
				sdk.Specification.CommandLineCommands.NoActions = !tc.actions
			} else {
				sdk.Specification.CommandLineCommands = nil
			}
			input, err := yaml.Marshal(sdk)
			require.NoError(t, err)
			manifest := filepath.Join(t.TempDir(), "generate.yml")
			require.NoError(t, os.WriteFile(manifest, input, 0600))
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input, cfg.Output = manifest, t.TempDir()
			cfg.ToolVersion = field.ToOptionalString("0.31.0")
			require.NoError(t, Generate(t.Context(), nil, cfg))
			read := func(name string) string {
				data, readErr := os.ReadFile(filepath.Join(cfg.Output, filepath.FromSlash(name)))
				require.NoError(t, readErr)
				return string(data)
			}
			for _, file := range []string{".github/workflows/skills.yml", "actions/validate-skills/action.yml", "actions/validate-skills/validate_skills.py", "actions/validate-skills/requirements.txt", "guides/agent-skills.md"} {
				if tc.cli {
					require.FileExists(t, filepath.Join(cfg.Output, file))
				} else {
					assert.NoFileExists(t, filepath.Join(cfg.Output, file))
				}
			}
			if !tc.cli {
				assert.NoDirExists(t, filepath.Join(cfg.Output, ".agents", "skills"))
				assert.NotContains(t, read("README.md"), "## Agent Skills")
				assert.NotContains(t, read("DOCUMENTATION.md"), "## Agent Skills")
				assert.NotContains(t, read(".goreleaser.yml"), "/skills")
				return
			}
			assert.Contains(t, read("README.md"), "./DOCUMENTATION.md#agent-skills")
			assert.NotContains(t, read("README.md"), "<details>")
			assert.Contains(t, read("DOCUMENTATION.md"), "npx skills add https://github.com/example/skills-product --list")
			assert.Contains(t, read("DOCUMENTATION.md"), "/etc/different-package/.agents/skills")
			assert.Contains(t, read("guides/agent-skills.md"), "--agent claude-code")
			assert.Contains(t, read(".github/workflows/skills.yml"), "uses: ./actions/validate-skills")
			assert.Contains(t, read("actions/validate-skills/action.yml"), "github.com/coolapso/agent-skills-validator@v0.2.0")
			assert.Equal(t, "/etc/different-package/.agents/skills", toolSkillsInstallPath(&sdk.Specification))
			actionPath := filepath.Join(cfg.Output, ".agents", "skills", "using-sample-tool-in-a-github-action", "SKILL.md")
			if tc.actions {
				assert.FileExists(t, actionPath)
			} else {
				assert.NoFileExists(t, actionPath)
				assert.NotContains(t, read("DOCUMENTATION.md"), "| [using-sample-tool-in-a-github-action]")
			}
			skills, err := filepath.Glob(filepath.Join(cfg.Output, ".agents", "skills", "*", "SKILL.md"))
			require.NoError(t, err)
			for _, skill := range skills {
				data, readErr := os.ReadFile(skill)
				require.NoError(t, readErr)
				require.True(t, strings.HasPrefix(strings.ReplaceAll(string(data), "\r\n", "\n"), "---\n"))
				var meta map[string]any
				require.NoError(t, yaml.Unmarshal([]byte(strings.SplitN(string(data), "---", 3)[1]), &meta))
				assert.Equal(t, filepath.Base(filepath.Dir(skill)), meta["name"])
				for key, value := range meta["metadata"].(map[string]any) {
					assert.IsType(t, "", value, key)
				}
				assert.Contains(t, read("DOCUMENTATION.md"), meta["description"])
			}
			if os.Getenv("CHECK_GENERATED_SKILLS") == "1" {
				// Supporting files and a subsequently added skill must be discovered
				// and installed too, without changing the generated workflow.
				extra := filepath.Join(cfg.Output, ".agents", "skills", "maintainer-added")
				require.NoError(t, os.MkdirAll(filepath.Join(extra, "references", "examples"), 0755))
				require.NoError(t, os.WriteFile(filepath.Join(extra, "SKILL.md"), []byte("---\nname: maintainer-added\ndescription: Use when testing an additional skill with bundled resources.\n---\nSee [example](references/examples/input.txt).\n"), 0600))
				require.NoError(t, os.WriteFile(filepath.Join(extra, "references", "examples", "input.txt"), []byte("complete bundled example\n"), 0600))
				cmd := exec.CommandContext(t.Context(), "python", "actions/validate-skills/validate_skills.py", "--install")
				cmd.Dir = cfg.Output
				out, checkErr := cmd.CombinedOutput()
				require.NoError(t, checkErr, "%s", out)
				// Verify an escaped resource is rejected even though the frontmatter is valid.
				resources, openErr := os.OpenRoot(extra)
				require.NoError(t, openErr)
				t.Cleanup(func() { require.NoError(t, resources.Close()) })
				data, readErr := resources.ReadFile("SKILL.md")
				require.NoError(t, readErr)
				require.NoError(t, resources.WriteFile("SKILL.md", append(data, []byte("\n[missing](../outside.md)\n")...), 0600))
				cmd = exec.CommandContext(t.Context(), "python", "actions/validate-skills/validate_skills.py")
				cmd.Dir = cfg.Output
				out, checkErr = cmd.CombinedOutput()
				require.Error(t, checkErr, "%s", out)
				assert.Contains(t, string(out), "missing or non-portable resource")
				require.NoError(t, resources.WriteFile("SKILL.md", data, 0600))
				duplicate := filepath.Join(cfg.Output, "extra-skills", "maintainer-added")
				require.NoError(t, os.MkdirAll(duplicate, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(duplicate, "SKILL.md"), []byte("---\nname: maintainer-added\ndescription: Use when checking a duplicate name.\n---\nInstructions.\n"), 0600))
				cmd = exec.CommandContext(t.Context(), "python", "actions/validate-skills/validate_skills.py")
				cmd.Dir = cfg.Output
				out, checkErr = cmd.CombinedOutput()
				require.Error(t, checkErr, "%s", out)
				assert.Contains(t, string(out), "Duplicate skill maintainer-added")
			}
		})
	}
}
