package generator

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/maps"
	"github.com/ARM-software/golang-utils/utils/ptr"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/strings/casing"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Arm-Debug/go-sdk-generator/generator/internal/acronyms"
)

var (
	// validArgTypes lists all the types of variable which can be used in specification
	// it tries to support special types such as the one defined by OpenAPI specification https://spec.openapis.org/registry/format/
	validArgTypes = []string{"time.Duration", "duration", "[]byte", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "decimal", "float32", "float64", "bool", "hex", "binary", "byte", "base64", "url", "uri", "host", "port", "version", "semver", "uuid", "email", "json", "validJSON", "ip", "datetime", "regex", "[]string", "filepath", "existingFilepath", "pathParameter", connectionLangType, onlineTargetLangType, protectedConnectionLangType, serviceLangType, serverLangType, protectedServiceLangType, protectedServerLangType, behindAuthConnectionLangType, behindAuthServiceLangType, behindAuthServerLangType, serviceBehindAuthLangType, serverBehindAuthLangType}

	namesWithLintErrorToIgnore = []string{"token", "password", "key"}
)

const (
	urlPathSeparator             = "/"
	defaultNamespacePrefix       = defaultGitHubOrganisation
	connectionLangType           = "connection"
	onlineTargetLangType         = "onlineTarget"
	protectedConnectionLangType  = "protectedConnection"
	serviceLangType              = "service"
	serverLangType               = "server"
	protectedServiceLangType     = "protectedService"
	protectedServerLangType      = "protectedServer"
	behindAuthConnectionLangType = "behindAuthConnection"
	behindAuthServiceLangType    = "behindAuthService"
	behindAuthServerLangType     = "behindAuthServer"
	serviceBehindAuthLangType    = "serviceBehindAuth"
	serverBehindAuthLangType     = "serverBehindAuth"
	requestConfigurationType     = "httpUtils.RequestConfiguration"
	defaultGeneratedToolVersion  = "0.0.0"
)

func argFlagType(s string) string {
	st := strings.TrimSpace(s)
	switch st {
	case "time.Duration", "duration":
		return "Duration"
	case "[]byte{}", "hex":
		return "BytesHex"
	case "base64", "byte":
		return "BytesBase64"
	case "host", "url", "uri", "version", "semver", "uuid", "email", "json", "validJSON", "datetime", "regex", "binary", "filepath", "existingFilepath", "pathParameter":
		return "String"
	case "ip":
		return "IP"
	case "port":
		return "Int"
	case "[]string":
		return "StringSlice"
	case "decimal":
		return "Float64"
	default:
		return cases.Title(language.English).String(st)
	}
}

