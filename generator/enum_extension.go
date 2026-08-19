/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

func hasEnum(enum []string) bool {
	return reflection.IsNotEmpty(enum)
}

func enumTypeName(ownerName, fieldName, langType string, enum []string) string {
	if !hasEnum(enum) {
		return convertLangType(langType)
	}
	return fmt.Sprintf("%s%s", toPascalCase(ownerName), toPascalCase(fieldName))
}

func enumConstName(typeName, value string) string {
	return fmt.Sprintf("%s%s", typeName, toPascalCase(value))
}

func enumConstRefs(ownerName, fieldName, langType string, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	typeName := enumTypeName(ownerName, fieldName, langType, enum)
	constRefs := collection.Map(enum, func(value string) string {
		return enumConstName(typeName, value)
	})
	return strings.Join(constRefs, ", ")
}

func enumStringConstRefs(ownerName, fieldName, langType string, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	typeName := enumTypeName(ownerName, fieldName, langType, enum)
	constRefs := collection.Map(enum, func(value string) string {
		return fmt.Sprintf("%s.String()", enumConstName(typeName, value))
	})
	return strings.Join(constRefs, ", ")
}

func descriptionWithEnum(description string, enum []string) string {
	if !hasEnum(enum) {
		return description
	}
	return fmt.Sprintf("%s Allowed values: %s.", strings.TrimSpace(description), strings.Join(enum, ", "))
}

func enumValidationRule(ownerName, fieldName, langType string, reference bool, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	return fmt.Sprintf(", validationUtils.Enum(%s)", enumConstRefs(ownerName, fieldName, langType, enum))
}

func fieldEnumValidationRule(ownerName, fieldName, langType string, reference bool, required bool, enum []string, fieldExpr string) string {
	if !hasEnum(enum) {
		return ""
	}
	baseRule := fmt.Sprintf("validationUtils.Enum(%s)", enumConstRefs(ownerName, fieldName, langType, enum))
	if required {
		return fmt.Sprintf(", %s", baseRule)
	}
	return fmt.Sprintf(", validation.When(reflection.IsNotEmpty(%s), %s)", fieldExpr, baseRule)
}

func enumStringValidationRule(ownerName, fieldName, langType string, reference bool, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	return fmt.Sprintf(", validationUtils.Enum(%s)", enumStringConstRefs(ownerName, fieldName, langType, enum))
}

func fieldEnumStringValidationRule(ownerName, fieldName, langType string, reference bool, required bool, enum []string, fieldExpr string) string {
	if !hasEnum(enum) {
		return ""
	}
	baseRule := fmt.Sprintf("validationUtils.Enum(%s)", enumStringConstRefs(ownerName, fieldName, langType, enum))
	if required {
		return fmt.Sprintf(", %s", baseRule)
	}
	return fmt.Sprintf(", validation.When(reflection.IsNotEmpty(%s), %s)", fieldExpr, baseRule)
}

func cliConfigFieldDefaultValue(ownerName string, cfg FunctionConfigFields) string {
	if !hasEnum(cfg.EffectiveEnum()) || cfg.DefaultValue == nil {
		return ""
	}

	enumType := enumTypeName(ownerName, cfg.Name, baseLangType(cfg.LangType), cfg.EffectiveEnum())
	value := fmt.Sprintf("%s.String()", enumConstName(enumType, field.OptionalString(cfg.DefaultValue, "")))
	if isReferenceLikeSecretField(cfg) {
		return fmt.Sprintf("func() *string {\nvalue := %s\nreturn &value\n}()", value)
	}
	return value
}

func cliConfigEnumGetterValueExpr(receiverName string, cfg FunctionConfigFields) string {
	fieldRef := fmt.Sprintf("%s.%s", receiverName, toPascalCase(cfg.Name))
	if isReferenceLikeSecretField(cfg) {
		return fmt.Sprintf(`field.OptionalString(%s, "")`, fieldRef)
	}
	return fieldRef
}
