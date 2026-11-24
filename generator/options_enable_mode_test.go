/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	jsonutils "github.com/ARM-software/golang-utils/utils/serialisation/json"
)

func TestOptionSetEnableModes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mode     OptionSetEnableMode
		wantMode OptionSetEnableMode
		together bool
		invalid  bool
	}{
		{name: "default", wantMode: OptionSetEnableModeIndependent},
		{name: "independent", mode: OptionSetEnableModeIndependent, wantMode: OptionSetEnableModeIndependent},
		{name: "together", mode: OptionSetEnableModeTogether, wantMode: OptionSetEnableModeTogether, together: true},
		{name: "unsupported", mode: "unsupported", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			optionSet := OptionSet{
				Name: "Build", Description: "Build settings", EnableMode: tc.mode,
				Functions: []OptionFunction{{Parameters: []OptionParameter{{
					Name: "Label", LangType: "string", Description: "Build label",
				}}}},
			}
			content, err := jsonutils.Marshal(optionSet)
			require.NoError(t, err)
			var decoded OptionSet
			err = jsonutils.Unmarshal(content, &decoded)
			if tc.invalid {
				require.Error(t, err)
				require.ErrorIs(t, optionSet.Validate(), commonerrors.ErrInvalid)
				return
			}
			require.NoError(t, err)
			require.NoError(t, optionSet.Validate())
			require.NoError(t, decoded.Validate())
			require.Equal(t, tc.wantMode, decoded.EnableMode)
			require.Equal(t, tc.together, optionSet.EnablesTogether())
			require.Equal(t, tc.together, decoded.EnablesTogether())
		})
	}
	require.False(t, (*OptionSet)(nil).EnablesTogether())
}

func TestIndependentOptionSetValidation(t *testing.T) {
	optionSet := OptionSet{Name: "Build", EnableMode: OptionSetEnableModeIndependent, Functions: []OptionFunction{
		{Parameters: []OptionParameter{{Name: "Label", LangType: "string"}}},
		{Parameters: []OptionParameter{{Name: "Tracing", LangType: "bool"}}},
	}}
	runOptionSetValidationFixture(t, optionSet, nil, "independent")
}

func TestTogetherOptionSetValidation(t *testing.T) {
	optionSet := OptionSet{Name: "Build", EnableMode: OptionSetEnableModeTogether, Functions: []OptionFunction{
		{Parameters: []OptionParameter{{Name: "Tags", LangType: "string", List: true, Enum: []string{"fast", "safe"}, DefaultValue: field.ToOptionalString("safe")}}},
		{Parameters: []OptionParameter{{Name: "References", LangType: "string", List: true, Reference: true, DefaultValue: field.ToOptionalString("original")}}},
		{Parameters: []OptionParameter{{Name: "Tracing", LangType: "bool", DefaultValue: field.ToOptionalString("false")}}},
		{Parameters: []OptionParameter{{Name: "Label", LangType: "string", Condition: "Tracing == true"}}},
		{FunctionName: field.ToOptionalString("WithWindow"), Parameters: []OptionParameter{{Name: "Width", LangType: "int"}, {Name: "Height", LangType: "int"}}},
		{Parameters: []OptionParameter{{Name: "Profile", LangType: "string", Constraints: &ParameterConstraints{
			ForbiddenWhen: &ParameterValueConstraint{Parameter: "Tracing", Operator: "not_equals", Value: true},
		}}}},
	}}
	runOptionSetValidationFixture(t, optionSet, nil, "together")
}

func TestTogetherOptionEnumListLiteralDefault(t *testing.T) {
	t.Parallel()

	optionSet := OptionSet{Name: "Build", EnableMode: OptionSetEnableModeTogether, Functions: []OptionFunction{
		{Parameters: []OptionParameter{{
			Name: "Locations", LangType: "string", List: true,
			Enum: []string{"home", "remote"}, DefaultValue: field.ToOptionalString("home"),
		}}},
		{Parameters: []OptionParameter{{Name: "Tracing", LangType: "bool"}}},
	}}

	runOptionSetValidationFixture(t, optionSet, nil, "together-enum-list-default")
}

func TestTogetherOptionDisabledCondition(t *testing.T) {
	t.Parallel()

	optionSet := OptionSet{Name: "Build", EnableMode: OptionSetEnableModeTogether, Functions: []OptionFunction{{
		Parameters: []OptionParameter{{Name: "Reason", LangType: "string", Condition: "Enabled == false"}},
	}}}

	runOptionSetValidationFixture(t, optionSet, nil, "together-disabled-condition")
}

func TestTogetherOptionSetRejectsWrapperOnlyMetadata(t *testing.T) {
	for _, tc := range []struct {
		name      string
		parameter OptionParameter
		invalid   bool
	}{
		{name: "direct value", parameter: OptionParameter{Name: "Tags", LangType: "string", List: true}},
		{name: "value field override", parameter: OptionParameter{Name: "Tags", LangType: "string", ValueField: &OptionValueField{Name: field.ToOptionalString("Values")}}, invalid: true},
		{name: "reserved enabled field", parameter: OptionParameter{Name: "Enabled", LangType: "bool"}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			optionSet := &OptionSet{Name: "Build", EnableMode: OptionSetEnableModeTogether, Functions: []OptionFunction{{Parameters: []OptionParameter{tc.parameter}}}}
			err := optionSet.ValidateOptionParameterNames()
			if tc.invalid {
				require.ErrorIs(t, err, commonerrors.ErrInvalid)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTogetherOptionSetEnabledConstraints(t *testing.T) {
	optionSet := OptionSet{Name: "Build", EnableMode: OptionSetEnableModeTogether, Functions: []OptionFunction{
		{Parameters: []OptionParameter{{Name: "Tags", LangType: "string", List: true, Constraints: &ParameterConstraints{
			ForbiddenWhen: &ParameterValueConstraint{Parameter: "Enabled", Operator: "not_equals", Value: true},
		}}}},
		{Parameters: []OptionParameter{{Name: "Label", LangType: "string", Constraints: &ParameterConstraints{
			RequiredWhen: &ParameterValueConstraint{Parameter: "Enabled", Operator: "equals", Value: true},
		}}}},
	}}
	runOptionSetValidationFixture(t, optionSet, nil, "enabled-constraints")
}
