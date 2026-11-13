/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
	"sync"

	"golang.org/x/tools/imports"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const (
	fieldImportPath           = "github.com/ARM-software/golang-utils/utils/field"
	filesystemImportPath      = "github.com/ARM-software/golang-utils/utils/filesystem"
	httpUtilsImportPath       = "github.com/ARM-software/golang-utils/utils/http"
	jsonUtilsImportPath       = "github.com/ARM-software/golang-utils/utils/serialisation/json" // canonical import path
	platformImportPath        = "github.com/ARM-software/golang-utils/utils/platform"
	validationUtilsImportPath = "github.com/ARM-software/golang-utils/utils/validation"
	xdgImportPath             = "github.com/adrg/xdg"
)

type additionalTypeImport struct {
	Alias string
	Path  string
}

var importsLocalPrefixMu sync.Mutex

var importSpecLinePattern = regexp.MustCompile(`^\s*(?:[\w\.]+\s+)?"([^"]+)"(?:\s*//.*)?\s*$`)

func renderImport(alias, importPath string, applyMisspellException bool) string {
	alias = strings.TrimSpace(alias)
	importPath = strings.TrimSpace(importPath)
	if reflection.IsEmpty(importPath) {
		return ""
	}
	comment := ""
	if applyMisspellException {
		comment = " //nolint:misspell"
	}
	if reflection.IsEmpty(alias) || alias == path.Base(importPath) {
		return fmt.Sprintf("%q%s", importPath, comment)
	}
	return fmt.Sprintf("%s %q%s", alias, importPath, comment)
}

func additionalTypeImports(spec *SDKSpecification, currentPackage string) []additionalTypeImport {
	if spec == nil {
		return nil
	}
	return collection.FilterRef(collection.Map(sortedAdditionalTypePackages(spec.AdditionalTypes), func(pkg string) additionalTypeImport {
		return additionalTypeImport{
			Alias: pkg,
			Path:  fmt.Sprintf("%s/%s", sdkModuleRoot(spec), pkg),
		}
	}), func(i *additionalTypeImport) bool { return i.Alias != currentPackage })
}

// filterExtraImportsByPath removes explicit import paths that a template already
// emits directly, so generated files do not end up with duplicate imports.
func filterExtraImportsByPath(extraImports ExtraImports, excludedPaths ...string) ExtraImports {
	excluded := collection.Filter(collection.Map(excludedPaths, strings.TrimSpace), func(path string) bool {
		return reflection.IsNotEmpty(path)
	})
	if len(excluded) == 0 {
		return extraImports
	}
	return collection.FilterRef(extraImports, func(extraImport *ExtraImport) bool {
		if extraImport == nil {
			return false
		}
		importPath := strings.TrimSpace(field.OptionalString(extraImport.ImportUrl, ""))
		return !slices.Contains(excluded, importPath)
	})
}

// configExtraImports returns extra imports for manager configuration templates
// after removing paths that the template already imports explicitly.
func configExtraImports(spec *SDKSpecification) ExtraImports {
	if spec == nil {
		return nil
	}
	excluded := []string{}
	excluded = append(excluded, validationUtilsImportPath)
	if !spec.GlobalConfiguration.UsesFilesystemValidationAlias() {
		excluded = append(excluded, filesystemImportPath)
	}
	if !spec.GlobalConfiguration.UsesJSONValidationAlias() {
		excluded = append(excluded, jsonUtilsImportPath)
	}
	return filterExtraImportsByPath(spec.ExtraImports, excluded...)
}

// cliConfigExtraImports returns extra imports for CLI configuration templates
// after removing paths that the template already imports explicitly.
func cliConfigExtraImports(spec *SDKSpecification) ExtraImports {
	if spec == nil || spec.CommandLineCommands == nil {
		return nil
	}
	excluded := []string{}
	if !spec.CommandLineCommands.UsesValidationUtilsImport() {
		excluded = append(excluded, validationUtilsImportPath)
	}
	if !spec.CommandLineCommands.UsesFilesystemValidationAlias() {
		excluded = append(excluded, filesystemImportPath)
	}
	if !spec.CommandLineCommands.UsesJSONValidationAlias() {
		excluded = append(excluded, jsonUtilsImportPath)
	}
	excluded = append(excluded, collection.Filter([]string{httpUtilsImportPath, platformImportPath, xdgImportPath}, func(importPath string) bool {
		return !cliConfigUsesImportPath(spec, importPath)
	})...)
	return filterExtraImportsByPath(spec.ExtraImports, excluded...)
}

func cliConfigUsesImportPath(spec *SDKSpecification, importPath string) bool {
	if spec == nil || spec.CommandLineCommands == nil {
		return false
	}
	return collection.AnyFunc(spec.CommandLineCommands.Functions, func(fct CliFunction) bool {
		return collection.AnyFunc(fct.Configuration, func(cfg FunctionConfigFields) bool {
			if reflection.IsEmpty(cfg.DefaultValue) {
				return false
			}
			expr := defaultFieldValue(nil, "", "tool", defaultGeneratedToolVersion, "", cfg.Name, cfg.LangType, cfg.EffectiveEnum(), field.OptionalString(cfg.DefaultValue, ""), isReferenceLikeSecretField(cfg))
			switch importPath {
			case httpUtilsImportPath:
				return strings.Contains(expr, "httpUtils.")
			case platformImportPath:
				return strings.Contains(expr, "platform.")
			case xdgImportPath:
				return strings.Contains(expr, "xdg.")
			default:
				return false
			}
		})
	})
}

