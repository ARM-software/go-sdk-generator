package generator

import (
	"strings"
	"testing"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}
