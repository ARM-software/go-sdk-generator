/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	builtInManifestLangTypes     []string
	builtInManifestLangTypesOnce sync.Once
)

// types_extended_logic centralises capability queries and small schema-model
// traversal logic that belongs on the generated manifest types themselves.

func (b *Binding) HasEnum() bool {
	return false
}

func (arg *CommandArgument) HasEnum() bool {
	if arg == nil {
		return false
	}
	return hasEnum(arg.EffectiveEnum())
}

func (arg *CommandArgument) EffectiveEnum() []string {
	if arg == nil {
		return nil
	}
	return effectiveParameterEnum(arg.Enum, arg.Constraints)
}

func (arg *CommandArgument) EffectivePattern() *string {
	if arg == nil {
		return nil
	}
	return effectiveParameterPattern(arg.Pattern, arg.Constraints)
}

func (arg *CommandArgument) IsRequired() bool {
	if arg == nil {
		return false
	}
	return effectiveParameterRequired(arg.Required, arg.Constraints)
}

func (arg *CommandArgument) ApplyConstraintMirrors() {
	if arg == nil || arg.Constraints == nil {
		return
	}
	if reflection.IsEmpty(arg.Enum) && reflection.IsNotEmpty(arg.Constraints.Enum) {
		arg.Enum = append([]string(nil), arg.Constraints.Enum...)
	}
	if reflection.IsEmpty(arg.Pattern) && reflection.IsNotEmpty(arg.Constraints.Pattern) {
		arg.Pattern = field.ToOptionalString(field.OptionalString(arg.Constraints.Pattern, ""))
	}
	if !arg.Required && arg.Constraints.Required {
		arg.Required = true
	}
}

func (fieldDef *EntityField) HasEnum() bool {
	if fieldDef == nil {
		return false
	}
	return hasEnum(fieldDef.EffectiveEnum())
}

func (fieldDef *EntityField) EffectiveEnum() []string {
	if fieldDef == nil {
		return nil
	}
	return effectiveParameterEnum(fieldDef.Enum, fieldDef.Constraints)
}

func (fieldDef *EntityField) EffectivePattern() *string {
	if fieldDef == nil {
		return nil
	}
	return effectiveParameterPattern(fieldDef.Pattern, fieldDef.Constraints)
}

func (fieldDef *EntityField) IsRequired() bool {
	if fieldDef == nil {
		return false
	}
	return effectiveParameterRequired(fieldDef.Required, fieldDef.Constraints)
}

func (fieldDef *EntityField) ApplyConstraintMirrors() {
	if fieldDef == nil || fieldDef.Constraints == nil {
		return
	}
	if reflection.IsEmpty(fieldDef.Enum) && reflection.IsNotEmpty(fieldDef.Constraints.Enum) {
		fieldDef.Enum = append([]string(nil), fieldDef.Constraints.Enum...)
	}
	if reflection.IsEmpty(fieldDef.Pattern) && reflection.IsNotEmpty(fieldDef.Constraints.Pattern) {
		fieldDef.Pattern = field.ToOptionalString(field.OptionalString(fieldDef.Constraints.Pattern, ""))
	}
	if !fieldDef.Required && fieldDef.Constraints.Required {
		fieldDef.Required = true
	}
}

func (cfg *ConfigurationOption) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.EffectiveEnum())
}

func (cfg *ConfigurationOption) EffectiveEnum() []string {
	if cfg == nil {
		return nil
	}
	return effectiveParameterEnum(cfg.Enum, cfg.Constraints)
}

func (cfg *ConfigurationOption) EffectivePattern() *string {
	if cfg == nil {
		return nil
	}
	return effectiveParameterPattern(cfg.Pattern, cfg.Constraints)
}

func (cfg *ConfigurationOption) IsRequired() bool {
	if cfg == nil {
		return false
	}
	return effectiveParameterRequired(cfg.Required, cfg.Constraints)
}

func (cfg *ConfigurationOption) ApplyConstraintMirrors() {
	if cfg == nil || cfg.Constraints == nil {
		return
	}
	if reflection.IsEmpty(cfg.Enum) && reflection.IsNotEmpty(cfg.Constraints.Enum) {
		cfg.Enum = append([]string(nil), cfg.Constraints.Enum...)
	}
	if reflection.IsEmpty(cfg.Pattern) && reflection.IsNotEmpty(cfg.Constraints.Pattern) {
		cfg.Pattern = field.ToOptionalString(field.OptionalString(cfg.Constraints.Pattern, ""))
	}
	if !cfg.Required && cfg.Constraints.Required {
		cfg.Required = true
	}
}

func (cfg *ConfigurationOption) SuppressesRequiredValidation() bool {
	if cfg == nil {
		return false
	}
	return cfg.HasEnum() || isBoolean(cfg.LangType)
}

func (cfg *ConfigurationOption) UsesFilesystemValidationAlias() bool {
	return cfg != nil && (usesFilesystemValidationAlias(cfg.LangType) || usesFilesystemConstraintRule(cfg.Constraints))
}

func (cfg *ConfigurationOption) UsesJSONValidationAlias() bool {
	return cfg != nil && usesJSONValidationAlias(cfg.LangType)
}

func (cfg *FunctionConfigFields) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.EffectiveEnum())
}

func (cfg *FunctionConfigFields) EffectiveEnum() []string {
	if cfg == nil {
		return nil
	}
	return effectiveParameterEnum(cfg.Enum, cfg.Constraints)
}

func (cfg *FunctionConfigFields) EffectivePattern() *string {
	if cfg == nil {
		return nil
	}
	return effectiveParameterPattern(cfg.Pattern, cfg.Constraints)
}

func (cfg *FunctionConfigFields) IsRequired() bool {
	if cfg == nil {
		return false
	}
	return effectiveParameterRequired(cfg.Required, cfg.Constraints)
}

func (cfg *FunctionConfigFields) ApplyConstraintMirrors() {
	if cfg == nil || cfg.Constraints == nil {
		return
	}
	if reflection.IsEmpty(cfg.Enum) && reflection.IsNotEmpty(cfg.Constraints.Enum) {
		cfg.Enum = append([]string(nil), cfg.Constraints.Enum...)
	}
	if reflection.IsEmpty(cfg.Pattern) && reflection.IsNotEmpty(cfg.Constraints.Pattern) {
		cfg.Pattern = field.ToOptionalString(field.OptionalString(cfg.Constraints.Pattern, ""))
	}
	if !cfg.Required && cfg.Constraints.Required {
		cfg.Required = true
	}
}

func (optionFunction *OptionFunction) singleParameter() *OptionParameter {
	if optionFunction == nil || len(optionFunction.Parameters) != 1 {
		return nil
	}
	return &optionFunction.Parameters[0]
}

func (parameter *OptionParameter) HasEnum() bool {
	if parameter == nil {
		return false
	}
	return hasEnum(parameter.EffectiveEnum())
}

func (parameter *OptionParameter) EffectiveEnum() []string {
	if parameter == nil {
		return nil
	}
	return effectiveParameterEnum(parameter.Enum, parameter.Constraints)
}

func (parameter *OptionParameter) EffectivePattern() *string {
	if parameter == nil {
		return nil
	}
	return effectiveParameterPattern(parameter.Pattern, parameter.Constraints)
}

func (parameter *OptionParameter) IsRequired() bool {
	if parameter == nil {
		return false
	}
	return effectiveParameterRequired(parameter.Required, parameter.Constraints)
}

func (parameter *OptionParameter) ApplyConstraintMirrors() {
	if parameter == nil || parameter.Constraints == nil {
		return
	}
	if reflection.IsEmpty(parameter.Enum) && reflection.IsNotEmpty(parameter.Constraints.Enum) {
		parameter.Enum = append([]string(nil), parameter.Constraints.Enum...)
	}
	if reflection.IsEmpty(parameter.Pattern) && reflection.IsNotEmpty(parameter.Constraints.Pattern) {
		parameter.Pattern = field.ToOptionalString(field.OptionalString(parameter.Constraints.Pattern, ""))
	}
	if !parameter.Required && parameter.Constraints.Required {
		parameter.Required = true
	}
}

func (optionFunction *OptionFunction) ApplyConstraintMirrors() {
	if optionFunction == nil {
		return
	}
	optionFunction.FunctionName = field.ToOptionalStringOrNilIfEmpty(normaliseOptionFunctionName(field.OptionalString(optionFunction.FunctionName, "")))
	for i := range optionFunction.Parameters {
		optionFunction.Parameters[i].ApplyConstraintMirrors()
	}
}

func (optionFunction *OptionFunction) GeneratedName() string {
	return optionFunctionName(optionFunction)
}

func (arg *CommandArgument) GetName() string {
	if arg == nil {
		return ""
	}
	return arg.Name
}

func (arg *CommandArgument) GetConstraints() *ParameterConstraints {
	if arg == nil {
		return nil
	}
	return arg.Constraints
}

func (fieldDef *EntityField) GetName() string {
	if fieldDef == nil {
		return ""
	}
	return fieldDef.Name
}

