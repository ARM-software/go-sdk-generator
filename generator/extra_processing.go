package generator

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	mapset "github.com/deckarep/golang-set/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func usualExtraImports() map[string]string {
	return map[string]string{
		"xdg":        "github.com/adrg/xdg",
		"filesystem": "github.com/ARM-software/golang-utils/utils/filesystem",
		"httpUtils":  "github.com/ARM-software/golang-utils/utils/http",
		"platform":   "github.com/ARM-software/golang-utils/utils/platform",
	}
}

func ProcessExtraImports(ctx context.Context, sdk *SDK) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if sdk == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}
	extra := usualExtraImports()
	collection.ForEachRef(sdk.Specification.ExtraImports, func(extraImport *ExtraImport) {
		if reflection.IsEmpty(extraImport.ImportUrl) {
			return
		}
		importPath := field.Optional(extraImport.ImportUrl, "")
		extra[field.OptionalString(extraImport.ImportRef, path.Base(importPath))] = importPath
	})
	extraImports := make(ExtraImports, 0, len(extra))
	for k, v := range extra {
		extraImports = append(extraImports, ExtraImport{
			ImportRef: field.ToOptionalStringOrNilIfEmpty(k),
			ImportUrl: field.ToOptionalStringOrNilIfEmpty(v),
		})
	}
	sdk.Specification.ExtraImports = extraImports
	return
}

func ProcessRepositoryAndModuleMetadata(ctx context.Context, sdk *SDK) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if sdk == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}

	sdk.Specification.Interface.Namespace = normaliseNamespace(sdk.Specification.Interface.Namespace)
	if sdk.Specification.CommandLineCommands != nil {
		sdk.Specification.CommandLineCommands.Package = normaliseCLIPackageName(sdk.Specification.CommandLineCommands.Package)
	}

	repoURL := repositoryURL(&sdk.Specification)
	err = validateURL(repoURL)
	if err != nil {
		return
	}

	sdk.Specification.RepositoryUrl = field.ToOptionalString(repoURL)
	disableActionGenerationWhenNotHostedOnGitHub(&sdk.Specification)
	return
}

func ProcessAdditionalTypePackages(ctx context.Context, sdk *SDK) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if sdk == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}

	additionalTypesPackages, err := sdk.Specification.AdditionalTypes.NormalisePackages()
	if err != nil {
		return
	}
	sdk.Specification.RewriteAdditionalTypeReferences(additionalTypesPackages)

	return
}

func validateURL(value string) error {
	err := validation.Validate(value, validation.Required, is.URL)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "expected a URL, got %q", value)
	}
	return nil
}

func disableActionGenerationWhenNotHostedOnGitHub(spec *SDKSpecification) {
	if spec == nil || spec.CommandLineCommands == nil || isHostedOnGitHub(spec) {
		return
	}

	spec.CommandLineCommands.NoActions = true
	for i := range spec.CommandLineCommands.Functions {
		spec.CommandLineCommands.Functions[i].NoAction = true
	}
}

func ProcessBindings(ctx context.Context, sdk *SDK) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if sdk == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}

	bindings := mapset.NewSet[string]()
	collection.ForEach(sdk.Specification.ExtraValidTypes, func(extra string) {
		bindings.Add(extra)
	})
	collection.ForEach(sdk.Specification.GlobalConfiguration, func(cfg ConfigurationOption) {
		bindings.Add(cfg.LangType)
		collection.ForEach(cfg.Bindings, func(binding Binding) {
			bindings.Add(field.OptionalString(binding.LangType, ""))
		})
	})
	collection.ForEach(sdk.Specification.Interface.Methods, func(method InterfaceMethod) {
		collection.ForEach(method.Args, func(argument CommandArgument) {
			collection.ForEach(argument.Bindings, func(binding Binding) {
				bindings.Add(field.OptionalString(binding.LangType, ""))
			})
		})
	})

	if !reflection.IsEmpty(sdk.Specification.CommandLineCommands) {
		collection.ForEach(sdk.Specification.CommandLineCommands.Functions, func(fct CliFunction) {
			collection.ForEach(fct.Configuration, func(fctCfg FunctionConfigFields) {
				bindings.Add(fctCfg.LangType)
				collection.ForEach(fctCfg.Bindings, func(binding Binding) {
					bindings.Add(field.OptionalString(binding.LangType, ""))
				})
			})
		})
	}
	types := bindings.ToSlice()
	validTypes := make(ExtraValidTypes, 0, len(types))
	for i := range types {
		validTypes = append(validTypes, types[i])
	}
	sdk.Specification.ExtraValidTypes = validTypes
	return
}