// interfaceExtraImports returns extra imports that are actually referenced by
// generated interface method signatures.
func interfaceExtraImports(spec *SDKSpecification) ExtraImports {
	if spec == nil {
		return nil
	}
	return collection.FilterRef(spec.ExtraImports, func(extraImport *ExtraImport) bool {
		if extraImport == nil {
			return false
		}
		alias := strings.TrimSpace(field.OptionalString(extraImport.ImportRef, ""))
		if reflection.IsEmpty(alias) {
			return false
		}
		return interfaceUsesImportAlias(spec, alias)
	})
}

func interfaceUsesImportAlias(spec *SDKSpecification, alias string) bool {
	if spec == nil {
		return false
	}
	aliasRef := fmt.Sprintf("%s.", alias)
	return collection.AnyFunc(spec.Interface.Methods, func(method InterfaceMethod) bool {
		return collection.AnyFunc(method.Args, func(arg CommandArgument) bool {
			if hasEnum(arg.Enum) {
				return strings.Contains(methodArgumentTypeFromBase(arg.Reference, arg.List, arg.Iterator, enumTypeName(method.Name, arg.Name, arg.LangType, arg.Enum), false), aliasRef)
			}
			return strings.Contains(methodArgumentType(arg.Reference, arg.List, arg.Iterator, arg.LangType, false), aliasRef)
		}) || collection.AnyFunc(method.Returns, func(ret CommandArgument) bool {
			if hasEnum(ret.Enum) {
				return strings.Contains(parameterTypeFromBase(ret.Reference, ret.List, ret.Iterator, enumTypeName(method.Name, ret.Name, ret.LangType, ret.Enum)), aliasRef)
			}
			return strings.Contains(parameterType(ret.Reference, ret.List, ret.Iterator, ret.LangType), aliasRef)
		})
	})
}

// typesExtraImports returns extra imports for generated additional-type Go files
// after removing paths that the template already imports explicitly.
func typesExtraImports(spec *SDKSpecification) ExtraImports {
	if spec == nil {
		return nil
	}
	excluded := []string{}
	excluded = append(excluded, validationUtilsImportPath)
	if !spec.AdditionalTypes.UsesFilesystemValidationAlias() {
		excluded = append(excluded, filesystemImportPath)
	}
	if !spec.AdditionalTypes.UsesJSONValidationAlias() {
		excluded = append(excluded, jsonUtilsImportPath)
	}
	return filterExtraImportsByPath(spec.ExtraImports, excluded...)
}

func formatImports(ctx context.Context, path string, localPrefix string) (bytes []byte, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	src, err := filesystem.ReadFile(path)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrNotFound, err, "file '%v' doesn't exist", path)
		return
	}

	// FIXME It would be great to also simplify code as part of the process (gofmt -s). However, this is not possible in current go implementation https://github.com/golang/go/issues/21476
	importsLocalPrefixMu.Lock()
	previousLocalPrefix := imports.LocalPrefix
	imports.LocalPrefix = localPrefix
	defer func() {
		imports.LocalPrefix = previousLocalPrefix
		importsLocalPrefixMu.Unlock()
	}()

	cleanedSrc, err := imports.Process(path, src, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   4,
		Fragment:   true,
		FormatOnly: false,
	})
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not cleanup imports for file '%v'", path)
		return
	}

	bytes = normaliseLocalImportGroups(cleanedSrc, localPrefix)
	return
}

func normaliseLocalImportGroups(src []byte, localPrefix string) []byte {
	if strings.TrimSpace(localPrefix) == "" {
		return src
	}

	lines := strings.Split(string(src), "\n")
	normalised := make([]string, 0, len(lines))
	inImportBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inImportBlock {
			normalised = append(normalised, line)
			if trimmed == "import (" {
				inImportBlock = true
			}
			continue
		}

		if trimmed == ")" {
			inImportBlock = false
			normalised = append(normalised, line)
			continue
		}

		if trimmed != "" {
			normalised = append(normalised, line)
			continue
		}

		prevImportPath, prevFound := nearestImportPath(normalised, len(normalised)-1, -1)
		nextImportPath, nextFound := nearestImportPath(lines, i+1, 1)
		if prevFound && nextFound && strings.HasPrefix(prevImportPath, localPrefix) && strings.HasPrefix(nextImportPath, localPrefix) {
			continue
		}

		normalised = append(normalised, line)
	}

	return []byte(strings.Join(normalised, "\n"))
}

func nearestImportPath(lines []string, start int, step int) (string, bool) {
	for i := start; i >= 0 && i < len(lines); i += step {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || trimmed == "import (" || trimmed == ")" {
			continue
		}
		matches := importSpecLinePattern.FindStringSubmatch(lines[i])
		if len(matches) == 2 {
			return matches[1], true
		}
		break
	}
	return "", false
}
