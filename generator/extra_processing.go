package generator

import (
	"context"

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
