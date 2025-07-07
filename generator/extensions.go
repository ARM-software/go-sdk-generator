package generator

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/sttk/stringcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	// validArgTypes lists all the types of variable which can be used in specification
	// it tries to support special types such as the one defined by OpenAPI specification https://spec.openapis.org/registry/format/
	validArgTypes = []string{"time.Duration", "duration", "[]byte", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "decimal", "float32", "float64", "bool", "hex", "binary", "byte", "base64", "url", "uri", "host", "port", "version", "semver", "uuid", "email", "json", "ip", "datetime", "regex"}

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
	"convertLangType":    convertLangType,
	"validationLangType": argValidationRule,
	"defaultValue":       defaultValue,
	"toCamel":            toCamelCase,
	"toSnake":            toSnakeCase,
	"toPascal":           toPascalCase,
	"ignoreLinting":      ignoreLintingError,
	"dereferenceString":  func(s *string) string { return field.Optional(s, "") },
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

func handleAcronyms(pC string) string {
	for i := range specialAcronyms {
		pC = strings.ReplaceAll(pC, specialAcronyms[i], strings.ToUpper(specialAcronyms[i]))
	}
	// The following is to deal with the ID specific acronym
	pC = strings.ReplaceAll(pC, "IDentifier", "Identifier")
	pC = strings.ReplaceAll(pC, "IDentification", "Identification")
	return pC
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
	st := convertLangType(s)
	switch {
	case s == "json":
		if strings.HasPrefix(dValueStr, `""`) {
			return dValueStr
		}
		return fmt.Sprintf(`"%v"`, dValue)
	case s == "datetime":
		return fmt.Sprintf(
			`func() time.Time {
t, _ := time.Parse(time.RFC3339, "%v")
return t
}()`, dValue)
	case st == "time.Duration":
		if strings.Contains(dValueStr, "time.") {
			return dValueStr
		}
		return fmt.Sprintf(
			`func() time.Duration {
duration, _ := time.ParseDuration("%v")
return duration
}()`, dValue)
	case st == "string":
		if isFunc(dValue) {
			return dValueStr
		}
		return fmt.Sprintf(`"%v"`, dValue)
	case s == "hex":
		return fmt.Sprintf(
			`func() []byte {
values, _ := hex.DecodeString("%v")
return values
}()`, dValue)
	case strings.HasPrefix(st, "float"):
		if strings.Contains(dValueStr, ".") {
			return dValueStr
		} else {
			return fmt.Sprintf("%v.0", dValue)
		}
	case st == "[]byte":
		return fmt.Sprintf(`[]byte("%v")`, dValue)
	default:
		return dValueStr
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
			return field.ToOptionalString(", is.Port")
		}
		return nil
	}(strings.TrimSpace(s)), "")
}
