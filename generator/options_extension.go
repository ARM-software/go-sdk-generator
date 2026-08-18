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

const defaultOptionPackage = "options"

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
	if parameter.Reference {
		return fmt.Sprintf("if opt.%s != nil {\n\t\tvalues := slices.Clone(*opt.%s)\n\t\tclone.%s = &values\n\t}", fieldName, fieldName, fieldName)
	}
	return fmt.Sprintf("clone.%s = slices.Clone(opt.%s)", fieldName, fieldName)
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
		return fmt.Sprintf("opt.%s = %s", toPascalCase(copied.Name), optionFunctionArgumentName(optionFunction, &copied))
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
