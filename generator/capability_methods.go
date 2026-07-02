package generator

import "github.com/ARM-software/golang-utils/utils/collection"

// capability_methods centralises small capability queries on the generated
// schema model types.
//
// The goal is to keep traversal logic close to the types it describes instead
// of repeating nested loops in several helpers. That makes enum, secret, and
// callCloud capability checks easier to reuse across generation, validation,
// and template wiring while staying aligned with the collection-helper style
// preferred in this repository.

func (b *Binding) HasEnum() bool {
	return false
}

func (arg *CommandArgument) HasEnum() bool {
	if arg == nil {
		return false
	}
	return hasEnum(arg.Enum)
}

func (fieldDef *EntityField) HasEnum() bool {
	if fieldDef == nil {
		return false
	}
	return hasEnum(fieldDef.Enum)
}

func (cfg *ConfigurationOption) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.Enum)
}

func (cfg *FunctionConfigFields) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.Enum)
}

func (typ *Type) HasEnum() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.HasEnum()
	})
}

func (types Types) HasEnum() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.HasEnum()
	})
}

func (cfgs Configuration) HasEnum() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.HasEnum()
	})
}

func (method *InterfaceMethod) HasEnum() bool {
	if method == nil {
		return false
	}
	return collection.AnyRefFunc(method.Args, func(arg *CommandArgument) bool {
		return arg.HasEnum()
	}) || collection.AnyRefFunc(method.Returns, func(ret *CommandArgument) bool {
		return ret.HasEnum()
	})
}

func (iface *Interface) HasEnum() bool {
	if iface == nil {
		return false
	}
	return collection.AnyRefFunc(iface.Methods, func(method *InterfaceMethod) bool {
		return method.HasEnum()
	})
}

func (fct *CliFunction) HasEnum() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.HasEnum()
	})
}

func (cmds *CommandLineCommands) HasConfigEnums() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.HasEnum()
	})
}

func (cmds *CommandLineCommands) HasSecretStringParameters() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
			return cfg.Secret && isSecretCompatibleLangType(cfg.LangType)
		})
	})
}

func (cmds *CommandLineCommands) HasReferenceLikeSecretStringParameters() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
			return cfg.Secret && isReferenceLikeSecretField(*cfg) && isSecretCompatibleLangType(cfg.LangType)
		})
	})
}

func (cmds *CommandLineCommands) HasCloudCommands() bool {
	if cmds == nil || len(cmds.Functions) == 0 {
		return false
	}
	return collection.AllFunc(cmds.Functions, func(fct CliFunction) bool {
		return fct.CallCloud || fct.StoreToKeyring
	})
}

func (spec *SDKSpecification) HasTypeEnums() bool {
	if spec == nil {
		return false
	}
	return spec.AdditionalTypes.HasEnum()
}

func (spec *SDKSpecification) HasConfigEnums() bool {
	if spec == nil {
		return false
	}
	return spec.GlobalConfiguration.HasEnum()
}

func (spec *SDKSpecification) HasInterfaceEnums() bool {
	if spec == nil {
		return false
	}
	return (&spec.Interface).HasEnum()
}
