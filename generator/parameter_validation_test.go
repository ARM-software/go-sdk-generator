/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestDeprecationNoticeValidateRejectsMissingIssueDate(t *testing.T) {
	t.Parallel()

	notice := &DeprecationNotice{
		RemovalDate: field.ToOptionalTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
	}

	require.Error(t, notice.Validate())
}

func TestDeprecationNoticeValidateRejectsRemovalBeforeIssue(t *testing.T) {
	t.Parallel()

	notice := &DeprecationNotice{
		IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
		RemovalDate: field.ToOptionalTime(time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC)),
	}

	require.Error(t, notice.Validate())
}

func TestFunctionConfigFieldsValidateAcceptsValidDeprecationNotice(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "LegacyToken",
		LangType:        "string",
		MapstructureTag: "legacy_token",
		Description:     "Legacy token.",
		DeprecationNotice: &DeprecationNotice{
			IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
			RemovalDate: field.ToOptionalTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
			Description: field.ToOptionalString("Prefer the scoped token path."),
			Replacement: field.ToOptionalString("ScopedToken"),
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestProcessConstraintMirrorsPreservesExistingPattern(t *testing.T) {
	t.Parallel()

	localFs := filesystem.NewStandardFileSystem()
	data, err := filesystem.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "verify.yml"))
	require.NoError(t, err)

	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", data, &sdk))
	require.GreaterOrEqual(t, len(sdk.Specification.GlobalConfiguration), 3)
	require.NotNil(t, sdk.Specification.GlobalConfiguration[2].Pattern)
	assertPattern := *sdk.Specification.GlobalConfiguration[2].Pattern
	require.Equal(t, `.+?[/\s][\d.]+`, assertPattern)

	require.NoError(t, ProcessConstraintMirrors(context.Background(), &sdk))
	require.NotNil(t, sdk.Specification.GlobalConfiguration[2].Pattern)
	require.Equal(t, `.+?[/\s][\d.]+`, *sdk.Specification.GlobalConfiguration[2].Pattern)
}

func TestConfigurationOptionValidateRejectsStringLengthConstraintOnNonString(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Count",
		LangType:        "int",
		MapstructureTag: "count",
		Description:     "Count.",
		Constraints: &ParameterConstraints{
			MinLength: field.ToOptional(1),
		},
	}

	require.Error(t, cfg.Validate())
}

