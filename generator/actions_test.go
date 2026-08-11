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
		require.Len(t, actions, 2)

		installAction := requireAction(t, actions, installToolActionDirectoryName)
		assert.True(t, installAction.InstallOnly)
		assert.Empty(t, installAction.Inputs)

		action := requireAction(t, actions, "generate")
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
		require.Len(t, actions, 1)
		assert.Equal(t, installToolActionDirectoryName, actions[0].DirectoryName)
		assert.True(t, actions[0].InstallOnly)
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
		require.Len(t, actions, 2)
		assert.Equal(t, installToolActionDirectoryName, actions[0].DirectoryName)
		assert.Equal(t, "validate", actions[1].DirectoryName)
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
		require.Len(t, actions, 2)
		verifyAction := requireAction(t, actions, "verify")
		require.Len(t, verifyAction.Inputs, 1)
		match, err := matchingActionInputFlagName(noKeyringArgument, verifyAction.Inputs[0].Name)
		require.NoError(t, err)
		assert.True(t, match)
		assert.Equal(t, "true", verifyAction.Inputs[0].Default)
	})

	t.Run("defaults summary action input to github for summary-aware commands", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Generate",
						Description: "generate output",
						Configuration: []FunctionConfigFields{
							{
								Name:            summaryArgument,
								LangType:        "string",
								MapstructureTag: "summary",
								Description:     "summary destination",
							},
							{
								Name:            summaryFileArg,
								LangType:        "string",
								MapstructureTag: "summary_file",
								Description:     "summary file",
							},
						},
					}},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		require.Len(t, actions, 2)

		action := requireAction(t, actions, "generate")
		require.Len(t, action.Inputs, 2)
		assert.Equal(t, summaryMapstructureTag, action.Inputs[0].FlagName)
		assert.Equal(t, string(SummaryDestinationGithub), action.Inputs[0].Default)
	})

	t.Run("explicitly disabling summary keeps action input unset", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				PackageName: "example-tool",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Generate",
						Description: "generate output",
						WithSummary: field.ToOptionalBool(false),
						Configuration: []FunctionConfigFields{{
							Name:            "Input",
							LangType:        "string",
							MapstructureTag: "input",
							Description:     "input",
						}},
					}},
				},
			},
		}

		actions, err := BuildGeneratedActions(sdk)
		require.NoError(t, err)
		action := requireAction(t, actions, "generate")
		require.Len(t, action.Inputs, 1)
		assert.Equal(t, "input", action.Inputs[0].FlagName)
	})
}

func requireAction(t *testing.T, actions []GeneratedAction, directoryName string) GeneratedAction {
	t.Helper()

	for i := range actions {
		if actions[i].DirectoryName == directoryName {
			return actions[i]
		}
	}

	require.Failf(t, "missing action", "could not find action %q", directoryName)
	return GeneratedAction{}
}

func actionInputNames(t *testing.T, inputs []GeneratedActionInput) (names []string) {
	t.Helper()

	names = make([]string, 0, len(inputs))

	for i := range inputs {
		names = append(names, inputs[i].Name)
	}

	return
}
