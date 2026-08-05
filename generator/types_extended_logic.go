package generator

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// types_extended_logic centralises capability queries and small schema-model
// traversal logic that belongs on the generated manifest types themselves.

func isBoolean(langType string) bool {
	return strings.EqualFold(strings.TrimSpace(langType), "bool")
}

func usesFilesystemValidationAlias(langType string) bool {
	switch strings.TrimSpace(langType) {
	case "filepath", "existingFilepath":
		return true
	default:
		return false
	}
}

func usesJSONValidationAlias(langType string) bool {
	return strings.TrimSpace(langType) == "validJSON"
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

func (cfg *ConfigurationOption) UsesFilesystemValidationAlias() bool {
	return cfg != nil && usesFilesystemValidationAlias(cfg.LangType)
}

func (cfg *ConfigurationOption) UsesJSONValidationAlias() bool {
	return cfg != nil && usesJSONValidationAlias(cfg.LangType)
}

func (cfg *FunctionConfigFields) HasEnum() bool {
	if cfg == nil {
		return false
	}
	return hasEnum(cfg.Enum)
}

func (arg *CommandArgument) GetName() string {
	if arg == nil {
		return ""
	}
	return arg.Name
}

func (arg *CommandArgument) GetConstraints() *ParameterConstraints {
	if arg == nil {
		return nil
	}
	return arg.Constraints
}

func (fieldDef *EntityField) GetName() string {
	if fieldDef == nil {
		return ""
	}
	return fieldDef.Name
}

func (fieldDef *EntityField) GetConstraints() *ParameterConstraints {
	if fieldDef == nil {
		return nil
	}
	return fieldDef.Constraints
}

func (cfg *ConfigurationOption) GetName() string {
	if cfg == nil {
		return ""
	}
	return cfg.Name
}

func (cfg *ConfigurationOption) GetConstraints() *ParameterConstraints {
	if cfg == nil {
		return nil
	}
	return cfg.Constraints
}

func (cfg *FunctionConfigFields) GetName() string {
	if cfg == nil {
		return ""
	}
	return cfg.Name
}

func (cfg *FunctionConfigFields) GetConstraints() *ParameterConstraints {
	if cfg == nil {
		return nil
	}
	return cfg.Constraints
}

func (cfg *FunctionConfigFields) SuppressesRequiredValidation() bool {
	if cfg == nil {
		return false
	}
	return cfg.HasEnum() || isBoolean(cfg.LangType)
}

func (cfg *FunctionConfigFields) UsesFilesystemValidationAlias() bool {
	return cfg != nil && usesFilesystemValidationAlias(cfg.LangType)
}

func (cfg *FunctionConfigFields) UsesJSONValidationAlias() bool {
	return cfg != nil && usesJSONValidationAlias(cfg.LangType)
}

func (typ *Type) HasEnum() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.HasEnum()
	})
}

func (typ *Type) UsesFilesystemValidationAlias() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.UsesFilesystemValidationAlias()
	})
}

func (typ *Type) UsesJSONValidationAlias() bool {
	if typ == nil {
		return false
	}
	return collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef.UsesJSONValidationAlias()
	})
}

func (fieldDef *EntityField) SuppressesRequiredValidation() bool {
	if fieldDef == nil {
		return false
	}
	return fieldDef.HasEnum() || isBoolean(fieldDef.LangType)
}

func (fieldDef *EntityField) UsesFilesystemValidationAlias() bool {
	return fieldDef != nil && usesFilesystemValidationAlias(fieldDef.LangType)
}

func (fieldDef *EntityField) UsesJSONValidationAlias() bool {
	return fieldDef != nil && usesJSONValidationAlias(fieldDef.LangType)
}

func (types Types) HasEnum() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.HasEnum()
	})
}

func (types Types) UsesFilesystemValidationAlias() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.UsesFilesystemValidationAlias()
	})
}

func (types Types) UsesJSONValidationAlias() bool {
	return collection.AnyRefFunc(types, func(typ *Type) bool {
		return typ.UsesJSONValidationAlias()
	})
}

func (cfgs Configuration) HasEnum() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.HasEnum()
	})
}

func (cfgs Configuration) UsesFilesystemValidationAlias() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.UsesFilesystemValidationAlias()
	})
}

func (cfgs Configuration) UsesJSONValidationAlias() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg.UsesJSONValidationAlias()
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

