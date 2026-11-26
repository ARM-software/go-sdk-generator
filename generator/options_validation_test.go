/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestOptionValidationNestedFields(t *testing.T) {
	t.Parallel()
	parameters := []OptionParameter{{
		Name: "Label", LangType: "string", Description: "Build label.",
		Pattern: field.ToOptional("^[a-z]+$"),
	}}
	runOptionValidationFixture(t, parameters, nil, "nested-fields")
}

func TestOptionValidationSelectedValidators(t *testing.T) {
	t.Parallel()
	parameters := []OptionParameter{
		{Name: "Pointer", LangType: "PointerValidated", Description: "Pointer-receiver validator."},
		{Name: "Value", LangType: "ValueValidated", Description: "Value-receiver validator."},
		{Name: "PointerRef", LangType: "PointerValidated", Reference: true, Description: "Pointer-backed option."},
	}
	runOptionValidationFixture(t, parameters, nil, "selected-validators")
}

func TestOptionValidationConditionExpression(t *testing.T) {
	t.Parallel()
	parameters := []OptionParameter{
		{Name: "Tracing", LangType: "bool", Description: "Enable tracing."},
		{Name: "Label", LangType: "string", Description: "Trace label.", Condition: "Tracing == true"},
	}
	runOptionValidationFixture(t, parameters, nil, "condition-expression")
}

func TestOptionValidationConflictsWith(t *testing.T) {
	t.Parallel()
	parameters := []OptionParameter{
		{
			Name: "Alpha", LangType: "string", Description: "Alpha option.",
			Constraints: &ParameterConstraints{ConflictsWith: []string{"Beta"}},
		},
		{Name: "Beta", LangType: "string", Description: "Beta option."},
	}
	runOptionValidationFixture(t, parameters, nil, "conflicts-with")
}

func TestOptionValidationCLIConstraints(t *testing.T) {
	t.Parallel()
	parameters := []OptionParameter{{
		Name: "PageSize", LangType: "uint32", Description: "Requested page size.",
		Constraints: &ParameterConstraints{Maximum: field.ToOptional(float64(100))},
	}}
	cli := &CommandLineCommands{Functions: []CliFunction{{
		Name: "Search", Configuration: []FunctionConfigFields{{
			Name: "PageSizeOption", LangType: "BuildOptionsPageSizeOption", MapstructureTag: "page_size",
			DefaultValue: field.ToOptional("DefaultBuildOptions().PageSize"),
			Bindings:     []Binding{{Name: field.ToOptional("Enabled"), LangType: field.ToOptional("bool")}},
		}},
	}}}
	runOptionValidationFixture(t, parameters, cli, "cli-constraints")
}

func TestTogetherListOptionValidationInfersEnabled(t *testing.T) {
	t.Parallel()
	optionSet := OptionSet{
		Name: "Overwrite", EnableMode: OptionSetEnableModeTogether, InferEnabledFromNonEmptyValues: true,
		Functions: []OptionFunction{
			{Parameters: []OptionParameter{{Name: "ExcludeGroups", LangType: "string", List: true, Enum: []string{"ci", "sdk"}, Description: "Groups to preserve."}}},
			{Parameters: []OptionParameter{{Name: "OnlyGroups", LangType: "string", List: true, Enum: []string{"ci", "sdk"}, Description: "Groups to generate."}}},
		},
	}
	runOptionSetValidationFixture(t, optionSet, nil, "together-list-enabled")
}

func TestOptionCollectionEnumParsing(t *testing.T) {
	parameters := []OptionParameter{
		{Name: "Modes", LangType: "string", List: true, Enum: []string{"fast", "safe"}},
		{Name: "States", LangType: "string", Iterator: true, Enum: []string{"ready", "waiting"}},
	}
	runOptionValidationFixture(t, parameters, nil, "collection-enums")
}

func runOptionValidationFixture(t *testing.T, parameters []OptionParameter, cli *CommandLineCommands, fixture string) {
	t.Helper()
	optionSet := OptionSet{Name: "Build"}
	for _, parameter := range parameters {
		optionSet.Functions = append(optionSet.Functions, OptionFunction{
			Parameters: []OptionParameter{parameter},
		})
	}
	runOptionSetValidationFixture(t, optionSet, cli, fixture)
}

func runOptionSetValidationFixture(t *testing.T, optionSet OptionSet, cli *CommandLineCommands, fixture string) {
	t.Helper()
	localFs := filesystem.GetGlobalFileSystem()
	// The .go.txt suffix keeps fixtures out of Go package and test discovery.
	// Only the selected fixture is copied to a temporary Go test file below.
	source, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "options-validation", fixture, "validation_test.go.txt"))
	require.NoError(t, err)
	data := &DataModel{CurrentOptionPackage: "config", CurrentOptionSets: OptionSets{optionSet}}
	data.Specification.CommandLineCommands = cli

	dir := t.TempDir()
	optionsPath := filesystem.FilePathJoin(localFs, dir, "options.go")
	require.NoError(t, populateTemplate(t.Context(), data, "templates/options.go.tmpl", optionsPath, nil))
	args := []string{"test", "-count=1", "-v", optionsPath}
	if optionSet.HasEnum() {
		enumsPath := filesystem.FilePathJoin(localFs, dir, "enums.go")
		require.NoError(t, populateTemplate(t.Context(), data, optionsEnumFile, enumsPath, nil))
		args = append(args, enumsPath)
		for _, parameter := range optionSet.Parameters() {
			if !parameter.HasEnum() {
				continue
			}
			enumType := enumTypeName(optionSetStructName(optionSet.Name), parameter.Name, parameter.LangType, parameter.EffectiveEnum())
			parserPath := filesystem.FilePathJoin(localFs, dir, strings.ToLower(enumType)+"_enumer.go")
			cmd := exec.CommandContext(t.Context(), "go", "tool", "enumer", "-type="+enumType, "-linecomment", "-text", "-json", "-yaml", "-output="+parserPath, enumsPath) //nolint:gosec // Fixed executable and arguments derived only from generated test fixtures.
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "enumer failed: %s", output)
			args = append(args, parserPath)
		}
	}

	if cli != nil {
		// Co-locate the generated types to avoid creating an SDK module just
		// for this test. Both Validate methods come from the real templates.
		configPath := filesystem.FilePathJoin(localFs, dir, "config.go")
		require.NoError(t, populateTemplate(t.Context(), data, "templates/config.cli.go.tmpl", configPath, nil))
		args = append(args, configPath)
	}

	testPath := filesystem.FilePathJoin(localFs, dir, "validation_test.go")
	require.NoError(t, localFs.WriteFile(testPath, source, 0600))
	args = append(args, testPath)

	// File-list mode compiles the generated code using this module's pinned dependencies.
	cmd := exec.CommandContext(t.Context(), "go", args...)
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	require.NoError(t, err, "generated code should compile and satisfy the regression assertions")
}
