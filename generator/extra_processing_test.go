/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

const (
	testVerifyFunctionName = "Verify"

	testInitFunctionName = "Init"

	testStringType = stringLangType
)

func TestProcessExtraImports(t *testing.T) {

	t.Run("happy", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				ExtraImports: ExtraImports{

					{

						ImportRef: field.ToOptionalString("custom"),

						ImportUrl: field.ToOptionalString("example.com/acme/custom"),
					},

					{

						ImportRef: field.ToOptionalString("xdg"),

						ImportUrl: field.ToOptionalString("example.com/acme/xdg"),
					},

					{

						ImportUrl: field.ToOptionalString("example.com/acme/test"),
					},

					{

						ImportRef: field.ToOptionalString("ignored"),
					},
				},
			},
		}

		require.NoError(t, ProcessExtraImports(context.Background(), sdk))

		assert.ElementsMatch(t, sdk.Specification.ExtraImports, []ExtraImport{

			{

				ImportRef: field.ToOptionalString("custom"),

				ImportUrl: field.ToOptionalString("example.com/acme/custom"),
			},

			{

				ImportRef: field.ToOptionalString("xdg"),

				ImportUrl: field.ToOptionalString("example.com/acme/xdg"),
			},

			{

				ImportRef: field.ToOptionalString("filesystem"),

				ImportUrl: field.ToOptionalString(filesystemImportPath),
			},

			{

				ImportRef: field.ToOptionalString("httpUtils"),

				ImportUrl: field.ToOptionalString(httpUtilsImportPath),
			},

			{

				ImportRef: field.ToOptionalString("jsonutils"),

				ImportUrl: field.ToOptionalString(jsonUtilsImportPath),
			},

			{

				ImportRef: field.ToOptionalString("platform"),

				ImportUrl: field.ToOptionalString(platformImportPath),
			},

			{

				ImportRef: field.ToOptionalString("validationUtils"),

				ImportUrl: field.ToOptionalString(validationUtilsImportPath),
			},

			{

				ImportRef: field.ToOptionalString("test"),

				ImportUrl: field.ToOptionalString("example.com/acme/test"),
			},
		})

	})

	t.Run("context cancelled", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		errortest.AssertError(t, ProcessExtraImports(ctx, nil), commonerrors.ErrCancelled)

	})

	t.Run("no spec", func(t *testing.T) {

		errortest.AssertError(t, ProcessExtraImports(context.Background(), nil), commonerrors.ErrUndefined)

	})

}

func TestProcessRepositoryAndModuleMetadata(t *testing.T) {

	t.Run("happy", func(t *testing.T) {

		sdk := &SDK{
			Specification: SDKSpecification{
				MainModule:  "verify",
				PackageName: "verify-cmsis-example-sdk",
				Interface: Interface{
					Namespace: "https://github.com/ARM-software",
				},

				RepositoryUrl: field.ToOptionalString("https://github.com/Arm-Examples/verify-cmsis-example-sdk"),
			},
		}

		require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk))

		assert.Equal(t, "github.com/ARM-software/", sdk.Specification.Interface.Namespace)

		assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk", field.OptionalString(sdk.Specification.RepositoryUrl, ""))

	})

	t.Run("defaults repository url", func(t *testing.T) {

		sdk := &SDK{
			Specification: SDKSpecification{
				MainModule:  "verify",
				PackageName: "verify-cmsis-example-sdk",
				Interface:   Interface{},
			},
		}

		require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk))

		assert.Equal(t, "github.com/Arm-Debug/", sdk.Specification.Interface.Namespace)

		assert.Equal(t, "https://github.com/Arm-Debug/verify-cmsis-example-sdk", field.OptionalString(sdk.Specification.RepositoryUrl, ""))

	})

	t.Run("invalid repository url", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				MainModule:    "verify",
				PackageName:   "verify-cmsis-example-sdk",
				Interface:     Interface{},
				RepositoryUrl: field.ToOptionalString("not a url"),
			},
		}

		errortest.AssertError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk), commonerrors.ErrInvalid)
	})

	t.Run("invalid main module", func(t *testing.T) {
		sdk := &SDK{
			Specification: SDKSpecification{
				MainModule:  "../verify",
				PackageName: "verify-cmsis-example-sdk",
				Interface:   Interface{},
			},
		}

		err := ProcessRepositoryAndModuleMetadata(context.Background(), sdk)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		errortest.AssertErrorDescription(t, err, "main module \"../verify\" must be a single package segment")
		errortest.AssertErrorDescription(t, err, "invalid main module metadata")
	})

	t.Run("disables actions when repository is not hosted on github", func(t *testing.T) {

		sdk := &SDK{
			Specification: SDKSpecification{
				MainModule:    "verify",
				PackageName:   "verify-cmsis-example-sdk",
				Interface:     Interface{},
				RepositoryUrl: field.ToOptionalString("https://gitlab.example.com/arm/verify-cmsis-example-sdk"),
				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{

						{Name: "Generate"},

						{Name: "Validate"},
					},
				},
			},
		}

		require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk))

		assert.True(t, sdk.Specification.CommandLineCommands.NoActions)

		for i := range sdk.Specification.CommandLineCommands.Functions {

			assert.True(t, sdk.Specification.CommandLineCommands.Functions[i].NoAction)

		}

	})

	t.Run("context cancelled", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		errortest.AssertError(t, ProcessRepositoryAndModuleMetadata(ctx, nil), commonerrors.ErrCancelled)

	})

	t.Run("no spec", func(t *testing.T) {

		errortest.AssertError(t, ProcessRepositoryAndModuleMetadata(context.Background(), nil), commonerrors.ErrUndefined)

	})

}