func (fct *CliFunction) UsesFilesystemValidationAlias() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesFilesystemValidationAlias()
	})
}

func (fct *CliFunction) UsesJSONValidationAlias() bool {
	if fct == nil {
		return false
	}
	return collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg.UsesJSONValidationAlias()
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

func (cmds *CommandLineCommands) UsesFilesystemValidationAlias() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesFilesystemValidationAlias()
	})
}

func (cmds *CommandLineCommands) UsesJSONValidationAlias() bool {
	if cmds == nil {
		return false
	}
	return collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool {
		return fct.UsesJSONValidationAlias()
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

func (spec *SDKSpecification) UsesFilesystemValidationAlias() bool {
	return spec != nil && (spec.GlobalConfiguration.UsesFilesystemValidationAlias() || spec.AdditionalTypes.UsesFilesystemValidationAlias())
}

func (spec *SDKSpecification) UsesJSONValidationAlias() bool {
	return spec != nil && (spec.GlobalConfiguration.UsesJSONValidationAlias() || spec.AdditionalTypes.UsesJSONValidationAlias())
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

func (cfg *FunctionConfigFields) IsStandaloneCLIFlag() bool {
	return cfg != nil && !cfg.Hidden && len(cfg.Bindings) == 0
}

func (cfg *FunctionConfigFields) CLIFlagArgName(functionName string) string {
	if cfg == nil {
		return ""
	}
	return getArgName(functionName, "", cfg.Name, cfg.IsRoot)
}

func (cfg *FunctionConfigFields) FlagConstraintRegistrations(commandVar string, parameterFlags map[string]string, configurationByName map[string]*FunctionConfigFields) (out []FlagConstraintRegistration, err error) {
	if cfg == nil || cfg.Constraints == nil {
		return nil, nil
	}
	if registration, hasAny, subErr := buildFlagConstraintRegistration(commandVar, cfg.Name, cfg.CLIRequiredTogetherTargets(configurationByName), parameterFlags, "MarkFlagsRequiredTogether"); subErr != nil {
		return nil, subErr
	} else if hasAny {
		out = append(out, registration)
	}
	if registration, hasAny, subErr := buildFlagConstraintRegistration(commandVar, cfg.Name, cfg.Constraints.ConflictsWith, parameterFlags, "MarkFlagsMutuallyExclusive"); subErr != nil {
		return nil, subErr
	} else if hasAny {
		out = append(out, registration)
	}
	return
}

func (fct *CliFunction) StandaloneCLIParameterFlags() map[string]string {
	if fct == nil {
		return nil
	}
	parameterFlags := make(map[string]string)
	collection.ForEach(fct.Configuration, func(cfg FunctionConfigFields) {
		if !cfg.IsStandaloneCLIFlag() {
			return
		}
		parameterFlags[cfg.Name] = cfg.CLIFlagArgName(fct.Name)
	})
	return parameterFlags
}

func (fct *CliFunction) ConfigurationByName() map[string]*FunctionConfigFields {
	if fct == nil {
		return nil
	}
	return collection.Reduce(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) *FunctionConfigFields {
		return &cfg
	}), map[string]*FunctionConfigFields{}, func(acc map[string]*FunctionConfigFields, cfg *FunctionConfigFields) map[string]*FunctionConfigFields {
		if cfg != nil {
			acc[cfg.Name] = cfg
		}
		return acc
	})
}

func (fct *CliFunction) FlagConstraintRegistrations() (out []FlagConstraintRegistration, err error) {
	if fct == nil {
		return nil, nil
	}
	configurationByName := fct.ConfigurationByName()
	parameterFlags := fct.StandaloneCLIParameterFlags()
	commandVar := fmt.Sprintf("%vCmd", strings.ToLower(fct.Name))
	err = collection.ForAll(fct.Configuration, func(cfg FunctionConfigFields) error {
		registrations, subErr := (&cfg).FlagConstraintRegistrations(commandVar, parameterFlags, configurationByName)
		if subErr != nil {
			return subErr
		}
		out = append(out, registrations...)
		return nil
	})
	return
}

func (cmds *CommandLineCommands) FlagConstraintRegistrations() (out []FlagConstraintRegistration, err error) {
	if cmds == nil {
		return nil, nil
	}
	err = collection.ForAll(cmds.Functions, func(fct CliFunction) error {
		registrations, subErr := (&fct).FlagConstraintRegistrations()
		if subErr != nil {
			return subErr
		}
		out = append(out, registrations...)
		return nil
	})
	slices.SortFunc(out, func(a, b FlagConstraintRegistration) int {
		return strings.Compare(a.Statement(), b.Statement())
	})
	return
}

func buildFlagConstraintRegistration(commandVar, sourceName string, targetNames []string, parameterFlags map[string]string, method string) (registration FlagConstraintRegistration, hasAny bool, err error) {
	if reflection.IsEmpty(targetNames) {
		return FlagConstraintRegistration{}, false, nil
	}
	sourceFlag, found := parameterFlags[sourceName]
	if !found {
		return FlagConstraintRegistration{}, false, commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q declares CLI constraints but does not map to a standalone flag", sourceName)
	}
	resolvedTargets, err := collection.MapWithError(collection.UniqueBy(targetNames, func(name string) string { return name }), func(targetName string) (string, error) {
		targetFlag, targetFound := parameterFlags[targetName]
		if !targetFound {
			return "", commonerrors.Newf(commonerrors.ErrInvalid, "parameter %q references %q in CLI constraints but that parameter does not map to a standalone flag", sourceName, targetName)
		}
		return targetFlag, nil
	})
	if err != nil {
		return FlagConstraintRegistration{}, false, err
	}
	resolvedTargets = append([]string{sourceFlag}, resolvedTargets...)
	return FlagConstraintRegistration{CommandVar: commandVar, Method: method, FlagArgs: resolvedTargets}, true, nil
}

type constrainedParameter struct {
	Name        string
	Constraints *ParameterConstraints
}

func (constraints *ParameterConstraints) HasAnyValidationRules() bool {
	if constraints == nil {
		return false
	}
	return len(constraints.DependentRequired) > 0 || len(constraints.ConflictsWith) > 0 || constraints.RequiredWhen != nil || constraints.ForbiddenWhen != nil
}

func (constraints *ParameterConstraints) UsesCLIRequiredTogetherFlags() bool {
	return constraints != nil && constraints.RequireTogetherAsCliFlags
}

func (cfg *FunctionConfigFields) CLIRequiredTogetherTargets(configurationByName map[string]*FunctionConfigFields) []string {
	if cfg == nil || cfg.Constraints == nil || !cfg.Constraints.UsesCLIRequiredTogetherFlags() {
		return nil
	}
	return collection.Filter(collection.UniqueBy(cfg.Constraints.DependentRequired, func(name string) string { return name }), func(name string) bool {
		target, found := configurationByName[name]
		return found && target != nil && target.Constraints.UsesCLIRequiredTogetherFlags()
	})
}

func (constraints *ParameterConstraints) UsesReflect() bool {
	if constraints == nil {
		return false
	}
	return collection.AnyTrue(constraints.RequiredWhen.UsesReflect(), constraints.ForbiddenWhen.UsesReflect())
}

func (constraints *ParameterConstraints) UsesCollection() bool {
	if constraints == nil {
		return false
	}
	return collection.AnyTrue(constraints.RequiredWhen.UsesCollection(), constraints.ForbiddenWhen.UsesCollection())
}

func (condition *ParameterValueConstraint) operatorOrDefault() ParameterValueConstraintOperator {
	if condition == nil || reflection.IsEmpty(condition.Operator) {
		return ParameterValueConstraintOperatorPresent
	}
	return condition.Operator
}

func (condition *ParameterValueConstraint) UsesReflect() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorEquals, ParameterValueConstraintOperatorNotEquals, ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterValueConstraint) UsesCollection() bool {
	if condition == nil {
		return false
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return true
	default:
		return false
	}
}