func processArguments(arg *string, argShort *string, mapstructureTag *string, seenShortArguments mapset.Set[string]) (err error) {
	if (arg == nil && mapstructureTag == nil) ||
		(mapstructureTag != nil && reflection.IsEmpty(*mapstructureTag)) {
		err = commonerrors.New(commonerrors.ErrUndefined, "A parameter must have at least one long argument (or mapstructure tag) to bind to")
		return
	}
	if argShort == nil {
		return
	}

	short := *argShort
	if seenShortArguments.Contains(short) {
		err = commonerrors.Newf(commonerrors.ErrConflict, "cannot specify duplicate short argument %v", short)
		return
	}
	seenShortArguments.Add(short)
	return
}

// processCallCloud enforces that
// callCloud at the root level commands sets the associated arguments as root for all functions except for those storing credentials to keyring (storeToKeyring true)
// callCloud at the function level sets the associated arguments as root for all functions that have callCloud true when all functions have either callCloud true or storeToKeyring true
// callCloud at the function level sets the associated arguments as non-root when some or all functions neither have callCloud true nor storeToKeyring true and root level callCloud is not set
func processCallCloud(cmds *CommandLineCommands) (err error) {
	if reflection.IsEmpty(cmds) {
		return
	}

	shouldPromoteCallCloudToRoot := cmds.HasCloudCommands()
	if shouldPromoteCallCloudToRoot {
		cmds.CallCloud = field.ToOptionalBool(true)
	}

	if hasMixedRootAuth(cmds) {
		err = commonerrors.New(commonerrors.ErrInvalid, "cannot mix function-level callCloud with explicit root APIKey/Host definitions; use top-level callCloud or make all relevant commands use callCloud")
		return
	}

	rootCallCloudEnabled := field.Optional(cmds.CallCloud, false)
	for i := range cmds.Functions {
		fct := cmds.Functions[i]
		if !shouldInjectCallCloudArgs(fct, rootCallCloudEnabled) {
			continue
		}

		subErr := updateCliFunctionArgumentsToCallCloud(&fct, rootCallCloudEnabled)
		if subErr != nil {
			err = subErr
			return
		}
		cmds.Functions[i] = fct
	}
	return
}

// shouldInjectCallCloudArgs reports whether cloud auth arguments should be injected for a command based on the active callCloud mode
func shouldInjectCallCloudArgs(fct CliFunction, callCloudPromotedToRoot bool) bool {
	if callCloudPromotedToRoot {
		return !fct.StoreToKeyring
	}

	return fct.CallCloud
}

// isCallCloudTrue checks if all functions either have CallCloud set to true or are storing credentials to keyring.
func isCallCloudTrue(fcts []CliFunction) bool {
	cmds := &CommandLineCommands{Functions: fcts}
	return cmds.HasCloudCommands()
}

// hasMixedRootAuth reports whether the CLI manifest mixes function-level callCloud usage with explicit root-scoped cloud auth configuration.
func hasMixedRootAuth(cmds *CommandLineCommands) bool {
	if cmds == nil || field.Optional(cmds.CallCloud, false) {
		return false
	}

	hasCallCloud := collection.AnyFunc(cmds.Functions, func(fct CliFunction) bool {
		return fct.CallCloud
	})

	hasExplicitAuth := collection.AnyFunc(cmds.Functions, func(fct CliFunction) bool {
		return collection.AnyFunc(fct.Configuration, func(cfg FunctionConfigFields) bool {
			if !cfg.IsRoot {
				return false
			}

			if cfg.Name == "APIKey" && cfg.MapstructureTag == "api_key" {
				return true
			}

			if cfg.Name == "Host" && cfg.MapstructureTag == "host" {
				return true
			}

			return false
		})
	})

	return hasCallCloud && hasExplicitAuth
}

