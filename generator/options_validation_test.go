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

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestOptionValidationConflictsWith(t *testing.T) {
	t.Parallel()

	parameters := []OptionParameter{
		{
			Name: "Alpha", LangType: "string", Description: "Alpha option.",
			Constraints: &ParameterConstraints{ConflictsWith: []string{"Beta"}},
		},
		{Name: "Beta", LangType: "string", Description: "Beta option."},
	}
	data := &DataModel{CurrentOptionPackage: "config", CurrentOptionSets: OptionSets{{Name: "Build"}}}

	for _, parameter := range parameters {
		data.CurrentOptionSets[0].Functions = append(data.CurrentOptionSets[0].Functions, OptionFunction{
			Parameters: []OptionParameter{parameter},
		})
	}

	dir := t.TempDir()
	optionsPath := filepath.Join(dir, "options.go")
	require.NoError(t, populateTemplate(t.Context(), data, "templates/options.go.tmpl", optionsPath))

	source := `package config

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
`
	testPath := filepath.Join(dir, "validation_test.go")
	require.NoError(t, filesystem.WriteFile(testPath, []byte(source), 0600))

	// File-list mode compiles the generated code using this module's pinned dependencies.
	cmd := exec.CommandContext(t.Context(), "go", "test", "-count=1", "-v", optionsPath, testPath)
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	require.NoError(t, err, "generated code should compile and satisfy the regression assertions")
}