var templateFuncMap = maps.Merge(template.FuncMap{
	"title": func(s string) string {
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		return fmt.Sprintf("%s%s", strings.ToUpper(string(s[0])), s[1:])
	},
	"lower":      strings.ToLower,
	"lowerFirst": lowerFirstDescription,
	"uuid": func() string {
		uuid, err := idgen.GenerateUUID4()
		for err != nil {
			uuid, err = idgen.GenerateUUID4()
		}
		return uuid
	},
	"addExtraValidType": func(extraType string) string {
		validArgTypes = append(validArgTypes, extraType)
		return "" // a return value is required
	},
	"isVariadicArg":                isVariadicLangType,
	"convertLangType":              convertLangType,
	"validationLangType":           argValidationRule,
	"defaultValue":                 defaultValue,
	"defaultFieldValue":            defaultFieldValue,
	"configFieldType":              configFieldType,
	"configFieldDefaultValue":      configFieldDefaultValue,
	"suppressesRequiredValidation": suppressesRequiredValidation,
	"enumTypeName":                 enumTypeName,
	"enumConstName":                enumConstName,
	"enumValidationRule":           enumValidationRule,
	"descriptionWithEnum":          descriptionWithEnum,
	"descriptionWithExamples":      descriptionWithExamples,
	"descriptionWithMetadata":      descriptionWithMetadata,
	"descriptionWithDeprecation":   descriptionWithDeprecation,
	"deprecationNoticeDetails":     deprecationNoticeDetails,
	"cobraDeprecationMessage":      cobraDeprecationMessage,
	"hasEnum":                      hasEnum,
	"parameterType":                parameterType,
	"documentedAdditionalTypeFieldType": func(typeName string, fieldDef EntityField) string {
		return documentedAdditionalTypeFieldType(typeName, &fieldDef)
	},
	"documentedConfigFieldType": func(specName string, cfg ConfigurationOption) string {
		return documentedConfigFieldType(specName, &cfg)
	},
	"documentedInterfaceMethodSignature": documentedInterfaceMethodSignature,
	"documentedCLIOptions":               documentedCLIOptions,
	"parameterTypeFromBase":              parameterTypeFromBase,
	"methodArgumentType":                 methodArgumentType,
	"methodArgumentTypeFromBase":         methodArgumentTypeFromBase,
	"isVariadicParameter":                isVariadicParameter,
	"isLastIndex":                        func(index, total int) bool { return total > 0 && index == total-1 },
	"toCamel":                            toCamelCase,
	"toSnake":                            toSnakeCase,
	"toKebab":                            toKebabCase,
	"toPascal":                           toPascalCase,
	"quote":                              withQuote,
	"quoteDescription":                   quoteDescription,
	"toPackageName":                      toPackageName,
	"additionalTypeImports":              additionalTypeImports,
	"additionalTypePackageName":          additionalTypePackageName,
	"groupAdditionalTypesByPackage":      groupAdditionalTypesByPackage,
	"hasSingleDefaultTypesPackage":       hasSingleDefaultTypesPackage,
	"sortedAdditionalTypePackages":       sortedAdditionalTypePackages,
	"goComment":                          goComment,
	"referenceType":                      referenceType,
	"ignoreLinting":                      ignoreLintingError,
	"nowYear":                            nowYear,
	"toURIEncoded":                       toURIEncoded,
	"linkToManifest":                     linkToManifest,
	"runSkillName":                       runSkillName,
	"installSkillName":                   installSkillName,
	"githubActionSkillName":              githubActionSkillName,
	"normaliseActionDirectoryName":       normaliseActionDirectoryName,
	"normaliseActionName":                normaliseActionName,
	"hasCloudCommands":                   isCallCloudTrue,
	"toCLIArgNameConstantRef": func(cliFuncName, arg string, isRoot bool) string {
		return getArgName(cliFuncName, arg, "", isRoot)
	},
	"dereferenceString":                      ptr.Deref[string],
	"moduleRoot":                             moduleRoot,
	"sdkModuleRoot":                          sdkModuleRoot,
	"cliModuleRoot":                          cliModuleRoot,
	"importPrefix":                           importPrefix,
	"repositoryURL":                          repositoryURL,
	"hasSecretStringParameters":              hasSecretStringParameters,
	"hasReferenceLikeSecretStringParameters": hasReferenceLikeSecretStringParameters,
}, gitHubFuncMap)

func stripHTTPScheme(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	return value
}

// normaliseNamespace coerces user input into a stable Go import-path prefix.
//
// Go expects fully qualified import paths to look like host/path values rather
// than repository URLs. See:
// https://pkg.go.dev/cmd/go@master#hdr-Fully_qualified_import_paths
//
// The schema intentionally keeps this field higher level, but the generator has
// to produce deterministic Go module and import paths. For that reason this
// function accepts a few common authoring variations and rewrites them into the
// import-path form used by the templates. In particular, it removes any HTTP(S)
// scheme, normalises path separators, trims surrounding whitespace, and ensures
// the result is treated as a namespace prefix with exactly one trailing slash.
func normaliseNamespace(value string) string {
	value = stripHTTPScheme(value)
	value = strings.ReplaceAll(value, "\\", urlPathSeparator)
	value = strings.TrimSpace(value)
	value = strings.Trim(value, urlPathSeparator)
	if reflection.IsEmpty(value) {
		return normaliseNamespace(defaultNamespacePrefix)
	}
	return fmt.Sprintf("%s%s", value, urlPathSeparator)
}

func interfaceNamespace(spec *SDKSpecification) string {
	return normaliseNamespace(spec.Interface.Namespace)
}

func moduleRoot(spec *SDKSpecification) string {
	return fmt.Sprintf("%s%s", interfaceNamespace(spec), spec.PackageName)
}

