package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
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
	if value == "" {
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
		validation.Field(&arg.Example, validation.By(func(value any) error {
			return validateExamplesDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.List, arg.Iterator, arg.Enum, arg.Example, arg.Examples)
		})),
		validation.Field(&arg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("command argument %q", arg.Name), arg.LangType, arg.Reference, arg.List, arg.Iterator, arg.Enum)
		})),
		validation.Field(&arg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("command argument %q", arg.Name), arg.Enum, arg.DefaultValue)
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
		validation.Field(&cfg.Example, validation.By(func(value any) error {
			return validateExamplesDefinition(fmt.Sprintf("global configuration %q", cfg.Name), cfg.LangType, cfg.List, cfg.Iterator, cfg.Enum, cfg.Example, cfg.Examples)
		})),
		validation.Field(&cfg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("global configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)
		})),
		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("global configuration %q", cfg.Name), cfg.Enum, cfg.DefaultValue)
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
		validation.Field(&fieldDef.Example, validation.By(func(value any) error {
			return validateExamplesDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.List, fieldDef.Iterator, fieldDef.Enum, fieldDef.Example, fieldDef.Examples)
		})),
		validation.Field(&fieldDef.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.LangType, fieldDef.Reference, fieldDef.List, fieldDef.Iterator, fieldDef.Enum)
		})),
		validation.Field(&fieldDef.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("entity field %q", fieldDef.Name), fieldDef.Enum, fieldDef.DefaultValue)
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
		validation.Field(&cfg.Example, validation.By(func(value any) error {
			return validateExamplesDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.List, cfg.Iterator, cfg.Enum, cfg.Example, cfg.Examples)
		})),
		validation.Field(&cfg.Enum, validation.By(func(value any) error {
			return validateStringEnumDefinition(fmt.Sprintf("function configuration %q", cfg.Name), cfg.LangType, cfg.Reference, cfg.List, cfg.Iterator, cfg.Enum)
		})),
		validation.Field(&cfg.DefaultValue, validation.By(func(value any) error {
			return validateStringEnumDefaultValue(fmt.Sprintf("function configuration %q", cfg.Name), cfg.Enum, cfg.DefaultValue)
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

func (fct *CliFunction) Validate() error {
	if fct == nil {
		return commonerrors.UndefinedVariable("cli function")
	}
	if err := configUtils.ValidateEmbedded(fct); err != nil {
		return commonerrors.DescribeCircumstance(err, "failed to validate embedded fields in cli function")
	}
	err := validation.ValidateStruct(fct,
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
	return nil
}