func TestProcessBindings(t *testing.T) {

	t.Run("context cancelled", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		errortest.AssertError(t, ProcessBindings(ctx, nil), commonerrors.ErrCancelled)

	})

	t.Run("no spec", func(t *testing.T) {

		errortest.AssertError(t, ProcessBindings(context.Background(), nil), commonerrors.ErrUndefined)

	})

	t.Run("happy", func(t *testing.T) {

		var sdk SDK

		err := faker.FakeData(&sdk,

			options.WithRandomMapAndSliceMinSize(1),

			options.WithRandomMapAndSliceMaxSize(3),

			options.WithMaxFieldDepthOption(2),
		)

		require.NoError(t, err)

		bindingTypes := fmt.Sprintf("types.%v-%v", faker.Word(), faker.Word())

		expectedBinding := Binding{

			Arg: field.ToOptionalString(faker.Name()),

			Description: field.ToOptionalString(faker.Sentence()),

			LangType: field.ToOptionalString(bindingTypes),

			MapstructureTag: field.ToOptionalString(faker.Word()),

			Name: field.ToOptionalString(faker.FirstName()),
		}

		originalValidTypes := sdk.Specification.ExtraValidTypes

		assert.NotContains(t, originalValidTypes, bindingTypes)

		configuration := ConfigurationOption{

			LangType: faker.Word(),

			Bindings: []Binding{expectedBinding},
		}

		sdk.Specification.GlobalConfiguration = append(sdk.Specification.GlobalConfiguration, configuration)

		require.NoError(t, ProcessBindings(context.Background(), &sdk))

		collection.ForEach(originalValidTypes, func(s string) {

			assert.Contains(t, sdk.Specification.ExtraValidTypes, s)

		})

		assert.Contains(t, sdk.Specification.ExtraValidTypes, bindingTypes)

		assert.Contains(t, sdk.Specification.ExtraValidTypes, configuration.LangType)

	})

}

