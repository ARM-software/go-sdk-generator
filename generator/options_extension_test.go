/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestNormaliseOptionPackageName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, defaultOptionPackage, normaliseOptionPackageName(""))
	assert.Equal(t, "clientoptions", normaliseOptionPackageName("Client-Options"))
	assert.Equal(t, "runtime", normaliseOptionPackageName("Runtime"))
}

func TestValidateOptionPackageName(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateOptionPackageName("options"))
	require.Error(t, validateOptionPackageName("foo/bar"))
	require.Error(t, validateOptionPackageName("foo\\bar"))
	require.Error(t, validateOptionPackageName("foo.bar"))
}

func TestProcessOptionPackagesNormalisesPackage(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{Options: OptionSets{{
		Name:        "Build",
		Package:     "Client-Options",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Active", LangType: "bool", Description: "Active."}},
		}},
	}}}}

	require.NoError(t, ProcessOptionPackages(context.Background(), sdk))
	assert.Equal(t, "clientoptions", sdk.Specification.Options[0].Package)
}

func TestProcessOptionPackagesRejectsGeneratedHelperCollisions(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{Options: OptionSets{{
		Name:        "Build",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Timeout", LangType: "string", Description: "Timeout."}},
		}},
	}, {
		Name:        "Run",
		Description: "Run options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Timeout", LangType: "string", Description: "Timeout."}},
		}},
	}}}}

	err := ProcessOptionPackages(context.Background(), sdk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `generated option identifier "WithTimeout"`)
}

func TestProcessOptionPackagesRejectsPackageCollisionWithAdditionalTypes(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{
		AdditionalTypes: Types{{
			Name:        "BuildMetadata",
			Package:     "options",
			Description: "Build metadata.",
			Fields:      []EntityField{{Name: "ID", LangType: "string", Description: "ID."}},
		}},
		Options: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{
				Parameters: []OptionParameter{{Name: "Active", LangType: "bool", Description: "Active."}},
			}},
		}},
	}}

	err := ProcessOptionPackages(context.Background(), sdk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `option package "options" conflicts with an additional type package`)
}

func TestOptionHelpers(t *testing.T) {
	t.Parallel()
	data := &DataModel{CurrentOptionSets: OptionSets{{Name: "Generate", Functions: []OptionFunction{{Parameters: []OptionParameter{{
		Name: "LLM", LangType: "protectedConnection", Description: "Connect to an LLM service.",
		DefaultValue: field.ToOptionalString("https://openai-api-proxy.geo.arm.com/api/providers/openai/v1"),
		ValueField:   &OptionValueField{Name: field.ToOptionalString("Connection"), MapstructureTag: field.ToOptionalString("connection")},
	}}}}}}}
	optionSet := &data.CurrentOptionSets[0]
	withLLM := &optionSet.Functions[0]
	llm := &withLLM.Parameters[0]
	assert.Equal(t, "GenerateOptionsLLMOption", optionParameterWrapperTypeName("GenerateOptions", llm))
	assert.Equal(t, "Connection", optionParameterValueFieldName(llm))
	assert.Equal(t, "connection", optionParameterValueFieldMapstructureTag(llm))
	assert.Equal(t, "llm", optionParameterMapstructureTag(llm))
	assert.Equal(t, "opt.LLM.Connection", optionParameterValueSelector(llm, "opt"))
	assert.Equal(t, "opt.LLM.Enabled", optionParameterEnabledSelector(llm, "opt"))
	defaultValueField := &OptionParameter{Name: "PageToken", LangType: "string"}
	assert.Equal(t, "PageToken", optionParameterValueFieldName(defaultValueField))
	assert.Equal(t, "page_token", optionParameterValueFieldMapstructureTag(defaultValueField))
	assert.Equal(t, "opt.PageToken.PageToken", optionParameterValueSelector(defaultValueField, "opt"))
	namedValueField := &OptionParameter{Name: "PageToken", LangType: "string", ValueField: &OptionValueField{Name: field.ToOptionalString("Token")}}
	assert.Equal(t, "Token", optionParameterValueFieldName(namedValueField))
	assert.Equal(t, "token", optionParameterValueFieldMapstructureTag(namedValueField))
	assert.Equal(t, "WithLLM", optionFunctionName(withLLM))
	assert.Equal(t, "value httpUtils.RequestConfiguration", optionFunctionSignature(optionSet, withLLM))
	assert.Equal(t, "opt.LLM.Connection = value\n\t\topt.LLM.Enabled = true", optionFunctionAssignments(withLLM))

	boolean := &OptionFunction{Parameters: []OptionParameter{{Name: "Verbose", LangType: "bool"}}}
	assert.True(t, optionFunctionUsesBooleanShortcut(boolean))
	assert.Equal(t, "WithVerbose", optionFunctionName(boolean))
	assert.Equal(t, "WithoutVerbose", optionWithoutFunctionName(boolean))
	assert.Equal(t, "true", optionBooleanParameterValueExpression("GenerateOptions", boolean.singleParameter(), true))
	assert.Equal(t, "false", optionBooleanParameterValueExpression("GenerateOptions", boolean.singleParameter(), false))

	tags := &OptionParameter{Name: "Tags", LangType: "string", List: true}
	assert.Equal(t, "clone.Tags.Tags = slices.Clone(opt.Tags.Tags)", optionParameterCloneStatements(tags))
	window := &OptionFunction{FunctionName: field.ToOptionalString("WithWindow"), Parameters: []OptionParameter{{Name: "Width", LangType: "int"}, {Name: "Height", LangType: "int"}}}
	assert.Equal(t, "width int, height int", optionFunctionSignature(&OptionSet{Name: "Generate"}, window))
	assert.Equal(t, "opt.Width.Width = width\n\t\topt.Width.Enabled = true\n\t\topt.Height.Height = height\n\t\topt.Height.Enabled = true", optionFunctionAssignments(window))
}

