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
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/sttk/stringcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// validArgTypes lists all the types of variable which can be used in specification
	// it tries to support special types such as the one defined by OpenAPI specification https://spec.openapis.org/registry/format/
	validArgTypes = []string{"time.Duration", "duration", "[]byte", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "decimal", "float32", "float64", "bool", "hex", "binary", "byte", "base64", "url", "uri", "host", "port", "version", "semver", "uuid", "email", "json", "ip", "datetime", "regex", "[]string"}

	namesWithLintErrorToIgnore = []string{"token", "password", "key"}

	specialAcronyms = []string{"Id", "Api", "Db", "Grpc", "Json", "Ddos", "Xml", "Ios", "Yaml", "Yml", "Url", "Fpga", "Afaas", "Https", "Http", "Ttl"} // see https://google.github.io/styleguide/go/decisions.html#initialisms
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

var templateFuncMap = template.FuncMap{
	"title": func(s string) string {
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		return fmt.Sprintf("%s%s", strings.ToUpper(string(s[0])), s[1:])
	},
	"lower": strings.ToLower,
	"lowerFirst": func(s string) string {
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
	"convertLangType":     convertLangType,
	"validationLangType":  argValidationRule,
	"defaultValue":        defaultValue,
	"toCamel":             toCamelCase,
	"toSnake":             toSnakeCase,
	"toKebab":             toKebabCase,
	"toPascal":            toPascalCase,
	"quote":               withQuote,
	"toPackageName":       toPackageName,
	"ignoreLinting":       ignoreLintingError,
	"nowYear":             nowYear,
	"toURIEncoded":        toURIEncoded,
	"normaliseActionName": normaliseActionName,
	"hasCloudCommands":    isCallCloudTrue,
	"toCLIArgNameConstantRef": func(cliFuncName, arg string, isRoot bool) string {
		return getArgName(cliFuncName, arg, "", isRoot)
	},
	"dereferenceString": func(s *string) string { return field.Optional(s, "") },
}

func toPascalCase(s string) string {
	return handleAcronyms(stringcase.PascalCase(s))
}

func toCamelCase(s string) string {
	return handleAcronyms(stringcase.CamelCase(s))
}

func toSnakeCase(s string) string {
	return stringcase.SnakeCase(s)
}

func toKebabCase(s string) string {
	return stringcase.KebabCase(s)
}

func toURIEncoded(s string) string {
	return url.PathEscape(s)
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
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(s)), func(r rune) bool {
		return collection.And(!unicode.IsLetter(r), !unicode.IsDigit(r))
	})
	if len(parts) == 0 {
		return "action"
	}
	name := strings.Join(parts, "-")
	if strings.HasSuffix(name, "-action") {
		return name
	}
	return fmt.Sprintf("%s-action", name)
}

func nowYear() string {
	return fmt.Sprint(time.Now().Year())
}
func handleAcronyms(pC string) string {
	for i := range specialAcronyms {
		pC = strings.ReplaceAll(pC, specialAcronyms[i], strings.ToUpper(specialAcronyms[i]))
	}
	// The following is to deal with the ID specific acronym
	pC = strings.ReplaceAll(pC, "IDentifier", "Identifier")
	pC = strings.ReplaceAll(pC, "IDentification", "Identification")
	return pC
}

func withQuote(s string) string {
	return strconv.Quote(s)
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