func TestProcessArguments(t *testing.T) {

	t.Run("context cancelled", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		errortest.AssertError(t, ProcessArguments(ctx, nil), commonerrors.ErrCancelled)

	})

	t.Run("no spec", func(t *testing.T) {

		errortest.AssertError(t, ProcessArguments(context.Background(), nil), commonerrors.ErrUndefined)

	})

	t.Run("no CommandLineCommands", func(t *testing.T) {

		sdk := SDK{

			Specification: SDKSpecification{

				PackageName: "pkg",

				MainModule: "main",

				Interface: Interface{},
			},
		}

		assert.NoError(t, ProcessArguments(context.Background(), &sdk))

	})

	t.Run("empty CommandLineCommands.Functions", func(t *testing.T) {

		sdk := SDK{

			Specification: SDKSpecification{

				PackageName: "pkg",

				MainModule: "main",

				Interface: Interface{},

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{},
				},
			},
		}

		assert.NoError(t, ProcessArguments(context.Background(), &sdk))

	})

	t.Run("error from processArguments in bindings", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{

				{

					Bindings: []Binding{{}},

					Description: faker.Sentence(),

					LangType: faker.Word(),

					MapstructureTag: faker.Word(),

					Name: faker.FirstName(),
				},
			},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{cliFunc},
				},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		err := ProcessArguments(context.Background(), sdk)

		errortest.AssertError(t, err, commonerrors.ErrUndefined)

	})

	t.Run("error from processArguments in param", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{

				{

					// No Arg, no MapstructureTag

					Description: faker.Sentence(),

					LangType: faker.Word(),

					MapstructureTag: "", // empty tag, so will error

					Name: faker.FirstName(),
				},
			},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{cliFunc},
				},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		err := ProcessArguments(context.Background(), sdk)

		errortest.AssertError(t, err, commonerrors.ErrUndefined)

	})

	t.Run("duplicate short args triggers error", func(t *testing.T) {

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString(faker.Word()),

			ArgShort: field.ToOptional("x"),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: faker.Word(),

			Name: faker.FirstName(),
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString(faker.Word()),

			ArgShort: field.ToOptional("x"),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: faker.Word(),

			Name: faker.FirstName(),
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: faker.Name(),

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{cliFunc},
				},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		err := ProcessArguments(context.Background(), sdk)

		errortest.AssertError(t, err, commonerrors.ErrConflict)

	})

	t.Run("duplicate long args triggers error", func(t *testing.T) {

		arg := faker.Word()

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString(arg),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: faker.Word(),

			Name: faker.FirstName(),
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString(arg),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: faker.Word(),

			Name: faker.FirstName(),
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: faker.Name(),

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		err := ProcessArguments(context.Background(), sdk)

		errortest.AssertError(t, err, commonerrors.ErrConflict)

	})

	t.Run("duplicate binding long args triggers error", func(t *testing.T) {

		arg := faker.Word()

		cfg := FunctionConfigFields{

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: faker.Word(),

			Name: faker.FirstName(),

			Bindings: []Binding{

				{Arg: field.ToOptionalString(arg), MapstructureTag: field.ToOptionalString(faker.Word())},

				{Arg: field.ToOptionalString(arg), MapstructureTag: field.ToOptionalString(faker.Word())},
			},
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{cfg},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		err := ProcessArguments(context.Background(), sdk)

		errortest.AssertError(t, err, commonerrors.ErrConflict)

	})

	t.Run("h and v short args triggers error", func(t *testing.T) {

		for _, arg := range []string{"h", "v"} {

			t.Run(arg, func(t *testing.T) {

				arg1 := FunctionConfigFields{

					Arg: field.ToOptionalString(faker.Word()),

					ArgShort: field.ToOptional(arg),

					Description: faker.Sentence(),

					LangType: faker.Word(),

					MapstructureTag: faker.Word(),

					Name: faker.FirstName(),
				}

				arg2 := FunctionConfigFields{

					Arg: field.ToOptionalString(faker.Word()),

					ArgShort: field.ToOptional(arg),

					Description: faker.Sentence(),

					LangType: faker.Word(),

					MapstructureTag: faker.Word(),

					Name: faker.FirstName(),
				}

				cliFunc := CliFunction{

					Configuration: []FunctionConfigFields{arg1, arg2},

					Description: faker.Sentence(),

					Name: faker.Name(),
				}

				sdk := &SDK{

					Specification: SDKSpecification{

						CommandLineCommands: &CommandLineCommands{

							Functions: []CliFunction{cliFunc},
						},

						GlobalConfiguration: []ConfigurationOption{},

						Interface: Interface{},

						MainModule: "main",

						PackageName: "pkg",
					},
				}

				err := ProcessArguments(context.Background(), sdk)

				errortest.AssertError(t, err, commonerrors.ErrConflict)

			})

		}

	})

	t.Run("happy path: all good", func(t *testing.T) {

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString("first-flag"),

			ArgShort: field.ToOptional("a"),

			Description: faker.Sentence(),

			LangType: "string",

			MapstructureTag: "first_option",

			Name: "FirstOption",
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString("second-flag"),

			ArgShort: field.ToOptional("b"),

			Description: faker.Sentence(),

			LangType: "string",

			MapstructureTag: "second_option",

			Name: "SecondOption",
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: "Example",

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{cliFunc},
				},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		assert.Len(t, sdk.Specification.CommandLineCommands.Functions[0].Configuration, 2)

		assert.NoError(t, ProcessArguments(context.Background(), sdk))

		assert.Len(t, sdk.Specification.CommandLineCommands.Functions[0].Configuration, 6)

	})

	t.Run("happy path: with calls to Cloud", func(t *testing.T) {

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString("first-flag"),

			ArgShort: field.ToOptional("a"),

			Description: faker.Sentence(),

			LangType: "string",

			MapstructureTag: "first_option",

			Name: "FirstOption",
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString("second-flag"),

			ArgShort: field.ToOptional("b"),

			Description: faker.Sentence(),

			LangType: "string",

			MapstructureTag: "second_option",

			Name: "SecondOption",
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: "CloudCommand",

			CallCloud: true,

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{cliFunc},
				},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		assert.Len(t, sdk.Specification.CommandLineCommands.Functions[0].Configuration, 2)

		assert.NoError(t, ProcessArguments(context.Background(), sdk))

		assert.Len(t, sdk.Specification.CommandLineCommands.Functions[0].Configuration, 9)

	})

	t.Run("injects runtime logging arguments for CLI commands", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("input"),

				Description: "input file",

				LangType: "string",

				MapstructureTag: "input",

				Name: "Input",
			}},

			Description: "generate output",

			Name: "Generate",
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		require.Contains(t, byName, loggerArgument)

		require.Contains(t, byName, loggerFileArg)

		require.Contains(t, byName, annotationArgument)

		assert.Equal(t, []string{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeCustom.String()}, byName[loggerArgument].Enum)

		require.NotNil(t, byName[loggerArgument].DefaultValue)

		assert.Equal(t, defaultCLILoggerMode, *byName[loggerArgument].DefaultValue)

		require.NotNil(t, byName[annotationArgument].DefaultValue)

		assert.Equal(t, defaultCLIAnnotationMode, *byName[annotationArgument].DefaultValue)

		require.NotNil(t, byName[loggerFileArg].Constraints)

		require.NotNil(t, byName[loggerFileArg].Constraints.RequiredWhen)

		assert.Equal(t, loggerArgument, byName[loggerFileArg].Constraints.RequiredWhen.Parameter)

		assert.Equal(t, runtimeLoggerModeFile.String(), byName[loggerFileArg].Constraints.RequiredWhen.Value)

		require.NotNil(t, byName[loggerFileArg].Constraints.ForbiddenWhen)

		assert.Equal(t, ParameterValueConstraintOperatorNotEquals, byName[loggerFileArg].Constraints.ForbiddenWhen.Operator)

	})

	t.Run("injects summary arguments for summary-aware commands", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("input"),

				Description: "input file",

				LangType: "string",

				MapstructureTag: "input",

				Name: "Input",
			}},

			Description: "generate output",

			Name: "Generate",

			WithSummary: field.ToOptionalBool(true),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		configs := sdk.Specification.CommandLineCommands.Functions[0].Configuration

		assert.Len(t, configs, 7)

		assert.ElementsMatch(t, []string{"Input", showConfigArgument, loggerArgument, loggerFileArg, annotationArgument, summaryArgument, summaryFileArg}, []string{configs[0].Name, configs[1].Name, configs[2].Name, configs[3].Name, configs[4].Name, configs[5].Name, configs[6].Name})

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		require.Contains(t, byName, showConfigArgument)

		assert.True(t, byName[showConfigArgument].IsRoot)

		require.Contains(t, byName, loggerArgument)

		require.Contains(t, byName, loggerFileArg)

		require.Contains(t, byName, annotationArgument)

		require.Contains(t, byName, summaryArgument)

		require.Contains(t, byName, summaryFileArg)

		assert.Equal(t, "Write a human-readable summary to the selected destination.", byName[summaryArgument].Description)

		assert.Equal(t, []string{SummaryDestinationFile.String(), SummaryDestinationGithub.String(), SummaryDestinationOutput.String()}, byName[summaryArgument].Enum)

		require.NotNil(t, byName[summaryArgument].DefaultValue)

		assert.Equal(t, SummaryDestinationOutput.String(), *byName[summaryArgument].DefaultValue)

		require.NotNil(t, byName[summaryFileArg].Constraints)

		assert.Equal(t, "filepath", byName[summaryFileArg].LangType)

		assert.Equal(t, []string{summaryArgument}, byName[summaryFileArg].Constraints.DependentRequired)

		require.NotNil(t, byName[summaryFileArg].Constraints.RequiredWhen)

		assert.Equal(t, ParameterValueConstraintOperatorEquals, byName[summaryFileArg].Constraints.RequiredWhen.Operator)

		assert.Equal(t, summaryArgument, byName[summaryFileArg].Constraints.RequiredWhen.Parameter)

		assert.Equal(t, SummaryDestinationFile.String(), byName[summaryFileArg].Constraints.RequiredWhen.Value)

		assert.False(t, byName[summaryFileArg].Required)

		rules := sdk.Specification.CommandLineCommands.Functions[0].ConstraintValidationRules()

		assert.Contains(t, rules, `validationUtils.WhenFieldEqualsValue(&cfg.Summary, GenerateSummaryFile.String(), validationUtils.RequiredFieldsBy(&cfg.SummaryFile))`)

		assert.NotContains(t, rules, `validationUtils.WhenFieldEqualsValue(&cfg.Summary, "file", validationUtils.RequiredFieldsBy(&cfg.SummaryFile))`)

	})

	t.Run("defaults summary arguments when action is generated", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("input"),

				Description: "input file",

				LangType: "string",

				MapstructureTag: "input",

				Name: "Input",
			}},

			Description: "generate output",

			Name: "Generate",
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		require.Contains(t, byName, showConfigArgument)

		require.Contains(t, byName, loggerArgument)

		require.Contains(t, byName, loggerFileArg)

		require.Contains(t, byName, annotationArgument)

		require.Contains(t, byName, summaryArgument)

		require.Contains(t, byName, summaryFileArg)

	})

	t.Run("does not default summary arguments when action is disabled", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("input"),

				Description: "input file",

				LangType: "string",

				MapstructureTag: "input",

				Name: "Input",
			}},

			Description: "generate output",

			Name: "Generate",

			NoAction: true,
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		assert.Contains(t, byName, showConfigArgument)

		assert.Contains(t, byName, loggerArgument)

		assert.Contains(t, byName, loggerFileArg)

		assert.Contains(t, byName, annotationArgument)

		assert.NotContains(t, byName, summaryArgument)

		assert.NotContains(t, byName, summaryFileArg)

	})

	t.Run("injects result output arguments for result-aware commands", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("target"),

				Description: "target",

				LangType: "string",

				MapstructureTag: "target",

				Name: "Target",
			}},

			Description: "Emit command results to the selected destination.",

			Name: "Inspect",

			WithResult: true,

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{NoActions: true, Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		require.Contains(t, byName, resultOutputArgument)

		require.Contains(t, byName, resultOutputFileArgument)

		require.Contains(t, byName, loggerArgument)

		assert.Equal(t, "output_format", byName[resultOutputArgument].MapstructureTag)

		assert.Equal(t, "OutputFormat", byName[resultOutputArgument].Name)

		assert.Equal(t, []string{resultOutputModeHuman.String(), resultOutputModeJSON.String(), resultOutputModeYAML.String()}, byName[resultOutputArgument].Enum)

		require.NotNil(t, byName[resultOutputArgument].DefaultValue)

		assert.Equal(t, defaultCLIResultOutputMode, *byName[resultOutputArgument].DefaultValue)

		require.NotNil(t, byName[loggerArgument].Constraints)

		require.Len(t, byName[loggerArgument].Constraints.RequiredValuesWhen, 2)

		assert.Equal(t, resultOutputFileArgument, byName[loggerArgument].Constraints.RequiredValuesWhen[0].Parameter)

		assert.Equal(t, ParameterRequiredValuesConstraintOperatorAbsent, byName[loggerArgument].Constraints.RequiredValuesWhen[0].Operator)

		assert.ElementsMatch(t, []any{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeCustom.String()}, byName[loggerArgument].Constraints.RequiredValuesWhen[0].RequiredValues)

		assert.Equal(t, ParameterRequiredValuesConstraintOperatorEquals, byName[loggerArgument].Constraints.RequiredValuesWhen[1].Operator)

		assert.Equal(t, stdoutArgumentValue, byName[loggerArgument].Constraints.RequiredValuesWhen[1].Value)

		assert.ElementsMatch(t, []any{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeCustom.String()}, byName[loggerArgument].Constraints.RequiredValuesWhen[1].RequiredValues)

		require.NotNil(t, byName[resultOutputFileArgument].Constraints)

		assert.Equal(t, []string{resultOutputArgument}, byName[resultOutputFileArgument].Constraints.DependentRequired)

		require.NotNil(t, byName[resultOutputFileArgument].Constraints.ForbiddenWhen)

		assert.Equal(t, ParameterValueConstraintOperatorEquals, byName[resultOutputFileArgument].Constraints.ForbiddenWhen.Operator)

		assert.Equal(t, resultOutputArgument, byName[resultOutputFileArgument].Constraints.ForbiddenWhen.Parameter)

		assert.Equal(t, resultOutputModeHuman.String(), byName[resultOutputFileArgument].Constraints.ForbiddenWhen.Value)

		require.NotNil(t, byName[resultOutputFileArgument].Constraints.ParameterValue)

		assert.Equal(t, loggerFileArg, byName[resultOutputFileArgument].Constraints.ParameterValue.Parameter)

		assert.Equal(t, ParameterValueComparisonConstraintOperatorNotEquals, byName[resultOutputFileArgument].Constraints.ParameterValue.Operator)

		assert.Equal(t, ParameterValueComparisonConstraintWhenNotEmpty, byName[resultOutputFileArgument].Constraints.ParameterValue.When)

		assert.NotContains(t, byName, summaryArgument)

		assert.NotContains(t, byName, summaryFileArg)

	})

	t.Run("does not inject result output arguments when with_result is omitted", func(t *testing.T) {

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{{

				Arg: field.ToOptionalString("target"),

				Description: "target",

				LangType: "string",

				MapstructureTag: "target",

				Name: "Target",
			}},

			Description: "Emit command results to the selected destination.",

			Name: "Inspect",

			WithSummary: field.ToOptionalBool(false),
		}

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{NoActions: true, Functions: []CliFunction{cliFunc}},

				GlobalConfiguration: []ConfigurationOption{},

				Interface: Interface{},

				MainModule: "main",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		byName := sdk.Specification.CommandLineCommands.Functions[0].ConfigurationByName()

		assert.NotContains(t, byName, resultOutputArgument)

		assert.NotContains(t, byName, resultOutputFileArgument)

	})

	t.Run("rejects parameters that set both list and iterator", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				Interface: Interface{

					Methods: []InterfaceMethod{

						{

							Name: "Collect",

							Description: "Collect values",

							Args: []CommandArgument{{

								Name: "values",

								LangType: "string",

								Description: "Values to collect",

								List: true,

								Iterator: true,
							}},
						},
					},
				},
			},
		}

		errortest.AssertError(t, ProcessArguments(context.Background(), sdk), commonerrors.ErrInvalid)

	})

	t.Run("rejects explicit variadic lang type combined with list", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				Interface: Interface{

					Methods: []InterfaceMethod{

						{

							Name: "Collect",

							Description: "Collect values",

							Args: []CommandArgument{{

								Name: "values",

								LangType: "...string",

								Description: "Values to collect",

								List: true,
							}},
						},
					},
				},
			},
		}

		errortest.AssertError(t, ProcessArguments(context.Background(), sdk), commonerrors.ErrInvalid)

	})

	t.Run("rejects explicit variadic lang type combined with iterator", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				Interface: Interface{

					Methods: []InterfaceMethod{

						{

							Name: "Collect",

							Description: "Collect values",

							Args: []CommandArgument{{

								Name: "values",

								LangType: "...string",

								Description: "Values to collect",

								Iterator: true,
							}},
						},
					},
				},
			},
		}

		errortest.AssertError(t, ProcessArguments(context.Background(), sdk), commonerrors.ErrInvalid)

	})

	t.Run("rejects enum default value outside allowed list", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				GlobalConfiguration: []ConfigurationOption{

					{

						Name: "Mode",

						LangType: "string",

						MapstructureTag: "mode",

						Description: "Execution mode",

						Enum: []string{"fast", "safe"},

						DefaultValue: field.ToOptionalString("turbo"),
					},
				},
			},
		}

		errortest.AssertError(t, ProcessArguments(context.Background(), sdk), commonerrors.ErrInvalid)

	})

}

