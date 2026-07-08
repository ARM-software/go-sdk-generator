package generator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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
		validation.Field(&arg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.Reference, arg.List, arg.Iterator, arg.Enum)
		})),
		validation.Field(&arg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("command argument %q", arg.Name), arg.Enum, arg.DefaultValue)
		})),
		validation.Field(&arg.Bindings, validation.By(func(value any) error {
			return collection.Each(slices.Values(arg.Bindings), func(binding Binding) error {
				return (&binding).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate command argument")
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
		validation.Field(&cfg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("global configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)
		})),
		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("global configuration %q", cfg.Name), cfg.Enum, cfg.DefaultValue)
		})),
		validation.Field(&cfg.Bindings, validation.By(func(value any) error {
			return collection.Each(slices.Values(cfg.Bindings), func(binding Binding) error {
				return (&binding).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate configuration option")
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
		validation.Field(&fieldDef.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.Reference, fieldDef.List, fieldDef.Iterator, fieldDef.Enum)
		})),
		validation.Field(&fieldDef.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.Enum, fieldDef.DefaultValue)
		})),
		validation.Field(&fieldDef.Bindings, validation.By(func(value any) error {
			return collection.Each(slices.Values(fieldDef.Bindings), func(binding Binding) error {
				return (&binding).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate entity field")
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
		validation.Field(&cfg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)
		})),
		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("function configuration %q", cfg.Name), cfg.Enum, cfg.DefaultValue)
		})),
		validation.Field(&cfg.Bindings, validation.By(func(value any) error {
			return collection.Each(slices.Values(cfg.Bindings), func(binding Binding) error {
				return (&binding).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate function configuration field")
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
		validation.Field(&method.Args, validation.By(func(value any) error {
			return collection.Each(slices.Values(method.Args), func(arg CommandArgument) error {
				return (&arg).Validate()
			})
		})),
		validation.Field(&method.Returns, validation.By(func(value any) error {
			return collection.Each(slices.Values(method.Returns), func(ret CommandArgument) error {
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
		validation.Field(&iface.Methods, validation.By(func(value any) error {
			return collection.Each(slices.Values(iface.Methods), func(method InterfaceMethod) error {
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
		validation.Field(&typ.Package, validation.By(func(value any) error {
			return validateAdditionalTypePackageName(typ.Package)
		})),
		validation.Field(&typ.Fields, validation.By(func(value any) error {
			return collection.Each(slices.Values(typ.Fields), func(fieldDef EntityField) error {
				return (&fieldDef).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate type")
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
		validation.Field(&fct.Configuration, validation.By(func(value any) error {
			return collection.Each(slices.Values(fct.Configuration), func(cfg FunctionConfigFields) error {
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
		validation.Field(&cmds.Functions, validation.By(func(value any) error {
			return collection.Each(slices.Values(cmds.Functions), func(fct CliFunction) error {
				return (&fct).Validate()
			})
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate command line commands")
	}
	return nil
}

func validateCollectionParameters(sdk *SDK) error {
	if sdk == nil {
		return commonerrors.UndefinedVariable("generator configuration")
	}
	err := validation.ValidateStruct(&sdk.Specification,
		validation.Field(&sdk.Specification.GlobalConfiguration, validation.By(func(value any) error {
			return collection.Each(slices.Values(sdk.Specification.GlobalConfiguration), func(cfg ConfigurationOption) error {
				return (&cfg).Validate()
			})
		})),
		validation.Field(&sdk.Specification.AdditionalTypes, validation.By(func(value any) error {
			return collection.Each(slices.Values(sdk.Specification.AdditionalTypes), func(typ Type) error {
				return (&typ).Validate()
			})
		})),
		validation.Field(&sdk.Specification.Interface, validation.By(func(value any) error {
			return (&sdk.Specification.Interface).Validate()
		})),
		validation.Field(&sdk.Specification.CommandLineCommands, validation.By(func(value any) error {
			return sdk.Specification.CommandLineCommands.Validate()
		})),
	)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.DescribeCircumstance(err, "failed to validate manifest parameter collections"), "invalid manifest parameter shape")
	}
	return nil
}
