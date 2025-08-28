package generator

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	mapset "github.com/deckarep/golang-set/v2"
)

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
	for i := range sdk.Specification.ExtraValidTypes {
		bindings.Add(sdk.Specification.ExtraValidTypes[i])
	}
	for i := range sdk.Specification.GlobalConfiguration {
		binds := sdk.Specification.GlobalConfiguration[i].Bindings
		bindings.Add(sdk.Specification.GlobalConfiguration[i].LangType)
		for j := range binds {
			binding := binds[j]
			bindings.Add(field.OptionalString(binding.LangType, ""))
		}
	}
	for i := range sdk.Specification.Interface.Methods {
		args := sdk.Specification.Interface.Methods[i].Args
		for j := range args {
			binds := args[j].Bindings
			for k := range binds {
				binding := binds[k]
				bindings.Add(field.OptionalString(binding.LangType, ""))
			}
		}
	}
	if !reflection.IsEmpty(sdk.Specification.CommandLineCommands) {
		for i := range sdk.Specification.CommandLineCommands.Functions {
			fctCfg := sdk.Specification.CommandLineCommands.Functions[i].Configuration
			for j := range fctCfg {
				bindings.Add(fctCfg[j].LangType)
				binds := fctCfg[j].Bindings
				for k := range binds {
					binding := binds[k]
					bindings.Add(field.OptionalString(binding.LangType, ""))
				}
			}
		}
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
	if !reflection.IsEmpty(sdk.Specification.CommandLineCommands) {
		for _, cliFunc := range sdk.Specification.CommandLineCommands.Functions {
			seenShortArguments := mapset.NewSetFromMapKeys(map[string]any{"h": nil, "v": nil}) // -h and -v come provided

			for _, param := range cliFunc.Configuration {
				for _, binding := range param.Bindings {
					err = processArguments(binding.Arg, binding.ArgShort, binding.MapstructureTag, seenShortArguments)
					if err != nil {
						return
					}
				}

				err = processArguments(param.Arg, param.ArgShort, &param.MapstructureTag, seenShortArguments)
				if err != nil {
					return
				}
			}
		}
	}

	return
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
		hiddenStr = fmt.Sprintf("\n%v.MarkHidden(%v)", cmdPrefix, argName)
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
						NoLint: ignoreLintingError(argName) || ignoreLintingError(value),
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
					NoLint: ignoreLintingError(argName) || ignoreLintingError(value),
				})
			}
		}
	}

	slices.SortFunc(out, func(a, b ArgDefinition) int {
		return strings.Compare(a.Name, b.Name)
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
		return strings.Compare(a.ArgVar, b.ArgVar)
	})

	slices.SortFunc(bindings, func(a, b ArgFlagBindingInfo) int {
		return strings.Compare(a.EnvVar, b.EnvVar)
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
		return strings.Compare(a.ArgVar, b.ArgVar)
	})

	slices.SortFunc(bindings, func(a, b ArgFlagBindingInfo) int {
		return strings.Compare(a.EnvVar, b.EnvVar)
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
