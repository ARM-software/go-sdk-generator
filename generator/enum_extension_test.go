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

func TestEnumConstNameCapitalisesLowercaseValues(t *testing.T) {
	assert.Equal(t, "PageStatusDraft", enumConstName("PageStatus", "draft"))
	assert.Equal(t, "PageStatusPublished", enumConstName("PageStatus", "published"))
}

func TestEnumStringValidationRuleUsesStringValues(t *testing.T) {
	rule := enumStringValidationRule("Generate", "Summary", string(BuiltInLangTypeString), false, []string{"file", "output"})
	assert.Equal(t, ", validationUtils.Enum(GenerateSummaryFile.String(), GenerateSummaryOutput.String())", rule)
}

func TestFieldEnumValidationRuleWithShapeUsesArrayItemsForList(t *testing.T) {
	rule := fieldEnumValidationRuleWithShape("Generate", "Mode", string(BuiltInLangTypeString), false, true, false, false, []string{"fast", "safe"}, "cfg.Mode")
	assert.Equal(t, ", validation.When(reflection.IsNotEmpty(cfg.Mode), validationUtils.ArrayItems(validationUtils.Enum(\"fast\", \"safe\")))", rule)
}

func TestFieldEnumValidationRuleWithShapeUsesArrayItemsForIterator(t *testing.T) {
	rule := fieldEnumValidationRuleWithShape("Generate", "Mode", string(BuiltInLangTypeString), false, false, true, false, []string{"fast", "safe"}, "cfg.Mode")
	assert.Equal(t, ", validation.When(reflection.IsNotEmpty(cfg.Mode), validationUtils.ArrayItems(validationUtils.Enum(\"fast\", \"safe\")))", rule)
}

func TestCliConfigFieldDefaultValueUsesEnumStringValue(t *testing.T) {
	value := cliConfigFieldDefaultValue("Generate", FunctionConfigFields{
		Name:         "Summary",
		LangType:     string(BuiltInLangTypeString),
		DefaultValue: field.ToOptionalString("output"),
		Enum:         []string{"file", "output"},
	})
	assert.Equal(t, "GenerateSummaryOutput.String()", value)
}

func TestCliConfigFieldDefaultValueSupportsReferenceLikeSecretFields(t *testing.T) {
	value := cliConfigFieldDefaultValue("Generate", FunctionConfigFields{
		Name:         "Summary",
		LangType:     string(BuiltInLangTypeString),
		DefaultValue: field.ToOptionalString("output"),
		Enum:         []string{"file", "output"},
		Reference:    true,
		Secret:       true,
	})
	require.Contains(t, value, "GenerateSummaryOutput.String()")
	require.Contains(t, value, "return &value")
}