func TestPrecomputeArgNamesNoKeyringIsNotMarkedNoLint(t *testing.T) {

	cmds := &CommandLineCommands{

		Functions: []CliFunction{

			{

				Name: "Verify",

				LoadFromKeyring: true,

				Configuration: []FunctionConfigFields{

					{

						Name: noKeyringArgument,

						Description: "Disable loading configuration from the system keyring",

						IsRoot: false,

						LangType: "bool",

						MapstructureTag: toSnakeCase(noKeyringArgument),
					},

					{

						Name: "APIKey",

						Description: "An API key enabling the use of the online services",

						IsRoot: false,

						LangType: "string",

						MapstructureTag: "api_key",
					},
				},
			},
		},
	}

	args, err := PrecomputeArgNames(cmds)

	require.NoError(t, err)

	noKeyring, found := collection.FirstBy(args, func(arg ArgDefinition) bool {

		return arg.Name == "verifyNoKeyringArg"

	})

	require.True(t, found)

	assert.False(t, noKeyring.NoLint)

	apiKey, found := collection.FirstBy(args, func(arg ArgDefinition) bool {

		return arg.Name == "verifyAPIKeyArg"

	})

	require.True(t, found)

	assert.True(t, apiKey.NoLint)

}

func TestPrecomputeArgNamesFromVerifyManifestNoKeyringIsNotMarkedNoLint(t *testing.T) {

	localFs := filesystem.NewStandardFileSystem()

	manifestPath := filesystem.FilePathJoin(localFs, "testdata", "verify.yml")

	content, err := filesystem.ReadFile(manifestPath)

	require.NoError(t, err)

	var sdk SDK

	require.NoError(t, unmarshalManifestContent(filesystem.FilePathExt(localFs, manifestPath), content, &sdk))

	require.NoError(t, ProcessExtraImports(context.Background(), &sdk))

	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), &sdk))

	require.NoError(t, ProcessAdditionalTypePackages(context.Background(), &sdk))

	require.NoError(t, ProcessBindings(context.Background(), &sdk))

	require.NoError(t, ProcessArguments(context.Background(), &sdk))

	args, err := PrecomputeArgNames(sdk.Specification.CommandLineCommands)

	require.NoError(t, err)

	noKeyring, found := collection.FirstBy(args, func(arg ArgDefinition) bool {

		return arg.Name == "verifyNoKeyringArg"

	})

	require.True(t, found)

	assert.False(t, noKeyring.NoLint)

}