func (condition *ParameterValueConstraint) fieldRef(refs map[string]string) string {
	if condition == nil {
		return ""
	}
	return refs[condition.Parameter]
}

func (condition *ParameterValueConstraint) expectedLiteral() string {
	if condition == nil {
		return ""
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorIn, ParameterValueConstraintOperatorNotIn:
		return goAnySliceLiteral(condition.Values)
	default:
		return goLiteral(condition.Value)
	}
}

func (condition *ParameterValueConstraint) ValidationRuleLine(refs map[string]string, validatorExpr string) string {
	conditionRef := condition.fieldRef(refs)
	if reflection.IsEmpty(conditionRef) {
		return ""
	}
	switch condition.operatorOrDefault() {
	case ParameterValueConstraintOperatorPresent:
		return fmt.Sprintf(`validationUtils.WhenFieldMatches(%s, false, func(actual any, expected bool) (bool, error) { return validation.IsEmpty(actual) == expected, nil }, %s)`, conditionRef, validatorExpr)
	case ParameterValueConstraintOperatorAbsent:
		return fmt.Sprintf(`validationUtils.WhenFieldMatches(%s, true, func(actual any, expected bool) (bool, error) { return validation.IsEmpty(actual) == expected, nil }, %s)`, conditionRef, validatorExpr)
	case ParameterValueConstraintOperatorEquals:
		return fmt.Sprintf(`validationUtils.WhenFieldEquals(%s, %s, %s)`, conditionRef, condition.expectedLiteral(), validatorExpr)
	case ParameterValueConstraintOperatorNotEquals:
		return fmt.Sprintf(`validationUtils.WhenFieldMatches(%s, %s, func(actual any, expected any) (bool, error) { return !reflect.DeepEqual(actual, expected), nil }, %s)`, conditionRef, condition.expectedLiteral(), validatorExpr)
	case ParameterValueConstraintOperatorIn:
		return fmt.Sprintf(`validationUtils.WhenFieldMatches(%s, %s, func(actual any, expected []any) (bool, error) { return collection.In(expected, actual, func(left any, right any) (bool, error) { return reflect.DeepEqual(left, right), nil }) }, %s)`, conditionRef, condition.expectedLiteral(), validatorExpr)
	case ParameterValueConstraintOperatorNotIn:
		return fmt.Sprintf(`validationUtils.WhenFieldMatches(%s, %s, func(actual any, expected []any) (bool, error) { return !collection.In(expected, actual, func(left any, right any) (bool, error) { return reflect.DeepEqual(left, right), nil }) }, %s)`, conditionRef, condition.expectedLiteral(), validatorExpr)
	default:
		return ""
	}
}

