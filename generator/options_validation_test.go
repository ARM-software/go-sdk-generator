/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"os/exec"
	"path/filepath"
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
	runOptionValidationTest(t, parameters, nil, `package config

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestEnabledValidValue(t *testing.T) {
	opt := DefaultBuildOptions()
	WithLabel("valid")(opt)
	require.NoError(t, opt.Validate())
}

func TestDisabledInvalidValue(t *testing.T) {
	opt := DefaultBuildOptions()
	opt.Label.Label = "123"
	require.NoError(t, opt.Validate())
}
`)
}

func TestOptionValidationSelectedValidators(t *testing.T) {
	t.Parallel()

	parameters := []OptionParameter{
		{Name: "Pointer", LangType: "PointerValidated", Description: "Pointer-receiver validator."},
		{Name: "Value", LangType: "ValueValidated", Description: "Value-receiver validator."},
		{Name: "PointerRef", LangType: "PointerValidated", Reference: true, Description: "Pointer-backed option."},
	}
	runOptionValidationTest(t, parameters, nil, `package config

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/require"
)

var errSelectedValue = errors.New("selected value is invalid")

type PointerValidated struct{}
func (*PointerValidated) Validate() error { return errSelectedValue }

type ValueValidated struct{}
func (ValueValidated) Validate() error { return errSelectedValue }

type ScalarValidated string
func (ScalarValidated) Validate() error { return errSelectedValue }

func TestSelectedValidators(t *testing.T) {
	for _, test := range []struct {
		name string
		option BuildOption
	}{
		{name: "pointer_receiver", option: WithPointer(PointerValidated{})},
		{name: "value_receiver", option: WithValue(ValueValidated{})},
		{name: "pointer_backed_option", option: WithPointerRef(&PointerValidated{})},
	} {
		t.Run(test.name, func(t *testing.T) {
			opt := DefaultBuildOptions()
			test.option(opt)
			require.ErrorIs(t, opt.Validate(), errSelectedValue)
		})
	}
}

func TestDisabledValidators(t *testing.T) {
	opt := DefaultBuildOptions()
	opt.PointerRef.PointerRef = &PointerValidated{}
	require.NoError(t, opt.Validate())
}

func TestScalarValidator(t *testing.T) {
	value := ScalarValidated("invalid")
	var err error

	require.NotPanics(t, func() {
		err = validateOptionValue(&value, value)
	})
	require.ErrorIs(t, err, errSelectedValue)
}
`)
}

func TestOptionValidationConditionExpression(t *testing.T) {
	t.Parallel()

	parameters := []OptionParameter{
		{Name: "Tracing", LangType: "bool", Description: "Enable tracing."},
		{Name: "Label", LangType: "string", Description: "Trace label.", Condition: "Tracing == true"},
	}
	runOptionValidationTest(t, parameters, nil, `package config

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestFalseCondition(t *testing.T) {
	opt := DefaultBuildOptions()
	WithoutTracing()(opt)
	WithLabel("")(opt)
	require.NoError(t, opt.Validate())
}

func TestTrueConditionRequiresLabel(t *testing.T) {
	opt := DefaultBuildOptions()
	WithTracing()(opt)
	WithLabel("")(opt)
	require.Error(t, opt.Validate())
}

func TestTrueConditionAcceptsLabel(t *testing.T) {
	opt := DefaultBuildOptions()
	WithTracing()(opt)
	WithLabel("trace")(opt)
	require.NoError(t, opt.Validate())
}
`)
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
	runOptionValidationTest(t, parameters, nil, `package config

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestConflictsWith(t *testing.T) {
	tests := []struct {
		name string
		options []BuildOption
		wantError bool
	}{
		{name: "neither enabled"},
		{name: "only alpha enabled", options: []BuildOption{WithAlpha("alpha")}},
		{name: "only beta enabled", options: []BuildOption{WithBeta("beta")}},
		{name: "both enabled", options: []BuildOption{WithAlpha("alpha"), WithBeta("beta")}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			opt := WithBuildOptions(test.options...)
			err := opt.Validate()
			if test.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
`)
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
	runOptionValidationTest(t, parameters, cli, `package config

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestEnabledValidPageSize(t *testing.T) {
	cfg := DefaultSearchConfiguration()
	cfg.PageSizeOption.PageSize = 25
	cfg.PageSizeOption.Enabled = true
	require.NoError(t, cfg.Validate())
}

func TestEnabledInvalidPageSize(t *testing.T) {
	cfg := DefaultSearchConfiguration()
	cfg.PageSizeOption.PageSize = 200
	cfg.PageSizeOption.Enabled = true
	require.Error(t, cfg.Validate(), "enabled page size must obey maximum: 100")
}

func TestDisabledInvalidPageSize(t *testing.T) {
	cfg := DefaultSearchConfiguration()
	cfg.PageSizeOption.PageSize = 200
	require.NoError(t, cfg.Validate())
}
`)
}

func runOptionValidationTest(t *testing.T, parameters []OptionParameter, cli *CommandLineCommands, source string) {
	t.Helper()

	data := &DataModel{CurrentOptionPackage: "config", CurrentOptionSets: OptionSets{{Name: "Build"}}}
	data.Specification.CommandLineCommands = cli

	for _, parameter := range parameters {
		data.CurrentOptionSets[0].Functions = append(data.CurrentOptionSets[0].Functions, OptionFunction{
			Parameters: []OptionParameter{parameter},
		})
	}

	dir := t.TempDir()
	optionsPath := filepath.Join(dir, "options.go")
	require.NoError(t, populateTemplate(t.Context(), data, "templates/options.go.tmpl", optionsPath))
	args := []string{"test", "-count=1", "-v", optionsPath}

	if cli != nil {
		// Co-locate the generated types to avoid creating an SDK module just
		// for this test. Both Validate methods come from the real templates.
		configPath := filepath.Join(dir, "config.go")
		require.NoError(t, populateTemplate(t.Context(), data, "templates/config.cli.go.tmpl", configPath))
		args = append(args, configPath)
	}

	testPath := filepath.Join(dir, "validation_test.go")
	require.NoError(t, filesystem.WriteFile(testPath, []byte(source), 0600))
	args = append(args, testPath)

	// File-list mode compiles the generated code using this module's pinned dependencies.
	cmd := exec.CommandContext(t.Context(), "go", args...)
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	require.NoError(t, err, "generated code should compile and satisfy the regression assertions")
}
