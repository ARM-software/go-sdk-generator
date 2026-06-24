package generator

import (
	"fmt"
	"net/url"
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
	validArgTypes = []string{"time.Duration", "duration", "[]byte", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "decimal", "float32", "float64", "bool", "hex", "binary", "byte", "base64", "url", "uri", "host", "port", "version", "semver", "uuid", "email", "json", "ip", "datetime", "regex", "[]string", connectionLangType, onlineTargetLangType, protectedConnectionLangType, serviceLangType, serverLangType, protectedServiceLangType, protectedServerLangType, behindAuthConnectionLangType, behindAuthServiceLangType, behindAuthServerLangType, serviceBehindAuthLangType, serverBehindAuthLangType}

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
	case "host", "url", "uri", "version", "semver", "uuid", "email", "json", "datetime", "regex", "binary":
		return "String"
	case "ip":
		return "IP"
	case "port":
		return "Int"
	case "[]string":
		return "StringArray"
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
	"lower": strings.ToLower,
	"lowerFirst": func(s string) string {
		s = normaliseDescription(s)
		if s == "" {
			return s
		}
		return strings.ToLower(string(s[0])) + s[1:]
	},
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
	"isVariadicArg": func(s string) bool {
		return strings.HasPrefix(strings.TrimSpace(s), "...")
	},
	"convertLangType":              convertLangType,
	"validationLangType":           argValidationRule,
	"defaultValue":                 defaultValue,
	"defaultFieldValue":            defaultFieldValue,
	"configFieldType":              configFieldType,
	"configFieldDefaultValue":      configFieldDefaultValue,
	"toCamel":                      toCamelCase,
	"toSnake":                      toSnakeCase,
	"toKebab":                      toKebabCase,
	"toPascal":                     toPascalCase,
	"quote":                        withQuote,
	"quoteDescription":             quoteDescription,
	"toPackageName":                toPackageName,
	"goComment":                    goComment,
	"referenceType":                referenceType,
	"ignoreLinting":                ignoreLintingError,
	"nowYear":                      nowYear,
	"toURIEncoded":                 toURIEncoded,
	"linkToManifest":               linkToManifest,
	"normaliseActionDirectoryName": normaliseActionDirectoryName,
	"normaliseActionName":          normaliseActionName,
	"hasCloudCommands":             isCallCloudTrue,
	"toCLIArgNameConstantRef": func(cliFuncName, arg string, isRoot bool) string {
		return getArgName(cliFuncName, arg, "", isRoot)
	},
	"dereferenceString":                      ptr.Deref[string],
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
	return fmt.Sprintf("%s/cli", moduleRoot(spec))
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
	fileName := fmt.Sprintf("%s.yml", schemaFileName)
	if isHostedOnGitHub(spec) {
		return linkToFileOnGitHub(spec, fileName)
	}

	return ""
}

func toPascalCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToPascalCase(s, replacer)
	}
	return casing.ToPascalCase(s)
}

func toCamelCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToCamelCase(s, replacer)
	}
	return casing.ToCamelCase(s)
}

func toSnakeCase(s string) string {
	if replacer, err := acronyms.Replacer(); err == nil && replacer != nil {
		return casing.ToSnakeCase(s, replacer)
	}
	return casing.ToSnakeCase(s)
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

func referenceType(reference bool, langType string) string {
	if reference {
		return fmt.Sprintf("*%s", langType)
	}
	return langType
}

func toPackageName(value string) string {
	value = strings.TrimSpace(value)

	var builder strings.Builder

	builder.Grow(len(value))
	for _, r := range value {
		if unicode.IsLetter(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}

	pkg := builder.String()
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

func defaultFieldValue(toolName, toolVersion, s string, dValue any, reference bool) string {
	if isConnectionLangType(s) {
		return requestConfigurationDefaultValue(toolName, toolVersion, fmt.Sprintf("%v", dValue), isProtectedConnectionLangType(s), reference)
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
	default:
		return st
	}
}

func argValidationRule(s string) string {
	return field.OptionalString(func(st string) *string {
		switch st {
		case "[]byte{}", "hex":
			return field.ToOptionalString(", is.Hexadecimal")
		case "base64", "byte":
			return field.ToOptionalString(", is.Base64")
		case "host", "url", "uri":
			return field.ToOptionalString(", is.URL")
		case "version", "semver":
			return field.ToOptionalString(", is.Semver")
		case "uuid":
			return field.ToOptionalString(", is.UUID")
		case "email":
			return field.ToOptionalString(", is.Email")
		case "json":
			return field.ToOptionalString(", is.JSON")
		case "ip":
			return field.ToOptionalString(", is.IP")
		case "port":
			return field.ToOptionalString(", validationUtils.IsPort")
		}
		return nil
	}(strings.TrimSpace(s)), "")
}
