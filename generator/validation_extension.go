package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

func effectivePatternTemplateValue(value any) *string {
	switch typed := value.(type) {
	case CommandArgument:
		return (&typed).EffectivePattern()
	case *CommandArgument:
		return typed.EffectivePattern()
	case EntityField:
		return (&typed).EffectivePattern()
	case *EntityField:
		return typed.EffectivePattern()
	case ConfigurationOption:
		return (&typed).EffectivePattern()
	case *ConfigurationOption:
		return typed.EffectivePattern()
	case FunctionConfigFields:
		return (&typed).EffectivePattern()
	case *FunctionConfigFields:
		return typed.EffectivePattern()
	case OptionParameter:
		return (&typed).EffectivePattern()
	case *OptionParameter:
		return typed.EffectivePattern()
	default:
		return nil
	}
}

func effectivePatternStringTemplateValue(value any) string {
	return field.OptionalString(effectivePatternTemplateValue(value), "")
}

func optionalConstraintLiteral[T any](value *T) string {
	if value == nil {
		return ""
	}
	return goLiteral(*value)
}

func constraintValidationRule(langType string, list bool, iterator bool, constraints *ParameterConstraints) string {
	if constraints == nil {
		return ""
	}
	rules := make([]string, 0, 18)
	if constraints.NotEmpty {
		rules = append(rules, "validationUtils.NotEmpty()")
	}
	if reflection.IsNotEmpty(constraints.Contains) {
		rules = append(rules, fmt.Sprintf("validationUtils.ContainsString(%s)", strconv.Quote(field.OptionalString(constraints.Contains, ""))))
	}
	if reflection.IsNotEmpty(constraints.NotContains) {
		rules = append(rules, fmt.Sprintf("validationUtils.NotContains(%s)", strconv.Quote(field.OptionalString(constraints.NotContains, ""))))
	}
	if reflection.IsNotEmpty(constraints.Prefix) {
		rules = append(rules, fmt.Sprintf("validationUtils.Prefix(%s)", strconv.Quote(field.OptionalString(constraints.Prefix, ""))))
	}
	if reflection.IsNotEmpty(constraints.Suffix) {
		rules = append(rules, fmt.Sprintf("validationUtils.Suffix(%s)", strconv.Quote(field.OptionalString(constraints.Suffix, ""))))
	}
	if constraints.MinLength != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.MinLength(%s)", optionalConstraintLiteral(constraints.MinLength)))
	}
	if constraints.MaxLength != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.MaxLength(%s)", optionalConstraintLiteral(constraints.MaxLength)))
	}
	if constraints.Minimum != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.Minimum(%s)", optionalConstraintLiteral(constraints.Minimum)))
	}
	if constraints.Maximum != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.Maximum(%s)", optionalConstraintLiteral(constraints.Maximum)))
	}
	if constraints.ExclusiveMinimum != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.ExclusiveMinimum(%s)", optionalConstraintLiteral(constraints.ExclusiveMinimum)))
	}
	if constraints.ExclusiveMaximum != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.ExclusiveMaximum(%s)", optionalConstraintLiteral(constraints.ExclusiveMaximum)))
	}
	if constraints.MultipleOf != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.MultipleOf(%s)", optionalConstraintLiteral(constraints.MultipleOf)))
	}
	if constraints.MinItems != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.MinItems(%s)", optionalConstraintLiteral(constraints.MinItems)))
	}
	if constraints.MaxItems != nil {
		rules = append(rules, fmt.Sprintf("validationUtils.MaxItems(%s)", optionalConstraintLiteral(constraints.MaxItems)))
	}
	if constraints.AtLeastOne {
		rules = append(rules, "validationUtils.MinItems(1)")
	}
	if constraints.ExactlyOne {
		rules = append(rules, "validationUtils.MinItems(1)", "validationUtils.MaxItems(1)")
	}
	if constraints.UniqueItems {
		rules = append(rules, `validation.By(func(value any) error { return validationUtils.UniqueItems[any, string](func(item any) string { return fmt.Sprintf("%#v", item) }).Validate(value) })`)
	}
	if reflection.IsNotEmpty(constraints.Extension) {
		rules = append(rules, fmt.Sprintf("filesystem.NewOSPathExtensionRule(true, %s)", strings.Join(collection.Map(constraints.Extension, strconv.Quote), ", ")))
	}
	rules = collection.UniqueBy(rules, func(rule string) string { return rule })
	if len(rules) == 0 {
		return ""
	}
	if constraints.Nullable {
		if len(rules) == 1 {
			return fmt.Sprintf(", validationUtils.Nullable(%s)", rules[0])
		}
		return fmt.Sprintf(", validationUtils.Nullable(validationUtils.AllOf(%s))", strings.Join(rules, ", "))
	}
	return ", " + strings.Join(rules, ", ")
}

func usesPresenceGuardForValidation(langType string) bool {
	switch strings.TrimSpace(baseLangType(langType)) {
	case string(BuiltInLangTypeFilepath), string(BuiltInLangTypeExistingFilepath):
		return true
	default:
		return false
	}
}

func argValidationRule(s string) string {
	return field.OptionalString(func(st string) *string {
		switch st {
		case "[]byte{}", "hex":
			return field.ToOptionalString(", is.Hexadecimal")
		case "base64", "byte":
			return field.ToOptionalString(", is.Base64")
		case "host", "url":
			return field.ToOptionalString(", is.URL")
		case "uri":
			return field.ToOptionalString(", validationUtils.IsURI")
		case "version", "semver":
			return field.ToOptionalString(", is.Semver")
		case "uuid":
			return field.ToOptionalString(", is.UUID")
		case "email":
			return field.ToOptionalString(", is.Email")
		case "json":
			return field.ToOptionalString(", is.JSON")
		case "validJSON":
			return field.ToOptionalString(", jsonutils.IsValidJSON")
		case "ip":
			return field.ToOptionalString(", is.IP")
		case "datetime":
			return field.ToOptionalString(", validationUtils.IsRFC3339Timestamp")
		case string(BuiltInLangTypeFilepath):
			return field.ToOptionalString(", filesystem.NewOSPathValidationRule(true)")
		case string(BuiltInLangTypeExistingFilepath):
			return field.ToOptionalString(", filesystem.NewOSPathExistRule(true)")
		case string(BuiltInLangTypePathParameter):
			return field.ToOptionalString(", validationUtils.IsPathParameter")
		case "port":
			return field.ToOptionalString(", validationUtils.IsPort")
		}
		return nil
	}(strings.TrimSpace(s)), "")
}