func (fieldDef *EntityField) GetConstraints() *ParameterConstraints {
	if fieldDef == nil {
		return nil
	}
	return fieldDef.Constraints
}

func (cfg *ConfigurationOption) GetName() string {
	if cfg == nil {
		return ""
	}
	return cfg.Name
}

func (cfg *ConfigurationOption) GetConstraints() *ParameterConstraints {
	if cfg == nil {
		return nil
	}
	return cfg.Constraints
}

func (cfg *FunctionConfigFields) GetName() string {
	if cfg == nil {
		return ""
	}
	return cfg.Name
}

func (cfg *FunctionConfigFields) GetConstraints() *ParameterConstraints {
	if cfg == nil {
		return nil
	}
	return cfg.Constraints
}

func (parameter *OptionParameter) GetName() string {
	if parameter == nil {
		return ""
	}
	return parameter.Name
}

func (parameter *OptionParameter) GetConstraints() *ParameterConstraints {
	if parameter == nil {
		return nil
	}
	return parameter.Constraints
}

func (cfg *FunctionConfigFields) SuppressesRequiredValidation() bool {
	if cfg == nil {
		return false
	}
	return cfg.HasEnum() || isBoolean(cfg.LangType)
}

func (parameter *OptionParameter) SuppressesRequiredValidation() bool {
	return parameter != nil && (parameter.HasEnum() || isBoolean(parameter.LangType))
}

func (cfg *FunctionConfigFields) UsesFilesystemValidationAlias() bool {
	return cfg != nil && (usesFilesystemValidationAlias(cfg.LangType) || usesFilesystemConstraintRule(cfg.Constraints))
}

func (cfg *FunctionConfigFields) UsesJSONValidationAlias() bool {
	return cfg != nil && usesJSONValidationAlias(cfg.LangType)
}

func (parameter *OptionParameter) UsesFilesystemValidationAlias() bool {
	return parameter != nil && (usesFilesystemValidationAlias(parameter.LangType) || usesFilesystemConstraintRule(parameter.Constraints))
}

func (parameter *OptionParameter) UsesFmtInConstraintValidationRules() bool {
	return parameter != nil && usesFmtConstraintRule(parameter.Constraints)
}

func (parameter *OptionParameter) UsesJSONValidationAlias() bool {
	return parameter != nil && usesJSONValidationAlias(parameter.LangType)
}

func (parameter *OptionParameter) UsesPresenceGuardForValidation() bool {
	return parameter != nil && !parameter.IsRequired() && (usesPresenceGuardForValidation(parameter.LangType) || parameter.HasEnum())
}

func (parameter *OptionParameter) constrainedOptionParameter() constrainedParameter {
	if parameter == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{
		Name:        parameter.Name,
		LangType:    parameter.LangType,
		Reference:   parameter.Reference,
		Enum:        parameter.EffectiveEnum(),
		Constraints: parameter.Constraints,
		ValueRef:    fmt.Sprintf("&%s", optionParameterValueSelector(parameter, "opt")),
		PresenceRef: fmt.Sprintf("&%s", optionParameterEnabledSelector(parameter, "opt")),
	}
}

func (cfg *FunctionConfigFields) UsesValidationUtilsImport() bool {
	if cfg == nil {
		return false
	}
	validationRule := argValidationRule(cfg.LangType)
	return cfg.IsRequired() || strings.Contains(validationRule, "validationUtils.") || hasEnum(cfg.EffectiveEnum()) || cfg.Pattern != nil || cfg.Constraints.HasAnyValidationRules()
}

func (typ *Type) HasEnum() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.HasEnum()
	})
}

func (typ *Type) UsesFilesystemValidationAlias() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.UsesFilesystemValidationAlias()
	})
}

func (typ *Type) UsesJSONValidationAlias() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.UsesJSONValidationAlias()
	})
}

func (typ *Type) ApplyConstraintMirrors() {
	if typ == nil {
		return
	}
	for i := range typ.Fields {
		typ.Fields[i].ApplyConstraintMirrors()
	}
}

func (fieldDef *EntityField) SuppressesRequiredValidation() bool {
	if fieldDef == nil {
		return false
	}
	return fieldDef.HasEnum() || isBoolean(fieldDef.LangType)
}

func (fieldDef *EntityField) UsesFilesystemValidationAlias() bool {
	return fieldDef != nil && (usesFilesystemValidationAlias(fieldDef.LangType) || usesFilesystemConstraintRule(fieldDef.Constraints))
}

func (fieldDef *EntityField) UsesFmtInConstraintValidationRules() bool {
	return fieldDef != nil && usesFmtConstraintRule(fieldDef.Constraints)
}

func (cfg *ConfigurationOption) UsesFmtInConstraintValidationRules() bool {
	return cfg != nil && usesFmtConstraintRule(cfg.Constraints)
}

func (cfg *FunctionConfigFields) UsesFmtInConstraintValidationRules() bool {
	return cfg != nil && usesFmtConstraintRule(cfg.Constraints)
}

func (fieldDef *EntityField) UsesJSONValidationAlias() bool {
	return fieldDef != nil && usesJSONValidationAlias(fieldDef.LangType)
}

func (types Types) HasEnum() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.HasEnum()
	})
}

func (types Types) UsesFilesystemValidationAlias() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.UsesFilesystemValidationAlias()
	})
}

func (types Types) UsesJSONValidationAlias() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.UsesJSONValidationAlias()
	})
}

func (typ *Type) UsesFmtInConstraintValidationRules() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.UsesFmtInConstraintValidationRules()
	})
}

func (types Types) UsesFmtInConstraintValidationRules() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.UsesFmtInConstraintValidationRules()
	})
}

func (types Types) ApplyConstraintMirrors() {
	for i := range types {
		types[i].ApplyConstraintMirrors()
	}
}

func (optionSet *OptionSet) Parameters() []OptionParameter {
	if optionSet == nil {
		return nil
	}
	return collection.Reduce(optionSet.Functions, []OptionParameter{}, func(acc []OptionParameter, optionFunction OptionFunction) []OptionParameter {
		return append(acc, optionFunction.Parameters...)
	})
}

func (optionSet *OptionSet) HasEnum() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).HasEnum()
	})
}

func (optionSet *OptionSet) UsesFilesystemValidationAlias() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).UsesFilesystemValidationAlias()
	})
}

func (optionSet *OptionSet) UsesJSONValidationAlias() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).UsesJSONValidationAlias()
	})
}

func (optionSet *OptionSet) UsesFmtInConstraintValidationRules() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).UsesFmtInConstraintValidationRules()
	})
}

func (optionSet *OptionSet) UsesPresenceGuardForValidation() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).UsesPresenceGuardForValidation()
	})
}

func (optionSet *OptionSet) HasConstraintValidationRules() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return parameter.Constraints.HasAnyValidationRules()
	})
}

func (optionSet *OptionSet) UsesReflectInConstraintValidationRules() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return parameter.Constraints.UsesReflect()
	})
}

func (optionSet *OptionSet) UsesCollectionInConstraintValidationRules() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return parameter.Constraints.UsesCollection()
	})
}

func (optionSet *OptionSet) UsesSlicesClone() bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return parameter.List && !parameter.Iterator
	})
}

func (optionSet *OptionSet) ConstraintValidationRules() []string {
	if optionSet == nil {
		return nil
	}
	return parameterConstraintValidationRuleLines(collection.Map(optionSet.Parameters(), func(parameter OptionParameter) constrainedParameter {
		return (&parameter).constrainedOptionParameter()
	}), false)
}

func (optionSet *OptionSet) ApplyConstraintMirrors() {
	if optionSet == nil {
		return
	}
	for i := range optionSet.Functions {
		optionSet.Functions[i].ApplyConstraintMirrors()
	}
	optionSet.Package = normaliseOptionPackageName(optionSet.Package)
}

func (optionSet *OptionSet) NormalisePackage() {
	if optionSet == nil {
		return
	}
	optionSet.Package = normaliseOptionPackageName(optionSet.Package)
}

func (optionSets OptionSets) HasEnum() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.HasEnum() })
}

func (optionSets OptionSets) UsesFilesystemValidationAlias() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.UsesFilesystemValidationAlias() })
}

func (optionSets OptionSets) UsesJSONValidationAlias() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.UsesJSONValidationAlias() })
}

func (optionSets OptionSets) UsesFmtInConstraintValidationRules() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool {
		return optionSet != nil && optionSet.UsesFmtInConstraintValidationRules()
	})
}

func (optionSets OptionSets) UsesPresenceGuardForValidation() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.UsesPresenceGuardForValidation() })
}

func (optionSets OptionSets) HasConstraintValidationRules() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.HasConstraintValidationRules() })
}

func (optionSets OptionSets) UsesReflectInConstraintValidationRules() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool {
		return optionSet != nil && optionSet.UsesReflectInConstraintValidationRules()
	})
}