// ProcessArguments enforces that arguments are specified on commands and short arguments aren't duplicated
func ProcessArguments(ctx context.Context, sdk *SDK) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if sdk == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}
	err = validateCollectionParameters(sdk)
	if err != nil {
		return
	}
	if !reflection.IsEmpty(sdk.Specification.CommandLineCommands) {
		err = processCallCloud(sdk.Specification.CommandLineCommands)
		if err != nil {
			return
		}

		err = validateSecretParameters(sdk.Specification.CommandLineCommands)
		if err != nil {
			return
		}

		err = processCredentialsCommand(sdk.Specification.CommandLineCommands)
		if err != nil {
			return
		}

		for i := range sdk.Specification.CommandLineCommands.Functions {
			fct := sdk.Specification.CommandLineCommands.Functions[i]
			seenShortArguments := mapset.NewSetFromMapKeys(map[string]any{"h": nil, "v": nil}) // -h and -v come provided
			if fct.StoreToKeyring && fct.LoadFromKeyring {
				return commonerrors.Newf(commonerrors.ErrInvalid, "Issue with function [%v]: CLI configuration can either be loaded from system's keyring or stored there but not both", fct.Name)
			}
			if fct.LoadFromKeyring {
				subErr := updateCliFunctionArgumentsToLoadFromKeyring(&fct)
				if subErr != nil {
					return subErr
				}
			}
			subErr := collection.Each(slices.Values(fct.Configuration), func(param FunctionConfigFields) error {
				subErr := collection.Each(slices.Values(param.Bindings), func(binding Binding) error {
					return processArguments(binding.Arg, binding.ArgShort, binding.MapstructureTag, seenShortArguments)
				})
				if subErr != nil {
					return subErr
				}
				return processArguments(param.Arg, param.ArgShort, &param.MapstructureTag, seenShortArguments)
			})
			if subErr != nil {
				return subErr
			}
			sdk.Specification.CommandLineCommands.Functions[i] = fct
		}
	}
	return
}

func updateCliFunctionArgumentsToLoadFromKeyring(fct *CliFunction) error {
	args := map[string]FunctionConfigFields{
		"NoKeyring": {
			ArgShort:        nil,
			Description:     "Disable loading configuration from the system keyring",
			IsRoot:          false,
			LangType:        "bool",
			MapstructureTag: "no_keyring",
			Name:            "NoKeyring",
			Pattern:         nil,
			Reference:       false,
			Required:        false,
		},
	}
	return updateCliFunctionArguments(fct, args)
}

func updateCliFunctionArgumentsToCallCloud(fct *CliFunction, isRoot bool) error {
	if fct == nil {
		return commonerrors.UndefinedVariable("command line function")
	}
	fct.LoadFromKeyring = !fct.StoreToKeyring
	args := map[string]FunctionConfigFields{
		"APIKey": {
			ArgShort:        field.ToOptionalStringOrNilIfEmpty("k"),
			Description:     "An API key enabling the use of the online services",
			IsRoot:          isRoot,
			LangType:        "string",
			MapstructureTag: "api_key",
			Name:            "APIKey",
			Pattern:         field.ToOptionalStringOrNilIfEmpty(`^[!-~\w]+[!-~\w\s]*$`),
			Reference:       false,
			Required:        true,
			Secret:          true,
		},
		"Host": {
			Arg:             nil,
			ArgShort:        nil,
			Bindings:        nil,
			Condition:       "",
			DefaultValue:    field.ToOptionalStringOrNilIfEmpty("https://all.api.keil.arm.com"),
			Description:     "The host to use for accessing the online services",
			Hidden:          false,
			IsRoot:          isRoot,
			JsonTag:         nil,
			LangType:        "host",
			MapstructureTag: "host",
			Name:            "Host",
			Pattern:         nil,
			Reference:       false,
			Required:        true,
		},
	}
	return updateCliFunctionArguments(fct, args)
}

func updateCliFunctionArguments(fct *CliFunction, additionalArguments map[string]FunctionConfigFields) error {
	if fct == nil {
		return commonerrors.UndefinedVariable("command line function")
	}
	if additionalArguments == nil {
		additionalArguments = make(map[string]FunctionConfigFields)
	}

	collection.ForEach(fct.Configuration, func(arg FunctionConfigFields) {
		additionalArguments[arg.Name] = arg
	})
	fct.Configuration = slices.Collect(maps.Values(additionalArguments))
	slices.SortFunc(fct.Configuration, func(a, b FunctionConfigFields) int {
		return compareArgs(a.Name, b.Name)
	})
	return nil
}