func TestOptionConditionHelpers(t *testing.T) {
	t.Parallel()
	optionSet := &OptionSet{Name: "Search", Functions: []OptionFunction{{Parameters: []OptionParameter{{
		Name: "UsePaging", LangType: "bool", Description: "Whether paging should be applied.",
	}}}, {Parameters: []OptionParameter{{
		Name: "PageToken", LangType: "string", Description: "Continuation token.",
		Condition:  "UsePaging == true",
		ValueField: &OptionValueField{Name: field.ToOptionalString("Token")},
	}}}}}

	assert.True(t, optionSetHasConditionalValidationRules(optionSet))
	assert.Equal(t, "opt.UsePaging.Enabled && opt.UsePaging.UsePaging == true", optionParameterConditionExpression(optionSet, "UsePaging == true", "opt"))
	assert.Equal(t, "opt.PageToken.Enabled && opt.PageToken.Token != \"\"", optionParameterConditionExpression(&OptionSet{Name: "Search", Functions: []OptionFunction{{Parameters: []OptionParameter{{
		Name: "PageToken", LangType: "string", Description: "Continuation token.",
		Condition:  "PageToken != \"\"",
		ValueField: &OptionValueField{Name: field.ToOptionalString("Token")},
	}}}}}, "PageToken != \"\"", "opt"))
	assert.Equal(t, "opt.Unknown == true", optionParameterConditionExpression(optionSet, "Unknown == true", "opt"))
	assert.False(t, optionSetHasConditionalValidationRules(&OptionSet{Name: "Search", Functions: []OptionFunction{{Parameters: []OptionParameter{{
		Name: "UsePaging", LangType: "bool", Description: "Whether paging should be applied.",
	}}}}}))
}

