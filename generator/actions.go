package generator

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

type GeneratedActionInput struct {
	Name        string
	EnvName     string
	FlagName    string
	Description string
	Required    bool
	IsBoolean   bool
	IsOperand   bool
}

type GeneratedAction struct {
	DirectoryName string
	Subcommand    string
	Name          string
	Description   string
	Inputs        []GeneratedActionInput
	HasOperands   bool
}

// BuildGeneratedActions derives generated GitHub Actions from the CLI schema.
func BuildGeneratedActions(sdk *SDK) (actions []GeneratedAction, err error) {
	if sdk == nil {
		err = commonerrors.UndefinedVariable("sdk")
		return
	}

	cmds := sdk.Specification.CommandLineCommands
	if cmds == nil {
		return
	}

	rootInputs, err := buildRootActionInputs(cmds)
	if err != nil {
		return
	}

	actionPrefix := field.OptionalString(
		sdk.Specification.Name,
		field.OptionalString(cmds.Binary, sdk.Specification.PackageName),
	)
	binaryName := field.OptionalString(cmds.Binary, sdk.Specification.PackageName)

	for _, cliFunc := range cmds.Functions {
		if cliFunc.Hidden || cliFunc.NoAction {
			continue
		}

		actionInputs, subErr := buildFunctionActionInputs(cliFunc)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, subErr, "subcommand %q action inputs", cliFunc.Name)
			return
		}

		inputs, subErr := mergeActionInputs(rootInputs, actionInputs)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrConflict, subErr, "subcommand %q action inputs conflict", cliFunc.Name)
			return
		}

		actionName := fmt.Sprintf("%s - %s Action", actionPrefix, toPascalCase(cliFunc.Name))
		actionDescription := fmt.Sprintf("Install and optionally run %s %s from GitHub Releases.", binaryName, strings.ToLower(cliFunc.Name))

		actions = append(actions, GeneratedAction{
			DirectoryName: normaliseActionDirectoryName(cliFunc.Name),
			Subcommand:    strings.ToLower(cliFunc.Name),
			Name:          actionName,
			Description:   actionDescription,
			Inputs:        inputs,
			HasOperands:   cliFunc.Operands,
		})
	}

	return
}

func buildRootActionInputs(cmds *CommandLineCommands) (inputs []GeneratedActionInput, err error) {
	if cmds == nil {
		return
	}

	seen := map[string]bool{}

	for _, cliFunc := range cmds.Functions {
		for _, cfg := range cliFunc.Configuration {
			if !cfg.IsRoot || cfg.Hidden {
				continue
			}

			cfgInputs, subErr := buildActionInputsFromConfig(cfg)
			if subErr != nil {
				err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, subErr, "root configuration %q", cfg.Name)
				return
			}

			for _, input := range cfgInputs {
				if seen[input.Name] {
					continue
				}

				seen[input.Name] = true
				inputs = append(inputs, input)
			}
		}
	}

	return
}

func buildFunctionActionInputs(cliFunc CliFunction) (inputs []GeneratedActionInput, err error) {
	for _, cfg := range cliFunc.Configuration {
		if cfg.IsRoot || cfg.Hidden {
			continue
		}

		cfgInputs, subErr := buildActionInputsFromConfig(cfg)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, subErr, "configuration %q", cfg.Name)
			return
		}

		inputs = append(inputs, cfgInputs...)
	}

	if cliFunc.Operands {
		inputs = append(inputs, GeneratedActionInput{
			Name:        "operands",
			EnvName:     "OPERANDS",
			Description: "Optional positional operands, one per line.",
			IsOperand:   true,
		})
	}

	return
}

func mergeActionInputs(bases []GeneratedActionInput, extras []GeneratedActionInput) (out []GeneratedActionInput, err error) {
	seen := make(map[string]bool, len(bases)+len(extras))

	for _, base := range bases {
		seen[base.Name] = true
		out = append(out, base)
	}

	for _, extra := range extras {
		if seen[extra.Name] {
			err = commonerrors.Newf(commonerrors.ErrConflict, "duplicate action input %q", extra.Name)
			return
		}

		seen[extra.Name] = true
		out = append(out, extra)
	}

	return
}

func buildActionInputsFromConfig(cfg FunctionConfigFields) (inputs []GeneratedActionInput, err error) {
	required := cfg.Required && cfg.DefaultValue == nil && cfg.Condition == ""

	if len(cfg.Bindings) == 0 {
		flagName, subErr := determineActionInputFlagName(cfg, nil)
		if subErr != nil {
			err = subErr
			return
		}

		inputs = append(inputs, GeneratedActionInput{
			Name:        flagName,
			EnvName:     toActionInputEnvName(flagName),
			FlagName:    flagName,
			Description: describeActionInput(cfg.Description, cfg.Condition),
			Required:    required,
			IsBoolean:   strings.TrimSpace(cfg.LangType) == "bool",
		})

		return
	}

	for _, binding := range cfg.Bindings {
		if binding.Hidden {
			continue
		}

		flagName, subErr := determineActionInputFlagName(cfg, &binding)
		if subErr != nil {
			err = subErr
			return
		}

		description := cfg.Description
		if !reflection.IsEmpty(binding.Description) {
			description = field.Optional(binding.Description, cfg.Description)
		}

		langType := field.Optional(binding.LangType, cfg.LangType)

		inputs = append(inputs, GeneratedActionInput{
			Name:        flagName,
			EnvName:     toActionInputEnvName(flagName),
			FlagName:    flagName,
			Description: describeActionInput(description, cfg.Condition),
			Required:    required,
			IsBoolean:   strings.TrimSpace(langType) == "bool",
		})
	}

	return
}

func determineActionInputFlagName(cfg FunctionConfigFields, binding *Binding) (flagName string, err error) {
	switch {
	case binding != nil:
		switch {
		case !reflection.IsEmpty(binding.Arg):
			flagName = field.Optional(binding.Arg, "")

		case !reflection.IsEmpty(binding.MapstructureTag):
			flagName = replaceUnderscores(field.Optional(binding.MapstructureTag, ""))

		default:
			flagName = replaceUnderscores(cfg.Name)
		}
	case !reflection.IsEmpty(cfg.Arg):
		flagName = field.Optional(cfg.Arg, "")
	default:
		flagName = replaceUnderscores(cfg.MapstructureTag)
	}

	if reflection.IsEmpty(flagName) {
		err = commonerrors.UndefinedVariable("flag name")
		return
	}

	return
}

func toActionInputEnvName(inputName string) (inputEnvName string) {
	replaced := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	inputEnvName = strings.ToUpper(replaced.Replace(inputName))
	return
}

func describeActionInput(description string, condition string) (newDescription string) {
	if strings.TrimSpace(condition) == "" {
		newDescription = description
		return
	}

	newDescription = fmt.Sprintf("%s Required when %s.", description, condition)
	return
}