type ArgDefinition struct {
	Name   string
	Value  string
	NoLint bool
}

type FlagRegistration struct {
	FlagLine   string
	BindingEnv string
	ArgVar     string
}

type ArgFlagBindingInfo struct {
	EnvVar       string
	FlagBindings []string
}

func getArgName(prefix, cfgName, bindingName string, isRoot bool) string {
	if isRoot {
		return fmt.Sprintf("root%vArg", bindingName)
	}

	return fmt.Sprintf("%v%v%vArg", toCamelCase(prefix), toPascalCase(cfgName), toPascalCase(bindingName))
}

func buildFlagLine(isRoot, hidden bool, funcName, flagType, argName, argShort, zeroVal, desc string) string {
	cmdPrefix := "rootCmd.PersistentFlags()"
	if !isRoot {
		cmdPrefix = fmt.Sprintf("%v.Flags()", funcName)
	}

	var hiddenStr string
	if hidden {
		hiddenStr = fmt.Sprintf("\n_ = %v.MarkHidden(%v)", cmdPrefix, argName)
	}

	if argShort != "" {
		return fmt.Sprintf(`%v.%vP(%v, "%v", %v, "%v")%v`, cmdPrefix, flagType, argName, argShort, zeroVal, desc, hiddenStr)
	}

	return fmt.Sprintf(`%v.%v(%v, %v, "%v")%v`, cmdPrefix, flagType, argName, zeroVal, desc, hiddenStr)
}

func getBindingEnv(cfgMapstructureTag string, bindingMapstructureTag *string, prefix string) string {
	if !reflection.IsEmpty(bindingMapstructureTag) {
		return fmt.Sprintf("%v%v", prefix, *bindingMapstructureTag)
	}

	return cfgMapstructureTag
}

func addFlagLookup(flagLookupMap map[string][]string, envVar, flagPtr string) {
	flagLookupMap[envVar] = append(flagLookupMap[envVar], flagPtr)
}

func determineNameFromPointer(arg *string, variableName string) (name string, err error) {
	name = field.Optional(arg, "")
	if name == "" {
		err = commonerrors.UndefinedVariable(variableName)
		return
	}
	return
}

func determineName(name string, variableName string) (nameOut string, err error) {
	if name == "" {
		err = commonerrors.UndefinedVariable(variableName)
		return
	}
	nameOut = name
	return
}

func parseBindingToFlag(cliFunc CliFunction, cfg FunctionConfigFields, binding Binding) (flagLine, envVar, flagString, argName string, err error) {
	funcName := fmt.Sprintf("%vCmd", strings.ToLower(cliFunc.Name))
	prefix := fmt.Sprintf("%v_", cfg.MapstructureTag)

	bindingName, err := determineNameFromPointer(binding.Name, "name")
	if err != nil {
		return
	}

	langType, err := determineNameFromPointer(binding.LangType, "lang_type")
	if err != nil {
		return
	}

	flagType := argFlagType(langType)
	zeroVal := determineZeroValue(langType)

	description := binding.Description
	if cfg.IsRoot {
		description = &cfg.Description
	}

	desc, err := determineNameFromPointer(description, "description")
	if err != nil {
		return
	}
	desc = descriptionWithEnum(desc, cfg.Enum)

	argShort := field.Optional(binding.ArgShort, "")

	argName = getArgName(cliFunc.Name, cfg.Name, bindingName, cfg.IsRoot)
	flagLine = buildFlagLine(cfg.IsRoot, binding.Hidden, funcName, flagType, argName, argShort, zeroVal, desc)

	if cfg.IsRoot {
		envVar, err = determineNameFromPointer(binding.MapstructureTag, "mapstructure_tag")
		if err != nil {
			return
		}
	} else {
		envVar = getBindingEnv(cfg.MapstructureTag, binding.MapstructureTag, prefix)
	}

	flagString = fmt.Sprintf("%v.Flags().Lookup(%v)", funcName, argName)

	return
}