func TestPrecomputeFlagsAndBindings(t *testing.T) {

	t.Run("duplicate is_root results in only one root command", func(t *testing.T) {

		name := faker.Word()

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			IsRoot: true,

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			IsRoot: true,

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		flagRegistrations, _, err := PreComputeRootFlagsAndBindings(&CommandLineCommands{

			Functions: []CliFunction{cliFunc},
		})

		assert.NoError(t, err)

		assert.Len(t, flagRegistrations, 1)

		assert.Equal(t, fmt.Sprintf("root%vArg", name), flagRegistrations[0].ArgVar)

	})

	t.Run("duplicate normal results in extra binding", func(t *testing.T) {

		name := faker.Word()

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		_, bindingInfo, err := PreComputeFlagsAndBindings(&CommandLineCommands{

			Functions: []CliFunction{cliFunc},
		})

		assert.NoError(t, err)

		require.Len(t, bindingInfo, 1)

		assert.Len(t, bindingInfo[0].FlagBindings, 2)

	})

	t.Run("duplicate is_root results in only one binding", func(t *testing.T) {

		name := faker.Word()

		arg1 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			IsRoot: true,

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		arg2 := FunctionConfigFields{

			Arg: field.ToOptionalString(name),

			IsRoot: true,

			Description: faker.Sentence(),

			LangType: faker.Word(),

			MapstructureTag: name,

			Name: name,
		}

		cliFunc := CliFunction{

			Configuration: []FunctionConfigFields{arg1, arg2},

			Description: faker.Sentence(),

			Name: faker.Name(),
		}

		_, bindingInfo, err := PreComputeRootFlagsAndBindings(&CommandLineCommands{

			Functions: []CliFunction{cliFunc},
		})

		assert.NoError(t, err)

		assert.Len(t, bindingInfo, 1)

		assert.Len(t, bindingInfo[0].FlagBindings, 1)

	})

}