func TestOptionCLIEnabledBinding(t *testing.T) {
	t.Parallel()
	option := OptionParameter{
		Name: "LLM", LangType: "protectedConnection", Description: "Connect to an LLM service.",
		DefaultValue: field.ToOptionalString("https://openai-api-proxy.geo.arm.com/api/providers/openai/v1"),
	}
	sdk := &SDK{Specification: SDKSpecification{
		Options: OptionSets{{Name: "Refine", Description: "Refinement options.", Functions: []OptionFunction{{Parameters: []OptionParameter{option}}}}},
		CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{{
			Name: "Refine", Description: "Refine document tags.",
			Configuration: []FunctionConfigFields{{
				Name: "LLM", LangType: "options.RefineOptionsLLMOption", MapstructureTag: "llm", Description: "LLM refinement.",
			}},
		}}},
	}}

	require.NoError(t, ProcessArguments(context.Background(), sdk))
	var cfg FunctionConfigFields
	for _, candidate := range sdk.Specification.CommandLineCommands.Functions[0].Configuration {
		if candidate.Name == "LLM" {
			cfg = candidate
			break
		}
	}
	require.NotNil(t, cfg.DefaultValue)
	assert.Equal(t, "options.DefaultRefineOptions().LLM", *cfg.DefaultValue)
	require.Len(t, cfg.Bindings, 1)
	assert.Equal(t, "Enabled", *cfg.Bindings[0].Name)
	assert.Equal(t, "enabled", *cfg.Bindings[0].MapstructureTag)
	args, err := PrecomputeArgNames(sdk.Specification.CommandLineCommands)
	require.NoError(t, err)
	assert.Contains(t, args, ArgDefinition{Name: "refineLLMEnabledArg", Value: "llm"})
	flags, envBindings, err := PreComputeFlagsAndBindings(sdk.Specification.CommandLineCommands)
	require.NoError(t, err)
	assert.Contains(t, flags[0].FlagLine, "refineCmd.Flags().Bool(refineLLMEnabledArg, false,")
	assert.Contains(t, envBindings, ArgFlagBindingInfo{EnvVar: "llm_enabled", FlagBindings: []string{"refineCmd.Flags().Lookup(refineLLMEnabledArg)"}})

	// The Enabled binding's arg can be overridden without renaming the field.
	for i := range sdk.Specification.CommandLineCommands.Functions[0].Configuration {
		if sdk.Specification.CommandLineCommands.Functions[0].Configuration[i].Name == "LLM" {
			sdk.Specification.CommandLineCommands.Functions[0].Configuration[i].Bindings[0].Arg = field.ToOptionalString("refine-with-llm")
		}
	}
	args, err = PrecomputeArgNames(sdk.Specification.CommandLineCommands)
	require.NoError(t, err)
	assert.Contains(t, args, ArgDefinition{Name: "refineLLMEnabledArg", Value: "refine-with-llm"})
}

func TestOptionCLIRegressionManifestHasDefaultAndRenamedEnabledFlags(t *testing.T) {
	t.Parallel()
	content, err := filesystem.ReadFile(filesystem.FilePathJoin(filesystem.NewStandardFileSystem(), "testdata", "petshop-purchase.yml"))
	require.NoError(t, err)
	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", content, &sdk))
	require.NoError(t, ProcessArguments(context.Background(), &sdk))

	args, err := PrecomputeArgNames(sdk.Specification.CommandLineCommands)
	require.NoError(t, err)
	flagNames := make(map[string]string, len(args))
	for _, arg := range args {
		flagNames[arg.Name] = arg.Value
	}
	assert.Equal(t, "page-token", flagNames["searchProductsPageTokenOptionEnabledArg"])
	assert.Equal(t, "use-paging", flagNames["searchProductsPageSizeOptionEnabledArg"])

	_, bindings, err := PreComputeFlagsAndBindings(sdk.Specification.CommandLineCommands)
	require.NoError(t, err)
	assert.Contains(t, bindings, ArgFlagBindingInfo{EnvVar: "page_token_enabled", FlagBindings: []string{"searchproductsCmd.Flags().Lookup(searchProductsPageTokenOptionEnabledArg)"}})
	assert.Contains(t, bindings, ArgFlagBindingInfo{EnvVar: "page_size_enabled", FlagBindings: []string{"searchproductsCmd.Flags().Lookup(searchProductsPageSizeOptionEnabledArg)"}})
}

