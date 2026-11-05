/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
	validationUtils "github.com/ARM-software/golang-utils/utils/validation"
)

const (
	defaultAdditionalTypePackage = "types"
	defaultCLIPackage            = "cli"
	fieldImportPath              = "github.com/ARM-software/golang-utils/utils/field"
	filesystemImportPath         = "github.com/ARM-software/golang-utils/utils/filesystem"
	httpUtilsImportPath          = "github.com/ARM-software/golang-utils/utils/http"
	jsonUtilsImportPath          = "github.com/ARM-software/golang-utils/utils/serialisation/json" // canonical import path
	platformImportPath           = "github.com/ARM-software/golang-utils/utils/platform"
	validationUtilsImportPath    = "github.com/ARM-software/golang-utils/utils/validation"
	xdgImportPath                = "github.com/adrg/xdg"
)

type additionalTypeImport struct {
	Alias string
	Path  string
}

type generatedAdditionalTypeReference struct {
	TypeName        string
	TargetPackage   string
	ExplicitPointer bool
}

var (
	generatedPackageSegmentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	generatedBinarySegmentPattern  = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_-]*$`)
)

func normaliseAdditionalTypePackageName(value string) string {
	if reflection.IsEmpty(value) {
		return defaultAdditionalTypePackage
	}
	normalised := toPackageName(value)
	if normalised == "main" {
		return defaultAdditionalTypePackage
	}
	return normalised
}

func validateAdditionalTypePackageName(value string) error {
	return validateGeneratedPackageName(value)
}

func validateGeneratedPackageName(value string) error {
	if reflection.IsEmpty(value) {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	if err := validation.Validate(trimmed, validationUtils.Pattern(generatedPackageSegmentPattern)); err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "package %q must be a single package segment containing only letters, numbers, and underscores", value)
	}
	return nil
}

func validateGeneratedBinaryName(value string) error {
	if reflection.IsEmpty(value) {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	if err := validation.Validate(trimmed, validationUtils.Pattern(generatedBinarySegmentPattern)); err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "binary %q must be a single executable segment containing only letters, numbers, underscores, or hyphens", value)
	}
	return nil
}

func validateGeneratedMainModuleName(value string) error {
	if reflection.IsEmpty(value) {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	if err := validation.Validate(trimmed, validationUtils.Pattern(generatedPackageSegmentPattern)); err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "main module %q must be a single package segment containing only letters, numbers, and underscores", value)
	}
	return nil
}

func normaliseCLIPackageName(value string) string {
	if reflection.IsEmpty(value) {
		return defaultCLIPackage
	}
	normalised := toPackageName(value)
	if normalised == "main" {
		return defaultCLIPackage
	}
	return normalised
}

func validateCLIPackageName(value string) error {
	return validateGeneratedPackageName(value)
}

func additionalTypePackageName(typ *Type) string {
	if typ == nil {
		return defaultAdditionalTypePackage
	}
	return normaliseAdditionalTypePackageName(typ.Package)
}

func groupAdditionalTypesByPackage(types Types) map[string]Types {
	grouped := collection.GroupByRef(types, additionalTypePackageName)
	result := make(map[string]Types, len(grouped))
	for pkg, groupedTypes := range grouped {
		result[pkg] = Types(groupedTypes)
	}
	return result
}

func sortedAdditionalTypePackages(types Types) []string {
	packages := slices.Collect(maps.Keys(groupAdditionalTypesByPackage(types)))
	slices.Sort(packages)
	return packages
}

func hasSingleDefaultTypesPackage(types Types) bool {
	packages := sortedAdditionalTypePackages(types)
	return len(packages) == 1 && packages[0] == defaultAdditionalTypePackage
}

func additionalTypePackagesByName(types Types) map[string]string {
	packages := make(map[string]string, len(types))
	for i := range types {
		packages[toPascalCase(types[i].Name)] = additionalTypePackageName(&types[i])
	}
	return packages
}

func resolveGeneratedAdditionalTypeReference(langType string, additionalTypesPackages map[string]string) (generatedAdditionalTypeReference, bool) {
	trimmed := strings.TrimSpace(langType)
	if reflection.IsEmpty(trimmed) || strings.HasPrefix(trimmed, "[]") || strings.HasPrefix(trimmed, "...") {
		return generatedAdditionalTypeReference{}, false
	}
	baseType := strings.TrimPrefix(trimmed, "*")
	packageAlias := ""
	typeName := baseType
	if dot := strings.LastIndex(baseType, "."); dot >= 0 {
		packageAlias = baseType[:dot]
		typeName = baseType[dot+1:]
	}
	normalisedTypeName := toPascalCase(typeName)
	targetPackage, found := additionalTypesPackages[normalisedTypeName]
	if !found {
		return generatedAdditionalTypeReference{}, false
	}
	if reflection.IsNotEmpty(packageAlias) && packageAlias != targetPackage && packageAlias != defaultAdditionalTypePackage {
		return generatedAdditionalTypeReference{}, false
	}
	return generatedAdditionalTypeReference{
		TypeName:        normalisedTypeName,
		TargetPackage:   targetPackage,
		ExplicitPointer: strings.HasPrefix(trimmed, "*"),
	}, true
}

func generatedAdditionalTypeDefaultConstructor(ref generatedAdditionalTypeReference, currentPackage string) string {
	if ref.TargetPackage == currentPackage {
		return fmt.Sprintf("Default%s()", ref.TypeName)
	}
	return fmt.Sprintf("%s.Default%s()", ref.TargetPackage, ref.TypeName)
}

func localAdditionalTypeDefaultValue(additionalTypesPackages map[string]string, currentPackage string, langType string, reference bool) (string, bool) {
	ref, ok := resolveGeneratedAdditionalTypeReference(langType, additionalTypesPackages)
	if !ok {
		return "", false
	}
	constructor := generatedAdditionalTypeDefaultConstructor(ref, currentPackage)
	if reference || ref.ExplicitPointer {
		return constructor, true
	}
	return fmt.Sprintf("*%s", constructor), true
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

func rewriteGeneratedAdditionalTypeReference(value string, currentPackage string, additionalTypesPackages map[string]string) string {
	value = strings.TrimSpace(value)
	if reflection.IsEmpty(value) {
		return value
	}
	if prefix, found := collection.FirstBy([]string{"[]", "*", "..."}, func(prefix string) bool {
		return strings.HasPrefix(value, prefix)
	}); found {
		return prefix + rewriteGeneratedAdditionalTypeReference(value[len(prefix):], currentPackage, additionalTypesPackages)
	}

	packageName := ""
	typeName := value
	if dot := strings.LastIndex(value, "."); dot >= 0 {
		packageName = value[:dot]
		typeName = value[dot+1:]
	}

	normalisedTypeName := toPascalCase(typeName)
	targetPackage, ok := additionalTypesPackages[normalisedTypeName]
	if !ok {
		return value
	}
	if reflection.IsNotEmpty(packageName) && packageName != defaultAdditionalTypePackage && packageName != targetPackage {
		return value
	}
	if targetPackage == currentPackage {
		return normalisedTypeName
	}
	return fmt.Sprintf("%s.%s", targetPackage, normalisedTypeName)
}

func rewriteGeneratedAdditionalTypeDefaultValue(value *string, currentPackage string, additionalTypesPackages map[string]string) *string {
	if reflection.IsEmpty(value) {
		return value
	}
	rewritten := field.OptionalString(value, "")
	for typeName := range additionalTypesPackages {
		ref, ok := resolveGeneratedAdditionalTypeReference(fmt.Sprintf("types.%s", typeName), additionalTypesPackages)
		if !ok {
			continue
		}
		typePattern := regexp.MustCompile(`\btypes\.` + regexp.QuoteMeta(typeName) + `\b`)
		defaultPattern := regexp.MustCompile(`\btypes\.Default` + regexp.QuoteMeta(typeName) + `\b`)
		replacement := fmt.Sprintf("%s.%s", ref.TargetPackage, ref.TypeName)
		defaultReplacement := strings.TrimSuffix(generatedAdditionalTypeDefaultConstructor(ref, currentPackage), "()")
		if ref.TargetPackage == currentPackage {
			replacement = typeName
		}
		rewritten = typePattern.ReplaceAllString(rewritten, replacement)
		rewritten = defaultPattern.ReplaceAllString(rewritten, defaultReplacement)
	}
	return field.ToOptionalString(rewritten)
}
