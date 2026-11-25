/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	defaultOptionPackage       = "options"
	optionEnabledFieldName     = "Enabled"
	optionAsListMethodName     = "AsList"
	optionAsIteratorMethodName = "AsIterator"
)

var optionConditionReferencePattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)(.*)$`)

type generatedOptionIdentifier struct {
	Name   string
	Source string
}

// EnablesTogether reports whether one Enabled flag selects the entire option set.
// An omitted mode uses the independent behaviour, including Go zero values.
func (optionSet *OptionSet) EnablesTogether() bool {
	return optionSet != nil && optionSet.EnableMode == OptionSetEnableModeTogether
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

func generatedOptionSetForObject(spec *SDKSpecification, langType string) *OptionSet {
	if spec == nil {
		return nil
	}
	name := convertLangType(baseLangType(langType))
	for i := range spec.Options {
		optionSet := &spec.Options[i]
		if name == fmt.Sprintf("%s.%s", optionPackageName(optionSet), optionSetStructName(optionSet.Name)) {
			return optionSet
		}
	}
	return nil
}

// processOptionObjectShapes keeps editable CLI objects separate from the
// functional options consumed by SDK methods.
func processOptionObjectShapes(spec *SDKSpecification) error {
	for i := range spec.Interface.Methods {
		for j := range spec.Interface.Methods[i].Args {
			arg := &spec.Interface.Methods[i].Args[j]
			optionSet := generatedOptionSetForObject(spec, arg.LangType)
			if optionSet == nil {
				continue
			}
			if err := validateParameterShape(arg.Name, arg.LangType, arg.List, arg.Iterator); err != nil {
				return err
			}
			arg.LangType = fmt.Sprintf("%s.%s", optionPackageName(optionSet), optionSetOptionTypeName(optionSet.Name))
			arg.Reference = false
			arg.List = !arg.Iterator
		}
	}
	if spec.CommandLineCommands != nil {
		for i := range spec.CommandLineCommands.Functions {
			for j := range spec.CommandLineCommands.Functions[i].Configuration {
				cfg := &spec.CommandLineCommands.Functions[i].Configuration[j]
				if generatedOptionSetForObject(spec, cfg.LangType) == nil {
					continue
				}
				if err := validateParameterShape(cfg.Name, cfg.LangType, cfg.List, cfg.Iterator); err != nil {
					return err
				}
				cfg.List, cfg.Iterator = false, false
			}
		}
	}
	return nil
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

func optionSetHasConditionalValidationRules(optionSet *OptionSet) bool {
	return optionSet != nil && collection.AnyFunc(optionSet.Parameters(), func(parameter OptionParameter) bool {
		return (&parameter).HasCondition()
	})
}

// optionParameterConditionExpression rewrites an option parameter condition so it
// can be evaluated against generated wrapper fields.
//
// Option parameter conditions are authored against the logical value name, such
// as `UsePaging == true`, but generated option structs wrap every parameter in a
// `{ValueField, Enabled}` pair. This helper rewrites the leading parameter
// reference to check both that the source option has been selected and that its
// wrapped value satisfies the original expression.
//
// For example, if `PageToken` has condition `UsePaging == true`, the generated
// validation should evaluate `opt.UsePaging.Enabled && opt.UsePaging.UsePaging == true`
// rather than incorrectly appending `.Enabled` to the full expression.
func optionParameterConditionExpression(optionSet *OptionSet, condition, receiver string) string {
	trimmedCondition := strings.TrimSpace(condition)
	if reflection.IsEmpty(trimmedCondition) {
		return ""
	}
	fallback := trimmedCondition
	if reflection.IsNotEmpty(receiver) {
		fallback = fmt.Sprintf("%s.%s", receiver, trimmedCondition)
	}
	if optionSet == nil || reflection.IsEmpty(receiver) {
		return fallback
	}
	matches := optionConditionReferencePattern.FindStringSubmatch(trimmedCondition)
	if len(matches) != 3 {
		return fallback
	}
	referenceName := matches[1]
	suffix := matches[2]
	for functionIndex := range optionSet.Functions {
		for parameterIndex := range optionSet.Functions[functionIndex].Parameters {
			parameter := &optionSet.Functions[functionIndex].Parameters[parameterIndex]
			if !strings.EqualFold(parameter.Name, referenceName) {
				continue
			}
			if optionSet.EnablesTogether() {
				return fmt.Sprintf("%s%s", optionSetParameterValueSelector(optionSet, parameter, receiver), suffix)
			}
			return fmt.Sprintf("%s && %s%s", optionParameterEnabledSelector(parameter, receiver), optionParameterValueSelector(parameter, receiver), suffix)
		}
	}
	return fallback
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
	if optionSet := generatedOptionSetForObject(spec, cfg.LangType); optionSet != nil {
		// CLI bindings edit one settings object; conversion happens at the SDK call.
		cfg.List, cfg.Iterator = false, false
		if reflection.IsEmpty(cfg.DefaultValue) {
			defaultExpr := fmt.Sprintf("%s.Default%s()", optionPackageName(optionSet), optionSetStructName(optionSet.Name))
			if !cfg.Reference && !isPointerLangType(cfg.LangType) {
				defaultExpr = "*" + defaultExpr
			}
			cfg.DefaultValue = field.ToOptionalString(defaultExpr)
		}
		return
	}
	enabledTag := strings.ToLower(optionEnabledFieldName)
	for _, optionSet := range spec.Options {
		structName := optionSetStructName(optionSet.Name)
		if optionSet.EnablesTogether() {
			continue
		}
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
		baseType = enumValueType(ownerName, parameter.Name, baseLangType(parameter.LangType), parameter.List, parameter.Iterator, parameter.EffectiveEnum())
	}
	return parameterTypeFromBase(parameter.Reference, parameter.List, parameter.Iterator, baseType)
}

func optionFieldEnumValidationRule(ownerName string, parameter *OptionParameter, fieldExpr string) string {
	enum := parameter.EffectiveEnum()
	if !parameter.List && !parameter.Iterator {
		return fieldEnumValidationRule(ownerName, parameter.Name, parameter.LangType, parameter.Reference, false, false, parameter.IsRequired(), enum, fieldExpr)
	}
	if !hasEnum(enum) {
		return ""
	}
	rule := wrapCollectionItemRule(parameter.List, parameter.Iterator, fmt.Sprintf("validationUtils.Enum(%s)", enumStringConstRefs(ownerName, parameter.Name, parameter.LangType, enum)))
	if parameter.IsRequired() {
		return ", " + rule
	}
	return fmt.Sprintf(", validation.When(reflection.IsNotEmpty(%s), %s)", fieldExpr, rule)
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

func optionSetParameterValueSelector(optionSet *OptionSet, parameter *OptionParameter, receiver string) string {
	if optionSet.EnablesTogether() {
		return fmt.Sprintf("%s.%s", receiver, toPascalCase(parameter.Name))
	}
	return optionParameterValueSelector(parameter, receiver)
}

func optionSetParameterEnabledSelector(optionSet *OptionSet, parameter *OptionParameter, receiver string) string {
	if optionSet.EnablesTogether() {
		return receiver + ".Enabled"
	}
	return optionParameterEnabledSelector(parameter, receiver)
}

func optionSetFunctionAssignments(optionSet *OptionSet, optionFunction *OptionFunction) string {
	if !optionSet.EnablesTogether() {
		return optionFunctionAssignments(optionFunction)
	}
	assignments := collection.Map(optionFunction.Parameters, func(parameter OptionParameter) string {
		return fmt.Sprintf("%s = %s", optionSetParameterValueSelector(optionSet, &parameter, "opt"), optionFunctionArgumentName(optionFunction, &parameter))
	})
	return strings.Join(append(assignments, "opt.Enabled = true"), "\n\t\t")
}

func optionSetParameterCloneStatements(optionSet *OptionSet, parameter *OptionParameter) string {
	if !optionSet.EnablesTogether() {
		return optionParameterCloneStatements(parameter)
	}
	if !parameter.List || parameter.Iterator {
		return ""
	}
	name := toPascalCase(parameter.Name)
	if parameter.Reference {
		return fmt.Sprintf("if opt.%s != nil { values := slices.Clone(*opt.%s); clone.%s = &values }", name, name, name)
	}
	return fmt.Sprintf("clone.%s = slices.Clone(opt.%s)", name, name)
}

func togetherOptionParameterDefaultValue(additionalTypesPackages map[string]string, currentPackage, toolName, toolVersion, ownerName string, parameter *OptionParameter) string {
	value := field.OptionalString(parameter.DefaultValue, "")
	if !parameter.List || parameter.Iterator {
		return defaultFieldValue(additionalTypesPackages, currentPackage, toolName, toolVersion, ownerName, parameter.Name, parameter.LangType, parameter.EffectiveEnum(), value, parameter.Reference)
	}
	var item string
	if parameter.HasEnum() {
		// Collection enums use literal strings, without scalar enum constants or semantic defaults.
		item = strconv.Quote(value)
	} else {
		item = defaultFieldValue(additionalTypesPackages, currentPackage, toolName, toolVersion, ownerName, parameter.Name, parameter.LangType, nil, value, false)
	}
	literal := fmt.Sprintf("%s{%s}", strings.TrimPrefix(optionParameterType(ownerName, parameter), "*"), item)
	if parameter.Reference {
		return "&" + literal
	}
	return literal
}

func optionSetConstraintValidationRules(optionSet *OptionSet) []string {
	if optionSet == nil {
		return nil
	}
	return optionSet.ConstraintValidationRules()
}

// optionSetConstraintParameters includes the generated selection field when
// resolving together-mode constraints, without generating it as a parameter.
func optionSetConstraintParameters(optionSet *OptionSet) []OptionParameter {
	parameters := optionSet.Parameters()
	if optionSet.EnablesTogether() {
		parameters = append(parameters, OptionParameter{Name: optionEnabledFieldName, LangType: "bool"})
	}
	return parameters
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
		if !optionSet.EnablesTogether() {
			identifiers = append(identifiers, generatedOptionIdentifier{
				Name:   optionParameterWrapperTypeName(ownerName, &parameter),
				Source: fmt.Sprintf("wrapper type for %q.%s", optionSet.Name, parameter.Name),
			})
		}
		if !parameter.HasEnum() {
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