func TestFunctionConfigFieldsValidateRejectsExactlyOneOnScalar(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "Mode",
		LangType:        "string",
		MapstructureTag: "mode",
		Description:     "Mode.",
		Constraints: &ParameterConstraints{
			ExactlyOne: true,
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateRejectsConflictingConstraintEnumAlias(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Format",
		LangType:        "string",
		MapstructureTag: "format",
		Description:     "Format.",
		Enum:            []string{"json", "yaml"},
		Constraints: &ParameterConstraints{
			Enum: []string{"xml"},
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateAcceptsConstraintValidationAliases(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Format",
		LangType:        "string",
		MapstructureTag: "format",
		Description:     "Format.",
		DefaultValue:    field.ToOptionalString("json"),
		Constraints: &ParameterConstraints{
			Enum:      []string{"json", "yaml"},
			Pattern:   field.ToOptionalString(`^(json|yaml)$`),
			Required:  true,
			MinLength: field.ToOptional(4),
			MaxLength: field.ToOptional(4),
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestConfigurationOptionValidateRejectsSubstringConstraintsOnNonString(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Count",
		LangType:        "int",
		MapstructureTag: "count",
		Description:     "Count.",
		Constraints: &ParameterConstraints{
			Contains: field.ToOptionalString("1"),
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateAcceptsNullableStringConstraints(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Name",
		LangType:        "string",
		Reference:       true,
		MapstructureTag: "name",
		Description:     "Name.",
		Constraints: &ParameterConstraints{
			Nullable:    true,
			NotEmpty:    true,
			Contains:    field.ToOptionalString("sdk"),
			NotContains: field.ToOptionalString("tmp"),
			Prefix:      field.ToOptionalString("arm-"),
			Suffix:      field.ToOptionalString("-gen"),
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestFunctionConfigFieldsValidateRejectsOptionalEnumWithoutDefault(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "Mode",
		LangType:        "string",
		MapstructureTag: "mode",
		Description:     "Mode.",
		Enum:            []string{"fast", "safe"},
	}

	err := cfg.Validate()
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "must define a default value")
}

func TestFunctionConfigFieldsValidateRejectsRequiredEnumWithoutDefault(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "Mode",
		LangType:        "string",
		MapstructureTag: "mode",
		Description:     "Mode.",
		Enum:            []string{"fast", "safe"},
		Required:        true,
	}

	err := cfg.Validate()
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "must define a default value")
}

func TestValidateProtectedCLIArgumentNamesRejectsProtectedConfigurationNames(t *testing.T) {
	t.Parallel()

	cmds := &CommandLineCommands{
		Functions: []CliFunction{{
			Name:        "Run",
			Description: "Run.",
			Configuration: []FunctionConfigFields{
				{
					Name:            loggerArgument,
					LangType:        "bool",
					MapstructureTag: "domain_logger",
					Description:     "Domain logger.",
				},
				{
					Name:            showConfigArgument,
					LangType:        "bool",
					MapstructureTag: "domain_show_config",
					Description:     "Domain show config.",
				},
				{
					Name:            summaryFileArg,
					LangType:        "string",
					MapstructureTag: "domain_summary_file",
					Description:     "Domain summary path.",
				},
			},
		}},
	}

	err := validateProtectedCLIArgumentNames(cmds)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "uses protected argument names [Logger, ShowConfig, SummaryFile]")
	errortest.AssertErrorDescription(t, err, "cannot be used in command_line_commands.functions[].configuration")
	errortest.AssertErrorDescription(t, err, "cli function \"Run\"")
}

func TestValidateProtectedCLIArgumentNamesRejectsResultProtectedConfigurationNames(t *testing.T) {
	t.Parallel()

	cmds := &CommandLineCommands{
		Functions: []CliFunction{{
			Name:        "Run",
			Description: "Run.",
			WithResult:  true,
			Configuration: []FunctionConfigFields{
				{
					Name:            resultOutputArgument,
					LangType:        "string",
					MapstructureTag: "domain_output",
					Description:     "Domain output.",
				},
			},
		}},
	}

	err := validateProtectedCLIArgumentNames(cmds)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "uses protected argument names [OutputFormat]")
}

func TestValidateProtectedCLIArgumentNamesAllowsOutputWhenResultOutputIsDisabled(t *testing.T) {
	t.Parallel()

	cmds := &CommandLineCommands{
		Functions: []CliFunction{{
			Name:        "Run",
			Description: "Run.",
			Configuration: []FunctionConfigFields{
				{
					Name:            resultOutputArgument,
					LangType:        "string",
					MapstructureTag: "domain_output",
					Description:     "Domain output.",
				},
			},
		}},
	}

	require.NoError(t, validateProtectedCLIArgumentNames(cmds))
}

func TestValidateCollectionParametersRejectsCaseInsensitiveDuplicateInterfaceArgumentNames(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{
			Name:        "Build",
			Description: "Build.",
			Args: []CommandArgument{{
				Name:        "Target",
				LangType:    "string",
				Description: "Target.",
			}, {
				Name:        "target",
				LangType:    "string",
				Description: "Target override.",
			}},
		}}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, `interface method "Build" defines duplicate argument names [Target]`)
	errortest.AssertErrorDescription(t, err, "invalid manifest parameter naming")
}

func TestValidateCollectionParametersRejectsDuplicateInterfaceArgumentNamesAcrossCaseAndSeparators(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{
			Name:        "Build",
			Description: "Build.",
			Args: []CommandArgument{{
				Name:        "a_name",
				LangType:    "string",
				Description: "Target.",
			}, {
				Name:        "a-name",
				LangType:    "string",
				Description: "Target override.",
			}},
		}}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, `interface method "Build" defines duplicate argument names [aName]`)
}

func TestValidateCollectionParametersRejectsCaseInsensitiveDuplicateCLIConfigurationNames(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{Name: "Build", Description: "Build."}}},
		CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{{
			Name:        "Run",
			Description: "Run.",
			Configuration: []FunctionConfigFields{{
				Name:            "Mode",
				LangType:        "string",
				MapstructureTag: "mode",
				Description:     "Mode.",
			}, {
				Name:            "mode",
				LangType:        "string",
				MapstructureTag: "mode_2",
				Description:     "Mode override.",
			}},
		}}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, `cli function "Run" defines duplicate configuration names [Mode]`)
}