func TestPreComputeFlagConstraintRegistrations(t *testing.T) {

	t.Run("builds required together and mutually exclusive registrations", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{{

				Name: "Verify",

				Description: "verify",

				Configuration: []FunctionConfigFields{

					{

						Name: "Profile",

						LangType: "string",

						MapstructureTag: "profile",

						Description: "profile",

						IsRoot: true,

						Constraints: &ParameterConstraints{

							DependentRequired: []string{"Output"},

							RequireTogetherAsCliFlags: true,
						},
					},

					{

						Name: "Mode",

						LangType: "string",

						MapstructureTag: "mode",

						Description: "mode",

						Constraints: &ParameterConstraints{

							ConflictsWith: []string{"Output"},
						},
					},

					{

						Name: "Output",

						LangType: "string",

						MapstructureTag: "output",

						Description: "output",

						Constraints: &ParameterConstraints{

							RequireTogetherAsCliFlags: true,
						},
					},
				},
			}},
		}

		registrations, err := PreComputeFlagConstraintRegistrations(cmds)

		require.NoError(t, err)

		assert.Equal(t, []FlagConstraintRegistration{

			{CommandVar: "verifyCmd", Method: "MarkFlagsMutuallyExclusive", FlagArgs: []string{"verifyModeArg", "verifyOutputArg"}},

			{CommandVar: "verifyCmd", Method: "MarkFlagsRequiredTogether", FlagArgs: []string{"rootProfileArg", "verifyOutputArg"}},
		}, registrations)

	})

	t.Run("rejects constraints on non-standalone flags", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{{

				Name: "Verify",

				Description: "verify",

				Configuration: []FunctionConfigFields{

					{

						Name: "Project",

						LangType: "string",

						MapstructureTag: "project",

						Description: "project",

						Bindings: []Binding{{

							Name: field.ToOptionalString("Path"),

							LangType: field.ToOptionalString("string"),

							MapstructureTag: field.ToOptionalString("path"),

							Description: field.ToOptionalString("path"),
						}},

						Constraints: &ParameterConstraints{

							DependentRequired: []string{"Output"},

							RequireTogetherAsCliFlags: true,
						},
					},

					{

						Name: "Output",

						LangType: "string",

						MapstructureTag: "output",

						Description: "output",

						Constraints: &ParameterConstraints{

							RequireTogetherAsCliFlags: true,
						},
					},
				},
			}},
		}

		_, err := PreComputeFlagConstraintRegistrations(cmds)

		errortest.AssertError(t, err, commonerrors.ErrInvalid)

	})

	t.Run("dependent required only generates required together when all referenced parameters opt in as CLI flags", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{{

				Name: "Verify",

				Description: "verify",

				Configuration: []FunctionConfigFields{

					{

						Name: "Profile",

						LangType: "string",

						MapstructureTag: "profile",

						Description: "profile",

						Constraints: &ParameterConstraints{

							DependentRequired: []string{"Output", "Manifest"},

							RequireTogetherAsCliFlags: true,
						},
					},

					{

						Name: "Output",

						LangType: "string",

						MapstructureTag: "output",

						Description: "output",

						Constraints: &ParameterConstraints{RequireTogetherAsCliFlags: true},
					},

					{

						Name: "Manifest",

						LangType: "string",

						MapstructureTag: "manifest",

						Description: "manifest",
					},
				},
			}},
		}

		registrations, err := PreComputeFlagConstraintRegistrations(cmds)

		require.NoError(t, err)

		assert.Equal(t, []FlagConstraintRegistration{

			{CommandVar: "verifyCmd", Method: "MarkFlagsRequiredTogether", FlagArgs: []string{"verifyProfileArg", "verifyOutputArg"}},
		}, registrations)

	})

}

func TestFlagConstraintRegistrationStatement(t *testing.T) {

	registration := FlagConstraintRegistration{

		CommandVar: "verifyCmd",

		Method: "MarkFlagsRequiredTogether",

		FlagArgs: []string{"verifyProfileArg", "verifyOutputArg"},
	}

	assert.Equal(t, `verifyCmd.MarkFlagsRequiredTogether(verifyProfileArg, verifyOutputArg)`, registration.Statement())

}

func TestConstraintValidationRuleLines(t *testing.T) {

	fct := &CliFunction{

		Name: "Verify",

		Description: "verify",

		Configuration: []FunctionConfigFields{

			{

				Name: "Mode",

				LangType: "string",

				MapstructureTag: "mode",

				Description: "mode",
			},

			{

				Name: "Output",

				LangType: "string",

				MapstructureTag: "output",

				Description: "output",

				Constraints: &ParameterConstraints{

					RequiredWhen: &ParameterValueConstraint{Parameter: "Mode", Operator: ParameterValueConstraintOperatorEquals, Value: "strict"},
				},
			},

			{

				Name: "Manifest",

				LangType: "string",

				MapstructureTag: "manifest",

				Description: "manifest",

				Constraints: &ParameterConstraints{

					ForbiddenWhen: &ParameterValueConstraint{Parameter: "Mode", Operator: ParameterValueConstraintOperatorIn, Values: []interface{}{"relaxed", "archive"}},
				},
			},

			{

				Name: "Profile",

				LangType: "string",

				MapstructureTag: "profile",

				Description: "profile",

				Constraints: &ParameterConstraints{

					DependentRequired: []string{"Output"},

					ConflictsWith: []string{"Manifest"},
				},
			},
		},
	}

	rules := fct.ConstraintValidationRules()

	assert.Contains(t, rules, `validationUtils.FieldDependencyBy(map[any][]any{&cfg.Profile: []any{&cfg.Output}})`)

	assert.Contains(t, rules, `validationUtils.MutuallyExclusiveFieldsBy(&cfg.Profile, &cfg.Manifest)`)

	assert.Contains(t, rules, `validationUtils.WhenFieldEqualsValue(&cfg.Mode, "strict", validationUtils.RequiredFieldsBy(&cfg.Output))`)

	assert.Contains(t, rules, `validationUtils.WhenFieldInValues(&cfg.Mode, []any{"relaxed", "archive"}, validationUtils.ForbiddenFieldsBy(&cfg.Manifest))`)

}

