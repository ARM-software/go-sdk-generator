package generator

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	mapset "github.com/deckarep/golang-set/v2"
)

type DocumentedCLIOption struct {
	FieldName    string
	FlagName     string
	ShortFlag    string
	EnvVar       string
	Description  string
	Required     bool
	Condition    string
	SharedAcross bool
}

type DocumentedCLIOptionSet []DocumentedCLIOption

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

func documentedInterfaceMethodSignature(method InterfaceMethod) string {
	functionName := toPascalCase(method.Name)
	arguments := append([]string{"ctx context.Context"}, collection.Map(method.Args, func(arg CommandArgument) string {
		argType := parameterType(arg.Reference, arg.List, arg.Iterator, arg.LangType)
		if hasEnum(arg.Enum) {
			argType = parameterTypeFromBase(arg.Reference, arg.List, arg.Iterator, enumTypeName(functionName, arg.Name, arg.LangType, arg.Enum))
		}
		return fmt.Sprintf("%s %s", toCamelCase(arg.Name), argType)
	})...)

	returns := collection.Map(method.Returns, func(ret CommandArgument) string {
		retType := parameterType(ret.Reference, ret.List, ret.Iterator, ret.LangType)
		if hasEnum(ret.Enum) {
			retType = parameterTypeFromBase(ret.Reference, ret.List, ret.Iterator, enumTypeName(functionName, ret.Name, ret.LangType, ret.Enum))
		}
		return fmt.Sprintf("%s %s", toCamelCase(ret.Name), retType)
	})

	signature := fmt.Sprintf("%s(%s)", functionName, strings.Join(arguments, ", "))
	if len(returns) == 0 {
		return signature
	}
	if len(returns) == 1 {
		return fmt.Sprintf("%s (%s)", signature, returns[0])
	}
	return fmt.Sprintf("%s (%s)", signature, strings.Join(returns, ", "))
}

func documentedCLIOptions(cmds *CommandLineCommands, cliFunc CliFunction) DocumentedCLIOptionSet {
	if cmds == nil {
		return nil
	}

	options := DocumentedCLIOptionSet{}
	seen := mapset.NewSet[string]()
	collection.ForEach(cmds.Functions, func(fct CliFunction) {
		collection.ForEach(fct.Configuration, func(cfg FunctionConfigFields) {
			if !cfg.IsRoot || cfg.Hidden {
				return
			}
			collection.ForEach(documentedCLIOptionsFromConfig(field.OptionalString(cmds.RootEnvironmentVar, ""), cfg, true), func(option DocumentedCLIOption) {
				key := option.FlagName + "|" + option.EnvVar
				if seen.Contains(key) {
					return
				}
				seen.Add(key)
				options = append(options, option)
			})
		})
	})

	collection.ForEach(cliFunc.Configuration, func(cfg FunctionConfigFields) {
		if cfg.IsRoot || cfg.Hidden {
			return
		}
		collection.ForEach(documentedCLIOptionsFromConfig(field.OptionalString(cmds.RootEnvironmentVar, ""), cfg, false), func(option DocumentedCLIOption) {
			key := option.FlagName + "|" + option.EnvVar
			if seen.Contains(key) {
				return
			}
			seen.Add(key)
			options = append(options, option)
		})
	})

	return options
}

func (options DocumentedCLIOptionSet) HasAny() bool {
	return len(options) > 0
}

func (options DocumentedCLIOptionSet) First() *DocumentedCLIOption {
	if option, ok := collection.First(options); ok {
		return &option
	}
	return nil
}

func documentedCLIOptionsFromConfig(rootPrefix string, cfg FunctionConfigFields, shared bool) []DocumentedCLIOption {
	if len(cfg.Bindings) == 0 {
		return []DocumentedCLIOption{{
			FieldName:    cfg.Name,
			FlagName:     normalisedCLIArgumentName(cfg, nil),
			ShortFlag:    field.OptionalString(cfg.ArgShort, ""),
			EnvVar:       documentedCLIEnvVar(rootPrefix, documentedCLIEnvSuffix(cfg, nil)),
			Description:  documentedCLIOptionDescription(cfg.Description, cfg.Condition, cfg.Enum, cfg.Example, cfg.Examples),
			Required:     cfg.Required,
			Condition:    cfg.Condition,
			SharedAcross: shared,
		}}
	}

	options := make([]DocumentedCLIOption, 0, len(cfg.Bindings))
	collection.ForEach(cfg.Bindings, func(binding Binding) {
		if binding.Hidden {
			return
		}
		description := cfg.Description
		if binding.Description != nil && strings.TrimSpace(*binding.Description) != "" {
			description = *binding.Description
		}
		example := cfg.Example
		if binding.Example != nil && strings.TrimSpace(*binding.Example) != "" {
			example = binding.Example
		}
		examples := cfg.Examples
		if len(binding.Examples) > 0 {
			examples = binding.Examples
		}
		fieldName := cfg.Name
		if binding.Name != nil && strings.TrimSpace(*binding.Name) != "" {
			fieldName = fmt.Sprintf("%s.%s", cfg.Name, *binding.Name)
		}
		options = append(options, DocumentedCLIOption{
			FieldName:    fieldName,
			FlagName:     normalisedCLIArgumentName(cfg, &binding),
			ShortFlag:    field.OptionalString(binding.ArgShort, field.OptionalString(cfg.ArgShort, "")),
			EnvVar:       documentedCLIEnvVar(rootPrefix, documentedCLIEnvSuffix(cfg, &binding)),
			Description:  documentedCLIOptionDescription(description, cfg.Condition, cfg.Enum, example, examples),
			Required:     cfg.Required,
			Condition:    cfg.Condition,
			SharedAcross: shared,
		})
	})
	slices.SortFunc(options, func(a, b DocumentedCLIOption) int {
		return strings.Compare(a.FlagName, b.FlagName)
	})
	return options
}

func documentedCLIEnvSuffix(cfg FunctionConfigFields, binding *Binding) string {
	if cfg.IsRoot {
		if binding != nil && binding.MapstructureTag != nil && strings.TrimSpace(*binding.MapstructureTag) != "" {
			return *binding.MapstructureTag
		}
		return cfg.MapstructureTag
	}
	var bindingTag *string
	if binding != nil {
		bindingTag = binding.MapstructureTag
	}
	return getBindingEnv(cfg.MapstructureTag, bindingTag, fmt.Sprintf("%v_", cfg.MapstructureTag))
}

func documentedCLIEnvVar(prefix, suffix string) string {
	if strings.TrimSpace(suffix) == "" {
		return ""
	}
	parts := collection.Filter([]string{strings.TrimSpace(prefix), strings.TrimSpace(suffix)}, func(part string) bool {
		return part != ""
	})
	joined := strings.Join(parts, "_")
	joined = strings.ReplaceAll(joined, "-", "_")
	return strings.ToUpper(joined)
}

func documentedCLIOptionDescription(description, condition string, enum []string, example *string, examples []string) string {
	desc := descriptionWithMetadata(description, enum, example, examples)
	if strings.TrimSpace(condition) == "" {
		return desc
	}
	return fmt.Sprintf("%s Required when %s.", desc, condition)
}