func TestValidateCollectionParametersRejectsInvalidMainModuleSegment(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		MainModule: "../verify",
		Interface:  Interface{Methods: []InterfaceMethod{{Name: "Build", Description: "Build."}}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "main module \"../verify\" must be a single package segment")
	errortest.AssertErrorDescription(t, err, "invalid manifest parameter shape")
}

func TestValidateCollectionParametersRejectsCaseInsensitiveDuplicateOptionFunctionNames(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{Name: "Build", Description: "Build."}}},
		Options: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{
				FunctionName: field.ToOptionalString("EnableTracing"),
				Parameters: []OptionParameter{{
					Name:        "Tracing",
					LangType:    "bool",
					Description: "Tracing.",
				}},
			}, {
				FunctionName: field.ToOptionalString("enableTracing"),
				Parameters: []OptionParameter{{
					Name:        "Verbose",
					LangType:    "bool",
					Description: "Verbose.",
				}},
			}},
		}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, `option set "Build" defines duplicate function names [EnableTracing]`)
}

func TestValidateParameterConditionReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		names       []string
		sourceName  string
		condition   string
		wantErrText string
	}{
		{name: "empty condition", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: ""},
		{name: "valid single reference", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: "UseWindow == true"},
		{name: "does not start with identifier", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: "(UseWindow == true)", wantErrText: "must start with the name of another parameter"},
		{name: "self reference", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: "WindowPreset != \"\"", wantErrText: "must reference another parameter"},
		{name: "unknown reference", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: "Missing == true", wantErrText: "starts with unknown parameter \"Missing\""},
		{name: "multiple parameters", names: []string{"UseWindow", "Verbose", "WindowPreset"}, sourceName: "WindowPreset", condition: "UseWindow == true && Verbose == false", wantErrText: "must depend on exactly one other parameter"},
		{name: "same parameter mentioned twice", names: []string{"UseWindow", "WindowPreset"}, sourceName: "WindowPreset", condition: "UseWindow == true || UseWindow == false", wantErrText: "must mention parameter \"UseWindow\" exactly once"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateParameterConditionReference(test.names, test.sourceName, test.condition)
			if test.wantErrText == "" {
				require.NoError(t, err)
				return
			}
			errortest.AssertError(t, err, commonerrors.ErrInvalid)
			errortest.AssertErrorDescription(t, err, test.wantErrText)
		})
	}
}

func TestValidateParameterCollectionConditionReferences(t *testing.T) {
	t.Parallel()

	parameters := []*OptionParameter{{
		Name:        "UseWindow",
		LangType:    "bool",
		Description: "Use a window preset.",
	}, {
		Name:        "Verbose",
		LangType:    "bool",
		Description: "Verbose output.",
	}, {
		Name:        "WindowPreset",
		LangType:    "string",
		Description: "Named window preset.",
		Condition:   "UseWindow == true",
	}}
	require.NoError(t, validateParameterCollectionConditionReferences(parameters))

	parameters[1].Condition = "UseWindow == true && Verbose == false"
	err := validateParameterCollectionConditionReferences(parameters)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "must depend on exactly one other parameter")
}

func TestValidateCollectionParametersRejectsConditionThatDoesNotStartWithFieldName(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{Name: "Build", Description: "Build."}}},
		Options: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{Parameters: []OptionParameter{{
				Name:        "UseWindow",
				LangType:    "bool",
				Description: "Use a window preset.",
			}}}, {Parameters: []OptionParameter{{
				Name:        "WindowPreset",
				LangType:    "string",
				Description: "Named window preset.",
				Condition:   "(UseWindow == true)",
			}}}},
		}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "must start with the name of another parameter")
}

func TestValidateCollectionParametersRejectsConditionReferencingMultipleFields(t *testing.T) {
	t.Parallel()

	sdk := &SDK{Specification: SDKSpecification{
		Interface: Interface{Methods: []InterfaceMethod{{Name: "Build", Description: "Build."}}},
		Options: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{Parameters: []OptionParameter{{
				Name:        "UseWindow",
				LangType:    "bool",
				Description: "Use a window preset.",
			}}}, {Parameters: []OptionParameter{{
				Name:        "Verbose",
				LangType:    "bool",
				Description: "Verbose output.",
			}}}, {Parameters: []OptionParameter{{
				Name:        "WindowPreset",
				LangType:    "string",
				Description: "Named window preset.",
				Condition:   "UseWindow == true && Verbose == false",
			}}}},
		}},
	}}

	err := validateCollectionParameters(sdk)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	errortest.AssertErrorDescription(t, err, "must depend on exactly one other parameter")
}
