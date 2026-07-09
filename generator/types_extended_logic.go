package generator

import (
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
)

// types_extended_logic centralises capability queries and small schema-model
// traversal logic that belongs on the generated manifest types themselves.

func isBoolean(langType string) bool {
	return strings.EqualFold(strings.TrimSpace(langType), "bool")
}

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

func (cfg *ConfigurationOption) SuppressesRequiredValidation() bool {
	if cfg == nil {
		return false
	}
	return cfg.HasEnum() || isBoolean(cfg.LangType)
}

func (cfg *FunctionConfigFields) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.Enum)
}

func (cfg *FunctionConfigFields) SuppressesRequiredValidation() bool {
	if cfg == nil {
		return false
	}
	return cfg.HasEnum() || isBoolean(cfg.LangType)
}

func (typ *Type) HasEnum() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.HasEnum()
	})
}

func (fieldDef *EntityField) SuppressesRequiredValidation() bool {
	if fieldDef == nil {
		return false
	}
	return fieldDef.HasEnum() || isBoolean(fieldDef.LangType)
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

func (cfgs Configuration) SuppressedRequiredValidationMap() map[string]bool {
	result := map[string]bool{}
	collection.ForEachRef(cfgs, func(cfg *ConfigurationOption) {
		if cfg.SuppressesRequiredValidation() {
			result[validationSuppressionKey(globalConfigurationValidationScope, "", cfg.Name)] = true
		}
	})
	return result
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

func (fct *CliFunction) SuppressedRequiredValidationMap() map[string]bool {
	if fct == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(fct.Configuration, func(cfg *FunctionConfigFields) {
		if cfg.SuppressesRequiredValidation() {
			result[validationSuppressionKey(cliConfigurationValidationScope, fct.Name, cfg.Name)] = true
		}
	})
	return result
}

func (cmds *CommandLineCommands) HasConfigEnums() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.HasEnum()
	})
}

func (cmds *CommandLineCommands) SuppressedRequiredValidationMap() map[string]bool {
	if cmds == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(cmds.Functions, func(fct *CliFunction) {
		for key := range fct.SuppressedRequiredValidationMap() {
			result[key] = true
		}
	})
	return result
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

func (typ *Type) SuppressedRequiredValidationMap() map[string]bool {
	if typ == nil {
		return nil
	}
	result := map[string]bool{}
	collection.ForEachRef(typ.Fields, func(fieldDef *EntityField) {
		if fieldDef.SuppressesRequiredValidation() {
			result[validationSuppressionKey(additionalTypeValidationScope, typ.Name, fieldDef.Name)] = true
		}
	})
	return result
}

func (types Types) SuppressedRequiredValidationMap() map[string]bool {
	result := map[string]bool{}
	collection.ForEachRef(types, func(typ *Type) {
		for key := range typ.SuppressedRequiredValidationMap() {
			result[key] = true
		}
	})
	return result
}

func (spec *SDKSpecification) SuppressedRequiredValidationMap() map[string]bool {
	if spec == nil {
		return nil
	}
	result := spec.GlobalConfiguration.SuppressedRequiredValidationMap()
	for key := range spec.CommandLineCommands.SuppressedRequiredValidationMap() {
		result[key] = true
	}
	for key := range spec.AdditionalTypes.SuppressedRequiredValidationMap() {
		result[key] = true
	}
	return result
}

func (b *Binding) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if b == nil {
		return
	}
	b.LangType = field.ToOptionalStringOrNilIfEmpty(rewriteGeneratedAdditionalTypeReference(field.OptionalString(b.LangType, ""), currentPackage, additionalTypesPackages))
}

