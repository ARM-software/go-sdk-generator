/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestValidateExamplesDefinition(t *testing.T) {
	t.Run("rejects both example and examples", func(t *testing.T) {
		err := validateExamplesDefinition("field \"mode\"", "string", false, false, nil, field.ToOptionalString("fast"), []string{"safe"})
		require.Error(t, err)
	})

	t.Run("rejects invalid int example", func(t *testing.T) {
		err := validateExamplesDefinition("field \"count\"", "int", false, false, nil, field.ToOptionalString("not-an-int"), nil)
		require.Error(t, err)
	})

	t.Run("rejects invalid enum example", func(t *testing.T) {
		err := validateExamplesDefinition("field \"mode\"", "string", false, false, []string{"fast", "safe"}, field.ToOptionalString("turbo"), nil)
		require.Error(t, err)
	})

	t.Run("accepts bool and url examples", func(t *testing.T) {
		require.NoError(t, validateExamplesDefinition("field \"enabled\"", "bool", false, false, nil, field.ToOptionalString("true"), nil))
		require.NoError(t, validateExamplesDefinition("field \"host\"", "host", false, false, nil, nil, []string{"https://example.com"}))
	})
}

func TestBuildGeneratedActionsIncludesExamplesInDescriptions(t *testing.T) {
	sdk := &SDK{
		Specification: SDKSpecification{
			PackageName: "example-tool",
			Name:        field.ToOptionalString("Example Tool"),
			CommandLineCommands: &CommandLineCommands{
				Binary: field.ToOptionalString("example-tool"),
				Functions: []CliFunction{{
					Name:        "Verify",
					Description: "verify output",
					Configuration: []FunctionConfigFields{{
						Name:            "Host",
						LangType:        "host",
						MapstructureTag: "host",
						Description:     "Host to verify.",
						Example:         field.ToOptionalString("https://example.com"),
					}},
				}},
			},
		},
	}

	actions, err := BuildGeneratedActions(sdk)
	require.NoError(t, err)
	require.Len(t, actions, 2)
	for _, action := range actions {
		if strings.Contains(action.Subcommand, "erify") {
			require.Len(t, action.Inputs, 1)
			assert.Contains(t, action.Inputs[0].Description, `Example: "https://example.com".`)
		}
	}

}

func TestEmbeddedDocumentationTemplateIncludesExamples(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/documentation.md.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("Example"),
				PackageName: "example-sdk",
				MainModule:  "example",
				GlobalConfiguration: Configuration{{
					Name:            "Host",
					LangType:        "host",
					MapstructureTag: "host",
					Description:     "Host to verify.",
					Example:         field.ToOptionalString("https://example.com"),
				}},
				Interface: Interface{Methods: []InterfaceMethod{{Name: "Verify", Description: "Verify.", Returns: []CommandArgument{{Name: "err", LangType: "error", Description: "Whether verify failed."}}}}},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "documentation.md.tmpl", data))
	assert.Contains(t, b.String(), `Example: "https://example.com".`)
}

