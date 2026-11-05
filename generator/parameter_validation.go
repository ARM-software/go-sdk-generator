/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
	validationUtils "github.com/ARM-software/golang-utils/utils/validation"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var conditionLeadingIdentifierPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\b`)

func validateParameterShape(name string, langType string, list bool, iterator bool) error {

	if list && iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot set both list and iterator", name)

	}

	if isVariadicLangType(langType) && (list || iterator) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot combine an explicit variadic lang_type %q with list or iterator", name, langType)

	}

	return nil

}

func validateStringEnumDefinition(name string, langType string, reference bool, list bool, iterator bool, enum []string) error {

	if len(enum) == 0 {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot define enum values together with list or iterator", name)

	}

	baseType := strings.TrimSpace(baseLangType(langType))

	if baseType != stringLangType {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define enum values for string or *string types", name)

	}

	if reference && isPointerLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s should use either reference or *string for enum values, not both", name)

	}

	return nil

}

func validateStringEnumDefaultValue(name string, enum []string, defaultValue *string) error {

	if len(enum) == 0 || reflection.IsEmpty(defaultValue) {

		return nil

	}

	value := field.OptionalString(defaultValue, "")

	if !collection.In(enum, value, collection.StringCleanCaseInsensitiveMatch) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s default value %q must be one of %v", name, value, enum)

	}

	return nil

}

func validateEnumDefaultValueRequired(name string, enum []string, defaultValue *string) error {

	if len(enum) == 0 {

		return nil

	}

	if reflection.IsEmpty(defaultValue) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s defines enum values, so it must define a default value", name)

	}

	return nil

}

func normalisedStringValues(values []string) []string {

	normalised := collection.Filter(collection.Map(values, strings.TrimSpace), func(value string) bool {

		return reflection.IsNotEmpty(value)

	})

	normalised = collection.UniqueBy(normalised, func(value string) string { return value })

	slices.Sort(normalised)

	return normalised

}

func sameStringValues(left []string, right []string) bool {

	return slices.Equal(normalisedStringValues(left), normalisedStringValues(right))

}

func isStringConstraintCompatibleLangType(langType string) bool {

	return strings.TrimSpace(convertLangType(baseLangType(langType))) == stringLangType

}

func isNumericConstraintCompatibleLangType(langType string) bool {

	switch strings.TrimSpace(baseLangType(langType)) {

	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "decimal", "port":

		return true

	default:

		return false

	}

}

func validatePatternDefinition(name string, langType string, list bool, iterator bool, pattern *string) error {

	if reflection.IsEmpty(pattern) {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define pattern for scalar string-compatible values", name)

	}

	if !isStringConstraintCompatibleLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define pattern for string-compatible values", name)

	}

	if _, err := regexp.Compile(field.OptionalString(pattern, "")); err != nil {

		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s pattern must be a valid regular expression", name)

	}

	return nil

}

func validateStringSubstringConstraintDefinition(name string, langType string, list bool, iterator bool, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	definesStringSubstringConstraint := reflection.IsNotEmpty(field.OptionalString(constraints.Contains, "")) || reflection.IsNotEmpty(field.OptionalString(constraints.NotContains, "")) || reflection.IsNotEmpty(field.OptionalString(constraints.Prefix, "")) || reflection.IsNotEmpty(field.OptionalString(constraints.Suffix, ""))

	if !definesStringSubstringConstraint {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define contains, not_contains, prefix, and suffix for scalar string-compatible values", name)

	}

	if !isStringConstraintCompatibleLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define contains, not_contains, prefix, and suffix for string-compatible values", name)

	}

	return nil

}

func validateMirroredConstraintDefinitions(name string, directEnum []string, directPattern *string, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	if hasEnum(directEnum) && hasEnum(constraints.Enum) && !sameStringValues(directEnum, constraints.Enum) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot define conflicting enum values at both the top level and under constraints", name)

	}

	topLevelPattern := strings.TrimSpace(field.OptionalString(directPattern, ""))

	constraintPattern := strings.TrimSpace(field.OptionalString(constraints.Pattern, ""))

	if reflection.IsNotEmpty(topLevelPattern) && reflection.IsNotEmpty(constraintPattern) && topLevelPattern != constraintPattern {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot define conflicting patterns at both the top level and under constraints", name)

	}

	return nil

}

func validateStringConstraintDefinition(name string, langType string, list bool, iterator bool, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	if constraints.MinLength == nil && constraints.MaxLength == nil {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define min_length and max_length for scalar string-compatible values", name)

	}

	if !isStringConstraintCompatibleLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define min_length and max_length for string-compatible values", name)

	}

	if constraints.MinLength != nil && constraints.MaxLength != nil && *constraints.MinLength > *constraints.MaxLength {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s min_length must be less than or equal to max_length", name)

	}

	return nil

}

func numericConstraintValue(value *float64) (float64, bool) {

	if value == nil {

		return 0, false

	}

	return *value, true

}

func validateNumericConstraintDefinition(name string, langType string, list bool, iterator bool, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	definesNumericConstraint := constraints.Minimum != nil || constraints.Maximum != nil || constraints.ExclusiveMinimum != nil || constraints.ExclusiveMaximum != nil || constraints.MultipleOf != nil

	if !definesNumericConstraint {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define numeric constraints for scalar numeric values", name)

	}

	if !isNumericConstraintCompatibleLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define numeric constraints for numeric values", name)

	}

	minimum, hasMinimum := numericConstraintValue(constraints.Minimum)

	maximum, hasMaximum := numericConstraintValue(constraints.Maximum)

	exclusiveMinimum, hasExclusiveMinimum := numericConstraintValue(constraints.ExclusiveMinimum)

	exclusiveMaximum, hasExclusiveMaximum := numericConstraintValue(constraints.ExclusiveMaximum)

	if hasMinimum && hasMaximum && minimum > maximum {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s minimum must be less than or equal to maximum", name)

	}

	if hasExclusiveMinimum && hasExclusiveMaximum && exclusiveMinimum >= exclusiveMaximum {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s exclusive_minimum must be less than exclusive_maximum", name)

	}

	if hasMinimum && hasExclusiveMaximum && minimum >= exclusiveMaximum {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s minimum must be less than exclusive_maximum", name)

	}

	if hasExclusiveMinimum && hasMaximum && exclusiveMinimum >= maximum {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s exclusive_minimum must be less than maximum", name)

	}

	if constraints.MultipleOf != nil && *constraints.MultipleOf <= 0 {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s multiple_of must be strictly positive", name)

	}

	return nil

}

func validateCollectionConstraintDefinition(name string, list bool, iterator bool, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	definesCollectionConstraint := constraints.MinItems != nil || constraints.MaxItems != nil || constraints.UniqueItems || constraints.AtLeastOne || constraints.ExactlyOne

	if !definesCollectionConstraint {

		return nil

	}

	if !list && !iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define min_items, max_items, unique_items, at_least_one, or exactly_one for list or iterator values", name)

	}

	if constraints.MinItems != nil && constraints.MaxItems != nil && *constraints.MinItems > *constraints.MaxItems {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s min_items must be less than or equal to max_items", name)

	}

	if constraints.AtLeastOne && constraints.MaxItems != nil && *constraints.MaxItems < 1 {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s at_least_one conflicts with max_items less than 1", name)

	}

	if constraints.ExactlyOne {

		if constraints.MinItems != nil && *constraints.MinItems > 1 {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s exactly_one conflicts with min_items greater than 1", name)

		}

		if constraints.MaxItems != nil && *constraints.MaxItems < 1 {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s exactly_one conflicts with max_items less than 1", name)

		}

	}

	return nil

}

func validateExtensionConstraintDefinition(name string, langType string, list bool, iterator bool, constraints *ParameterConstraints) error {

	if constraints == nil || reflection.IsEmpty(constraints.Extension) {

		return nil

	}

	if list || iterator {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define extension for scalar string-compatible path values", name)

	}

	if !isStringConstraintCompatibleLangType(langType) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s can only define extension for string-compatible path values", name)

	}

	normalised := normalisedStringValues(constraints.Extension)

	if len(normalised) == 0 {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s extension must define at least one non-empty extension", name)

	}

	return nil

}

func validateExamplesDefinition(name string, langType string, list bool, iterator bool, enum []string, example *string, examples []string) error {

	if reflection.IsNotEmpty(example) && reflection.IsNotEmpty(examples) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s cannot define both example and examples", name)

	}

	if reflection.IsNotEmpty(example) {

		if err := validateExampleValue(fmt.Sprintf("%s example", name), langType, list, iterator, field.OptionalString(example, "")); err != nil {

			return err

		}

	}

	for index, value := range collection.Enumerate(examples) {

		if err := validateExampleValue(fmt.Sprintf("%s examples[%d]", name, index), langType, list, iterator, value); err != nil {

			return err

		}

	}

	values := append(collection.Filter([]string{field.OptionalString(example, "")}, func(v string) bool { return reflection.IsNotEmpty(strings.TrimSpace(v)) }), examples...)

	if err := collection.ForAll(values, func(value string) error {

		if len(enum) > 0 && !collection.In(enum, value, collection.StringCleanCaseInsensitiveMatch) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s example value %q must be one of %v", name, value, enum)

		}

		return nil

	}); err != nil {

		return err

	}

	return nil

}

func validateExampleValue(name string, langType string, list bool, iterator bool, value string) error {

	baseType := strings.TrimSpace(baseLangType(langType))

	if list || iterator {

		baseType = strings.TrimSpace(baseLangType(baseType))

	}

	value = strings.TrimSpace(value)

	if reflection.IsEmpty(value) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must not be empty", name)

	}

	switch baseType {

	case stringLangType:

		return nil

	case "bool":

		_, err := strconv.ParseBool(value)

		return wrapExampleParseError(name, err)

	case "int", "int8", "int16", "int32", "int64":

		_, err := strconv.ParseInt(value, 10, 64)

		return wrapExampleParseError(name, err)

	case "uint", "uint8", "uint16", "uint32", "uint64", "port":

		_, err := strconv.ParseUint(value, 10, 64)

		return wrapExampleParseError(name, err)

	case "float32", "float64", "decimal":

		_, err := strconv.ParseFloat(value, 64)

		return wrapExampleParseError(name, err)

	case "duration", "time.Duration":

		_, err := time.ParseDuration(value)

		return wrapExampleParseError(name, err)

	case "uuid":

		return wrapExampleValidationError(name, validation.Validate(value, is.UUID))

	case "email":

		return wrapExampleValidationError(name, validation.Validate(value, is.Email))

	case "host", "url", "uri":

		return wrapExampleValidationError(name, validation.Validate(value, is.URL))

	case "version", "semver":

		return wrapExampleValidationError(name, validation.Validate(value, is.Semver))

	case "json":

		var tmp any

		return wrapExampleParseError(name, json.Unmarshal([]byte(value), &tmp))

	case "datetime":

		_, err := time.Parse(time.RFC3339, value)

		return wrapExampleParseError(name, err)

	case "regex":

		_, err := regexp.Compile(value)

		return wrapExampleParseError(name, err)

	default:

		return nil

	}

}

func wrapExampleParseError(name string, err error) error {

	if err == nil {

		return nil

	}

	return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s does not comply with its lang_type", name)

}

func wrapExampleValidationError(name string, err error) error {

	if err == nil {

		return nil

	}

	return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s does not comply with its lang_type", name)

}

func (notice *DeprecationNotice) Validate() error {

	if notice == nil {

		return nil

	}

	if err := configUtils.ValidateEmbedded(notice); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in deprecation notice")

	}

	err := validation.ValidateStruct(notice,

		validation.Field(&notice.IssueDate,

			validation.Required,

			validationUtils.IsRFC3339Timestamp,
		),

		validation.Field(&notice.RemovalDate,

			validationUtils.NilTimestampOrNotEmpty,

			validation.When(reflection.IsNotEmpty(notice.RemovalDate),

				validationUtils.IsRFC3339Timestamp,

				validationUtils.TimestampMinimum(notice.IssueDate),
			),
		),

		validation.Field(&notice.Description,

			validation.NilOrNotEmpty.Error("deprecation notice description must not be empty when provided"),
		),

		validation.Field(&notice.Replacement,

			validation.NilOrNotEmpty.Error("deprecation notice replacement must not be empty when provided"),
		),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate deprecation notice")

	}

	return nil

}

func (b *Binding) Validate() error {

	if b == nil {

		return commonerrors.UndefinedVariable("binding")

	}

	if err := configUtils.ValidateEmbedded(b); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in binding")

	}

	if err := validateParameterShape(fmt.Sprintf("binding %q", field.OptionalString(b.Name, "")), field.OptionalString(b.LangType, ""), b.List, b.Iterator); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate binding shape")

	}

	if err := validateExamplesDefinition(fmt.Sprintf("binding %q", field.OptionalString(b.Name, "")), field.OptionalString(b.LangType, ""), b.List, b.Iterator, nil, b.Example, b.Examples); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate binding examples")

	}

	return nil

}

func (arg *CommandArgument) Validate() error {

	if arg == nil {

		return commonerrors.UndefinedVariable("command argument")

	}

	if err := configUtils.ValidateEmbedded(arg); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in command argument")

	}

	err := validation.ValidateStruct(arg,

		validation.Field(&arg.LangType, validation.By(func(value any) error {

			langType, _ := value.(string)

			return validateParameterShape(fmt.Sprintf("command argument %q", arg.Name), langType, arg.List, arg.Iterator)

		})),

		validation.Field(&arg.DeprecationNotice, validation.By(func(value any) error {

			return arg.DeprecationNotice.Validate()

		})),

		validation.Field(&arg.Example, validation.By(func(value any) error {

			return validateExamplesDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.List, arg.Iterator, arg.EffectiveEnum(), arg.Example, arg.Examples)

		})),

		validation.Field(&arg.Enum, validation.By(func(value any) error {

			return validateStringEnumDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.Reference, arg.List, arg.Iterator, arg.Enum)

		})),

		validation.Field(&arg.DefaultValue, validation.By(func(value any) error {

			if err := validateEnumDefaultValueRequired(fmt.Sprintf("command argument %q", arg.Name), arg.EffectiveEnum(), arg.DefaultValue); err != nil {

				return err

			}

			return validateStringEnumDefaultValue(fmt.Sprintf("command argument %q", arg.Name), arg.EffectiveEnum(), arg.DefaultValue)

		})),

		validation.Field(&arg.Bindings, validation.By(func(value any) error {

			return collection.ForAll(arg.Bindings, func(binding Binding) error {

				return (&binding).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate command argument")

	}

	if err := validateParameterConstraintsDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.Reference, arg.List, arg.Iterator, arg.Enum, arg.Pattern, arg.Constraints); err != nil {

		return err

	}

	return nil

}

func (cfg *ConfigurationOption) Validate() error {

	if cfg == nil {

		return commonerrors.UndefinedVariable("configuration option")

	}

	if err := configUtils.ValidateEmbedded(cfg); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in configuration option")

	}

	err := validation.ValidateStruct(cfg,

		validation.Field(&cfg.LangType, validation.By(func(value any) error {

			langType, _ := value.(string)

			return validateParameterShape(fmt.Sprintf("global configuration %q", cfg.Name), langType, cfg.List, cfg.Iterator)

		})),

		validation.Field(&cfg.DeprecationNotice, validation.By(func(value any) error {

			return cfg.DeprecationNotice.Validate()

		})),

		validation.Field(&cfg.Example, validation.By(func(value any) error {

			return validateExamplesDefinition(fmt.Sprintf("global configuration %q", cfg.Name), cfg.LangType, cfg.List, cfg.Iterator, cfg.EffectiveEnum(), cfg.Example, cfg.Examples)

		})),

		validation.Field(&cfg.Enum, validation.By(func(value any) error {

			return validateStringEnumDefinition(fmt.Sprintf("global configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)

		})),

		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {

			if err := validateEnumDefaultValueRequired(fmt.Sprintf("global configuration %q", cfg.Name), cfg.EffectiveEnum(), cfg.DefaultValue); err != nil {

				return err

			}

			return validateStringEnumDefaultValue(fmt.Sprintf("global configuration %q", cfg.Name), cfg.EffectiveEnum(), cfg.DefaultValue)

		})),

		validation.Field(&cfg.Bindings, validation.By(func(value any) error {

			return collection.ForAll(cfg.Bindings, func(binding Binding) error {

				return (&binding).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate configuration option")

	}

	if err := validateParameterConstraintsDefinition(fmt.Sprintf("configuration option %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum, cfg.Pattern, cfg.Constraints); err != nil {

		return err

	}

	return nil

}

func (fieldDef *EntityField) Validate() error {

	if fieldDef == nil {

		return commonerrors.UndefinedVariable("entity field")

	}

	if err := configUtils.ValidateEmbedded(fieldDef); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in entity field")

	}

	err := validation.ValidateStruct(fieldDef,

		validation.Field(&fieldDef.LangType, validation.By(func(value any) error {

			langType, _ := value.(string)

			return validateParameterShape(fmt.Sprintf("entity field %q", fieldDef.Name), langType, fieldDef.List, fieldDef.Iterator)

		})),

		validation.Field(&fieldDef.DeprecationNotice, validation.By(func(value any) error {

			return fieldDef.DeprecationNotice.Validate()

		})),

		validation.Field(&fieldDef.Example, validation.By(func(value any) error {

			return validateExamplesDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.List, fieldDef.Iterator, fieldDef.EffectiveEnum(), fieldDef.Example, fieldDef.Examples)

		})),

		validation.Field(&fieldDef.Enum, validation.By(func(value any) error {

			return validateStringEnumDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.Reference, fieldDef.List, fieldDef.Iterator, fieldDef.Enum)

		})),

		validation.Field(&fieldDef.DefaultValue, validation.By(func(value any) error {

			if err := validateEnumDefaultValueRequired(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.EffectiveEnum(), fieldDef.DefaultValue); err != nil {

				return err

			}

			return validateStringEnumDefaultValue(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.EffectiveEnum(), fieldDef.DefaultValue)

		})),

		validation.Field(&fieldDef.Bindings, validation.By(func(value any) error {

			return collection.ForAll(fieldDef.Bindings, func(binding Binding) error {

				return (&binding).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate entity field")

	}

	if err := validateParameterConstraintsDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.Reference, fieldDef.List, fieldDef.Iterator, fieldDef.Enum, fieldDef.Pattern, fieldDef.Constraints); err != nil {

		return err

	}

	return nil

}

func (cfg *FunctionConfigFields) Validate() error {

	if cfg == nil {

		return commonerrors.UndefinedVariable("function configuration field")

	}

	if err := configUtils.ValidateEmbedded(cfg); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in function configuration field")

	}

	err := validation.ValidateStruct(cfg,

		validation.Field(&cfg.LangType, validation.By(func(value any) error {

			langType, _ := value.(string)

			return validateParameterShape(fmt.Sprintf("function configuration %q", cfg.Name), langType, cfg.List, cfg.Iterator)

		})),

		validation.Field(&cfg.DeprecationNotice, validation.By(func(value any) error {

			return cfg.DeprecationNotice.Validate()

		})),

		validation.Field(&cfg.Example, validation.By(func(value any) error {

			return validateExamplesDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.List, cfg.Iterator, cfg.EffectiveEnum(), cfg.Example, cfg.Examples)

		})),

		validation.Field(&cfg.Enum, validation.By(func(value any) error {

			return validateStringEnumDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)

		})),

		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {

			if err := validateEnumDefaultValueRequired(fmt.Sprintf("function configuration %q", cfg.Name), cfg.EffectiveEnum(), cfg.DefaultValue); err != nil {

				return err

			}

			return validateStringEnumDefaultValue(fmt.Sprintf("function configuration %q", cfg.Name), cfg.EffectiveEnum(), cfg.DefaultValue)

		})),

		validation.Field(&cfg.Bindings, validation.By(func(value any) error {

			return collection.ForAll(cfg.Bindings, func(binding Binding) error {

				return (&binding).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate function configuration field")

	}

	if err := validateParameterConstraintsDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum, cfg.Pattern, cfg.Constraints); err != nil {

		return err

	}

	return nil

}

func (method *InterfaceMethod) Validate() error {

	if method == nil {

		return commonerrors.UndefinedVariable("interface method")

	}

	if err := configUtils.ValidateEmbedded(method); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in interface method")

	}

	err := validation.ValidateStruct(method,

		validation.Field(&method.DeprecationNotice, validation.By(func(value any) error {

			return method.DeprecationNotice.Validate()

		})),

		validation.Field(&method.Args, validation.By(func(value any) error {

			return collection.ForAll(method.Args, func(arg CommandArgument) error {

				return (&arg).Validate()

			})

		})),

		validation.Field(&method.Returns, validation.By(func(value any) error {

			return collection.ForAll(method.Returns, func(ret CommandArgument) error {

				return (&ret).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate interface method")

	}

	return nil

}

func (iface *Interface) Validate() error {

	if iface == nil {

		return commonerrors.UndefinedVariable("interface")

	}

	if err := configUtils.ValidateEmbedded(iface); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in interface")

	}

	err := validation.ValidateStruct(iface,

		validation.Field(&iface.DeprecationNotice, validation.By(func(value any) error {

			return iface.DeprecationNotice.Validate()

		})),

		validation.Field(&iface.Methods, validation.By(func(value any) error {

			return collection.ForAll(iface.Methods, func(method InterfaceMethod) error {

				return (&method).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate interface")

	}

	return nil

}

func (typ *Type) Validate() error {

	if typ == nil {

		return commonerrors.UndefinedVariable("type")

	}

	if err := configUtils.ValidateEmbedded(typ); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in type")

	}

	err := validation.ValidateStruct(typ,

		validation.Field(&typ.DeprecationNotice, validation.By(func(value any) error {

			return typ.DeprecationNotice.Validate()

		})),

		validation.Field(&typ.Package, validation.By(func(value any) error {

			return validateAdditionalTypePackageName(typ.Package)

		})),

		validation.Field(&typ.Fields, validation.By(func(value any) error {

			return collection.ForAll(typ.Fields, func(fieldDef EntityField) error {

				return (&fieldDef).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate type")

	}

	return nil

}

func (parameter *OptionParameter) Validate() error {

	if parameter == nil {

		return commonerrors.UndefinedVariable("option parameter")

	}

	if strings.EqualFold(optionParameterValueFieldName(parameter), optionEnabledFieldName) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "option parameter %q cannot use %s as its value field; set value_field.name to a different name", parameter.Name, optionEnabledFieldName)

	}

	if err := configUtils.ValidateEmbedded(parameter); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in option parameter")

	}

	err := validation.ValidateStruct(parameter,

		validation.Field(&parameter.LangType, validation.By(func(value any) error {

			langType, _ := value.(string)

			return validateParameterShape(fmt.Sprintf("option parameter %q", parameter.Name), langType, parameter.List, parameter.Iterator)

		})),

		validation.Field(&parameter.DeprecationNotice, validation.By(func(value any) error {

			return parameter.DeprecationNotice.Validate()

		})),

		validation.Field(&parameter.Example, validation.By(func(value any) error {

			return validateExamplesDefinition(fmt.Sprintf("option parameter %q", parameter.Name), parameter.LangType, parameter.List, parameter.Iterator, parameter.EffectiveEnum(), parameter.Example, parameter.Examples)

		})),

		validation.Field(&parameter.Enum, validation.By(func(value any) error {

			return validateStringEnumDefinition(fmt.Sprintf("option parameter %q", parameter.Name), parameter.LangType, parameter.Reference, parameter.List, parameter.Iterator, parameter.Enum)

		})),

		validation.Field(&parameter.DefaultValue, validation.By(func(value any) error {

			if err := validateEnumDefaultValueRequired(fmt.Sprintf("option parameter %q", parameter.Name), parameter.EffectiveEnum(), parameter.DefaultValue); err != nil {

				return err

			}

			return validateStringEnumDefaultValue(fmt.Sprintf("option parameter %q", parameter.Name), parameter.EffectiveEnum(), parameter.DefaultValue)

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate option parameter")

	}

	if err := validateParameterConstraintsDefinition(fmt.Sprintf("option parameter %q", parameter.Name), parameter.LangType, parameter.Reference, parameter.List, parameter.Iterator, parameter.Enum, parameter.Pattern, parameter.Constraints); err != nil {

		return err

	}

	return nil

}

func (optionFunction *OptionFunction) Validate() error {

	if optionFunction == nil {

		return commonerrors.UndefinedVariable("option function")

	}

	if err := configUtils.ValidateEmbedded(optionFunction); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in option function")

	}

	err := validation.ValidateStruct(optionFunction,

		validation.Field(&optionFunction.Parameters, validation.By(func(value any) error {

			return collection.ForAll(optionFunction.Parameters, func(parameter OptionParameter) error {

				return (&parameter).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate option function")

	}

	if len(optionFunction.Parameters) > 1 && reflection.IsEmpty(optionFunction.FunctionName) {

		return commonerrors.New(commonerrors.ErrInvalid, "option functions with more than one parameter must declare function_name")

	}

	return nil

}

func (optionSet *OptionSet) Validate() error {

	if optionSet == nil {

		return commonerrors.UndefinedVariable("option set")

	}

	if err := configUtils.ValidateEmbedded(optionSet); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in option set")

	}

	err := validation.ValidateStruct(optionSet,

		validation.Field(&optionSet.DeprecationNotice, validation.By(func(value any) error {

			return optionSet.DeprecationNotice.Validate()

		})),

		validation.Field(&optionSet.Package, validation.By(func(value any) error {

			return validateOptionPackageName(optionSet.Package)

		})),

		validation.Field(&optionSet.Functions, validation.By(func(value any) error {

			return collection.ForAll(optionSet.Functions, func(optionFunction OptionFunction) error {

				return (&optionFunction).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate option set")

	}

	if err := optionSet.ValidateOptionParameterNames(); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate option set parameters")

	}

	return nil

}

func (fct *CliFunction) Validate() error {

	if fct == nil {

		return commonerrors.UndefinedVariable("cli function")

	}

	if err := configUtils.ValidateEmbedded(fct); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in cli function")

	}

	err := validation.ValidateStruct(fct,

		validation.Field(&fct.DeprecationNotice, validation.By(func(value any) error {

			return fct.DeprecationNotice.Validate()

		})),

		validation.Field(&fct.Configuration, validation.By(func(value any) error {

			return collection.ForAll(fct.Configuration, func(cfg FunctionConfigFields) error {

				return (&cfg).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate cli function")

	}

	return nil

}

func (cmds *CommandLineCommands) Validate() error {

	if cmds == nil {

		return nil

	}

	if err := configUtils.ValidateEmbedded(cmds); err != nil {

		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in command line commands")

	}

	err := validation.ValidateStruct(cmds,
		validation.Field(&cmds.Binary, validation.By(func(value any) error {
			return validateGeneratedBinaryName(field.OptionalString(cmds.Binary, ""))
		})),
		validation.Field(&cmds.Package, validation.By(func(value any) error {
			return validateCLIPackageName(cmds.Package)
		})),
		validation.Field(&cmds.Functions, validation.By(func(value any) error {

			return collection.ForAll(cmds.Functions, func(fct CliFunction) error {

				return (&fct).Validate()

			})

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate command line commands")

	}

	return nil

}

func validateProtectedCLIArgumentNames(cmds *CommandLineCommands) error {

	if cmds == nil {

		return nil

	}

	for i := range cmds.Functions {

		if err := validateProtectedCLIArgumentNamesForFunction(&cmds.Functions[i]); err != nil {

			return err

		}

	}

	return nil

}

func validateProtectedCLIArgumentNamesForFunction(fct *CliFunction) error {

	if fct == nil {

		return nil

	}

	protectedNames, err := embeddedProtectedCLIArgumentNames(alwaysProtectedCLIArgumentNamesKey)

	if err != nil {

		return err

	}

	if fct.UsesResult() {

		resultProtectedNames, err := embeddedProtectedCLIArgumentNames(resultProtectedCLIArgumentNamesKey)

		if err != nil {

			return err

		}

		protectedNames = append(protectedNames, resultProtectedNames...)

	}

	protected := collection.Filter(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) string {

		return cfg.Name

	}), func(name string) bool {

		return slices.Contains(protectedNames, name)

	})

	if len(protected) == 0 {

		return nil

	}

	slices.Sort(protected)

	protected = slices.Compact(protected)

	return commonerrors.Newf(commonerrors.ErrInvalid, "cli function %q uses protected argument names [%s]; these names are reserved by the generator and cannot be used in command_line_commands.functions[].configuration", fct.Name, strings.Join(protected, ", "))

}

func validateOCIStringMap(name string, entries OciMapStringString) error {

	for key, value := range entries {

		if err := validation.Validate(key, validation.Required); err != nil {

			return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s contains an empty key", name)

		}

		stringValue, err := validation.EnsureString(value)

		if err != nil {

			return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s[%q] must be a string", name, key)

		}

		if err := validation.Validate(stringValue, validation.Required, validation.Length(1, 0)); err != nil {

			return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "%s[%q] must not be empty", name, key)

		}

	}

	return nil

}

func (distribution *DistributeOci) Validate() error {

	if distribution == nil {

		return nil

	}

	err := validation.ValidateStruct(distribution,

		validation.Field(&distribution.Description, validation.Required),

		validation.Field(&distribution.Attributes, validation.By(func(value any) error {

			return validateOCIStringMap("distribute_oci.attributes", distribution.Attributes)

		})),

		validation.Field(&distribution.Labels, validation.By(func(value any) error {

			return validateOCIStringMap("distribute_oci.labels", distribution.Labels)

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate distribute_oci")

	}

	return nil

}

func validateDistributeOCI(spec *SDKSpecification) error {

	if spec == nil || spec.DistributeOci == nil {

		return nil

	}

	if err := spec.DistributeOci.Validate(); err != nil {

		return err

	}

	if spec.CommandLineCommands == nil {

		return commonerrors.New(commonerrors.ErrInvalid, "distribute_oci requires command_line_commands because OCI images publish the generated CLI")

	}

	if !isHostedOnGitHub(spec) {

		return commonerrors.New(commonerrors.ErrInvalid, "distribute_oci currently requires a GitHub-hosted repository so OCI images can be published to GHCR")

	}

	return nil

}

func validateParameterValueConstraintDefinition(name string, condition *ParameterValueConstraint) error {

	if condition == nil {

		return nil

	}

	operator := condition.Operator

	if reflection.IsEmpty(operator) {

		operator = ParameterValueConstraintOperatorPresent

	}

	if reflection.IsEmpty(condition.Parameter) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must specify a parameter", name)

	}

	switch operator {

	case ParameterValueConstraintOperatorEquals, ParameterValueConstraintOperatorNotEquals:

		if reflection.IsEmpty(condition.Value) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define value", name, operator)

		}

	case ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:

		if reflection.IsEmpty(condition.Values) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define values", name, operator)

		}

	}

	return nil

}

func validateParameterRequiredValuesConstraintDefinition(name string, condition *ParameterRequiredValuesConstraint) error {

	if condition == nil {

		return nil

	}

	operator := condition.Operator

	if reflection.IsEmpty(operator) {

		operator = ParameterRequiredValuesConstraintOperatorPresent

	}

	if reflection.IsEmpty(condition.Parameter) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must specify a parameter", name)

	}

	if reflection.IsEmpty(condition.RequiredValues) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must define required_values", name)

	}

	switch operator {

	case ParameterRequiredValuesConstraintOperatorEquals, ParameterRequiredValuesConstraintOperatorNotEquals:

		if reflection.IsEmpty(condition.Value) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define value", name, operator)

		}

	case ParameterRequiredValuesConstraintOperatorIn, ParameterRequiredValuesConstraintOperatorNotIn:

		if reflection.IsEmpty(condition.Values) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define values", name, operator)

		}

	}

	return nil

}

func validateParameterForbiddenValuesConstraintDefinition(name string, condition *ParameterForbiddenValuesConstraint) error {

	if condition == nil {

		return nil

	}

	operator := condition.Operator

	if reflection.IsEmpty(operator) {

		operator = ParameterForbiddenValuesConstraintOperatorPresent

	}

	if reflection.IsEmpty(condition.Parameter) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must specify a parameter", name)

	}

	if reflection.IsEmpty(condition.ForbiddenValues) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must define forbidden_values", name)

	}

	switch operator {

	case ParameterForbiddenValuesConstraintOperatorEquals, ParameterForbiddenValuesConstraintOperatorNotEquals:

		if reflection.IsEmpty(condition.Value) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define value", name, operator)

		}

	case ParameterForbiddenValuesConstraintOperatorIn, ParameterForbiddenValuesConstraintOperatorNotIn:

		if reflection.IsEmpty(condition.Values) {

			return commonerrors.Newf(commonerrors.ErrInvalid, "%s with operator %q must define values", name, operator)

		}

	}

	return nil

}

func validateParameterValueComparisonConstraintDefinition(name string, condition *ParameterValueComparisonConstraint) error {

	if condition == nil {

		return nil

	}

	if reflection.IsEmpty(condition.Parameter) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "%s must specify a parameter", name)

	}

	return nil

}

func validateParameterConstraintsDefinition(name string, langType string, reference bool, list bool, iterator bool, directEnum []string, directPattern *string, constraints *ParameterConstraints) error {

	effectiveEnum := effectiveParameterEnum(directEnum, constraints)

	effectivePattern := effectiveParameterPattern(directPattern, constraints)

	if err := validateStringEnumDefinition(name, langType, reference, list, iterator, effectiveEnum); err != nil {

		return err

	}

	if err := validatePatternDefinition(name, langType, list, iterator, effectivePattern); err != nil {

		return err

	}

	if err := validateMirroredConstraintDefinitions(name, directEnum, directPattern, constraints); err != nil {

		return err

	}

	if err := validateStringConstraintDefinition(name, langType, list, iterator, constraints); err != nil {

		return err

	}

	if err := validateStringSubstringConstraintDefinition(name, langType, list, iterator, constraints); err != nil {

		return err

	}

	if err := validateNumericConstraintDefinition(name, langType, list, iterator, constraints); err != nil {

		return err

	}

	if err := validateCollectionConstraintDefinition(name, list, iterator, constraints); err != nil {

		return err

	}

	if err := validateExtensionConstraintDefinition(name, langType, list, iterator, constraints); err != nil {

		return err

	}

	if constraints == nil {

		return nil

	}

	return validation.ValidateStruct(constraints,

		validation.Field(&constraints.RequiredWhen, validation.By(func(value any) error {

			return validateParameterValueConstraintDefinition(fmt.Sprintf("%s required_when", name), constraints.RequiredWhen)

		})),

		validation.Field(&constraints.ForbiddenWhen, validation.By(func(value any) error {

			return validateParameterValueConstraintDefinition(fmt.Sprintf("%s forbidden_when", name), constraints.ForbiddenWhen)

		})),

		validation.Field(&constraints.RequiredValuesWhen, validation.By(func(value any) error {

			return collection.ForAll(constraints.RequiredValuesWhen, func(condition ParameterRequiredValuesConstraint) error {

				return validateParameterRequiredValuesConstraintDefinition(fmt.Sprintf("%s required_values_when", name), &condition)

			})

		})),

		validation.Field(&constraints.ForbiddenValuesWhen, validation.By(func(value any) error {

			return collection.ForAll(constraints.ForbiddenValuesWhen, func(condition ParameterForbiddenValuesConstraint) error {

				return validateParameterForbiddenValuesConstraintDefinition(fmt.Sprintf("%s forbidden_values_when", name), &condition)

			})

		})),

		validation.Field(&constraints.ParameterValue, validation.By(func(value any) error {

			return validateParameterValueComparisonConstraintDefinition(fmt.Sprintf("%s parameter_value", name), constraints.ParameterValue)

		})),
	)

}

func validateConstraintTargets(known []string, sourceName, fieldName string, targets []string) error {

	return collection.ForAll(collection.UniqueEntries(targets), func(target string) error {

		if collection.In(known, target, collection.StringMatch) {

			return nil

		}

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares %s target %q that does not exist in the same parameter collection", sourceName, fieldName, target)

	})

}

func validateParameterConstraintReferences(names []string, sourceName string, constraints *ParameterConstraints) error {

	if constraints == nil {

		return nil

	}

	known := collection.UniqueEntries(names)

	if err := validateConstraintTargets(known, sourceName, "dependent_required", constraints.DependentRequired); err != nil {

		return err

	}

	if err := validateConstraintTargets(known, sourceName, "conflicts_with", constraints.ConflictsWith); err != nil {

		return err

	}

	if constraints.RequiredWhen != nil && !collection.In(known, constraints.RequiredWhen.Parameter, collection.StringMatch) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares required_when parameter %q that does not exist in the same parameter collection", sourceName, constraints.RequiredWhen.Parameter)

	}

	if constraints.ForbiddenWhen != nil && !collection.In(known, constraints.ForbiddenWhen.Parameter, collection.StringMatch) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares forbidden_when parameter %q that does not exist in the same parameter collection", sourceName, constraints.ForbiddenWhen.Parameter)

	}

	if constraints.ParameterValue != nil && !collection.In(known, constraints.ParameterValue.Parameter, collection.StringMatch) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares parameter_value parameter %q that does not exist in the same parameter collection", sourceName, constraints.ParameterValue.Parameter)

	}

	if err := collection.ForAll(constraints.RequiredValuesWhen, func(condition ParameterRequiredValuesConstraint) error {

		if collection.In(known, condition.Parameter, collection.StringMatch) {

			return nil

		}

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares required_values_when parameter %q that does not exist in the same parameter collection", sourceName, condition.Parameter)

	}); err != nil {

		return err

	}

	if err := collection.ForAll(constraints.ForbiddenValuesWhen, func(condition ParameterForbiddenValuesConstraint) error {

		if collection.In(known, condition.Parameter, collection.StringMatch) {

			return nil

		}

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares forbidden_values_when parameter %q that does not exist in the same parameter collection", sourceName, condition.Parameter)

	}); err != nil {

		return err

	}

	return nil

}

func conditionReferenceCount(condition, name string) int {

	if reflection.IsEmpty(condition) || reflection.IsEmpty(name) {

		return 0

	}

	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)

	return len(pattern.FindAllStringIndex(condition, -1))

}

func validateParameterConditionReference(names []string, sourceName, condition string) error {

	trimmedCondition := strings.TrimSpace(condition)

	if reflection.IsEmpty(trimmedCondition) {

		return nil

	}

	matches := conditionLeadingIdentifierPattern.FindStringSubmatch(trimmedCondition)

	if len(matches) != 2 {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares condition %q that must start with the name of another parameter in the same parameter collection", sourceName, condition)

	}

	referenceName := matches[1]

	if referenceName == sourceName {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares condition %q that must reference another parameter in the same parameter collection", sourceName, condition)

	}

	if !collection.In(names, referenceName, collection.StringMatch) {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares condition %q that starts with unknown parameter %q", sourceName, condition, referenceName)

	}

	if conditionReferenceCount(trimmedCondition, referenceName) != 1 {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares condition %q that must mention parameter %q exactly once", sourceName, condition, referenceName)

	}

	referencedNames := collection.Filter(collection.UniqueEntries(names), func(name string) bool {

		return conditionReferenceCount(trimmedCondition, name) > 0

	})

	if len(referencedNames) != 1 || referencedNames[0] != referenceName {

		return commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares condition %q that must depend on exactly one other parameter in the same parameter collection", sourceName, condition)

	}

	return nil

}

func validateParameterCollectionConditionReferences[T interface {
	GetName() string

	GetCondition() string
}](parameters []T) error {

	names := collection.Map(parameters, func(parameter T) string { return parameter.GetName() })

	return collection.ForAll(parameters, func(parameter T) error {

		return validateParameterConditionReference(names, parameter.GetName(), parameter.GetCondition())

	})

}

func validateParameterCollectionConstraintReferences[T interface {
	GetName() string

	GetConstraints() *ParameterConstraints
}](parameters []T) (err error) {

	names := collection.Map(parameters, func(parameter T) string { return parameter.GetName() })

	return collection.ForAll(parameters, func(parameter T) error {

		return validateParameterConstraintReferences(names, parameter.GetName(), parameter.GetConstraints())

	})

}

func validateCollectionParameters(sdk *SDK) error {

	if sdk == nil {
		return commonerrors.UndefinedVariable("generator configuration")
	}
	err := validation.ValidateStruct(&sdk.Specification,
		validation.Field(&sdk.Specification.GlobalConfiguration, validation.By(func(value any) error {
			return collection.ForAll(sdk.Specification.GlobalConfiguration, func(cfg ConfigurationOption) error {
				return (&cfg).Validate()
			})

		})),

		validation.Field(&sdk.Specification.AdditionalTypes, validation.By(func(value any) error {

			return collection.ForAll(sdk.Specification.AdditionalTypes, func(typ Type) error {

				return (&typ).Validate()

			})

		})),

		validation.Field(&sdk.Specification.Options, validation.By(func(value any) error {

			return collection.ForAll(sdk.Specification.Options, func(optionSet OptionSet) error {

				return (&optionSet).Validate()

			})

		})),

		validation.Field(&sdk.Specification.Interface, validation.By(func(value any) error {

			return (&sdk.Specification.Interface).Validate()

		})),

		validation.Field(&sdk.Specification.CommandLineCommands, validation.By(func(value any) error {

			return sdk.Specification.CommandLineCommands.Validate()

		})),

		validation.Field(&sdk.Specification.DistributeOci, validation.By(func(value any) error {

			return validateDistributeOCI(&sdk.Specification)

		})),
	)

	if err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.DescribeCircumstance(err, "failed to validate manifest parameter collections"), "invalid manifest parameter shape")

	}

	if err := sdk.Specification.ValidateGeneratedNameUniqueness(); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid manifest parameter naming")

	}

	if err := validateParameterCollectionConstraintReferences(collection.Map(sdk.Specification.GlobalConfiguration, func(cfg ConfigurationOption) *ConfigurationOption { return &cfg })); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid global configuration constraints")

	}

	if err := validateParameterCollectionConditionReferences(collection.Map(sdk.Specification.GlobalConfiguration, func(cfg ConfigurationOption) *ConfigurationOption { return &cfg })); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid global configuration conditions")

	}

	if err := collection.ForAll(sdk.Specification.Interface.Methods, func(method InterfaceMethod) error {

		return validateParameterCollectionConstraintReferences(collection.Map(method.Args, func(arg CommandArgument) *CommandArgument { return &arg }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid interface argument constraints")

	}

	if err := collection.ForAll(sdk.Specification.Interface.Methods, func(method InterfaceMethod) error {

		return validateParameterCollectionConditionReferences(collection.Map(method.Args, func(arg CommandArgument) *CommandArgument { return &arg }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid interface argument conditions")

	}

	if sdk.Specification.CommandLineCommands != nil {

		if err := collection.ForAll(sdk.Specification.CommandLineCommands.Functions, func(fct CliFunction) error {

			return validateParameterCollectionConstraintReferences(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) *FunctionConfigFields { return &cfg }))

		}); err != nil {

			return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid CLI configuration constraints")

		}

		if err := collection.ForAll(sdk.Specification.CommandLineCommands.Functions, func(fct CliFunction) error {

			return validateParameterCollectionConditionReferences(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) *FunctionConfigFields { return &cfg }))

		}); err != nil {

			return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid CLI configuration conditions")

		}

	}

	if err := collection.ForAll(sdk.Specification.AdditionalTypes, func(typ Type) error {

		return validateParameterCollectionConstraintReferences(collection.Map(typ.Fields, func(fieldDef EntityField) *EntityField { return &fieldDef }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid additional type constraints")

	}

	if err := collection.ForAll(sdk.Specification.AdditionalTypes, func(typ Type) error {

		return validateParameterCollectionConditionReferences(collection.Map(typ.Fields, func(fieldDef EntityField) *EntityField { return &fieldDef }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid additional type conditions")

	}

	if err := collection.ForAll(sdk.Specification.Options, func(optionSet OptionSet) error {

		return validateParameterCollectionConstraintReferences(collection.Map((&optionSet).Parameters(), func(parameter OptionParameter) *OptionParameter { return &parameter }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid option parameter constraints")

	}

	if err := collection.ForAll(sdk.Specification.Options, func(optionSet OptionSet) error {

		return validateParameterCollectionConditionReferences(collection.Map((&optionSet).Parameters(), func(parameter OptionParameter) *OptionParameter { return &parameter }))

	}); err != nil {

		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid option parameter conditions")

	}

	return nil

}
