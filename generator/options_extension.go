/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	defaultOptionPackage   = "options"
	optionEnabledFieldName = "Enabled"
)

type generatedOptionIdentifier struct {
	Name   string
	Source string
}

func normaliseOptionPackageName(value string) string {
	if reflection.IsEmpty(value) {
		return defaultOptionPackage
	}
	normalised := toPackageName(value)
	if normalised == "main" {
		return defaultOptionPackage
	}
	return normalised
}

func validateOptionPackageName(value string) error {
	return validateGeneratedPackageName(value)
}

func optionPackageName(optionSet *OptionSet) string {
	if optionSet == nil {
		return defaultOptionPackage
	}
	return normaliseOptionPackageName(optionSet.Package)
}

func groupOptionSetsByPackage(optionSets OptionSets) map[string]OptionSets {
	grouped := collection.GroupByRef(optionSets, optionPackageName)
	result := make(map[string]OptionSets, len(grouped))
	for pkg, groupedOptionSets := range grouped {
		result[pkg] = OptionSets(groupedOptionSets)
	}
	return result
}

func sortedOptionPackages(optionSets OptionSets) []string {
	packages := slices.Collect(maps.Keys(groupOptionSetsByPackage(optionSets)))
	slices.Sort(packages)
	return packages
}

func optionImports(spec *SDKSpecification, currentPackage string) []additionalTypeImport {
	if spec == nil {
		return nil
	}
	return collection.FilterRef(collection.Map(sortedOptionPackages(spec.Options), func(pkg string) additionalTypeImport {
		return additionalTypeImport{
			Alias: pkg,
			Path:  fmt.Sprintf("%s/%s", sdkModuleRoot(spec), pkg),
		}
	}), func(i *additionalTypeImport) bool { return i.Alias != currentPackage })
}

func optionSetStructName(name string) string {
	return fmt.Sprintf("%sOptions", toPascalCase(name))
}

func optionSetOptionTypeName(name string) string {
	return fmt.Sprintf("%sOption", toPascalCase(name))
}

func optionParameterWrapperTypeName(ownerName string, parameter *OptionParameter) string {
	if parameter == nil {
		return ""
	}
	return fmt.Sprintf("%s%sOption", ownerName, toPascalCase(parameter.Name))
}

func optionParameterValueFieldName(parameter *OptionParameter) string {
	if parameter == nil {
		return "Value"
	}
	if parameter.ValueField != nil && reflection.IsNotEmpty(parameter.ValueField.Name) {
		return toPascalCase(field.OptionalString(parameter.ValueField.Name, ""))
	}
	return toPascalCase(parameter.Name)
}

func optionParameterMapstructureTag(parameter *OptionParameter) string {
	if parameter == nil {
		return ""
	}
	if reflection.IsNotEmpty(parameter.MapstructureTag) {
		return field.OptionalString(parameter.MapstructureTag, "")
	}
	return toCamelCase(parameter.Name)
}

func optionParameterValueFieldMapstructureTag(parameter *OptionParameter) string {
	if parameter != nil && parameter.ValueField != nil && reflection.IsNotEmpty(parameter.ValueField.MapstructureTag) {
		return field.OptionalString(parameter.ValueField.MapstructureTag, "")
	}
	return toSnakeCase(optionParameterValueFieldName(parameter))
}