func TestEmbeddedCLISkillTemplateIncludesExamples(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/cli.skill.md.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("Example"),
				Description: field.ToOptionalString("Example CLI skill."),
				PackageName: "example-tool",
				MainModule:  "example",
				Licence:     "MIT",
				CommandLineCommands: &CommandLineCommands{
					Binary:             field.ToOptionalString("example-tool"),
					RootEnvironmentVar: field.ToOptionalString("EXAMPLE"),
					Functions: []CliFunction{{
						Name:        "Verify",
						Description: "Verify output.",
						Configuration: []FunctionConfigFields{{
							Name:            "Host",
							LangType:        "host",
							MapstructureTag: "host",
							Description:     "Host to verify.",
							Example:         field.ToOptionalString("https://example.com"),
						}},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "cli.skill.md.tmpl", data))
	assert.Contains(t, b.String(), `Example: "https://example.com".`)
	assert.NotContains(t, b.String(), "example-tool credentials")
}

func TestEmbeddedCLISkillTemplateUsesBindingDescriptions(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, cliSkillFile)
	require.NoError(t, err)
	for _, required := range []bool{false, true} {
		t.Run(fmt.Sprintf("required=%t", required), func(t *testing.T) {
			data := &DataModel{SDK: SDK{Specification: SDKSpecification{
				Name: field.ToOptionalString("Example"), PackageName: "example", MainModule: "example",
				Description: field.ToOptionalString("Example CLI skill."),
				CommandLineCommands: &CommandLineCommands{Binary: field.ToOptionalString("example"), Functions: []CliFunction{{
					Name: "Generate", Configuration: []FunctionConfigFields{{
						Name: "Overwrite", LangType: "options.OverwriteOptions", MapstructureTag: "overwrite",
						Description: "Parent option-set description.", Required: required,
						DefaultValue: field.ToOptionalString("*options.DefaultOverwriteOptions()"),
						Bindings: []Binding{
							{Name: field.ToOptionalString("Enabled"), Arg: field.ToOptionalString("overwrite"), Description: field.ToOptionalString("overwrite state files")},
							{Name: field.ToOptionalString("ExcludeFiles"), Arg: field.ToOptionalString("exclude-files"), Description: field.ToOptionalString("Files to preserve."), Example: field.ToOptionalString("README.md")},
							{Name: field.ToOptionalString("Internal"), Arg: field.ToOptionalString("internal-flag"), Description: field.ToOptionalString("Hidden binding."), Hidden: true},
						},
					}, {
						Name: "Internal", MapstructureTag: "hidden-parent", Description: "Hidden parent.", Hidden: true,
					}},
				}}},
			}}}
			var b strings.Builder
			require.NoError(t, tmpl.ExecuteTemplate(&b, "cli.skill.md.tmpl", data))
			output := strings.ReplaceAll(b.String(), "\r\n", "\n")
			assert.Contains(t, output, "- `overwrite`: overwrite state files")
			assert.Contains(t, output, "- `exclude-files`: Files to preserve. Example: \"README.md\".")
			assert.NotContains(t, output, "Parent option-set description.")
			assert.NotContains(t, output, "internal-flag")
			assert.NotContains(t, output, "hidden-parent")
			assert.NotContains(t, output, "DefaultOverwriteOptions")
			if required {
				assert.Contains(t, output, "Optional:\n- none")
			} else {
				assert.Contains(t, output, "Required:\n- none")
			}
		})
	}
}

func TestEmbeddedCLISkillTemplateIncludesCredentialsGuidanceWhenManifestDefinesSecretInputs(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/cli.skill.md.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("Example"),
				Description: field.ToOptionalString("Example CLI skill."),
				PackageName: "example-tool",
				MainModule:  "example",
				Licence:     "MIT",
				CommandLineCommands: &CommandLineCommands{
					Binary:             field.ToOptionalString("example-tool"),
					RootEnvironmentVar: field.ToOptionalString("EXAMPLE"),
					Functions: []CliFunction{{
						Name:        "Verify",
						Description: "Verify output.",
						Configuration: []FunctionConfigFields{{
							Name:            "Token",
							LangType:        stringLangType,
							MapstructureTag: "token",
							Description:     "Token to verify.",
							Secret:          true,
						}},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "cli.skill.md.tmpl", data))
	assert.Contains(t, b.String(), "This tool may expose a `example-tool credentials` command when its manifest defines secret inputs.")
	assert.Contains(t, b.String(), "- Use `example-tool credentials` to manage secrets")
}

func TestEmbeddedCLISkillTemplateIncludesResultOutputCombinationTable(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/cli.skill.md.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("Result Output SDK"),
				Description: field.ToOptionalString("Example CLI skill."),
				PackageName: "result-output-sdk",
				MainModule:  "resultoutput",
				Licence:     "MIT",
				CommandLineCommands: &CommandLineCommands{
					Binary:             field.ToOptionalString("result-output"),
					RootEnvironmentVar: field.ToOptionalString("RESULT_OUTPUT"),
					Functions: []CliFunction{{
						Name:        "Inspect",
						Description: "Emit command results to the selected destination.",
						WithResult:  true,
						Configuration: []FunctionConfigFields{{
							Name:            resultOutputArgument,
							LangType:        stringLangType,
							MapstructureTag: resultOutputMapstructureTag,
							Description:     "output format",
						}},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "cli.skill.md.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "Result output and logging combinations:")
	assert.Contains(t, output, "| `output_format` | `output_file` | Allowed `logger` values | Behaviour | Allowed |")
	assert.Contains(t, output, "| `human` | empty or `-` | `console`, `plain`, `rich`, `json`, `quiet`, `file`, `none`, `custom` |")
	assert.Contains(t, output, "| `json` or `yaml` | empty or `-` | `console`, `plain`, `rich`, `json`, `quiet`, `file`, `none`, `custom` |")
	assert.Contains(t, output, "`logger=file` still requires `logger_file` in every allowed combination.")
}