func sdkModuleRoot(spec *SDKSpecification) string {
	return fmt.Sprintf("%s/%s", moduleRoot(spec), spec.MainModule)
}

func cliModuleRoot(spec *SDKSpecification) string {
	if spec == nil || spec.CommandLineCommands == nil {
		return fmt.Sprintf("%s/%s", moduleRoot(spec), defaultCLIPackage)
	}
	return fmt.Sprintf("%s/%s", moduleRoot(spec), spec.CommandLineCommands.Package)
}

func importPrefix(spec *SDKSpecification) string {
	return strings.TrimSuffix(interfaceNamespace(spec), urlPathSeparator)
}

func defaultRepositoryURL(packageName string) string {
	return fmt.Sprintf("https://%s/%s", defaultGitHubOrganisation, packageName)
}

func repositoryURL(spec *SDKSpecification) string {
	if spec == nil {
		return ""
	}
	value := strings.TrimSpace(field.OptionalString(spec.RepositoryUrl, ""))
	if reflection.IsEmpty(value) {
		return defaultRepositoryURL(spec.PackageName)
	}
	return value
}

func linkToManifest(spec *SDKSpecification, schemaFileName string) string {
	fileName := normaliseManifestFileName(schemaFileName)
	if reflection.IsEmpty(fileName) {
		return ""
	}
	if isHostedOnGitHub(spec) {
		return linkToFileOnGitHub(spec, fileName)
	}

	return ""
}

func normaliseManifestFileName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	if reflection.IsEmpty(trimmed) {
		return ""
	}
	if reflection.IsNotEmpty(path.Ext(trimmed)) {
		return trimmed
	}
	return fmt.Sprintf("%s.yml", trimmed)
}

func toPascalCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToPascalCase(s, replacer)
	}
	return casing.ToPascalCase(s)
}

func toCamelCase(s string) string {
	var result string
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		result = casing.ToCamelCase(s, replacer)
	} else {
		result = casing.ToCamelCase(s)
	}
	return applyKnownAcronymCasing(result)
}

func toSnakeCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToSnakeCase(s, replacer)
	}
	return casing.ToSnakeCase(s)
}

func applyKnownAcronymCasing(value string) string {
	applied, err := acronyms.ApplyKnownAcronyms(value)
	if err != nil || reflection.IsEmpty(applied) {
		return value
	}
	return applied
}

func toKebabCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToKebabCase(s, replacer)
	}
	return casing.ToKebabCase(s)
}

func toURIEncoded(s string) string {
	return url.PathEscape(s)
}

// normaliseActionDirectoryName converts a package or project name into a stable
// GitHub Action directory name by lower-casing it and replacing
// non-alphanumeric separators with hyphens.
func normaliseActionDirectoryName(s string) string {
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(s)), func(r rune) bool {
		return collection.And(!unicode.IsLetter(r), !unicode.IsDigit(r))
	})
	if len(parts) == 0 {
		return "action"
	}

	return strings.Join(parts, "-")
}

// normaliseActionName converts a package or project name into a stable GitHub
// Action directory name by lower-casing it, replacing non-alphanumeric
// separators with hyphens, and ensuring there is exactly one trailing
// `-action` suffix.
//
// The generated name is intended for the action directory and `uses:` path,
// not the free-form `name` field in `action.yml`. See GitHub's metadata syntax
// guidance for actions:
// https://docs.github.com/actions/creating-actions/metadata-syntax-for-github-actions
func normaliseActionName(s string) string {
	name := normaliseActionDirectoryName(s)

	if name == "action" {
		return name
	}

	if strings.HasSuffix(name, "-action") {
		return name
	}
	return fmt.Sprintf("%s-action", name)
}

func nowYear() string {
	return fmt.Sprint(time.Now().Year())
}

func normaliseDescription(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimSpace(s)
}

func withQuote(s string) string {
	return strconv.Quote(s)
}

func quoteDescription(s string) string {
	return withQuote(normaliseDescription(s))
}