func TestBuildFlagLine(t *testing.T) {

	t.Run("normal flag", func(t *testing.T) {

		result := buildFlagLine(true, false, "someCmd", "Bool", "someArg", "a", "false", "normal flag", "")

		expected := `rootCmd.PersistentFlags().BoolP(someArg, "a", false, "normal flag")`

		assert.Equal(t, expected, result)

	})

	t.Run("hidden flag", func(t *testing.T) {

		result := buildFlagLine(true, true, "someCmd", "Bool", "someArg", "a", "false", "normal flag", "")

		expected := `rootCmd.PersistentFlags().BoolP(someArg, "a", false, "normal flag")

_ = rootCmd.PersistentFlags().MarkHidden(someArg)`

		assert.Equal(t, expected, result)

	})

	t.Run("deprecated flag", func(t *testing.T) {

		result := buildFlagLine(false, false, "someCmd", "String", "someArg", "", `""`, "normal flag", "use replacement instead; deprecated since 2026-07-24T00:00:00Z")

		expected := `someCmd.Flags().String(someArg, "", "normal flag")

_ = someCmd.Flags().MarkDeprecated(someArg, "use replacement instead; deprecated since 2026-07-24T00:00:00Z")`

		assert.Equal(t, expected, result)

	})

}

func TestCliFlagTypeAndZeroValue(t *testing.T) {

	t.Run("built-in type keeps its cobra flag type", func(t *testing.T) {

		flagType, zeroVal := cliFlagTypeAndZeroValue("[]string")

		assert.Equal(t, "StringSlice", flagType)

		assert.Equal(t, "[]string{}", zeroVal)

	})

	t.Run("custom type falls back to string flag", func(t *testing.T) {

		flagType, zeroVal := cliFlagTypeAndZeroValue("models.FileEntry")

		assert.Equal(t, "String", flagType)

		assert.Equal(t, `""`, zeroVal)

	})

}