func optionParameterValueSelector(parameter *OptionParameter, receiver string) string {
	if parameter == nil || reflection.IsEmpty(receiver) {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s", receiver, toPascalCase(parameter.Name), optionParameterValueFieldName(parameter))
}

func optionParameterEnabledSelector(parameter *OptionParameter, receiver string) string {
	if parameter == nil || reflection.IsEmpty(receiver) {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s", receiver, toPascalCase(parameter.Name), optionEnabledFieldName)
}

// addOptionCLIEnabledBindings bridges generated functional options and CLI
// configuration. A parameter such as options.RefineOptionsLLMOption has both a
// value and an Enabled field, but CLI bindings normally exist only for fields
// declared explicitly on the command configuration. When that wrapper is used
// as a CLI field with mapstructure_tag: llm, create a bool binding for Enabled:
// the flag is --llm and its environment key is llm_enabled. The wrapper's
// default constructor supplies the value without enabling the option.
// An explicit Enabled binding is preserved so its arg can rename --llm.
func addOptionCLIEnabledBindings(spec *SDKSpecification, cfg *FunctionConfigFields) {
	if spec == nil || cfg == nil {
		return
	}
	enabledTag := strings.ToLower(optionEnabledFieldName)
	for _, optionSet := range spec.Options {
		structName := optionSetStructName(optionSet.Name)
		for _, parameter := range (&optionSet).Parameters() {
			typeName := fmt.Sprintf("%s.%s", optionPackageName(&optionSet), optionParameterWrapperTypeName(structName, &parameter))
			if baseLangType(cfg.LangType) != typeName {
				continue
			}
			if reflection.IsEmpty(cfg.DefaultValue) {
				defaultExpr := fmt.Sprintf("%s.Default%s().%s", optionPackageName(&optionSet), structName, toPascalCase(parameter.Name))
				if cfg.Reference || isPointerLangType(cfg.LangType) {
					defaultExpr = "&" + defaultExpr
				}
				cfg.DefaultValue = field.ToOptionalString(defaultExpr)
			}
			arg := field.OptionalString(cfg.Arg, replaceUnderscores(cfg.MapstructureTag))
			for i := range cfg.Bindings {
				if strings.EqualFold(field.OptionalString(cfg.Bindings[i].Name, ""), optionEnabledFieldName) {
					if reflection.IsEmpty(cfg.Bindings[i].Arg) {
						cfg.Bindings[i].Arg = field.ToOptionalString(arg)
					}
					if cfg.IsRoot && strings.EqualFold(field.OptionalString(cfg.Bindings[i].MapstructureTag, ""), enabledTag) {
						cfg.Bindings[i].MapstructureTag = field.ToOptionalString(cfg.MapstructureTag + "_" + enabledTag)
					}
					return
				}
			}
			bindingTag := enabledTag
			if cfg.IsRoot {
				bindingTag = cfg.MapstructureTag + "_" + enabledTag
			}
			cfg.Bindings = append(cfg.Bindings, Binding{
				Name:            field.ToOptionalString(optionEnabledFieldName),
				LangType:        field.ToOptionalString("bool"),
				Description:     field.ToOptionalString(fmt.Sprintf("Enable %s", cfg.Description)),
				MapstructureTag: field.ToOptionalString(bindingTag),
				Arg:             field.ToOptionalString(arg),
			})
			return
		}
	}
}

func optionSetBuilderFunctionName(name string) string {
	return fmt.Sprintf("With%sOptions", toPascalCase(name))
}

func normaliseOptionFunctionName(value string) string {
	trimmed := strings.TrimSpace(value)
	if reflection.IsEmpty(trimmed) {
		return ""
	}
	return toPascalCase(trimmed)
}

func isBooleanOptionParameter(parameter *OptionParameter) bool {
	return parameter != nil && !parameter.List && !parameter.Iterator && isBoolean(baseLangType(parameter.LangType))
}

func optionFunctionUsesBooleanShortcut(optionFunction *OptionFunction) bool {
	parameter := optionFunction.singleParameter()
	return parameter != nil && isBooleanOptionParameter(parameter) && reflection.IsEmpty(optionFunction.FunctionName)
}

func optionFunctionSingleParameter(optionFunction *OptionFunction) *OptionParameter {
	return optionFunction.singleParameter()
}

func optionFunctionName(optionFunction *OptionFunction) string {
	if optionFunction == nil {
		return ""
	}
	if functionName := normaliseOptionFunctionName(field.OptionalString(optionFunction.FunctionName, "")); reflection.IsNotEmpty(functionName) {
		return functionName
	}
	parameter := optionFunction.singleParameter()
	if parameter == nil {
		return ""
	}
	return fmt.Sprintf("With%s", toPascalCase(parameter.Name))
}

func optionWithoutFunctionName(optionFunction *OptionFunction) string {
	parameter := optionFunction.singleParameter()
	if parameter == nil || !optionFunctionUsesBooleanShortcut(optionFunction) {
		return ""
	}
	return fmt.Sprintf("Without%s", toPascalCase(parameter.Name))
}

func optionParameterType(ownerName string, parameter *OptionParameter) string {
	if parameter == nil {
		return ""
	}
	baseType := convertLangType(baseLangType(parameter.LangType))
	if parameter.HasEnum() {
		baseType = enumTypeName(ownerName, parameter.Name, baseLangType(parameter.LangType), parameter.EffectiveEnum())
	}
	return parameterTypeFromBase(parameter.Reference, parameter.List, parameter.Iterator, baseType)
}

func optionBooleanParameterValueExpression(ownerName string, parameter *OptionParameter, value bool) string {
	valueExpression := "false"
	if value {
		valueExpression = "true"
	}
	if parameter == nil || !parameter.Reference {
		return valueExpression
	}
	return fmt.Sprintf("func() %s {\nvalue := %s\nreturn &value\n}()", optionParameterType(ownerName, parameter), valueExpression)
}

func optionParameterCloneStatements(parameter *OptionParameter) string {
	if parameter == nil || !parameter.List || parameter.Iterator {
		return ""
	}
	fieldName := toPascalCase(parameter.Name)
	valueFieldName := optionParameterValueFieldName(parameter)
	if parameter.Reference {
		return fmt.Sprintf("if opt.%s.%s != nil {\n\t\tvalues := slices.Clone(*opt.%s.%s)\n\t\tclone.%s.%s = &values\n\t}", fieldName, valueFieldName, fieldName, valueFieldName, fieldName, valueFieldName)
	}
	return fmt.Sprintf("clone.%s.%s = slices.Clone(opt.%s.%s)", fieldName, valueFieldName, fieldName, valueFieldName)
}

func optionFunctionSignature(optionSet *OptionSet, optionFunction *OptionFunction) string {
	if optionSet == nil || optionFunction == nil {
		return ""
	}
	parameter := optionFunction.singleParameter()
	if parameter != nil && reflection.IsEmpty(optionFunction.FunctionName) && !optionFunctionUsesBooleanShortcut(optionFunction) {
		return fmt.Sprintf("value %s", optionParameterType(optionSetStructName(optionSet.Name), parameter))
	}
	return strings.Join(collection.Map(optionFunction.Parameters, func(parameter OptionParameter) string {
		copied := parameter
		return fmt.Sprintf("%s %s", optionFunctionArgumentName(optionFunction, &copied), optionParameterType(optionSetStructName(optionSet.Name), &copied))
	}), ", ")
}

func optionFunctionArgumentName(optionFunction *OptionFunction, parameter *OptionParameter) string {
	if optionFunction == nil || parameter == nil {
		return ""
	}
	if optionFunction.singleParameter() != nil && reflection.IsEmpty(optionFunction.FunctionName) && !optionFunctionUsesBooleanShortcut(optionFunction) {
		return "value"
	}
	return toCamelCase(parameter.Name)
}

func optionFunctionAssignments(optionFunction *OptionFunction) string {
	if optionFunction == nil {
		return ""
	}
	return strings.Join(collection.Map(optionFunction.Parameters, func(parameter OptionParameter) string {
		copied := parameter
		return fmt.Sprintf("%s = %s\n\t\t%s = true", optionParameterValueSelector(&copied, "opt"), optionFunctionArgumentName(optionFunction, &copied), optionParameterEnabledSelector(&copied, "opt"))
	}), "\n\t\t")
}

func optionSetConstraintValidationRules(optionSet *OptionSet) []string {
	if optionSet == nil {
		return nil
	}
	return collection.Map(optionSet.ConstraintValidationRules(), func(rule string) string {
		return strings.ReplaceAll(rule, "&cfg.", "&opt.")
	})
}

func optionSetIdentifiers(optionSet *OptionSet) []generatedOptionIdentifier {
	if optionSet == nil {
		return nil
	}
	identifiers := []generatedOptionIdentifier{{
		Name:   optionSetStructName(optionSet.Name),
		Source: fmt.Sprintf("options structure for %q", optionSet.Name),
	}, {
		Name:   optionSetOptionTypeName(optionSet.Name),
		Source: fmt.Sprintf("option function type for %q", optionSet.Name),
	}, {
		Name:   fmt.Sprintf("Default%s", optionSetStructName(optionSet.Name)),
		Source: fmt.Sprintf("default constructor for %q", optionSet.Name),
	}, {
		Name:   optionSetBuilderFunctionName(optionSet.Name),
		Source: fmt.Sprintf("options builder for %q", optionSet.Name),
	}}
	for i := range optionSet.Functions {
		optionFunction := &optionSet.Functions[i]
		identifiers = append(identifiers, generatedOptionIdentifier{
			Name:   optionFunctionName(optionFunction),
			Source: fmt.Sprintf("option helper for %q function %d", optionSet.Name, i+1),
		})
		if withoutName := optionWithoutFunctionName(optionFunction); reflection.IsNotEmpty(withoutName) {
			identifiers = append(identifiers, generatedOptionIdentifier{
				Name:   withoutName,
				Source: fmt.Sprintf("negative option helper for %q function %d", optionSet.Name, i+1),
			})
		}
	}
	ownerName := optionSetStructName(optionSet.Name)
	for _, parameter := range optionSet.Parameters() {
		parameter := parameter
		identifiers = append(identifiers, generatedOptionIdentifier{
			Name:   optionParameterWrapperTypeName(ownerName, &parameter),
			Source: fmt.Sprintf("wrapper type for %q.%s", optionSet.Name, parameter.Name),
		})
		if !(&parameter).HasEnum() {
			continue
		}
		enumType := enumTypeName(ownerName, parameter.Name, baseLangType(parameter.LangType), parameter.EffectiveEnum())
		identifiers = append(identifiers, generatedOptionIdentifier{
			Name:   enumType,
			Source: fmt.Sprintf("enum type for %q.%s", optionSet.Name, parameter.Name),
		})
		for _, value := range parameter.EffectiveEnum() {
			identifiers = append(identifiers, generatedOptionIdentifier{
				Name:   enumConstName(enumType, value),
				Source: fmt.Sprintf("enum value for %q.%s", optionSet.Name, parameter.Name),
			})
		}
	}
	return identifiers
}