func (optionSets OptionSets) UsesCollectionInConstraintValidationRules() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool {
		return optionSet != nil && optionSet.UsesCollectionInConstraintValidationRules()
	})
}

func (optionSets OptionSets) UsesSlicesClone() bool {
	return collection.AnyRefFunc(optionSets, func(optionSet *OptionSet) bool { return optionSet != nil && optionSet.UsesSlicesClone() })
}

func (optionSets OptionSets) ApplyConstraintMirrors() {
	for i := range optionSets {
		optionSets[i].ApplyConstraintMirrors()
	}
}

func (optionSets OptionSets) NormalisePackages() {
	for i := range optionSets {
		optionSets[i].NormalisePackage()
	}
}

func (optionSet *OptionSet) ValidateOptionParameterNames() error {
	if optionSet == nil {
		return nil
	}
	parametersByName := map[string]OptionParameter{}
	for i := range optionSet.Functions {
		for j := range optionSet.Functions[i].Parameters {
			parameter := optionSet.Functions[i].Parameters[j]
			if strings.EqualFold(optionParameterValueFieldName(&parameter), optionEnabledFieldName) {
				return commonerrors.Newf(commonerrors.ErrInvalid, "option parameter %q in option set %q cannot use %s as its value field; set value_field.name to a different name", parameter.Name, optionSet.Name, optionEnabledFieldName)
			}
			generatedName := toPascalCase(parameter.Name)
			key := normaliseKeyNames(generatedName)
			if existing, found := parametersByName[key]; found {
				if existing.LangType != parameter.LangType || existing.Reference != parameter.Reference || existing.List != parameter.List || existing.Iterator != parameter.Iterator {
					return commonerrors.Newf(commonerrors.ErrConflict, "option parameter %q in option set %q is defined more than once with different shapes", generatedName, optionSet.Name)
				}
				return commonerrors.Newf(commonerrors.ErrConflict, "option parameter %q in option set %q is defined more than once; parameter names must be unique within an option set", generatedName, optionSet.Name)
			}
			parametersByName[key] = parameter
		}
	}
	return nil
}

func (optionSet *OptionSet) ValidateGeneratedNameUniqueness() error {
	if optionSet == nil {
		return nil
	}
	if err := optionSet.ValidateOptionParameterNames(); err != nil {
		return err
	}
	if err := validateUniqueCaseInsensitiveNames(fmt.Sprintf("option set %q", optionSet.Name), "function", collection.Map(optionSet.Functions, func(optionFunction OptionFunction) string {
		return (&optionFunction).GeneratedName()
	})); err != nil {
		return err
	}
	return validateUniqueCaseInsensitiveNames(fmt.Sprintf("option set %q", optionSet.Name), "parameter", collection.Map(optionSet.Parameters(), func(parameter OptionParameter) string {
		return toPascalCase(parameter.Name)
	}))
}

func (optionSets OptionSets) ValidateGeneratedNameUniqueness() error {
	if err := validateUniqueCaseInsensitiveNames("options", "option set", collection.Map(optionSets, func(optionSet OptionSet) string {
		return toPascalCase(optionSet.Name)
	})); err != nil {
		return err
	}
	return collection.ForAll(optionSets, func(optionSet OptionSet) error {
		return (&optionSet).ValidateGeneratedNameUniqueness()
	})
}

func (optionSets OptionSets) ValidateGeneratedIdentifiers() error {
	grouped := groupOptionSetsByPackage(optionSets)
	for pkg, groupedOptionSets := range grouped {
		identifiers := map[string]string{}
		for i := range groupedOptionSets {
			optionSet := &groupedOptionSets[i]
			if err := optionSet.ValidateGeneratedNameUniqueness(); err != nil {
				return err
			}
			for _, identifier := range optionSetIdentifiers(optionSet) {
				if reflection.IsEmpty(identifier.Name) {
					continue
				}
				key := normaliseKeyNames(identifier.Name)
				if existing, found := identifiers[key]; found {
					return commonerrors.Newf(commonerrors.ErrConflict, "generated option identifier %q in package %q is defined by both %s and %s", identifier.Name, pkg, existing, identifier.Source)
				}
				identifiers[key] = identifier.Source
			}
		}
	}
	return nil
}

func (cfgs Configuration) ValidateGeneratedNameUniqueness() error {
	return validateUniqueCaseInsensitiveNames("global configuration", "field", collection.Map(cfgs, func(cfg ConfigurationOption) string {
		return toPascalCase(cfg.Name)
	}))
}

func (typ *Type) ValidateGeneratedNameUniqueness() error {
	if typ == nil {
		return nil
	}
	return validateUniqueCaseInsensitiveNames(fmt.Sprintf("additional type %q", typ.Name), "field", collection.Map(typ.Fields, func(fieldDef EntityField) string {
		return toPascalCase(fieldDef.Name)
	}))
}

func (types Types) ValidateGeneratedNameUniqueness() error {
	if err := validateUniqueCaseInsensitiveNames("additional types", "type", collection.Map(types, func(typ Type) string {
		return toPascalCase(typ.Name)
	})); err != nil {
		return err
	}
	return collection.ForAll(types, func(typ Type) error {
		return (&typ).ValidateGeneratedNameUniqueness()
	})
}

func (cfgs Configuration) HasEnum() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.HasEnum()
	})
}

func (cfgs Configuration) UsesFilesystemValidationAlias() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.UsesFilesystemValidationAlias()
	})
}

func (cfgs Configuration) UsesJSONValidationAlias() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.UsesJSONValidationAlias()
	})
}

func (cfgs Configuration) UsesFmtInConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.UsesFmtInConstraintValidationRules()
	})
}

func (cfgs Configuration) SuppressedRequiredValidationMap() map[string]bool {
	result := map[string]bool{}
	collection.ForEachRef(cfgs, func(cfg *ConfigurationOption) {
		if cfg.SuppressesRequiredValidation() {
			result[validationSuppressionKey(globalConfigurationValidationScope, "", cfg.Name)] = true
		}
	})
	return result
}

func (cfgs Configuration) ApplyConstraintMirrors() {
	for i := range cfgs {
		cfgs[i].ApplyConstraintMirrors()
	}
}

func (method *InterfaceMethod) HasEnum() bool {
	if method == nil {
		return false
	}
	return collection.AnyRefFunc(method.Args, func(arg *CommandArgument) bool {
		return arg.HasEnum()
	}) || collection.AnyRefFunc(method.Returns, func(ret *CommandArgument) bool {
		return ret.HasEnum()
	})
}

func (method *InterfaceMethod) ApplyConstraintMirrors() {
	if method == nil {
		return
	}
	for i := range method.Args {
		method.Args[i].ApplyConstraintMirrors()
	}
	for i := range method.Returns {
		method.Returns[i].ApplyConstraintMirrors()
	}
}

func (iface *Interface) HasEnum() bool {
	if iface == nil {
		return false
	}
	return collection.AnyRefFunc(iface.Methods, func(method *InterfaceMethod) bool {
		return method.HasEnum()
	})
}

func (iface *Interface) ApplyConstraintMirrors() {
	if iface == nil {
		return
	}
	for i := range iface.Methods {
		iface.Methods[i].ApplyConstraintMirrors()
	}
}

func (method *InterfaceMethod) ValidateGeneratedNameUniqueness() error {
	if method == nil {
		return nil
	}
	if err := validateUniqueCaseInsensitiveNames(fmt.Sprintf("interface method %q", method.Name), "argument", collection.Map(method.Args, func(arg CommandArgument) string {
		return toCamelCase(arg.Name)
	})); err != nil {
		return err
	}
	return validateUniqueCaseInsensitiveNames(fmt.Sprintf("interface method %q", method.Name), "return value", collection.Map(method.Returns, func(ret CommandArgument) string {
		return toCamelCase(ret.Name)
	}))
}

func (iface *Interface) ValidateGeneratedNameUniqueness() error {
	if iface == nil {
		return nil
	}
	if err := validateUniqueCaseInsensitiveNames("interface", "method", collection.Map(iface.Methods, func(method InterfaceMethod) string {
		return toPascalCase(method.Name)
	})); err != nil {
		return err
	}
	return collection.ForAll(iface.Methods, func(method InterfaceMethod) error {
		return (&method).ValidateGeneratedNameUniqueness()
	})
}

func (fct *CliFunction) HasEnum() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.HasEnum()
	})
}

func (fct *CliFunction) HasGeneratedAction(cmds *CommandLineCommands) bool {
	if fct == nil || cmds == nil {
		return false
	}
	return !cmds.NoActions && !fct.Hidden && !fct.NoAction
}

func (fct *CliFunction) UsesSummary(cmds *CommandLineCommands) bool {
	if fct == nil {
		return false
	}
	if fct.WithSummary != nil {
		return field.Optional(fct.WithSummary, false)
	}
	return fct.HasGeneratedAction(cmds)
}

func (fct *CliFunction) UsesResult() bool {
	return fct != nil && fct.WithResult
}

