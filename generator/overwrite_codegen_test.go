/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestOverwriteListCLIHelpers(t *testing.T) {
	cfg := FunctionConfigFields{Name: "ExcludeFiles", LangType: "string", List: true, MapstructureTag: "exclude_files", Description: "Files to preserve"}
	_, line, _, _, err := parseNonBindingToFlag(CliFunction{Name: "Generate"}, cfg)
	require.NoError(t, err)
	require.Contains(t, line, `.StringSlice(generateExcludeFilesArg, []string{},`)
	binding := Binding{Name: field.ToOptionalString("Files"), LangType: field.ToOptionalString("string"), List: true, Description: field.ToOptionalString("Files to preserve"), MapstructureTag: field.ToOptionalString("files")}
	line, _, _, _, err = parseBindingToFlag(CliFunction{Name: "Generate"}, cfg, binding)
	require.NoError(t, err)
	require.Contains(t, line, `.StringSlice(`)
	require.Equal(t, `[]string{"value"}`, configFieldTestValue(nil, "", "tool", "1.0.0", "Generate", cfg))
	cfg.Enum = []string{"project", "agent"}
	require.Equal(t, `[]string{"project"}`, configFieldTestValue(nil, "", "tool", "1.0.0", "Generate", cfg))
	cfg.Enum = nil
	cfg.Constraints = &ParameterConstraints{ForbiddenWhen: &ParameterValueConstraint{Parameter: "Generate", Operator: "not_equals", Value: true}}
	require.True(t, shouldAssignCLIConfigTestValue(cfg, `[]string{"value"}`), "forbidden_when is evaluated against the completed test configuration")
	cfg = FunctionConfigFields{
		Name: "Overwrite", LangType: "options.OverwriteOptions", MapstructureTag: "overwrite",
		DefaultValue: field.ToOptionalString("*options.DefaultOverwriteOptions()"),
		Bindings:     []Binding{binding},
	}
	require.False(t, shouldAssignCLIConfigTestValue(cfg, "options.OverwriteOptions{}"), "optional bound fields retain their generated defaults")
	fct := &CliFunction{Name: "Generate", WithSummary: field.ToOptionalBool(true), Configuration: []FunctionConfigFields{cfg}}
	require.True(t, fct.CanGenerateSummaryValidationTest(&CommandLineCommands{}))
}