func parseNonBindingToFlag(cliFunc CliFunction, cfg FunctionConfigFields) (envVar, flagLine, flagString, argName string, err error) {
	langType, err := determineName(cfg.LangType, "lang_type")
	if err != nil {
		return
	}

	flagType := argFlagType(langType)
	zeroVal := determineZeroValue(langType)

	desc, err := determineName(cfg.Description, "description")
	if err != nil {
		return
	}
	desc = descriptionWithEnum(desc, cfg.Enum)

	argShort := ""
	if cfg.ArgShort != nil {
		argShort = *cfg.ArgShort
	}

	funcName := fmt.Sprintf("%vCmd", strings.ToLower(cliFunc.Name))
	argName = getArgName(cliFunc.Name, "", cfg.Name, cfg.IsRoot)
	flagLine = buildFlagLine(cfg.IsRoot, cfg.Hidden, funcName, flagType, argName, argShort, zeroVal, desc)

	envVar = cfg.MapstructureTag
	if !cfg.IsRoot {
		envVar, err = determineName(cfg.MapstructureTag, "mapstructure_tag")
		if err != nil {
			return
		}
	}

	flagString = fmt.Sprintf("%v.Flags().Lookup(%v)", funcName, argName)

	return
}

func preComputeArgNames(cmds *CommandLineCommands, isRoot bool) (out []ArgDefinition, err error) {
	if cmds == nil {
		return
	}

	seen := make(map[string]bool)

	for _, cliFunc := range cmds.Functions {
		for _, cfg := range cliFunc.Configuration {
			if cfg.IsRoot != isRoot {
				continue
			}

			parentName := replaceUnderscores(cfg.Name)
			if len(cfg.Bindings) > 0 {
				for _, binding := range cfg.Bindings {
					bindingName, subErr := determineNameFromPointer(binding.Name, "name")
					if subErr != nil {
						err = subErr
						return
					}

					argName := getArgName(cliFunc.Name, cfg.Name, bindingName, isRoot)
					if seen[argName] {
						continue
					}
					seen[argName] = true

					var value string
					switch {
					case !reflection.IsEmpty(binding.Arg):
						value = *binding.Arg
					case !reflection.IsEmpty(binding.MapstructureTag):
						value = replaceUnderscores(*binding.MapstructureTag)
					default:
						value = parentName
					}

					if value == "" {
						err = commonerrors.UndefinedVariable("value")
						return
					}

					out = append(out, ArgDefinition{
						Name:   argName,
						Value:  value,
						NoLint: binding.Secret || ignoreLintingError(argName) || ignoreLintingError(value),
					})
				}
			} else {
				argName := getArgName(cliFunc.Name, "", cfg.Name, isRoot)
				if seen[argName] {
					continue
				}
				seen[argName] = true

				var value string
				if !reflection.IsEmpty(cfg.Arg) {
					value = *cfg.Arg
				} else {
					value = replaceUnderscores(cfg.MapstructureTag)
				}

				if value == "" {
					err = commonerrors.UndefinedVariable("value")
					return
				}

				out = append(out, ArgDefinition{
					Name:   argName,
					Value:  value,
					NoLint: cfg.Secret || ignoreLintingError(argName) || ignoreLintingError(value),
				})
			}
		}
	}

	slices.SortFunc(out, func(a, b ArgDefinition) int {
		return compareArgs(a.Name, b.Name)
	})

	return
}