func (cfg *FunctionConfigFields) CanGenerateValidationTestValue() bool {
	if cfg == nil {
		return false
	}
	return configFieldTestValue(nil, "", "tool", defaultGeneratedToolVersion, "", *cfg) != ""
}

func isBuiltInManifestLangType(langType string) bool {
	builtInManifestLangTypesOnce.Do(func() {
		builtInManifestLangTypes = collection.Map(enumValues_BuiltInLangType, func(value any) string {
			return strings.TrimSpace(fmt.Sprintf("%v", value))
		})
	})

	trimmed := strings.TrimSpace(baseLangType(langType))
	return collection.In(builtInManifestLangTypes, trimmed, collection.StringMatch)
}

func (fct *CliFunction) CanGenerateSummaryValidationTest(cmds *CommandLineCommands) bool {
	if fct == nil || !fct.UsesSummary(cmds) {
		return false
	}
	err := collection.ForAll(fct.Configuration, func(cfg FunctionConfigFields) error {
		if cfg.Name == summaryArgument || cfg.Name == summaryFileArg {
			return nil
		}
		if len(cfg.Bindings) > 0 && !isBuiltInManifestLangType(cfg.LangType) {
			return commonerrors.Newf(commonerrors.ErrUnsupported, "cannot generate validation test for bound custom type %s.%s", fct.Name, cfg.Name)
		}
		if !cfg.IsRequired() || cfg.DefaultValue != nil || (&cfg).CanGenerateValidationTestValue() {
			return nil
		}
		return commonerrors.Newf(commonerrors.ErrUnsupported, "cannot generate validation test value for %s.%s", fct.Name, cfg.Name)
	})
	return err == nil
}

func (cmds *CommandLineCommands) HasSummaryValidationTests() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.CanGenerateSummaryValidationTest(cmds)
	})
}

func (cfg *ConfigurationOption) UsesPresenceGuardForValidation() bool {
	return cfg != nil && !cfg.IsRequired() && (usesPresenceGuardForValidation(cfg.LangType) || cfg.HasEnum())
}

func (cfg *FunctionConfigFields) UsesPresenceGuardForValidation() bool {
	return cfg != nil && !cfg.IsRequired() && (usesPresenceGuardForValidation(cfg.LangType) || cfg.HasEnum())
}

func (fieldDef *EntityField) UsesPresenceGuardForValidation() bool {
	return fieldDef != nil && !fieldDef.IsRequired() && (usesPresenceGuardForValidation(fieldDef.LangType) || fieldDef.HasEnum())
}

func (fct *CliFunction) UsesFmtInConstraintValidationRules() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesFmtInConstraintValidationRules()
	})
}

func (fct *CliFunction) UsesFilesystemValidationAlias() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesFilesystemValidationAlias()
	})
}

func (fct *CliFunction) UsesJSONValidationAlias() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesJSONValidationAlias()
	})
}

func (fct *CliFunction) UsesValidationUtilsImport() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesValidationUtilsImport()
	})
}

func (fct *CliFunction) SuppressedRequiredValidationMap() map[string]bool {
	if fct == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(fct.Configuration, func(cfg *FunctionConfigFields) {
		if cfg.SuppressesRequiredValidation() {
			result[validationSuppressionKey(cliConfigurationValidationScope, fct.Name, cfg.Name)] = true
		}
	})
	return result
}

func (fct *CliFunction) ApplyConstraintMirrors() {
	if fct == nil {
		return
	}
	for i := range fct.Configuration {
		fct.Configuration[i].ApplyConstraintMirrors()
	}
}

func (fct *CliFunction) ValidateGeneratedNameUniqueness() error {
	if fct == nil {
		return nil
	}
	return validateUniqueCaseInsensitiveNames(fmt.Sprintf("cli function %q", fct.Name), "configuration", collection.Map(fct.Configuration, func(cfg FunctionConfigFields) string {
		return toPascalCase(cfg.Name)
	}))
}

func (cmds *CommandLineCommands) HasConfigEnums() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.HasEnum()
	})
}

func (cmds *CommandLineCommands) UsesFmtInConstraintValidationRules() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesFmtInConstraintValidationRules()
	})
}

func (cmds *CommandLineCommands) ValidateGeneratedNameUniqueness() error {
	if cmds == nil {
		return nil
	}
	if err := validateUniqueCaseInsensitiveNames("command line commands", "function", collection.Map(cmds.Functions, func(fct CliFunction) string {
		return toPascalCase(fct.Name)
	})); err != nil {
		return err
	}
	return collection.ForAll(cmds.Functions, func(fct CliFunction) error {
		return (&fct).ValidateGeneratedNameUniqueness()
	})
}

func (cmds *CommandLineCommands) UsesFilesystemValidationAlias() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesFilesystemValidationAlias()
	})
}

func (cmds *CommandLineCommands) UsesJSONValidationAlias() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesJSONValidationAlias()
	})
}

func (cmds *CommandLineCommands) UsesValidationUtilsImport() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesValidationUtilsImport()
	})
}

func (cfgs Configuration) UsesPresenceGuardForValidation() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool { return cfg != nil && cfg.UsesPresenceGuardForValidation() })
}

func (fct *CliFunction) UsesPresenceGuardForValidation() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool { return cfg != nil && cfg.UsesPresenceGuardForValidation() })
}

func (cmds *CommandLineCommands) UsesPresenceGuardForValidation() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct != nil && fct.UsesPresenceGuardForValidation() })
}

func (typ *Type) UsesPresenceGuardForValidation() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool { return fieldDef != nil && fieldDef.UsesPresenceGuardForValidation() })
}

func (types Types) UsesPresenceGuardForValidation() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool { return typ != nil && typ.UsesPresenceGuardForValidation() })
}

func (cmds *CommandLineCommands) UsesSummary() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesSummary(cmds)
	})
}

func (cmds *CommandLineCommands) UsesResult() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesResult()
	})
}

func (cmds *CommandLineCommands) SuppressedRequiredValidationMap() map[string]bool {
	if cmds == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(cmds.Functions, func(fct *CliFunction) {
		for key := range fct.SuppressedRequiredValidationMap() {
			result[key] = true
		}
	})
	return result
}

func (cmds *CommandLineCommands) ApplyConstraintMirrors() {
	if cmds == nil {
		return
	}
	for i := range cmds.Functions {
		cmds.Functions[i].ApplyConstraintMirrors()
	}
}

func (cmds *CommandLineCommands) HasSecretStringParameters() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
			return cfg.Secret && isSecretCompatibleLangType(cfg.LangType)
		})
	})
}

func (cmds *CommandLineCommands) HasCloudCommands() bool {
	if cmds == nil || len(cmds.Functions) == 0 {
		return false
	}
	return collection.AllFunc(cmds.Functions, func(fct CliFunction) bool {
		return fct.CallCloud || fct.StoreToKeyring
	})
}

func (spec *SDKSpecification) HasTypeEnums() bool {
	if spec == nil {
		return false
	}
	return spec.AdditionalTypes.HasEnum()
}

func (spec *SDKSpecification) HasConfigEnums() bool {
	if spec == nil {
		return false
	}
	return spec.GlobalConfiguration.HasEnum()
}

func (spec *SDKSpecification) HasInterfaceEnums() bool {
	if spec == nil {
		return false
	}
	return (&spec.Interface).HasEnum()
}

func (spec *SDKSpecification) UsesFmtInConstraintValidationRules() bool {
	return spec != nil && (spec.GlobalConfiguration.UsesFmtInConstraintValidationRules() || spec.AdditionalTypes.UsesFmtInConstraintValidationRules())
}

func (spec *SDKSpecification) UsesFilesystemValidationAlias() bool {
	return spec != nil && (spec.GlobalConfiguration.UsesFilesystemValidationAlias() || spec.AdditionalTypes.UsesFilesystemValidationAlias())
}

func (spec *SDKSpecification) UsesJSONValidationAlias() bool {
	return spec != nil && (spec.GlobalConfiguration.UsesJSONValidationAlias() || spec.AdditionalTypes.UsesJSONValidationAlias())
}

func (typ *Type) SuppressedRequiredValidationMap() map[string]bool {
	if typ == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(typ.Fields, func(fieldDef *EntityField) {
		if fieldDef.SuppressesRequiredValidation() {
			result[validationSuppressionKey(additionalTypeValidationScope, typ.Name, fieldDef.Name)] = true
		}
	})
	return result
}

func (types Types) SuppressedRequiredValidationMap() map[string]bool {
	result := map[string]bool{}
	collection.ForEachRef(types, func(typ *Type) {
		for key := range typ.SuppressedRequiredValidationMap() {
			result[key] = true
		}
	})
	return result
}

func (spec *SDKSpecification) SuppressedRequiredValidationMap() map[string]bool {
	if spec == nil {
		return nil
	}
	result := spec.GlobalConfiguration.SuppressedRequiredValidationMap()
	for key := range spec.CommandLineCommands.SuppressedRequiredValidationMap() {
		result[key] = true
	}
	for key := range spec.AdditionalTypes.SuppressedRequiredValidationMap() {
		result[key] = true
	}
	return result
}

