/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestCliStringListFlagType(t *testing.T) {
	for _, langType := range []string{"string", "*string", "filepath", "uuid", "host"} {
		t.Run(langType, func(t *testing.T) {
			flagType, zero := cliFlagTypeAndZeroValue(langType, true)
			assert.Equal(t, "StringSlice", flagType)
			assert.Equal(t, "[]string{}", zero)
		})
	}
	flagType, zero := cliFlagTypeAndZeroValue("string", false)
	assert.Equal(t, "String", flagType)
	assert.Equal(t, `""`, zero)
}

func TestStringListFlagRegistrations(t *testing.T) {
	command := CliFunction{Name: "Check"}
	_, statement, _, _, err := parseNonBindingToFlag(command, FunctionConfigFields{
		Name: "Tags", LangType: "string", List: true,
		MapstructureTag: "tags", Description: "Tags to select.",
	})
	require.NoError(t, err)
	assert.Contains(t, statement, ".StringSlice(")
	assert.Contains(t, statement, "[]string{}")

	statement, _, _, _, err = parseBindingToFlag(command, FunctionConfigFields{
		Name: "Filter", LangType: "types.Filter", MapstructureTag: "filter", Description: "Filter.",
	}, Binding{
		Name: field.ToOptionalString("Tags"), LangType: field.ToOptionalString("string"), List: true,
		MapstructureTag: field.ToOptionalString("tags"), Description: field.ToOptionalString("Tags to select."),
	})
	require.NoError(t, err)
	assert.Contains(t, statement, ".StringSlice(")
	assert.Contains(t, statement, "[]string{}")
}
