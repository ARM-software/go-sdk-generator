package generator

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGeneratedActions(t *testing.T) {
	t.Run("builds visible subcommand actions with repeated root inputs", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				Name:        field.ToOptionalString("Example Tool"),
				CommandLineCommands: &CommandLineCommands{
					Binary: field.ToOptionalString("example-tool"),
					Functions: []CliFunction{
						{
							Name:        "Hidden",
							Description: "hidden command",
							Hidden:      true,
							Configuration: []FunctionConfigFields{
								{
									Name:            "APIKey",
									LangType:        "string",
									MapstructureTag: "api_key",
									Description:     "API key",
									IsRoot:          true,
									Required:        true,
								},
							},
						},
						{
							Name:        "Generate",
							Description: "generate output",
							Operands:    true,
							Configuration: []FunctionConfigFields{
								{
									Name:            "Debug",
									LangType:        "bool",
									MapstructureTag: "debug",
									Description:     "Enable debug logging",
									IsRoot:          true,
								},
								{
									Name:            "Input",
									LangType:        "string",
									MapstructureTag: "input",
									Description:     "Input file",
									Required:        true,
								},
								{
									Name:            "HiddenFlag",
									LangType:        "bool",
									MapstructureTag: "hidden_flag",
									Description:     "Hidden flag",
									Hidden:          true,
								},
								{
									Name:            "Board",
									LangType:        "types.BoardID",
									MapstructureTag: "board",
									Description:     "Board override",
									Bindings: []Binding{
										{
											Name:            field.ToOptionalString("Name"),
											Arg:             field.ToOptionalString("board-name"),
											Description:     field.ToOptionalString("Board name"),
											LangType:        field.ToOptionalString("string"),
											MapstructureTag: field.ToOptionalString("name"),
										},
										{
											Name:            field.ToOptionalString("Secret"),
											Description:     field.ToOptionalString("Hidden secret"),
											LangType:        field.ToOptionalString("string"),
											MapstructureTag: field.ToOptionalString("secret"),
											Hidden:          true,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		require.Len(t, actions, 1)

		action := actions[0]
		assert.Equal(t, "generate", action.DirectoryName)
		assert.Equal(t, "generate", action.Subcommand)
		assert.Equal(t, "Example Tool - Generate Action", action.Name)
		assert.True(t, action.HasOperands)

		inputNames := actionInputNames(t, action.Inputs)

		assert.Contains(t, inputNames, "api-key")
		assert.Contains(t, inputNames, "debug")
		assert.Contains(t, inputNames, "input")
		assert.Contains(t, inputNames, "board-name")
		assert.Contains(t, inputNames, "operands")
		assert.NotContains(t, inputNames, "hidden-flag")
		assert.NotContains(t, inputNames, "secret")
	})

	t.Run("skips hidden commands", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:          "Hidden",
						Description:   "hidden command",
						Hidden:        true,
						Configuration: []FunctionConfigFields{},
					}},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		assert.Empty(t, actions)
	})

	t.Run("skips commands with no action enabled", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{
						{
							Name:          "Generate",
							Description:   "generate output",
							Configuration: []FunctionConfigFields{},
							NoAction:      true,
						},
						{
							Name:          "Validate",
							Description:   "validate output",
							Configuration: []FunctionConfigFields{},
						},
					},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		require.Len(t, actions, 1)
		assert.Equal(t, "validate", actions[0].DirectoryName)
	})

	t.Run("sets no-keyring action input default to true", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Verify",
						Description: "verify output",
						Configuration: []FunctionConfigFields{{
							Name:            "NoKeyring",
							LangType:        "bool",
							MapstructureTag: "no_keyring",
							Description:     "Disable keyring",
							Bindings: []Binding{{
								Arg:             field.ToOptionalString("no-keyring"),
								MapstructureTag: field.ToOptionalString("no_keyring"),
								Description:     field.ToOptionalString("Disable keyring"),
								LangType:        field.ToOptionalString("bool"),
							}},
						}},
					}},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		require.Len(t, actions, 1)
		require.Len(t, actions[0].Inputs, 1)
		match, err := matchingActionInputFlagName(noKeyringArgument, actions[0].Inputs[0].Name)
		require.NoError(t, err)
		assert.True(t, match)
		assert.Equal(t, "true", actions[0].Inputs[0].Default)
	})
}

func actionInputNames(t *testing.T, inputs []GeneratedActionInput) (names []string) {
	t.Helper()

	names = make([]string, 0, len(inputs))

	for i := range inputs {
		names = append(names, inputs[i].Name)
	}

	return
}