func goComment(linePrefix, s string) string {
	s = normaliseDescription(s)
	if s == "" {
		return fmt.Sprintf("%s//", linePrefix)
	}

	commentLines := collection.Map(strings.Split(s, "\n"), func(line string) string {
		line = strings.TrimSpace(line)
		if line == "" {
			return fmt.Sprintf("%s//", linePrefix)
		}
		return fmt.Sprintf("%s// %s", linePrefix, line)
	})

	return strings.Join(commentLines, "\n")
}

func isVariadicLangType(s string) bool {
	return strings.HasPrefix(strings.TrimSpace(s), "...")
}

func hasEnum(enum []string) bool {
	return reflection.IsNotEmpty(enum)
}

func enumTypeName(ownerName, fieldName, langType string, enum []string) string {
	if !hasEnum(enum) {
		return convertLangType(langType)
	}
	return fmt.Sprintf("%s%s", toPascalCase(ownerName), toPascalCase(fieldName))
}

func enumConstName(typeName, value string) string {
	return fmt.Sprintf("%s%s", typeName, toPascalCase(value))
}

func enumConstRefs(ownerName, fieldName, langType string, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	typeName := enumTypeName(ownerName, fieldName, langType, enum)
	constRefs := collection.Map(enum, func(value string) string {
		return enumConstName(typeName, value)
	})
	return strings.Join(constRefs, ", ")
}

func enumValidationRule(ownerName, fieldName, langType string, reference bool, enum []string) string {
	if !hasEnum(enum) {
		return ""
	}
	constRefs := enumConstRefs(ownerName, fieldName, langType, enum)
	return fmt.Sprintf(", validationUtils.Enum(%s)", constRefs)
}

func descriptionWithEnum(description string, enum []string) string {
	if !hasEnum(enum) {
		return description
	}
	return fmt.Sprintf("%s Allowed values: %s.", strings.TrimSpace(description), strings.Join(enum, ", "))
}

func descriptionWithExamples(description string, example *string, examples []string) string {
	parts := make([]string, 0, 2)
	if trimmed := strings.TrimSpace(description); trimmed != "" {
		parts = append(parts, trimmed)
	}
	values := make([]string, 0, 1+len(examples))
	if reflection.IsNotEmpty(example) {
		values = append(values, field.OptionalString(example, ""))
	}
	values = append(values, examples...)
	values = collection.Filter(values, func(value string) bool {
		return reflection.IsNotEmpty(strings.TrimSpace(value))
	})
	if len(values) == 1 {
		parts = append(parts, fmt.Sprintf("Example: %s.", withQuote(strings.TrimSpace(values[0]))))
	} else if len(values) > 1 {
		quoted := collection.Map(values, func(value string) string {
			return withQuote(strings.TrimSpace(value))
		})
		parts = append(parts, fmt.Sprintf("Examples: %s.", strings.Join(quoted, ", ")))
	}
	return strings.Join(parts, " ")
}

func descriptionWithMetadata(description string, enum []string, example *string, examples []string) string {
	return descriptionWithExamples(descriptionWithEnum(description, enum), example, examples)
}

func referenceType(reference bool, langType string) string {
	if reference {
		return fmt.Sprintf("*%s", langType)
	}
	return langType
}

func parameterType(reference bool, list bool, iterator bool, langType string) string {
	baseType := convertLangType(langType)
	return parameterTypeFromBase(reference, list, iterator, baseType)
}

func parameterTypeFromBase(reference bool, list bool, iterator bool, baseType string) string {

	switch {
	case iterator:
		baseType = fmt.Sprintf("iter.Seq[%s]", baseType)
	case list:
		baseType = fmt.Sprintf("[]%s", baseType)
	}

	if reference {
		return fmt.Sprintf("*%s", baseType)
	}

	return baseType
}

func methodArgumentType(reference bool, list bool, iterator bool, langType string, variadic bool) string {
	baseType := convertLangType(langType)
	return methodArgumentTypeFromBase(reference, list, iterator, baseType, variadic)
}

func methodArgumentTypeFromBase(reference bool, list bool, iterator bool, baseType string, variadic bool) string {

	if variadic {
		return fmt.Sprintf("...%s", baseType)
	}

	return parameterTypeFromBase(reference, list, iterator, baseType)
}

