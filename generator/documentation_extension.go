package generator

import (
	"strings"
	"unicode"
)

func lowerFirstDescription(s string) string {
	s = normaliseDescription(s)
	if s == "" {
		return s
	}
	leading := firstDescriptionToken(s)
	if len(leadingUppercasePrefix(leading)) >= 2 {
		return s
	}
	if len(leading) >= 2 && leading == strings.ToUpper(leading) {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func firstDescriptionToken(s string) string {
	parts := strings.Fields(normaliseDescription(s))
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func leadingUppercasePrefix(s string) string {
	var token []rune
	for _, r := range s {
		if !unicode.IsUpper(r) {
			break
		}
		token = append(token, r)
	}
	return string(token)
}

func documentedAdditionalTypeFieldType(typeName string, fieldDef *EntityField) string {
	if fieldDef == nil {
		return ""
	}
	if hasEnum(fieldDef.Enum) {
		return parameterTypeFromBase(fieldDef.Reference, fieldDef.List, fieldDef.Iterator, enumTypeName(typeName, fieldDef.Name, fieldDef.LangType, fieldDef.Enum))
	}
	return parameterType(fieldDef.Reference, fieldDef.List, fieldDef.Iterator, fieldDef.LangType)
}

func documentedConfigFieldType(specName string, cfg *ConfigurationOption) string {
	if cfg == nil {
		return ""
	}
	if hasEnum(cfg.Enum) {
		return parameterTypeFromBase(cfg.Reference, cfg.List, cfg.Iterator, enumTypeName(specName, cfg.Name, cfg.LangType, cfg.Enum))
	}
	return parameterType(cfg.Reference, cfg.List, cfg.Iterator, cfg.LangType)
}