func (spec *SDKSpecification) ApplyConstraintMirrors() {
	if spec == nil {
		return
	}
	spec.GlobalConfiguration.ApplyConstraintMirrors()
	spec.AdditionalTypes.ApplyConstraintMirrors()
	spec.Options.ApplyConstraintMirrors()
	(&spec.Interface).ApplyConstraintMirrors()
	if spec.CommandLineCommands != nil {
		spec.CommandLineCommands.ApplyConstraintMirrors()
	}
}

func (spec *SDKSpecification) ValidateGeneratedNameUniqueness() error {
	if spec == nil {
		return nil
	}
	if err := spec.GlobalConfiguration.ValidateGeneratedNameUniqueness(); err != nil {
		return err
	}
	if err := spec.AdditionalTypes.ValidateGeneratedNameUniqueness(); err != nil {
		return err
	}
	if err := (&spec.Interface).ValidateGeneratedNameUniqueness(); err != nil {
		return err
	}
	if err := spec.CommandLineCommands.ValidateGeneratedNameUniqueness(); err != nil {
		return err
	}
	return spec.Options.ValidateGeneratedNameUniqueness()
}

func (b *Binding) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if b == nil {
		return
	}
	b.LangType = field.ToOptionalStringOrNilIfEmpty(rewriteGeneratedAdditionalTypeReference(field.OptionalString(b.LangType, ""), currentPackage, additionalTypesPackages))
}

func (arg *CommandArgument) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if arg == nil {
		return
	}
	arg.LangType = rewriteGeneratedAdditionalTypeReference(arg.LangType, currentPackage, additionalTypesPackages)
	arg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(arg.DefaultValue, currentPackage, additionalTypesPackages)
	arg.Bindings = collection.MapRef(arg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (cfg *ConfigurationOption) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if cfg == nil {
		return
	}
	cfg.LangType = rewriteGeneratedAdditionalTypeReference(cfg.LangType, currentPackage, additionalTypesPackages)
	cfg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(cfg.DefaultValue, currentPackage, additionalTypesPackages)
	cfg.Bindings = collection.MapRef(cfg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (fieldDef *EntityField) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if fieldDef == nil {
		return
	}
	fieldDef.LangType = rewriteGeneratedAdditionalTypeReference(fieldDef.LangType, currentPackage, additionalTypesPackages)
	fieldDef.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(fieldDef.DefaultValue, currentPackage, additionalTypesPackages)
	fieldDef.Bindings = collection.MapRef(fieldDef.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (cfg *FunctionConfigFields) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if cfg == nil {
		return
	}
	cfg.LangType = rewriteGeneratedAdditionalTypeReference(cfg.LangType, currentPackage, additionalTypesPackages)
	cfg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(cfg.DefaultValue, currentPackage, additionalTypesPackages)
	cfg.Bindings = collection.MapRef(cfg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (parameter *OptionParameter) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if parameter == nil {
		return
	}
	parameter.LangType = rewriteGeneratedAdditionalTypeReference(parameter.LangType, currentPackage, additionalTypesPackages)
	parameter.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(parameter.DefaultValue, currentPackage, additionalTypesPackages)
}

func (optionFunction *OptionFunction) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if optionFunction == nil {
		return
	}
	optionFunction.Parameters = collection.MapRef(optionFunction.Parameters, func(parameter *OptionParameter) *OptionParameter {
		parameter.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return parameter
	})
}

func (method *InterfaceMethod) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if method == nil {
		return
	}
	method.Args = collection.MapRef(method.Args, func(arg *CommandArgument) *CommandArgument {
		arg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return arg
	})
	method.Returns = collection.MapRef(method.Returns, func(ret *CommandArgument) *CommandArgument {
		ret.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return ret
	})
}

func (iface *Interface) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if iface == nil {
		return
	}
	iface.Methods = collection.MapRef(iface.Methods, func(method *InterfaceMethod) *InterfaceMethod {
		method.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return method
	})
}

func (typ *Type) NormalisePackage(additionalTypesPackages map[string]string) error {
	if typ == nil {
		return nil
	}
	typ.Package = normaliseAdditionalTypePackageName(typ.Package)
	typeName := toPascalCase(typ.Name)
	if existing, found := additionalTypesPackages[typeName]; found && existing != typ.Package {
		return commonerrors.Newf(commonerrors.ErrConflict, "additional type %q is defined in multiple packages (%q and %q)", typeName, existing, typ.Package)
	}
	additionalTypesPackages[typeName] = typ.Package
	return nil
}

func (typ *Type) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if typ == nil {
		return
	}
	currentPackage := typ.Package
	typ.Fields = collection.MapRef(typ.Fields, func(fieldDef *EntityField) *EntityField {
		fieldDef.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return fieldDef
	})
}

func (types Types) NormalisePackages() (map[string]string, error) {
	additionalTypesPackages := make(map[string]string, len(types))
	for i := range types {
		if err := types[i].NormalisePackage(additionalTypesPackages); err != nil {
			return nil, err
		}
	}
	return additionalTypesPackages, nil
}

func (types Types) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	updated := collection.MapRef(types, func(typ *Type) *Type {
		typ.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return typ
	})
	copy(types, updated)
}

func (optionSet *OptionSet) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if optionSet == nil {
		return
	}
	currentPackage := optionPackageName(optionSet)
	optionSet.Functions = collection.MapRef(optionSet.Functions, func(optionFunction *OptionFunction) *OptionFunction {
		optionFunction.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return optionFunction
	})
}

func (optionSets OptionSets) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	updated := collection.MapRef(optionSets, func(optionSet *OptionSet) *OptionSet {
		optionSet.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return optionSet
	})
	copy(optionSets, updated)
}

func (fct *CliFunction) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if fct == nil {
		return
	}
	fct.Configuration = collection.MapRef(fct.Configuration, func(cfg *FunctionConfigFields) *FunctionConfigFields {
		cfg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return cfg
	})
}

func (cmds *CommandLineCommands) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if cmds == nil {
		return
	}
	cmds.Functions = collection.MapRef(cmds.Functions, func(fct *CliFunction) *CliFunction {
		fct.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return fct
	})
}

func (spec *SDKSpecification) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if spec == nil {
		return
	}
	spec.GlobalConfiguration = collection.MapRef(spec.GlobalConfiguration, func(cfg *ConfigurationOption) *ConfigurationOption {
		cfg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return cfg
	})
	(&spec.Interface).RewriteAdditionalTypeReferences(additionalTypesPackages)
	spec.CommandLineCommands.RewriteAdditionalTypeReferences(additionalTypesPackages)
	spec.AdditionalTypes.RewriteAdditionalTypeReferences(additionalTypesPackages)
	spec.Options.RewriteAdditionalTypeReferences(additionalTypesPackages)
}

func (cfg *FunctionConfigFields) IsStandaloneCLIFlag() bool {
	return cfg != nil && !cfg.Hidden && len(cfg.Bindings) == 0
}

func (cfg *FunctionConfigFields) CLIFlagArgName(functionName string) string {
	if cfg == nil {
		return ""
	}
	return getArgName(functionName, "", cfg.Name, cfg.IsRoot)
}

func (cfg *FunctionConfigFields) FlagConstraintRegistrations(commandVar string, parameterFlags map[string]string, configurationByName map[string]*FunctionConfigFields) (out []FlagConstraintRegistration, err error) {
	if cfg == nil || cfg.Constraints == nil {
		return nil, nil
	}
	if registration, hasAny, subErr := buildFlagConstraintRegistration(commandVar, cfg.Name, cfg.CLIRequiredTogetherTargets(configurationByName), parameterFlags, "MarkFlagsRequiredTogether"); subErr != nil {
		return nil, subErr
	} else if hasAny {
		out = append(out, registration)
	}
	if registration, hasAny, subErr := buildFlagConstraintRegistration(commandVar, cfg.Name, cfg.Constraints.ConflictsWith, parameterFlags, "MarkFlagsMutuallyExclusive"); subErr != nil {
		return nil, subErr
	} else if hasAny {
		out = append(out, registration)
	}
	return
}

func (fct *CliFunction) StandaloneCLIParameterFlags() map[string]string {
	if fct == nil {
		return nil
	}
	parameterFlags := make(map[string]string)
	collection.ForEach(fct.Configuration, func(cfg FunctionConfigFields) {
		if !cfg.IsStandaloneCLIFlag() {
			return
		}
		parameterFlags[cfg.Name] = cfg.CLIFlagArgName(fct.Name)
	})
	return parameterFlags
}

func (fct *CliFunction) ConfigurationByName() map[string]*FunctionConfigFields {
	if fct == nil {
		return nil
	}
	return collection.Reduce(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) *FunctionConfigFields {
		return &cfg
	}), map[string]*FunctionConfigFields{}, func(acc map[string]*FunctionConfigFields, cfg *FunctionConfigFields) map[string]*FunctionConfigFields {
		if cfg != nil {
			acc[cfg.Name] = cfg
		}
		return acc
	})
}