func isVariadicParameter(reference bool, list bool, iterator bool, isLast bool) bool {
	return isLast && list && !iterator && !reference
}

func hasConfigEnums(spec *SDKSpecification) bool {
	return spec.HasConfigEnums()
}

func hasCLIConfigEnums(cmds *CommandLineCommands) bool {
	return cmds.HasConfigEnums()
}

func hasInterfaceEnums(spec *SDKSpecification) bool {
	return spec.HasInterfaceEnums()
}

func toPackageName(value string) string {
	pkg := strings.ReplaceAll(casing.ToSnakeCase(strings.TrimSpace(value)), "_", "")
	pkg = strings.TrimLeftFunc(pkg, unicode.IsDigit)
	if reflection.IsEmpty(pkg) {
		return "main"
	}

	return pkg
}

func ignoreLintingError(name string) bool {
	nameLower := strings.ToLower(name)
	for i := range namesWithLintErrorToIgnore {
		if strings.Contains(nameLower, namesWithLintErrorToIgnore[i]) {
			return true
		}
	}
	return false
}

func determineZeroValue(s string) string {
	st := convertLangType(s)
	switch {
	case s == "json":
		return `"{}"`
	case s == "datetime":
		return fmt.Sprintf("\"%v\"", time.Now().UTC().String())
	case st == "string":
		return "\"\""
	case st == "bool":
		return "false"
	case strings.HasPrefix(st, "int") || strings.HasPrefix(st, "uint"):
		return "0"
	case strings.HasPrefix(s, "float"):
		return "0.0"
	case st == "time.Duration":
		return "0"
	case st == "[]byte":
		return "[]byte{}"
	case st == "[]string":
		return "[]string{}"
	default:
		return "nil"
	}
}

func isConnectionLangType(s string) bool {
	return collection.In([]string{
		connectionLangType,
		onlineTargetLangType,
		protectedConnectionLangType,
		serviceLangType,
		serverLangType,
		protectedServiceLangType,
		protectedServerLangType,
		behindAuthConnectionLangType,
		behindAuthServiceLangType,
		behindAuthServerLangType,
		serviceBehindAuthLangType,
		serverBehindAuthLangType,
	}, s, collection.StringCleanCaseInsensitiveMatch)
}

func isProtectedConnectionLangType(s string) bool {
	normalised := strings.ToLower(strings.TrimSpace(s))
	return isConnectionLangType(s) && (strings.Contains(normalised, "protected") || strings.Contains(normalised, "behindauth"))
}

func requestConfigurationUserAgent(toolName, toolVersion *string) string {
	name := strings.TrimSpace(field.OptionalString(toolName, "tool"))
	version := strings.TrimSpace(field.OptionalString(toolVersion, defaultGeneratedToolVersion))
	if reflection.IsEmpty(name) {
		name = "tool"
	}
	if reflection.IsEmpty(version) {
		version = defaultGeneratedToolVersion
	}
	return fmt.Sprintf("%s/%s", name, version)
}

func requestConfigurationDefaultValue(toolName, toolVersion, host string, protected bool, reference bool) string {
	constructor := "httpUtils.DefaultHTTPRequestConfiguration"
	if protected {
		constructor = "httpUtils.DefaultHTTPRequestWithAuthorisationConfigurationEnforced"
	}

	host = strings.TrimSpace(host)
	userAgent := requestConfigurationUserAgent(field.ToOptionalStringOrNilIfEmpty(toolName), field.ToOptionalStringOrNilIfEmpty(toolVersion))

	if reference {
		var builder strings.Builder
		builder.WriteString("func() *httpUtils.RequestConfiguration {\n")
		_, _ = fmt.Fprintf(&builder, "request := %s(%q)\n", constructor, userAgent)
		if protected {
			builder.WriteString("request.Authorisation.Scheme = httpUtils.AuthorisationSchemeBearer\n")
		}
		if reflection.IsNotEmpty(host) {
			_, _ = fmt.Fprintf(&builder, "request.Host = %q\n", host)
		}
		builder.WriteString("return request\n")
		builder.WriteString("}()")
		return builder.String()
	}

	var builder strings.Builder
	builder.WriteString("func() httpUtils.RequestConfiguration {\n")
	_, _ = fmt.Fprintf(&builder, "request := *%s(%q)\n", constructor, userAgent)
	if protected {
		builder.WriteString("request.Authorisation.Scheme = httpUtils.AuthorisationSchemeBearer\n")
	}
	if reflection.IsNotEmpty(host) {
		_, _ = fmt.Fprintf(&builder, "request.Host = %q\n", host)
	}
	builder.WriteString("return request\n")
	builder.WriteString("}()")
	return builder.String()
}

