/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"maps"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
)

const (
	credentialsCommandName        = "Credentials"
	credentialsCommandDescription = "Store credentials securely in the platform" //nolint:gosec // descriptive constant, not a credential
	stringLangType                = "string"
	secretProtectedStringType     = "secretutil.ProtectedString" //nolint:gosec // type name, not a credential
	secretTypeValidationMessage   = "must resolve to a string-compatible type when marked as secret"
)

func isPointerLangType(langType string) bool {
	return strings.HasPrefix(strings.TrimSpace(langType), "*")
}

func baseLangType(langType string) string {
	return strings.TrimPrefix(strings.TrimSpace(langType), "*")
}

func isReferenceLikeSecretField(cfg FunctionConfigFields) bool {
	return cfg.Reference || isPointerLangType(cfg.LangType)
}

var secretCompatibleLangTypeRule = validation.By(func(value any) error {
	langType, ok := value.(string)
	if !ok {
		return commonerrors.Newf(commonerrors.ErrInvalid, "expected a string language type, got %T", value)
	}
	if isSecretCompatibleLangType(langType) {
		return nil
	}
	return commonerrors.New(commonerrors.ErrInvalid, secretTypeValidationMessage)
})

// hasSecretStringParameters reports whether any generated CLI configuration field
// is both marked as secret and represented as a string. This is used to decide
// whether the generated CLI needs the local `secretutil` helper package.
func hasSecretStringParameters(cmds *CommandLineCommands) bool {
	return cmds.HasSecretStringParameters()
}

func validateSecretParameters(cmds *CommandLineCommands) error {
	if cmds == nil {
		return nil
	}

	return collection.ForAll(cmds.Functions, func(fct CliFunction) error {
		return collection.ForAll(fct.Configuration, func(cfg FunctionConfigFields) error {
			if cfg.Secret {
				err := validation.Validate(cfg.LangType, validation.Required, secretCompatibleLangTypeRule)
				if err != nil {
					return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "function %q configuration %q is marked as secret but has incompatible type %q", fct.Name, cfg.Name, cfg.LangType)
				}
			}

			return collection.ForAll(cfg.Bindings, func(binding Binding) error {
				if !binding.Secret {
					return nil
				}

				bindingLangType := field.OptionalString(binding.LangType, cfg.LangType)
				err := validation.Validate(bindingLangType, validation.Required, secretCompatibleLangTypeRule)
				if err != nil {
					return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "function %q configuration %q binding %q is marked as secret but has incompatible type %q", fct.Name, cfg.Name, field.OptionalString(binding.Name, ""), bindingLangType)
				}
				return nil
			})
		})
	})
}

func isSecretCompatibleLangType(langType string) bool {
	return convertLangType(baseLangType(langType)) == stringLangType
}

func processCredentialsCommand(cmds *CommandLineCommands) error {
	if cmds == nil {
		return nil
	}

	secrets := collectSecretConfiguration(cmds.Functions)
	if len(secrets) == 0 {
		return nil
	}

	credentialsIndex := slices.IndexFunc(cmds.Functions, func(fct CliFunction) bool {
		return fct.Name == credentialsCommandName
	})

	if credentialsIndex < 0 {
		cmds.Functions = append(cmds.Functions, CliFunction{
			Name:           credentialsCommandName,
			Description:    credentialsCommandDescription,
			Configuration:  secrets,
			StoreToKeyring: true,
			NoAction:       true,
		})
		return nil
	}

	credentials := cmds.Functions[credentialsIndex]
	credentials.StoreToKeyring = true
	credentials.LoadFromKeyring = false
	credentials.NoAction = true
	credentials.Configuration = mergeCredentialConfiguration(credentials.Configuration, secrets)
	cmds.Functions[credentialsIndex] = credentials
	return nil
}

func collectSecretConfiguration(functions []CliFunction) []FunctionConfigFields {
	collected := make(map[string]FunctionConfigFields)
	collection.ForEach(functions, func(fct CliFunction) {
		if fct.Name == credentialsCommandName {
			return
		}
		collection.ForEach(fct.Configuration, func(cfg FunctionConfigFields) {
			secretCfg, ok := secretConfiguration(cfg)
			if !ok {
				return
			}
			if existing, found := collected[secretCfg.Name]; found {
				collected[secretCfg.Name] = mergeCredentialField(existing, secretCfg)
				return
			}
			collected[secretCfg.Name] = secretCfg
		})
	})

	return sortCredentialConfiguration(slices.Collect(maps.Values(collected)))
}

func secretConfiguration(cfg FunctionConfigFields) (FunctionConfigFields, bool) {
	secretBindings := make([]Binding, 0, len(cfg.Bindings))
	collection.ForEach(cfg.Bindings, func(binding Binding) {
		if binding.Secret {
			secretBindings = append(secretBindings, binding)
		}
	})

	if !cfg.Secret && len(secretBindings) == 0 {
		return FunctionConfigFields{}, false
	}

	secretCfg := cfg
	secretCfg.IsRoot = false
	if len(secretBindings) > 0 && !cfg.Secret {
		secretCfg.Bindings = secretBindings
	}
	return secretCfg, true
}

func mergeCredentialConfiguration(existing []FunctionConfigFields, additional []FunctionConfigFields) []FunctionConfigFields {
	merged := make(map[string]FunctionConfigFields)
	collection.ForEach(existing, func(cfg FunctionConfigFields) {
		merged[cfg.Name] = cfg
	})
	collection.ForEach(additional, func(cfg FunctionConfigFields) {
		if current, found := merged[cfg.Name]; found {
			merged[cfg.Name] = mergeCredentialField(current, cfg)
			return
		}
		merged[cfg.Name] = cfg
	})
	return sortCredentialConfiguration(slices.Collect(maps.Values(merged)))
}

func mergeCredentialField(existing FunctionConfigFields, additional FunctionConfigFields) FunctionConfigFields {
	merged := existing
	merged.IsRoot = false
	merged.Secret = existing.Secret || additional.Secret
	if len(existing.Bindings) == 0 {
		merged.Bindings = additional.Bindings
		return merged
	}
	if len(additional.Bindings) == 0 {
		return merged
	}
	bindings := make(map[string]Binding)
	collection.ForEach(existing.Bindings, func(binding Binding) {
		bindings[field.OptionalString(binding.Name, "")] = binding
	})
	collection.ForEach(additional.Bindings, func(binding Binding) {
		name := field.OptionalString(binding.Name, "")
		if current, ok := bindings[name]; ok {
			current.Secret = current.Secret || binding.Secret
			bindings[name] = current
			return
		}
		bindings[name] = binding
	})
	merged.Bindings = slices.Collect(maps.Values(bindings))
	slices.SortFunc(merged.Bindings, func(a, b Binding) int {
		return compareArgs(field.OptionalString(a.Name, ""), field.OptionalString(b.Name, ""))
	})
	return merged
}

func sortCredentialConfiguration(fields []FunctionConfigFields) []FunctionConfigFields {
	result := append([]FunctionConfigFields(nil), fields...)
	slices.SortFunc(result, func(a, b FunctionConfigFields) int {
		return compareArgs(a.Name, b.Name)
	})
	return result
}