func (fct *CliFunction) FlagConstraintRegistrations() (out []FlagConstraintRegistration, err error) {
	if fct == nil {
		return nil, nil
	}
	configurationByName := fct.ConfigurationByName()
	parameterFlags := fct.StandaloneCLIParameterFlags()
	commandVar := fmt.Sprintf("%vCmd", strings.ToLower(fct.Name))
	err = collection.ForAll(fct.Configuration, func(cfg FunctionConfigFields) error {
		registrations, subErr := (&cfg).FlagConstraintRegistrations(commandVar, parameterFlags, configurationByName)
		if subErr != nil {
			return subErr
		}
		out = append(out, registrations...)
		return nil
	})
	return
}

func (cmds *CommandLineCommands) FlagConstraintRegistrations() (out []FlagConstraintRegistration, err error) {
	if cmds == nil {
		return nil, nil
	}
	err = collection.ForAll(cmds.Functions, func(fct CliFunction) error {
		registrations, subErr := (&fct).FlagConstraintRegistrations()
		if subErr != nil {
			return subErr
		}
		out = append(out, registrations...)
		return nil
	})
	slices.SortFunc(out, func(a, b FlagConstraintRegistration) int {
		return strings.Compare(a.Statement(), b.Statement())
	})
	return
}

func buildFlagConstraintRegistration(commandVar, sourceName string, targetNames []string, parameterFlags map[string]string, method string) (registration FlagConstraintRegistration, hasAny bool, err error) {
	if reflection.IsEmpty(targetNames) {
		return FlagConstraintRegistration{}, false, nil
	}
	sourceFlag, found := parameterFlags[sourceName]
	if !found {
		return FlagConstraintRegistration{}, false, commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares CLI constraints but does not map to a standalone flag", sourceName)
	}
	resolvedTargets, err := collection.MapWithError(collection.UniqueBy(targetNames, func(name string) string { return name }), func(targetName string) (string, error) {
		targetFlag, targetFound := parameterFlags[targetName]
		if !targetFound {
			return "", commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q references %q in CLI constraints but that parameter does not map to a standalone flag", sourceName, targetName)
		}
		return targetFlag, nil
	})
	if err != nil {
		return FlagConstraintRegistration{}, false, err
	}
	resolvedTargets = append([]string{sourceFlag}, resolvedTargets...)
	return FlagConstraintRegistration{CommandVar: commandVar, Method: method, FlagArgs: resolvedTargets}, true, nil
}

type constrainedParameter struct {
	OwnerName   string
	Name        string
	LangType    string
	Reference   bool
	Enum        []string
	Constraints *ParameterConstraints
	ValueRef    string
	PresenceRef string
}

func (constraints *ParameterConstraints) HasAnyValidationRules() bool {
	if constraints == nil {
		return false
	}
	return collection.AnyTrue(
		reflection.IsNotEmpty(constraints.DependentRequired),
		reflection.IsNotEmpty(constraints.ConflictsWith),
		reflection.IsNotEmpty(constraints.RequiredValuesWhen),
		reflection.IsNotEmpty(constraints.ForbiddenValuesWhen),
		reflection.IsNotEmpty(constraints.RequiredWhen),
		reflection.IsNotEmpty(constraints.ForbiddenWhen),
		reflection.IsNotEmpty(constraints.ParameterValue),
	)
}

func (constraints *ParameterConstraints) UsesCLIRequiredTogetherFlags() bool {
	return constraints != nil && constraints.RequireTogetherAsCliFlags
}

func (cfg *FunctionConfigFields) CLIRequiredTogetherTargets(configurationByName map[string]*FunctionConfigFields) []string {
	if cfg == nil || cfg.Constraints == nil || !cfg.Constraints.UsesCLIRequiredTogetherFlags() {
		return nil
	}
	return collection.Filter(collection.UniqueBy(cfg.Constraints.DependentRequired, func(name string) string { return name }), func(name string) bool {
		target, found := configurationByName[name]
		return found && target != nil && target.Constraints.UsesCLIRequiredTogetherFlags()
	})
}

func (constraints *ParameterConstraints) UsesReflect() bool {
	if constraints == nil {
		return false
	}
	return collection.AnyTrue(
		constraints.RequiredWhen.UsesReflect(),
		constraints.ForbiddenWhen.UsesReflect(),
		collection.AnyFunc(constraints.RequiredValuesWhen, func(condition ParameterRequiredValuesConstraint) bool { return (&condition).UsesReflect() }),
		collection.AnyFunc(constraints.ForbiddenValuesWhen, func(condition ParameterForbiddenValuesConstraint) bool { return (&condition).UsesReflect() }),
	)
}

func (constraints *ParameterConstraints) UsesCollection() bool {
	if constraints == nil {
		return false
	}
	return collection.AnyTrue(
		constraints.RequiredWhen.UsesCollection(),
		constraints.ForbiddenWhen.UsesCollection(),
		collection.AnyFunc(constraints.RequiredValuesWhen, func(condition ParameterRequiredValuesConstraint) bool { return (&condition).UsesCollection() }),
		collection.AnyFunc(constraints.ForbiddenValuesWhen, func(condition ParameterForbiddenValuesConstraint) bool { return (&condition).UsesCollection() }),
	)
}

func (condition *ParameterValueConstraint) operatorOrDefault() ParameterValueConstraintOperator {
	if condition == nil || reflection.IsEmpty(condition.Operator) {
		return ParameterValueConstraintOperatorPresent
	}
	return condition.Operator
}

func (condition *ParameterValueConstraint) UsesReflect() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorEquals, ParameterValueConstraintOperatorNotEquals, ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterValueConstraint) UsesCollection() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterValueConstraint) fieldRef(refs map[string]string) string {
	if condition == nil {
		return ""
	}
	return refs[condition.Parameter]
}

func renderConstraintLiteral(parameter constrainedParameter, value any, enumValuesAsStrings bool) string {
	if enumValuesAsStrings && hasEnum(parameter.Enum) {
		enumType := enumTypeName(parameter.OwnerName, parameter.Name, parameter.LangType, parameter.Enum)
		return fmt.Sprintf(`%s.String()`, enumConstName(enumType, fmt.Sprintf("%v", value)))
	}
	return goLiteral(value)
}

func renderConstraintValueList(parameter constrainedParameter, values []interface{}, enumValuesAsStrings bool) []string {
	return collection.Map(values, func(value any) string {
		return renderConstraintLiteral(parameter, value, enumValuesAsStrings)
	})
}

func (condition *ParameterValueConstraint) expectedLiteral(parameters map[string]constrainedParameter, enumValuesAsStrings bool) string {
	if condition == nil {
		return ""
	}
	parameter := parameters[condition.Parameter]
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return fmt.Sprintf("[]any{%s}", strings.Join(collection.Map(condition.Values, func(value any) string {
			return renderConstraintLiteral(parameter, value, enumValuesAsStrings)
		}), ", "))
	default:
		return renderConstraintLiteral(parameter, condition.Value, enumValuesAsStrings)
	}
}

func (condition *ParameterValueConstraint) ValidationRuleLine(valueRefs map[string]string, presenceRefs map[string]string, parameters map[string]constrainedParameter, validatorExpr string, enumValuesAsStrings bool) string {
	conditionValueRef := condition.fieldRef(valueRefs)
	conditionPresenceRef := condition.fieldRef(presenceRefs)
	if reflection.IsEmpty(conditionValueRef) && reflection.IsEmpty(conditionPresenceRef) {
		return ""
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorPresent:
		return fmt.Sprintf(`validationUtils.WhenFieldPresent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterValueConstraintOperatorAbsent:
		return fmt.Sprintf(`validationUtils.WhenFieldAbsent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterValueConstraintOperatorEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterValueConstraintOperatorNotEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterValueConstraintOperatorIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterValueConstraintOperatorNotIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	default:
		return ""
	}
}

func (condition *ParameterRequiredValuesConstraint) operatorOrDefault() ParameterRequiredValuesConstraintOperator {
	if condition == nil || reflection.IsEmpty(condition.Operator) {
		return ParameterRequiredValuesConstraintOperatorPresent
	}
	return condition.Operator
}

func (condition *ParameterRequiredValuesConstraint) UsesReflect() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterRequiredValuesConstraintOperatorEquals, ParameterRequiredValuesConstraintOperatorNotEquals, ParameterRequiredValuesConstraintOperatorIn, ParameterRequiredValuesConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterRequiredValuesConstraint) UsesCollection() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterRequiredValuesConstraintOperatorIn, ParameterRequiredValuesConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterRequiredValuesConstraint) fieldRef(refs map[string]string) string {
	if condition == nil {
		return ""
	}
	return refs[condition.Parameter]
}