func (arg *CommandArgument) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if arg == nil {
		return
	}
	arg.LangType = rewriteGeneratedAdditionalTypeReference(arg.LangType, currentPackage, additionalTypesPackages)
	arg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(arg.DefaultValue, currentPackage, additionalTypesPackages)
	arg.Bindings = collection.MapRef(arg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (cfg *ConfigurationOption) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if cfg == nil {
		return
	}
	cfg.LangType = rewriteGeneratedAdditionalTypeReference(cfg.LangType, currentPackage, additionalTypesPackages)
	cfg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(cfg.DefaultValue, currentPackage, additionalTypesPackages)
	cfg.Bindings = collection.MapRef(cfg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (fieldDef *EntityField) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if fieldDef == nil {
		return
	}
	fieldDef.LangType = rewriteGeneratedAdditionalTypeReference(fieldDef.LangType, currentPackage, additionalTypesPackages)
	fieldDef.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(fieldDef.DefaultValue, currentPackage, additionalTypesPackages)
	fieldDef.Bindings = collection.MapRef(fieldDef.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (cfg *FunctionConfigFields) RewriteAdditionalTypeReferences(currentPackage string, additionalTypesPackages map[string]string) {
	if cfg == nil {
		return
	}
	cfg.LangType = rewriteGeneratedAdditionalTypeReference(cfg.LangType, currentPackage, additionalTypesPackages)
	cfg.DefaultValue = rewriteGeneratedAdditionalTypeDefaultValue(cfg.DefaultValue, currentPackage, additionalTypesPackages)
	cfg.Bindings = collection.MapRef(cfg.Bindings, func(binding *Binding) *Binding {
		binding.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return binding
	})
}

func (method *InterfaceMethod) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if method == nil {
		return
	}
	method.Args = collection.MapRef(method.Args, func(arg *CommandArgument) *CommandArgument {
		arg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return arg
	})
	method.Returns = collection.MapRef(method.Returns, func(ret *CommandArgument) *CommandArgument {
		ret.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return ret
	})
}

func (iface *Interface) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if iface == nil {
		return
	}
	iface.Methods = collection.MapRef(iface.Methods, func(method *InterfaceMethod) *InterfaceMethod {
		method.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return method
	})
}

func (typ *Type) NormalisePackage(additionalTypesPackages map[string]string) error {
	if typ == nil {
		return nil
	}
	typ.Package = normaliseAdditionalTypePackageName(typ.Package)
	typeName := toPascalCase(typ.Name)
	if existing, found := additionalTypesPackages[typeName]; found && existing != typ.Package {
		return commonerrors.Newf(commonerrors.ErrConflict, "additional type %q is defined in multiple packages (%q and %q)", typeName, existing, typ.Package)
	}
	additionalTypesPackages[typeName] = typ.Package
	return nil
}

func (typ *Type) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if typ == nil {
		return
	}
	currentPackage := typ.Package
	typ.Fields = collection.MapRef(typ.Fields, func(fieldDef *EntityField) *EntityField {
		fieldDef.RewriteAdditionalTypeReferences(currentPackage, additionalTypesPackages)
		return fieldDef
	})
}

func (types Types) NormalisePackages() (map[string]string, error) {
	additionalTypesPackages := make(map[string]string, len(types))
	for i := range types {
		if err := types[i].NormalisePackage(additionalTypesPackages); err != nil {
			return nil, err
		}
	}
	return additionalTypesPackages, nil
}

func (types Types) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	updated := collection.MapRef(types, func(typ *Type) *Type {
		typ.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return typ
	})
	copy(types, updated)
}

func (fct *CliFunction) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if fct == nil {
		return
	}
	fct.Configuration = collection.MapRef(fct.Configuration, func(cfg *FunctionConfigFields) *FunctionConfigFields {
		cfg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return cfg
	})
}

func (cmds *CommandLineCommands) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if cmds == nil {
		return
	}
	cmds.Functions = collection.MapRef(cmds.Functions, func(fct *CliFunction) *CliFunction {
		fct.RewriteAdditionalTypeReferences(additionalTypesPackages)
		return fct
	})
}

func (spec *SDKSpecification) RewriteAdditionalTypeReferences(additionalTypesPackages map[string]string) {
	if spec == nil {
		return
	}
	spec.GlobalConfiguration = collection.MapRef(spec.GlobalConfiguration, func(cfg *ConfigurationOption) *ConfigurationOption {
		cfg.RewriteAdditionalTypeReferences("", additionalTypesPackages)
		return cfg
	})
	(&spec.Interface).RewriteAdditionalTypeReferences(additionalTypesPackages)
	spec.CommandLineCommands.RewriteAdditionalTypeReferences(additionalTypesPackages)
	spec.AdditionalTypes.RewriteAdditionalTypeReferences(additionalTypesPackages)
}