func isFunc(s any) bool {
	return strings.Contains(fmt.Sprint(s), "func()")
}
func replaceUnderscores(value string) string {
	return strings.ReplaceAll(value, "_", "-")
}
func defaultValue(s string, dValue any) string {
	dValueStr := fmt.Sprintf("%v", dValue)
	dValueStr = strings.TrimSpace(dValueStr)
	st := convertLangType(s)
	switch {
	case s == "json":
		if strings.HasPrefix(dValueStr, `""`) {
			return dValueStr
		}
		return fmt.Sprintf(`"%v"`, dValueStr)
	case s == "datetime":
		if strings.Contains(dValueStr, "now") {
			return "time.Now().UTC()"
		}
		return fmt.Sprintf(
			`func() time.Time {
t, _ := time.Parse(time.RFC3339, "%v")
return t
}()`, dValueStr)
	case st == "time.Duration":
		if strings.Contains(dValueStr, "time.") {
			return dValueStr
		}
		return fmt.Sprintf(
			`func() time.Duration {
duration, _ := time.ParseDuration("%v")
return duration
}()`, dValueStr)
	case st == "string":
		return defaultValueForStrings(dValueStr)
	case s == "hex":
		return fmt.Sprintf(
			`func() []byte {
values, _ := hex.DecodeString("%v")
return values
}()`, dValueStr)
	case strings.HasPrefix(st, "float"), st == "decimal":
		if strings.Contains(dValueStr, ".") {
			return dValueStr
		} else {
			return fmt.Sprintf("%v.0", dValueStr)
		}
	case st == "[]byte":
		return fmt.Sprintf(`[]byte("%v")`, dValueStr)
	default:
		return dValueStr
	}
}

func defaultFieldValue(additionalTypesPackages map[string]string, currentPackage, toolName, toolVersion, ownerName, fieldName, s string, enum []string, dValue any, reference bool) string {
	if hasEnum(enum) {
		enumType := enumTypeName(ownerName, fieldName, s, enum)
		value := enumConstName(enumType, strings.TrimSpace(fmt.Sprintf("%v", dValue)))
		if reference {
			return fmt.Sprintf("func() *%s {\nvalue := %s\nreturn &value\n}()", enumType, value)
		}
		return value
	}
	if isConnectionLangType(s) {
		return requestConfigurationDefaultValue(toolName, toolVersion, fmt.Sprintf("%v", dValue), isProtectedConnectionLangType(s), reference)
	}
	if reflection.IsEmpty(dValue) {
		if value, ok := localAdditionalTypeDefaultValue(additionalTypesPackages, currentPackage, s, reference); ok {
			return value
		}
	}
	return defaultValue(s, dValue)
}

