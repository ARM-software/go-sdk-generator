package generator

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/field"
)

func compositeConfigFieldTestValue(additionalTypesPackages map[string]string, currentPackage, toolName, toolVersion, ownerName string, cfg FunctionConfigFields) (string, bool) {
	if len(cfg.Bindings) == 0 || cfg.List || cfg.Iterator {
		return "", false
	}

	typeName := strings.TrimSpace(baseLangType(cfg.LangType))
	if typeName == "" {
		return "", false
	}

	fieldValues := make([]string, 0, len(cfg.Bindings))
	for _, binding := range cfg.Bindings {
		bindingName := strings.TrimSpace(field.OptionalString(binding.Name, ""))
		if bindingName == "" {
			return "", false
		}

		bindingCfg := FunctionConfigFields{
			Name:     bindingName,
			LangType: field.OptionalString(binding.LangType, string(BuiltInLangTypeString)),
			List:     binding.List,
			Iterator: binding.Iterator,
			Secret:   binding.Secret,
			Example:  binding.Example,
			Examples: binding.Examples,
		}

		bindingValue := configFieldTestValue(additionalTypesPackages, currentPackage, toolName, toolVersion, ownerName, bindingCfg)
		if bindingValue == "" {
			return "", false
		}

		fieldValues = append(fieldValues, fmt.Sprintf("%s: %s", toPascalCase(bindingName), bindingValue))
	}

	value := fmt.Sprintf("%s{%s}", typeName, strings.Join(fieldValues, ", "))
	if cfg.Reference || isPointerLangType(cfg.LangType) {
		return fmt.Sprintf("func() *%s {\nvalue := %s\nreturn &value\n}()", typeName, value), true
	}

	return value, true
}

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

func cliConfigFieldType(cfg FunctionConfigFields) string {
	langType := convertLangType(baseLangType(cfg.LangType))
	if cfg.Secret && langType == stringLangType {
		langType = secretProtectedStringType
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

func configFieldTestValue(additionalTypesPackages map[string]string, currentPackage, toolName, toolVersion, ownerName string, cfg FunctionConfigFields) string {
	if value, ok := compositeConfigFieldTestValue(additionalTypesPackages, currentPackage, toolName, toolVersion, ownerName, cfg); ok {
		return value
	}

	baseType := strings.TrimSpace(baseLangType(cfg.LangType))
	valueSource := field.OptionalString(cfg.DefaultValue, "")
	if valueSource == "" {
		switch {
		case cfg.Example != nil && strings.TrimSpace(*cfg.Example) != "":
			valueSource = strings.TrimSpace(*cfg.Example)
		case len(cfg.Examples) > 0 && strings.TrimSpace(cfg.Examples[0]) != "":
			valueSource = strings.TrimSpace(cfg.Examples[0])
		case len(cfg.EffectiveEnum()) > 0:
			valueSource = cfg.EffectiveEnum()[0]
		default:
			switch baseType {
			case string(BuiltInLangTypeString):
				valueSource = "value"
			case string(BuiltInLangTypeFilepath):
				valueSource = "summary.md"
			case string(BuiltInLangTypePathParameter):
				valueSource = "/resource"
			case string(BuiltInLangTypeHost), string(BuiltInLangTypeUrl), string(BuiltInLangTypeUri):
				valueSource = "https://example.com"
			case string(BuiltInLangTypeVersion), string(BuiltInLangTypeSemver):
				valueSource = "1.2.3"
			case string(BuiltInLangTypeUuid):
				valueSource = "123e4567-e89b-12d3-a456-426614174000"
			case string(BuiltInLangTypeEmail):
				valueSource = "user@example.com"
			case string(BuiltInLangTypeJson), string(BuiltInLangTypeValidJSON):
				valueSource = "{}"
			case string(BuiltInLangTypeIp):
				valueSource = "127.0.0.1"
			case string(BuiltInLangTypeDatetime):
				valueSource = "2026-01-01T00:00:00Z"
			case string(BuiltInLangTypeRegex):
				valueSource = ".*"
			case string(BuiltInLangTypeTimeDuration), string(BuiltInLangTypeDuration):
				valueSource = "5s"
			case string(BuiltInLangTypeBool):
				valueSource = "true"
			case string(BuiltInLangTypeInt), string(BuiltInLangTypeInt8), string(BuiltInLangTypeInt16), string(BuiltInLangTypeInt32), string(BuiltInLangTypeInt64), string(BuiltInLangTypeUint), string(BuiltInLangTypeUint8), string(BuiltInLangTypeUint16), string(BuiltInLangTypeUint32), string(BuiltInLangTypeUint64), string(BuiltInLangTypePort):
				valueSource = "1"
			case string(BuiltInLangTypeFloat32), string(BuiltInLangTypeFloat64), string(BuiltInLangTypeDecimal):
				valueSource = "1.5"
			case "[]string":
				return `[]string{"value"}`
			default:
				return ""
			}
		}
	}

	if valueSource == "" {
		return ""
	}

	return defaultFieldValue(additionalTypesPackages, currentPackage, toolName, toolVersion, ownerName, cfg.Name, baseType, cfg.EffectiveEnum(), valueSource, isReferenceLikeSecretField(cfg))
}