func TestCliRootLevelCallCloud(t *testing.T) {

	t.Run("nil CommandLineCommands does not panic", func(t *testing.T) {

		cmds := &CommandLineCommands{}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

	})

	t.Run("functions is unchanged when callCloud is not set at both root and function levels", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{

				{Name: "func", Description: "test", Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		assert.Empty(t, cmds.Functions[0].Configuration)

	})

	t.Run("callCloud is non-root when root-level callCloud and some function-level callCloud are not set to true", func(t *testing.T) {

		cmds := &CommandLineCommands{

			CallCloud: field.ToOptionalBool(false),

			Functions: []CliFunction{

				{Name: "someFunc", Description: "test", CallCloud: true, Configuration: []FunctionConfigFields{}},

				{Name: "anotherFunc", Description: "test", Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		// Verify the first function has non-root configuration

		for _, cfg := range cmds.Functions[0].Configuration {

			if cfg.Name == "APIKey" || cfg.Name == "Host" {

				assert.False(t, cfg.IsRoot, "%s should be non-root", cfg.Name)

			}

		}

		// Verify the second function has no configuration

		assert.Empty(t, cmds.Functions[1].Configuration)

	})

	t.Run("mixed function-level callCloud with explicit root auth is rejected", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{

				{

					Name: "credentials",

					Description: "test",

					Configuration: []FunctionConfigFields{

						{

							Description: "API key",

							IsRoot: true,

							LangType: "string",

							MapstructureTag: "api_key",

							Name: "APIKey",

							Required: true,
						},

						{

							Description: "Host",

							IsRoot: true,

							LangType: "host",

							MapstructureTag: "host",

							Name: "Host",

							Required: true,
						},
					},
				},

				{Name: "discover", Description: "test", CallCloud: true, Configuration: []FunctionConfigFields{}},

				{Name: "local", Description: "test", Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		errortest.AssertError(t, err, commonerrors.ErrInvalid)

	})

	t.Run("callCloud is promoted to root level when all functions have callCloud set to true", func(t *testing.T) {

		cmds := &CommandLineCommands{

			Functions: []CliFunction{

				{Name: "someFunc", Description: "test", CallCloud: true, Configuration: []FunctionConfigFields{}},

				{Name: "anotherFunc", Description: "test", CallCloud: true, Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		assertRootCallCloudConfiguration(t, cmds.Functions)

	})

	t.Run("callCloud is promoted to root level when all functions have either callCloud or storeToKeyring set to true", func(t *testing.T) {

		cmds := &CommandLineCommands{

			CallCloud: field.ToOptionalBool(true),

			Functions: []CliFunction{

				{Name: "someFunc", Description: "test", CallCloud: true, Configuration: []FunctionConfigFields{}},

				{Name: "anotherFunc", Description: "test", StoreToKeyring: true, Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		// Verify the first function has root configuration

		for _, cfg := range cmds.Functions[0].Configuration {

			if cfg.Name == "APIKey" || cfg.Name == "Host" {

				assert.True(t, cfg.IsRoot, "%s should be root", cfg.Name)

			}

		}

		// Verify the second function has no configuration

		assert.Empty(t, cmds.Functions[1].Configuration)

	})

	t.Run("root callCloud applies to all functions", func(t *testing.T) {

		cmds := &CommandLineCommands{

			CallCloud: field.ToOptionalBool(true),

			Functions: []CliFunction{

				{Name: "someFunc", Description: "test", Configuration: []FunctionConfigFields{}},

				{Name: "anotherFunc", Description: "test", Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		assertRootCallCloudConfiguration(t, cmds.Functions)

	})

	t.Run("root callCloud sets LoadFromKeyring", func(t *testing.T) {

		cmds := &CommandLineCommands{

			CallCloud: field.ToOptionalBool(true),

			Functions: []CliFunction{

				{Name: "func", Description: "test", Configuration: []FunctionConfigFields{}},
			},
		}

		err := processCallCloud(cmds)

		assert.NoError(t, err)

		assert.True(t, cmds.Functions[0].LoadFromKeyring)

	})

	t.Run("creates credentials command from secret fields and secret bindings", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{

						{

							Name: testVerifyFunctionName,

							Description: "verify",

							Configuration: []FunctionConfigFields{

								{

									Name: "APIKey",

									LangType: testStringType,

									MapstructureTag: "api_key",

									Description: "API key",

									Arg: field.ToOptionalString("api-key"),

									Required: true,

									Secret: true,
								},

								{

									Name: "ServiceCredentials",

									LangType: "types.ServiceCredentials",

									MapstructureTag: "service_credentials",

									Description: "Service credentials",

									DefaultValue: field.ToOptionalString("types.ServiceCredentials{}"),

									Bindings: []Binding{

										{

											Name: field.ToOptionalString("Token"),

											Arg: field.ToOptionalString("service-token"),

											MapstructureTag: field.ToOptionalString("token"),

											LangType: field.ToOptionalString(testStringType),

											Description: field.ToOptionalString("Service token"),

											Secret: true,
										},
									},
								},
							},
						},
					},
				},

				Interface: Interface{},

				MainModule: "verify",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		credentialsIndex := slices.IndexFunc(sdk.Specification.CommandLineCommands.Functions, func(f CliFunction) bool { return f.Name == credentialsCommandName })

		require.GreaterOrEqual(t, credentialsIndex, 0)

		credentials := sdk.Specification.CommandLineCommands.Functions[credentialsIndex]

		assert.True(t, credentials.StoreToKeyring)

		assert.True(t, credentials.NoAction)

		byName := credentials.ConfigurationByName()

		require.Contains(t, byName, showConfigArgument)

		require.Contains(t, byName, "APIKey")

		require.Contains(t, byName, "ServiceCredentials")

		require.Contains(t, byName, loggerArgument)

		require.Contains(t, byName, loggerFileArg)

		require.Contains(t, byName, annotationArgument)

		require.Len(t, byName["ServiceCredentials"].Bindings, 1)

		assert.Equal(t, "Token", field.OptionalString(byName["ServiceCredentials"].Bindings[0].Name, ""))

	})

	t.Run("rejects secret parameters that are not string compatible", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{

						{

							Name: testVerifyFunctionName,

							Description: "verify",

							Configuration: []FunctionConfigFields{

								{

									Name: "SecretPort",

									LangType: "port",

									MapstructureTag: "secret_port",

									Description: "Secret port",

									Secret: true,
								},
							},
						},
					},
				},

				Interface: Interface{},

				MainModule: "verify",

				PackageName: "pkg",
			},
		}

		errortest.AssertError(t, ProcessArguments(context.Background(), sdk), commonerrors.ErrInvalid)

	})

	t.Run("accepts optional secret strings", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{

						{

							Name: testVerifyFunctionName,

							Description: "verify",

							Configuration: []FunctionConfigFields{

								{

									Name: "OptionalSecret",

									LangType: testStringType,

									MapstructureTag: "optional_secret",

									Description: "Optional secret",

									Required: false,

									Secret: true,
								},
							},
						},
					},
				},

				Interface: Interface{},

				MainModule: "verify",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

	})

	t.Run("merges secrets into existing credentials command", func(t *testing.T) {

		sdk := &SDK{

			Specification: SDKSpecification{

				CommandLineCommands: &CommandLineCommands{

					Functions: []CliFunction{

						{

							Name: testInitFunctionName,

							Description: "init",

							Configuration: []FunctionConfigFields{

								{

									Name: "AgentAPIToken",

									LangType: testStringType,

									MapstructureTag: "agent_api_token",

									Description: "Agent API token",

									Arg: field.ToOptionalString("agent-api-token"),

									Secret: true,
								},
							},
						},

						{

							Name: credentialsCommandName,

							Description: "store credentials",

							StoreToKeyring: true,

							Configuration: []FunctionConfigFields{

								{

									Name: "AFAASToken",

									LangType: testStringType,

									MapstructureTag: "afaas_service_token",

									Description: "AFaaS token",

									Arg: field.ToOptionalString("afaas-service-token"),

									Secret: true,
								},
							},
						},
					},
				},

				Interface: Interface{},

				MainModule: "agent",

				PackageName: "pkg",
			},
		}

		require.NoError(t, ProcessArguments(context.Background(), sdk))

		credentialsIndex := slices.IndexFunc(sdk.Specification.CommandLineCommands.Functions, func(f CliFunction) bool { return f.Name == credentialsCommandName })

		require.GreaterOrEqual(t, credentialsIndex, 0)

		credentials := sdk.Specification.CommandLineCommands.Functions[credentialsIndex]

		assert.True(t, credentials.StoreToKeyring)

		assert.ElementsMatch(t, []string{"AFAASToken", "AgentAPIToken"}, []string{credentials.Configuration[0].Name, credentials.Configuration[1].Name})

	})

}