func defaultValueForStrings(dValueStr string) string {
	usernameHome := regexp.MustCompile(`^\s*(.+?)\s*'\s*s\s+home\s*$`)
	switch {
	case isFunc(dValueStr):
		return dValueStr
	case strings.EqualFold(dValueStr, "hostname"):
		return `
		func() string {
			hostname, _ := os.Hostname()
			return hostname
		}()`
	case strings.EqualFold(dValueStr, "uuid"):
		return `
		func() string {
			uuid, _ := idgen.GenerateUUID4()
			return uuid
		}()`
	case collection.In([]string{"temp", "tempdir", "temporary"}, dValueStr, collection.StringCleanCaseInsensitiveMatch):
		return "filesystem.TempDirectory()"
	case collection.In([]string{"bin", "bindir"}, dValueStr, collection.StringCleanCaseInsensitiveMatch):
		return "xdg.BinHome"
	case collection.In([]string{"cache", "cachedir"}, dValueStr, collection.StringCleanCaseInsensitiveMatch):
		return "xdg.CacheHome"
	case collection.In([]string{"config", "configdir", "conf", "configuration"}, dValueStr, collection.StringCleanCaseInsensitiveMatch):
		return "xdg.ConfigHome"
	case strings.EqualFold(dValueStr, "here"):
		return `
		func() string {
			pwd, _ := filesystem.CurrentDirectory()
			return pwd
		}()`
	case strings.EqualFold(dValueStr, "home"):
		return `
		func() string {
			home, _ := os.UserHomeDir()
			return home
		}()`
	case usernameHome.MatchString(dValueStr):
		m := usernameHome.FindStringSubmatch(dValueStr)
		if len(m) != 2 {
			return dValueStr
		}
		return fmt.Sprintf(`
		func() string {
			home, _ := platform.GetHomeDirectory("%v")
			return home
		}()`, m[1])
	case strings.EqualFold(dValueStr, "random"):
		return "rand.Text()"
	case strings.EqualFold(dValueStr, "now"):
		return "time.Now().UTC().Format(time.RFC3339)"

	default:
		return fmt.Sprintf(`"%v"`, dValueStr)
	}
}

func convertLangType(s string) string {
	st := strings.TrimSpace(s)
	switch st {
	case connectionLangType, onlineTargetLangType, protectedConnectionLangType,
		serviceLangType, serverLangType,
		protectedServiceLangType, protectedServerLangType,
		behindAuthConnectionLangType, behindAuthServiceLangType, behindAuthServerLangType,
		serviceBehindAuthLangType, serverBehindAuthLangType:
		return requestConfigurationType
	case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "any", "error", "[]string":
		return st
	case "decimal":
		return "float64"
	case "time.Duration", "duration":
		return "time.Duration"
	case "[]byte", "hex":
		return "[]byte"
	case "base64", "byte", "host", "url", "uri", "version", "semver", "uuid", "email", "json", "ip", "datetime", "binary", "regex":
		return "string"
	case "port":
		return "uint32"
	}
	return normaliseCustomLangType(st)
}

func normaliseCustomLangType(st string) string {
	switch {
	case strings.HasPrefix(st, "[]"):
		return fmt.Sprintf("[]%s", convertLangType(st[2:]))
	case strings.HasPrefix(st, "*"):
		return fmt.Sprintf("*%s", convertLangType(st[1:]))
	case strings.HasPrefix(st, "..."):
		return fmt.Sprintf("...%s", convertLangType(st[3:]))
	case strings.Contains(st, "."):
		lastDot := strings.LastIndex(st, ".")
		return fmt.Sprintf("%s%s", st[:lastDot+1], toPascalCase(st[lastDot+1:]))
	default:
		return toPascalCase(st)
	}
}

func argValidationRule(s string) string {
	return field.OptionalString(func(st string) *string {
		switch st {
		case "[]byte{}", "hex":
			return field.ToOptionalString(", is.Hexadecimal")
		case "base64", "byte":
			return field.ToOptionalString(", is.Base64")
		case "host", "url":
			return field.ToOptionalString(", is.URL")
		case "uri":
			return field.ToOptionalString(", validationUtils.IsURI")
		case "version", "semver":
			return field.ToOptionalString(", is.Semver")
		case "uuid":
			return field.ToOptionalString(", is.UUID")
		case "email":
			return field.ToOptionalString(", is.Email")
		case "json":
			return field.ToOptionalString(", is.JSON")
		case "validJSON":
			return field.ToOptionalString(", jsonutils.IsValidJSON")
		case "ip":
			return field.ToOptionalString(", is.IP")
		case "datetime":
			return field.ToOptionalString(", validationUtils.IsRFC3339Timestamp")
		case "filepath":
			return field.ToOptionalString(", filesystem.NewOSPathValidationRule(true)")
		case "existingFilepath":
			return field.ToOptionalString(", filesystem.NewOSPathExistRule(true)")
		case "pathParameter":
			return field.ToOptionalString(", validationUtils.IsPathParameter")
		case "port":
			return field.ToOptionalString(", validationUtils.IsPort")
		}
		return nil
	}(strings.TrimSpace(s)), "")
}