func (condition *ParameterRequiredValuesConstraint) expectedLiteral(parameters map[string]constrainedParameter, enumValuesAsStrings bool) string {
	if condition == nil {
		return ""
	}
	parameter := parameters[condition.Parameter]
	switch condition.operatorOrDefault() {
	case ParameterRequiredValuesConstraintOperatorIn, ParameterRequiredValuesConstraintOperatorNotIn:
		return fmt.Sprintf("[]any{%s}", strings.Join(collection.Map(condition.Values, func(value any) string {
			return renderConstraintLiteral(parameter, value, enumValuesAsStrings)
		}), ", "))
	default:
		return renderConstraintLiteral(parameter, condition.Value, enumValuesAsStrings)
	}
}

func (condition *ParameterRequiredValuesConstraint) ValidationRuleLine(source constrainedParameter, valueRefs map[string]string, presenceRefs map[string]string, parameters map[string]constrainedParameter, enumValuesAsStrings bool) string {
	conditionValueRef := condition.fieldRef(valueRefs)
	conditionPresenceRef := condition.fieldRef(presenceRefs)
	targetRef := valueRefs[source.Name]
	targetPresenceRef := presenceRefs[source.Name]
	if reflection.IsEmpty(targetRef) || (reflection.IsEmpty(conditionValueRef) && reflection.IsEmpty(conditionPresenceRef)) {
		return ""
	}
	trimmedTargetRef := strings.TrimPrefix(targetRef, "&")
	validatorExpr := guardValidationByPresence(targetRef, targetPresenceRef, fmt.Sprintf(`validation.By(func(value any) error { return validation.Validate(%s, validationUtils.Enum(%s)) })`, trimmedTargetRef, strings.Join(renderConstraintValueList(source, condition.RequiredValues, enumValuesAsStrings), ", ")))
	switch condition.operatorOrDefault() {
	case ParameterRequiredValuesConstraintOperatorPresent:
		return fmt.Sprintf(`validationUtils.WhenFieldPresent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterRequiredValuesConstraintOperatorAbsent:
		return fmt.Sprintf(`validationUtils.WhenFieldAbsent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterRequiredValuesConstraintOperatorEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterRequiredValuesConstraintOperatorNotEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterRequiredValuesConstraintOperatorIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterRequiredValuesConstraintOperatorNotIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	default:
		return ""
	}
}

func (condition *ParameterForbiddenValuesConstraint) operatorOrDefault() ParameterForbiddenValuesConstraintOperator {
	if condition == nil || reflection.IsEmpty(condition.Operator) {
		return ParameterForbiddenValuesConstraintOperatorPresent
	}
	return condition.Operator
}

func (condition *ParameterForbiddenValuesConstraint) UsesReflect() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterForbiddenValuesConstraintOperatorEquals, ParameterForbiddenValuesConstraintOperatorNotEquals, ParameterForbiddenValuesConstraintOperatorIn, ParameterForbiddenValuesConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterForbiddenValuesConstraint) UsesCollection() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterForbiddenValuesConstraintOperatorIn, ParameterForbiddenValuesConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterForbiddenValuesConstraint) fieldRef(refs map[string]string) string {
	if condition == nil {
		return ""
	}
	return refs[condition.Parameter]
}

func (condition *ParameterForbiddenValuesConstraint) expectedLiteral(parameters map[string]constrainedParameter, enumValuesAsStrings bool) string {
	if condition == nil {
		return ""
	}
	parameter := parameters[condition.Parameter]
	switch condition.operatorOrDefault() {
	case ParameterForbiddenValuesConstraintOperatorIn, ParameterForbiddenValuesConstraintOperatorNotIn:
		return fmt.Sprintf("[]any{%s}", strings.Join(collection.Map(condition.Values, func(value any) string {
			return renderConstraintLiteral(parameter, value, enumValuesAsStrings)
		}), ", "))
	default:
		return renderConstraintLiteral(parameter, condition.Value, enumValuesAsStrings)
	}
}