func TestOptionParameterRejectsBindingsInManifest(t *testing.T) {
	t.Parallel()
	content, err := filesystem.ReadFile(filesystem.FilePathJoin(filesystem.NewStandardFileSystem(), "testdata", "options-variadic.yml"))
	require.NoError(t, err)
	invalid := strings.Replace(string(content), "              description: Whether tracing output is enabled.", "              description: Whether tracing output is enabled.\n              bindings:\n                - name: Tracing", 1)
	require.NotEqual(t, string(content), invalid)
	err = validateManifestContent(context.Background(), ".yml", []byte(invalid))
	require.Error(t, err)
}

func TestOptionParameterRejectsValueFieldNamedEnabled(t *testing.T) {
	t.Parallel()
	parameter := &OptionParameter{Name: "Enabled", LangType: "bool", Description: "Enabled."}
	err := parameter.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value_field.name")
}

func TestRootOptionCLIEnabledBindingUsesNestedEnvironmentKey(t *testing.T) {
	t.Parallel()
	spec := &SDKSpecification{Options: OptionSets{{Name: "Refine", Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: "LLM", LangType: "string"}}}}}}}
	cfg := &FunctionConfigFields{Name: "LLM", LangType: "options.RefineOptionsLLMOption", MapstructureTag: "llm", Description: "LLM refinement.", IsRoot: true}
	addOptionCLIEnabledBindings(spec, cfg)
	require.Len(t, cfg.Bindings, 1)
	assert.Equal(t, "llm_enabled", *cfg.Bindings[0].MapstructureTag)
	assert.Equal(t, "llm", *cfg.Bindings[0].Arg)
}

func TestEmbeddedOptionEnumsTemplateUsesTypedFirstConstantOnly(t *testing.T) {
	t.Parallel()
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/options.enums.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{CurrentOptionPackage: "options", CurrentOptionSets: OptionSets{{
		Name:        "Build",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{
				Name:        "Mode",
				LangType:    "string",
				Enum:        []string{"fast", "safe"},
				Description: "Mode.",
			}},
		}},
	}}}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "options.enums.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "BuildOptionsModeFast BuildOptionsMode = iota // fast")
	assert.Contains(t, output, "BuildOptionsModeSafe // safe")
	assert.NotContains(t, output, "BuildOptionsModeSafe BuildOptionsMode")
}

func TestOptionsTemplateGeneratesWrapperValidation(t *testing.T) {
	t.Parallel()
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/options.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{SDK: SDK{Specification: SDKSpecification{PackageName: "petshop"}}, CurrentOptionPackage: "options", CurrentOptionSets: OptionSets{{
		Name:        "Search",
		Description: "Search options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{
				Name:        "NoAPI",
				LangType:    "bool",
				Description: "Whether the API is disabled.",
			}},
		}, {
			Parameters: []OptionParameter{{
				Name:        "Connection",
				LangType:    "protectedConnection",
				Description: "Connection details.",
				ValueField:  &OptionValueField{Name: field.ToOptionalString("Connection")},
			}},
		}, {
			Parameters: []OptionParameter{{
				Name:        "APIHost",
				LangType:    "string",
				Description: "API host.",
				Condition:   "NoAPI == false",
			}},
		}},
	}}}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "options.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "func validateOptionValue(address, value any) error {")
	assert.Contains(t, output, "return validator.Validate()")
	assert.Contains(t, output, "func (opt *SearchOptionsConnectionOption) Validate() (err error) {")
	assert.Contains(t, output, "err = validateOptionValue(&opt.Connection, opt.Connection)")
	assert.Contains(t, output, "err = configUtils.ValidateEmbedded(opt)")
	assert.Contains(t, output, "validation.When(opt.NoAPI.Enabled && opt.NoAPI.NoAPI == false, validation.By(func(value any) error {")
	assert.NotContains(t, output, "validateSelectedOption")
	assert.NotContains(t, output, "opt.NoAPI == false.Enabled")
}

func TestOptionFunctionValidateRequiresFunctionNameForMultipleParameters(t *testing.T) {
	t.Parallel()
	optionFunction := &OptionFunction{Parameters: []OptionParameter{{
		Name:        "Width",
		LangType:    "int",
		Description: "Width.",
	}, {
		Name:        "Height",
		LangType:    "int",
		Description: "Height.",
	}}}

	err := optionFunction.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function_name")
}