func (cfg *ConfigurationOption) constrainedParameter() constrainedParameter {
	if cfg == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: cfg.Name, Constraints: cfg.Constraints}
}

func (cfg *FunctionConfigFields) constrainedParameter() constrainedParameter {
	if cfg == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: cfg.Name, Constraints: cfg.Constraints}
}

func (fieldDef *EntityField) constrainedParameter() constrainedParameter {
	if fieldDef == nil {
		return constrainedParameter{}
	}
	return constrainedParameter{Name: fieldDef.Name, Constraints: fieldDef.Constraints}
}

func parameterRefs(parameters []constrainedParameter) map[string]string {
	return collection.Reduce(parameters, map[string]string{}, func(acc map[string]string, parameter constrainedParameter) map[string]string {
		if reflection.IsNotEmpty(parameter.Name) {
			acc[parameter.Name] = fmt.Sprintf("&cfg.%s", toPascalCase(parameter.Name))
		}
		return acc
	})
}

func goLiteral(value any) string {
	switch typed := value.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(typed)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", typed)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", typed)
	case float32:
		return strconv.FormatFloat(float64(typed), 'g', -1, 32)
	case float64:
		if math.Trunc(typed) == typed {
			return strconv.FormatFloat(typed, 'f', 0, 64)
		}
		return strconv.FormatFloat(typed, 'g', -1, 64)
	default:
		return fmt.Sprintf("%#v", typed)
	}
}

func goAnySliceLiteral(values []interface{}) string {
	return fmt.Sprintf("[]any{%s}", strings.Join(collection.Map(values, goLiteral), ", "))
}

