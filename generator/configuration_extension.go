package generator

import (
	"fmt"

	"github.com/ARM-software/golang-utils/utils/field"
)

func configFieldType(cfg FunctionConfigFields) string {
	langType := convertLangType(baseLangType(cfg.LangType))
	if cfg.Secret && langType == stringLangType {
		langType = secretProtectedStringType
	}
	if hasEnum(cfg.EffectiveEnum()) && !cfg.Secret {
		langType = enumTypeName("", cfg.Name, baseLangType(cfg.LangType), cfg.EffectiveEnum())
	}
	return parameterTypeFromBase(isReferenceLikeSecretField(cfg), cfg.List, cfg.Iterator, langType)
}

func configFieldDefaultValue(additionalTypesPackages map[string]string, currentPackage, toolName, toolVersion string, cfg FunctionConfigFields) string {
	baseType := baseLangType(cfg.LangType)
	value := defaultFieldValue(additionalTypesPackages, currentPackage, toolName, toolVersion, "", cfg.Name, baseType, cfg.EffectiveEnum(), field.OptionalString(cfg.DefaultValue, ""), isReferenceLikeSecretField(cfg))
	if cfg.Secret && convertLangType(baseType) == stringLangType {
		protectedValue := fmt.Sprintf("%s(%s)", secretProtectedStringType, value)
		if isReferenceLikeSecretField(cfg) {
			return fmt.Sprintf("field.ToOptional(%s)", protectedValue)
		}
		return protectedValue
	}
	return value
}
