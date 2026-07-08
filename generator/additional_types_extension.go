package generator

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const defaultAdditionalTypePackage = "types"

type additionalTypeImport struct {
	Alias string
	Path  string
}

type generatedAdditionalTypeReference struct {
	TypeName        string
	TargetPackage   string
	ExplicitPointer bool
}

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
	if reflection.IsEmpty(value) {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, ".") {
		return commonerrors.Newf(commonerrors.ErrInvalid, "package %q must be a single package segment, not a nested path", value)
	}
	return nil
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
	if packageAlias != "" && packageAlias != targetPackage && packageAlias != defaultAdditionalTypePackage {
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

func rewriteGeneratedAdditionalTypeReference(value string, currentPackage string, additionalTypesPackages map[string]string) string {
	value = strings.TrimSpace(value)
	if reflection.IsEmpty(value) {
		return value
	}
	for _, prefix := range []string{"[]", "*", "..."} {
		if strings.HasPrefix(value, prefix) {
			return prefix + rewriteGeneratedAdditionalTypeReference(value[len(prefix):], currentPackage, additionalTypesPackages)
		}
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
	if packageName != "" && packageName != defaultAdditionalTypePackage && packageName != targetPackage {
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