// PreComputeFlagsAndBindings will determine the arguments and how to bind them to the environment variables for the non-root command line functions
func PreComputeFlagsAndBindings(cmds *CommandLineCommands) (flagRegistrations []FlagRegistration, bindings []ArgFlagBindingInfo, err error) {
	if cmds == nil {
		return
	}

	flagLookupMap := make(map[string][]string)

	for _, cliFunc := range cmds.Functions {
		for _, cfg := range cliFunc.Configuration {
			if cfg.IsRoot {
				continue
			}

			if len(cfg.Bindings) > 0 {
				for _, binding := range cfg.Bindings {
					flagLine, envVar, flagName, _, subErr := parseBindingToFlag(cliFunc, cfg, binding)
					if subErr != nil {
						err = subErr
						return
					}

					flagRegistrations = append(flagRegistrations, FlagRegistration{FlagLine: flagLine})
					addFlagLookup(flagLookupMap, envVar, flagName)
				}
			} else {
				envVar, flagLine, flagString, _, subErr := parseNonBindingToFlag(cliFunc, cfg)
				if subErr != nil {
					err = subErr
					return
				}

				flagRegistrations = append(flagRegistrations, FlagRegistration{FlagLine: flagLine})
				addFlagLookup(flagLookupMap, envVar, flagString)
			}
		}
	}

	for envVar, flagBindings := range flagLookupMap {
		bindings = append(bindings, ArgFlagBindingInfo{
			EnvVar:       envVar,
			FlagBindings: flagBindings,
		})
	}

	slices.SortFunc(flagRegistrations, func(a, b FlagRegistration) int {
		return compareArgs(a.ArgVar, b.ArgVar)
	})

	slices.SortFunc(bindings, func(a, b ArgFlagBindingInfo) int {
		return compareEnv(a.EnvVar, b.EnvVar)
	})
	return
}

// PreComputeRootFlagsAndBindings will determine the arguments and how to bind them to the environment variables for the root command line functions
func PreComputeRootFlagsAndBindings(cmds *CommandLineCommands) (flagRegistration []FlagRegistration, bindings []ArgFlagBindingInfo, err error) {
	if cmds == nil {
		return
	}

	flagLookupMap := make(map[string][]string)

	for _, cliFunc := range cmds.Functions {
		for _, cfg := range cliFunc.Configuration {
			if !cfg.IsRoot {
				continue
			}

			if len(cfg.Bindings) > 0 {
				for _, binding := range cfg.Bindings {
					flagLine, bindingEnv, _, argName, subErr := parseBindingToFlag(cliFunc, cfg, binding)
					if subErr != nil {
						err = subErr
						return
					}

					if !reflection.IsEmpty(flagLookupMap[bindingEnv]) {
						continue
					}

					flagRegistration = append(flagRegistration, FlagRegistration{
						FlagLine:   flagLine,
						BindingEnv: bindingEnv,
						ArgVar:     argName,
					})

					flagLookup := fmt.Sprintf("rootCmd.PersistentFlags().Lookup(%v)", argName)
					addFlagLookup(flagLookupMap, bindingEnv, flagLookup)
				}
			} else {

				bindingEnv, flagLine, _, argName, subErr := parseNonBindingToFlag(cliFunc, cfg)
				if subErr != nil {
					err = subErr
					return
				}

				if !reflection.IsEmpty(flagLookupMap[bindingEnv]) {
					continue
				}

				flagRegistration = append(flagRegistration, FlagRegistration{
					FlagLine:   flagLine,
					BindingEnv: bindingEnv,
					ArgVar:     argName,
				})

				flagLookup := fmt.Sprintf("rootCmd.PersistentFlags().Lookup(%v)", argName)
				addFlagLookup(flagLookupMap, bindingEnv, flagLookup)
			}
		}
	}

	for envVar, matchingArgs := range flagLookupMap {
		bindings = append(bindings, ArgFlagBindingInfo{
			EnvVar:       envVar,
			FlagBindings: matchingArgs,
		})
	}

	slices.SortFunc(flagRegistration, func(a, b FlagRegistration) int {
		return compareArgs(a.ArgVar, b.ArgVar)
	})

	slices.SortFunc(bindings, func(a, b ArgFlagBindingInfo) int {
		return compareEnv(a.EnvVar, b.EnvVar)
	})

	return
}

// PrecomputeArgNames will determine the names of the command line arguments so they match the flag bindings for the non-root command line functions
func PrecomputeArgNames(cmds *CommandLineCommands) (out []ArgDefinition, err error) {
	return preComputeArgNames(cmds, false)
}

// PrecomputeRootArgNames will determine the names of the command line arguments so they match the flag bindings for the root command line functions
func PrecomputeRootArgNames(cmds *CommandLineCommands) (out []ArgDefinition, err error) {
	return preComputeArgNames(cmds, true)
}

func compareArgs(a, b string) int {
	return strings.Compare(toPascalCase(a), toPascalCase(b))
}

func compareEnv(a, b string) int {
	return strings.Compare(toSnakeCase(a), toSnakeCase(b))
}