func (condition *ParameterForbiddenValuesConstraint) ValidationRuleLine(source constrainedParameter, valueRefs map[string]string, presenceRefs map[string]string, parameters map[string]constrainedParameter, enumValuesAsStrings bool) string {
	conditionValueRef := condition.fieldRef(valueRefs)
	conditionPresenceRef := condition.fieldRef(presenceRefs)
	targetRef := valueRefs[source.Name]
	targetPresenceRef := presenceRefs[source.Name]
	if reflection.IsEmpty(targetRef) || (reflection.IsEmpty(conditionValueRef) && reflection.IsEmpty(conditionPresenceRef)) {
		return ""
	}
	trimmedTargetRef := strings.TrimPrefix(targetRef, "&")
	validatorExpr := ""
	if hasEnum(source.Enum) {
		forbidden := collection.Map(condition.ForbiddenValues, func(value any) string { return fmt.Sprintf("%v", value) })
		allowed := collection.Filter(source.Enum, func(value string) bool {
			return !collection.In(forbidden, value, collection.StringMatch)
		})
		validatorExpr = fmt.Sprintf(`validation.By(func(value any) error { return validation.Validate(%s, validationUtils.Enum(%s)) })`, trimmedTargetRef, strings.Join(collection.Map(allowed, strconv.Quote), ", "))
	} else {
		validatorExpr = fmt.Sprintf(`validation.By(func(value any) error { return validation.Validate(%s, validation.NotIn(%s)) })`, trimmedTargetRef, strings.Join(renderConstraintValueList(source, condition.ForbiddenValues, enumValuesAsStrings), ", "))
	}
	validatorExpr = guardValidationByPresence(targetRef, targetPresenceRef, validatorExpr)
	switch condition.operatorOrDefault() {
	case ParameterForbiddenValuesConstraintOperatorPresent:
		return fmt.Sprintf(`validationUtils.WhenFieldPresent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterForbiddenValuesConstraintOperatorAbsent:
		return fmt.Sprintf(`validationUtils.WhenFieldAbsent(%s, %s)`, conditionPresenceRef, validatorExpr)
	case ParameterForbiddenValuesConstraintOperatorEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterForbiddenValuesConstraintOperatorNotEquals:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotEqualsValue(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterForbiddenValuesConstraintOperatorIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	case ParameterForbiddenValuesConstraintOperatorNotIn:
		return guardValidationByPresence(conditionValueRef, conditionPresenceRef, fmt.Sprintf(`validationUtils.WhenFieldNotInValues(%s, %s, %s)`, conditionValueRef, condition.expectedLiteral(parameters, enumValuesAsStrings), validatorExpr))
	default:
		return ""
	}
}

func parameterValueComparisonValidationRuleLine(comparisonExpr string, message string, negated bool) string {
	if negated {
		return fmt.Sprintf(`validation.By(func(value any) error {
	if !%s {
		return nil
	}
	return commonerrors.New(commonerrors.ErrInvalid, %q)
})`, comparisonExpr, message)
	}
	return fmt.Sprintf(`validation.By(func(value any) error {
	if %s {
		return nil
	}
	return commonerrors.New(commonerrors.ErrInvalid, %q)
})`, comparisonExpr, message)
}

func (condition *ParameterValueComparisonConstraint) ValidationRuleLine(source constrainedParameter, valueRefs map[string]string, presenceRefs map[string]string) string {
	if condition == nil {
		return ""
	}
	targetRef := valueRefs[source.Name]
	targetPresenceRef := presenceRefs[source.Name]
	comparisonRef := valueRefs[condition.Parameter]
	comparisonPresenceRef := presenceRefs[condition.Parameter]
	if reflection.IsEmpty(targetRef) || reflection.IsEmpty(comparisonRef) {
		return ""
	}
	trimmedTargetRef := strings.TrimPrefix(targetRef, "&")
	trimmedComparisonRef := strings.TrimPrefix(comparisonRef, "&")
	comparisonExpr := fmt.Sprintf("reflect.DeepEqual(any(%s), any(%s))", trimmedTargetRef, trimmedComparisonRef)
	message := fmt.Sprintf("parameter %q must satisfy the configured value comparison against %q", source.Name, condition.Parameter)
	line := ""
	switch condition.Operator {
	case ParameterValueComparisonConstraintOperatorEquals:
		line = parameterValueComparisonValidationRuleLine(comparisonExpr, message, false)
	case ParameterValueComparisonConstraintOperatorNotEquals:
		line = parameterValueComparisonValidationRuleLine(comparisonExpr, message, true)
	default:
		return ""
	}
	if reflection.IsEmpty(line) {
		return ""
	}
	line = guardValidationByPresence(comparisonRef, comparisonPresenceRef, line)
	switch condition.whenOrDefault() {
	case ParameterValueComparisonConstraintWhenAlways:
		return guardValidationByPresence(targetRef, targetPresenceRef, line)
	case ParameterValueComparisonConstraintWhenEmpty:
		return fmt.Sprintf(`validationUtils.WhenFieldAbsent(%s, %s)`, targetPresenceRef, line)
	case ParameterValueComparisonConstraintWhenNotEmpty:
		return fmt.Sprintf(`validationUtils.WhenFieldPresent(%s, %s)`, targetPresenceRef, line)
	default:
		return line
	}
}

func (condition *ParameterValueComparisonConstraint) whenOrDefault() ParameterValueComparisonConstraintWhen {
	if condition == nil || reflection.IsEmpty(condition.When) {
		return ParameterValueComparisonConstraintWhenAlways
	}
	return condition.When
}

func (cfg *ConfigurationOption) constrainedParameter() constrainedParameter {
	if cfg == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: cfg.Name, LangType: cfg.LangType, Reference: cfg.Reference, Enum: cfg.EffectiveEnum(), Constraints: cfg.Constraints}
}

func (cfg *FunctionConfigFields) constrainedParameter() constrainedParameter {
	if cfg == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: cfg.Name, LangType: cfg.LangType, Reference: cfg.Reference, Enum: cfg.EffectiveEnum(), Constraints: cfg.Constraints}
}

func (fieldDef *EntityField) constrainedParameter() constrainedParameter {
	if fieldDef == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: fieldDef.Name, LangType: fieldDef.LangType, Reference: fieldDef.Reference, Enum: fieldDef.EffectiveEnum(), Constraints: fieldDef.Constraints}
}

func parameterRefs(parameters []constrainedParameter) map[string]string {
	return collection.Reduce(parameters, map[string]string{}, func(acc map[string]string, parameter constrainedParameter) map[string]string {
		if reflection.IsNotEmpty(parameter.Name) {
			if reflection.IsNotEmpty(parameter.ValueRef) {
				acc[parameter.Name] = parameter.ValueRef
			} else {
				acc[parameter.Name] = fmt.Sprintf("&cfg.%s", toPascalCase(parameter.Name))
			}
		}
		return acc
	})
}

func parameterPresenceRefs(parameters []constrainedParameter) map[string]string {
	return collection.Reduce(parameters, map[string]string{}, func(acc map[string]string, parameter constrainedParameter) map[string]string {
		if reflection.IsNotEmpty(parameter.Name) {
			switch {
			case reflection.IsNotEmpty(parameter.PresenceRef):
				acc[parameter.Name] = parameter.PresenceRef
			case reflection.IsNotEmpty(parameter.ValueRef):
				acc[parameter.Name] = parameter.ValueRef
			default:
				acc[parameter.Name] = fmt.Sprintf("&cfg.%s", toPascalCase(parameter.Name))
			}
		}
		return acc
	})
}

func guardValidationByPresence(valueRef, presenceRef, validatorExpr string) string {
	if reflection.IsEmpty(validatorExpr) || reflection.IsEmpty(presenceRef) || presenceRef == valueRef {
		return validatorExpr
	}
	return fmt.Sprintf(`validationUtils.WhenFieldPresent(%s, %s)`, presenceRef, validatorExpr)
}

func parameterByName(parameters []constrainedParameter) map[string]constrainedParameter {
	return collection.Reduce(parameters, map[string]constrainedParameter{}, func(acc map[string]constrainedParameter, parameter constrainedParameter) map[string]constrainedParameter {
		if reflection.IsNotEmpty(parameter.Name) {
			acc[parameter.Name] = parameter
		}
		return acc
	})
}

func goLiteral(value any) string {
	switch typed := value.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(typed)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", typed)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", typed)
	case float32:
		return strconv.FormatFloat(float64(typed), 'g', -1, 32)
	case float64:
		if math.Trunc(typed) == typed {
			return strconv.FormatFloat(typed, 'f', 0, 64)
		}
		return strconv.FormatFloat(typed, 'g', -1, 64)
	default:
		return fmt.Sprintf("%#v", typed)
	}
}

func parameterConstraintValidationRuleLines(parameters []constrainedParameter, enumValuesAsStrings bool) []string {
	valueRefs := parameterRefs(parameters)
	presenceRefs := parameterPresenceRefs(parameters)
	parametersByName := parameterByName(parameters)
	result := make([]string, 0)
	collection.ForEach(parameters, func(parameter constrainedParameter) {
		if !parameter.Constraints.HasAnyValidationRules() {
			return
		}
		targetRef, found := presenceRefs[parameter.Name]
		if !found {
			return
		}
		if reflection.IsNotEmpty(parameter.Constraints.DependentRequired) {
			dependentRefs := collection.Filter(collection.Map(collection.UniqueBy(parameter.Constraints.DependentRequired, func(name string) string { return name }), func(name string) string {
				return presenceRefs[name]
			}), func(ref string) bool { return reflection.IsNotEmpty(ref) })
			if len(dependentRefs) > 0 {
				result = append(result, fmt.Sprintf(`validationUtils.FieldDependencyBy(map[any][]any{%s: []any{%s}})`, targetRef, strings.Join(dependentRefs, ", ")))
			}
		}
		if reflection.IsNotEmpty(parameter.Constraints.ConflictsWith) {
			conflictRefs := collection.Filter(collection.Map(collection.UniqueBy(parameter.Constraints.ConflictsWith, func(name string) string { return name }), func(name string) string {
				return presenceRefs[name]
			}), func(ref string) bool { return reflection.IsNotEmpty(ref) })
			if len(conflictRefs) > 0 {
				refsWithSource := append([]string{targetRef}, conflictRefs...)
				result = append(result, fmt.Sprintf(`validationUtils.MutuallyExclusiveFieldsBy(%s)`, strings.Join(refsWithSource, ", ")))
			}
		}
		if line := parameter.Constraints.RequiredWhen.ValidationRuleLine(valueRefs, presenceRefs, parametersByName, fmt.Sprintf(`validationUtils.RequiredFieldsBy(%s)`, targetRef), enumValuesAsStrings); reflection.IsNotEmpty(line) {
			result = append(result, line)
		}
		if line := parameter.Constraints.ForbiddenWhen.ValidationRuleLine(valueRefs, presenceRefs, parametersByName, fmt.Sprintf(`validationUtils.ForbiddenFieldsBy(%s)`, targetRef), enumValuesAsStrings); reflection.IsNotEmpty(line) {
			result = append(result, line)
		}
		collection.ForEach(parameter.Constraints.RequiredValuesWhen, func(condition ParameterRequiredValuesConstraint) {
			if line := (&condition).ValidationRuleLine(parameter, valueRefs, presenceRefs, parametersByName, enumValuesAsStrings); reflection.IsNotEmpty(line) {
				result = append(result, line)
			}
		})
		collection.ForEach(parameter.Constraints.ForbiddenValuesWhen, func(condition ParameterForbiddenValuesConstraint) {
			if line := (&condition).ValidationRuleLine(parameter, valueRefs, presenceRefs, parametersByName, enumValuesAsStrings); reflection.IsNotEmpty(line) {
				result = append(result, line)
			}
		})
		if line := parameter.Constraints.ParameterValue.ValidationRuleLine(parameter, valueRefs, presenceRefs); reflection.IsNotEmpty(line) {
			result = append(result, line)
		}
	})
	return collection.UniqueBy(result, func(line string) string { return line })
}

func (cfgs Configuration) HasConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool { return cfg != nil && cfg.Constraints.HasAnyValidationRules() })
}

func (cfgs Configuration) UsesReflectInConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg != nil && cfg.Constraints.UsesReflect()
	})
}

func (cfgs Configuration) UsesCollectionInConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg != nil && cfg.Constraints.UsesCollection()
	})
}

func (cfgs Configuration) ConstraintValidationRules() []string {
	return parameterConstraintValidationRuleLines(collection.Map(cfgs, func(cfg ConfigurationOption) constrainedParameter { return (&cfg).constrainedParameter() }), false)
}

func (fct *CliFunction) HasConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool { return cfg != nil && cfg.Constraints.HasAnyValidationRules() })
}

func (fct *CliFunction) UsesReflectInConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg != nil && cfg.Constraints.UsesReflect()
	})
}

func (fct *CliFunction) UsesCollectionInConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg != nil && cfg.Constraints.UsesCollection()
	})
}

func (fct *CliFunction) ConstraintValidationRules() []string {
	if fct == nil {
		return nil
	}
	return parameterConstraintValidationRuleLines(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) constrainedParameter {
		parameter := (&cfg).constrainedParameter()
		parameter.OwnerName = fct.Name
		return parameter
	}), true)
}

func (typ *Type) HasConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.HasAnyValidationRules()
	})
}

func (typ *Type) UsesReflectInConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.UsesReflect()
	})
}

func (typ *Type) UsesCollectionInConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.UsesCollection()
	})
}

func (typ *Type) ConstraintValidationRules() []string {
	if typ == nil {
		return nil
	}
	return parameterConstraintValidationRuleLines(collection.Map(typ.Fields, func(fieldDef EntityField) constrainedParameter { return (&fieldDef).constrainedParameter() }), false)
}

func (cmds *CommandLineCommands) HasConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.HasConstraintValidationRules() })
}

func (cmds *CommandLineCommands) UsesReflectInConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.UsesReflectInConstraintValidationRules() })
}

func (cmds *CommandLineCommands) UsesCollectionInConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.UsesCollectionInConstraintValidationRules() })
}
