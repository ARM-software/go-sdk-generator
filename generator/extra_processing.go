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
