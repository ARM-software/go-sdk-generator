package generator

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLangTypeValidationRule(t *testing.T) {
	assert.Equal(t, ", validationUtils.IsPort", argValidationRule("port"))
	assert.Equal(t, ", is.UUID", argValidationRule("uuid"))
	assert.Equal(t, ", is.URL", argValidationRule("host"))
	assert.Empty(t, argValidationRule("string"))
	assert.Empty(t, argValidationRule(faker.Word()))
}

func TestArgFlagType(t *testing.T) {
	assert.Equal(t, "StringArray", argFlagType("[]string"))
	assert.Equal(t, "StringArray", argFlagType(" []string "))
}

func TestZeros(t *testing.T) {
	assert.Equal(t, "0", determineZeroValue("port"))
	assert.Equal(t, `""`, determineZeroValue("uuid"))
	assert.Equal(t, `""`, determineZeroValue("host"))
	assert.Equal(t, `""`, determineZeroValue("email"))
	assert.Equal(t, `""`, determineZeroValue("string"))
	assert.Equal(t, "0.0", determineZeroValue("float64"))
	assert.Equal(t, "nil", determineZeroValue("any"))
	assert.Equal(t, `"{}"`, determineZeroValue("json"))
	assert.Equal(t, "[]byte{}", determineZeroValue("hex"))
	assert.Equal(t, "false", determineZeroValue("bool"))
	assert.Equal(t, "[]string{}", determineZeroValue("[]string"))
}

func TestNowYear(t *testing.T) {
	require.NotEmpty(t, nowYear())
}

func TestToURIEncoded(t *testing.T) {
	assert.Equal(t, "Apache%202.0", toURIEncoded("Apache 2.0"))
}

func TestNormaliseActionName(t *testing.T) {
	assert.Equal(t, "go-sdk-generator-action", normaliseActionName("Go SDK Generator"))
	assert.Equal(t, "verify-cmsis-example-sdk-action", normaliseActionName("verify_cmsis example.sdk"))
	assert.Equal(t, "go-sdk-generator-action", normaliseActionName("go-sdk-generator-action"))
	assert.Equal(t, "go-sdk-generator-action", normaliseActionName("Go SDK Generator Action"))
	assert.Equal(t, "action", normaliseActionName("   ---   "))
}

func TestDefaults(t *testing.T) {
	assert.Equal(t, "56", defaultValue("port", "56"))
	assert.Equal(t, "56", defaultValue("port", 56))
	assert.Equal(t, `"bf095257-14b5-4e4d-ac40-2134fe27f331"`, defaultValue("uuid", "bf095257-14b5-4e4d-ac40-2134fe27f331"))
	assert.Equal(t, `"https://blah.com"`, defaultValue("host", "https://blah.com"))
	assert.Equal(t, `"adrien.cabarbaye@arm.com"`, defaultValue("email", "adrien.cabarbaye@arm.com"))
	assert.Equal(t, `""`, defaultValue("string", ""))
	assert.Equal(t, `func() string{return ""}`, defaultValue("string", `func() string{return ""}`))
	assert.Equal(t, "0.0", defaultValue("float64", "0.0"))
	assert.Equal(t, "0.0", defaultValue("float64", 0.0))
	assert.Equal(t, "51.45494", defaultValue("float64", 51.45494))
	assert.Equal(t, "0.0", defaultValue("decimal", "0.0"))
	assert.Equal(t, "0.0", defaultValue("decimal", 0.0))
	assert.Equal(t, "51.45494", defaultValue("decimal", 51.45494))
	assert.Equal(t, "nil", defaultValue("any", "nil"))
	assert.Equal(t, `"{"test":1}"`, defaultValue("json", `{"test":1}`))
	assert.Equal(t, `[]byte("aaffbb77")`, defaultValue("[]byte", "aaffbb77"))
	assert.Equal(t, "func() []byte {\nvalues, _ := hex.DecodeString(\"aaffbb77\")\nreturn values\n}()", defaultValue("hex", "aaffbb77"))
	assert.Equal(t, "func() time.Duration {\nduration, _ := time.ParseDuration(\"5m\")\nreturn duration\n}()", defaultValue("duration", "5m"))
	assert.Equal(t, "5 * time.Minute", defaultValue("duration", "5 * time.Minute"))
	assert.Equal(t, "false", defaultValue("bool", "false"))
}