func parameterConstraintValidationRuleLines(parameters []constrainedParameter) []string {
	refs := parameterRefs(parameters)
	result := make([]string, 0)
	collection.ForEach(parameters, func(parameter constrainedParameter) {
		if !parameter.Constraints.HasAnyValidationRules() {
			return
		}
		targetRef, found := refs[parameter.Name]
		if !found {
			return
		}
		if reflection.IsNotEmpty(parameter.Constraints.DependentRequired) {
			dependentRefs := collection.Filter(collection.Map(collection.UniqueBy(parameter.Constraints.DependentRequired, func(name string) string { return name }), func(name string) string {
				return refs[name]
			}), func(ref string) bool { return reflection.IsNotEmpty(ref) })
			if len(dependentRefs) > 0 {
				result = append(result, fmt.Sprintf(`validationUtils.DependentRequiredBy(map[any]any{%s: []any{%s}})`, targetRef, strings.Join(dependentRefs, ", ")))
			}
		}
		if reflection.IsNotEmpty(parameter.Constraints.ConflictsWith) {
			conflictRefs := collection.Filter(collection.Map(collection.UniqueBy(parameter.Constraints.ConflictsWith, func(name string) string { return name }), func(name string) string {
				return refs[name]
			}), func(ref string) bool { return reflection.IsNotEmpty(ref) })
			if len(conflictRefs) > 0 {
				result = append(result, fmt.Sprintf(`validationUtils.MutuallyExclusiveWithBy(%s, %s)`, targetRef, strings.Join(conflictRefs, ", ")))
			}
		}
		if line := parameter.Constraints.RequiredWhen.ValidationRuleLine(refs, fmt.Sprintf(`validationUtils.RequiredPropertiesBy(%s)`, targetRef)); reflection.IsNotEmpty(line) {
			result = append(result, line)
		}
		if line := parameter.Constraints.ForbiddenWhen.ValidationRuleLine(refs, fmt.Sprintf(`validationUtils.ForbiddenPropertiesBy(%s)`, targetRef)); reflection.IsNotEmpty(line) {
			result = append(result, line)
		}
	})
	return collection.UniqueBy(result, func(line string) string { return line })
}

func (cfgs Configuration) HasConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool { return cfg != nil && cfg.Constraints.HasAnyValidationRules() })
}

func (cfgs Configuration) UsesReflectInConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg != nil && cfg.Constraints.UsesReflect()
	})
}

func (cfgs Configuration) UsesCollectionInConstraintValidationRules() bool {
	return collection.AnyRefFunc(cfgs, func(cfg *ConfigurationOption) bool {
		return cfg != nil && cfg.Constraints.UsesCollection()
	})
}

func (cfgs Configuration) ConstraintValidationRules() []string {
	return parameterConstraintValidationRuleLines(collection.Map(cfgs, func(cfg ConfigurationOption) constrainedParameter { return (&cfg).constrainedParameter() }))
}

func (fct *CliFunction) HasConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool { return cfg != nil && cfg.Constraints.HasAnyValidationRules() })
}

func (fct *CliFunction) UsesReflectInConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg != nil && cfg.Constraints.UsesReflect()
	})
}

func (fct *CliFunction) UsesCollectionInConstraintValidationRules() bool {
	return fct != nil && collection.AnyRefFunc(fct.Configuration, func(cfg *FunctionConfigFields) bool {
		return cfg != nil && cfg.Constraints.UsesCollection()
	})
}

func (fct *CliFunction) ConstraintValidationRules() []string {
	if fct == nil {
		return nil
	}
	return parameterConstraintValidationRuleLines(collection.Map(fct.Configuration, func(cfg FunctionConfigFields) constrainedParameter { return (&cfg).constrainedParameter() }))
}

func (typ *Type) HasConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.HasAnyValidationRules()
	})
}

func (typ *Type) UsesReflectInConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.UsesReflect()
	})
}

func (typ *Type) UsesCollectionInConstraintValidationRules() bool {
	return typ != nil && collection.AnyRefFunc(typ.Fields, func(fieldDef *EntityField) bool {
		return fieldDef != nil && fieldDef.Constraints.UsesCollection()
	})
}

func (typ *Type) ConstraintValidationRules() []string {
	if typ == nil {
		return nil
	}
	return parameterConstraintValidationRuleLines(collection.Map(typ.Fields, func(fieldDef EntityField) constrainedParameter { return (&fieldDef).constrainedParameter() }))
}

func (cmds *CommandLineCommands) HasConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.HasConstraintValidationRules() })
}

func (cmds *CommandLineCommands) UsesReflectInConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.UsesReflectInConstraintValidationRules() })
}

func (cmds *CommandLineCommands) UsesCollectionInConstraintValidationRules() bool {
	return cmds != nil && collection.AnyRefFunc(cmds.Functions, func(fct *CliFunction) bool { return fct.UsesCollectionInConstraintValidationRules() })
}
