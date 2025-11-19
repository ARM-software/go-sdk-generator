package generator

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestLangTypeValidationRule(t *testing.T) {
	assert.Equal(t, ", validationUtils.IsPort", argValidationRule("port"))
	assert.Equal(t, ", is.UUID", argValidationRule("uuid"))
	assert.Equal(t, ", is.URL", argValidationRule("host"))
	assert.Empty(t, argValidationRule("string"))
	assert.Empty(t, argValidationRule(faker.Word()))
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
}

func TestDefaults(t *testing.T) {
	assert.Equal(t, "56", defaultValue("port", "56"))
	assert.Equal(t, "56", defaultValue("port", 56))
	assert.Equal(t, `"bf095257-14b5-4e4d-ac40-2134fe27f331"`, defaultValue("uuid", "bf095257-14b5-4e4d-ac40-2134fe27f331"))
	assert.Equal(t, `"https://blah.com"`, defaultValue("host", "https://blah.com"))
	assert.Equal(t, `"adrien.cabarbaye@arm.com"`, defaultValue("email", "adrien.cabarbaye@arm.com"))
	assert.Equal(t, `""`, defaultValue("string", ""))
	assert.Equal(t, `func() string{return ""\n}`, defaultValue("string", `func() string{return ""\n}`))
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