func TestDefaultsWithMultilines(t *testing.T) {
	assert.Equal(t, `"test value"`, defaultValue("string", "test value\n"))
	assert.Equal(t, `"test value"`, defaultValue("string", "test value\n\n"))
	assert.Equal(t, `"test value"`, defaultValue("string", "  test value  \n"))
	assert.Equal(t, `"test value"`, defaultValue("string", "  test value\r\n"))
	assert.Equal(t, "func(){}", defaultValue("string", "func(){}\n"))
	assert.Equal(t, "123", defaultValue("int", "123\n"))
	assert.Equal(t, "42.5", defaultValue("float64", "42.5\n"))
	assert.Equal(t, "5 * time.Minute", defaultValue("duration", "5 * time.Minute\n"))
	assert.Equal(t, `"{"test":1}"`, defaultValue("json", "{\"test\":1}\n"))
}

func TestDefaultValueForStrings(t *testing.T) {
	assert.Equal(t, `func() string{return ""}`, defaultValueForStrings(`func() string{return ""}`))
	assert.Equal(t, "filesystem.TempDirectory()", defaultValueForStrings(" tempdir "))
	assert.Equal(t, "xdg.BinHome", defaultValueForStrings("bindir"))
	assert.Equal(t, "xdg.CacheHome", defaultValueForStrings("cache"))
	assert.Equal(t, "xdg.ConfigHome", defaultValueForStrings("configuration"))
	assert.Equal(t, "rand.Text()", defaultValueForStrings("random"))
	assert.Equal(t, "time.Now().UTC().Format(time.RFC3339)", defaultValueForStrings("now"))
	assert.Equal(t, `"literal"`, defaultValueForStrings("literal"))
	assert.Contains(t, defaultValueForStrings("hostname"), "os.Hostname()")
	assert.Contains(t, defaultValueForStrings("uuid"), "idgen.GenerateUUID4()")
	assert.Contains(t, defaultValueForStrings("here"), "filesystem.CurrentDirectory()")
	assert.Contains(t, defaultValueForStrings("home"), "os.UserHomeDir()")
	assert.Contains(t, defaultValueForStrings("      Maddy      's home"), `platform.GetHomeDirectory("Maddy")`)
}

func Test_toCase(t *testing.T) {
	assert.Equal(t, "Port", toPascalCase("Port"))
	assert.Equal(t, "Port", toPascalCase("port"))
	assert.Equal(t, "APITestAPIXMLTest", toPascalCase("ApiTestApiXMLTest"))
	assert.Equal(t, "ThisIsOneMessedUpStringCanWeReallyPascalCaseIt", toPascalCase("ThisIsOne___messed up string. Can we Really pascal-case It"))
	assert.Equal(t, "AdrienID", toPascalCase("adrien_id"))
	assert.Equal(t, "AdrienIdentifier", toPascalCase("adrien_identifier"))
	assert.Equal(t, "port", toCamelCase("Port"))
	assert.Equal(t, "port", toCamelCase("port"))
	assert.Equal(t, "apiTestAPIXMLTest", toCamelCase("ApiTestApiXMLTest"))
	assert.Equal(t, "thisIsOneMessedUpStringCanWeReallyCamelCaseIt", toCamelCase("ThisIsOne___messed up string. Can we Really camel-case It"))
	assert.Equal(t, "sourceName", toCamelCase("sourceName"))
	assert.Equal(t, "sourceName", toCamelCase("SourceName"))
	assert.Equal(t, "adrienID", toCamelCase("adrien_id"))
	assert.Equal(t, "adrienIdentifier", toCamelCase("adrien_identifier"))
	assert.Equal(t, "port", toSnakeCase("Port"))
	assert.Equal(t, "port", toSnakeCase("port"))
	assert.Equal(t, "this_is_one_messed_up_string_can_we_really_snake_case_it_test", toSnakeCase("ThisIsOne___messed up string. Can we Really Snake Case It,test"))
}

func Test_ignoreLintingError(t *testing.T) {
	assert.True(t, ignoreLintingError("afaas-store-token"))
	assert.True(t, ignoreLintingError("serveAFAASStoreTokenArg"))
	assert.True(t, ignoreLintingError("apiKey"))
	assert.False(t, ignoreLintingError(faker.Paragraph()))
	assert.False(t, ignoreLintingError(faker.Password()))
}

func Test_packageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already_valid",
			input:    "sdk",
			expected: "sdk",
		},
		{
			name:     "removes_dashes_and_upper_cases",
			input:    "My-SDK",
			expected: "mysdk",
		},
		{
			name:     "invalid_characters",
			input:    "sdk--tool!!",
			expected: "sdktool",
		},
		{
			name:     "leading_digit",
			input:    "123module",
			expected: "module",
		},
		{
			name:     "multiple_underscores",
			input:    "__sdk___module__",
			expected: "sdkmodule",
		},
		{
			name:     "default_only_invalid",
			input:    "----",
			expected: "main",
		},
		{
			name:     "default_empty",
			input:    "",
			expected: "main",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, toPackageName(test.input))
		})
	}
}
