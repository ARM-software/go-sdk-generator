package generator

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormaliseKeyNamesCoalescesIdentifierVariants(t *testing.T) {
	t.Parallel()

	aNameKey := normaliseKeyNames("a_name")
	assert.Equal(t, aNameKey, normaliseKeyNames("AName"))
	assert.Equal(t, aNameKey, normaliseKeyNames("aName"))
	assert.Equal(t, aNameKey, normaliseKeyNames("a-name"))
	prettySDKKey := normaliseKeyNames("prettySDK")
	assert.Equal(t, prettySDKKey, normaliseKeyNames("pretty_sdk"))
	assert.Equal(t, prettySDKKey, normaliseKeyNames("prettysdk"))
	assert.Equal(t, prettySDKKey, normaliseKeyNames("pretty-sdk"))
}

func TestDuplicateCaseInsensitiveNamesTreatsIdentifierVariantsAsDuplicates(t *testing.T) {
	t.Parallel()

	duplicates := duplicateCaseInsensitiveNames([]string{"a_name", "AName", "aName", "a-name"})
	require.Len(t, duplicates, 1)
	assert.Equal(t, "a_name", duplicates[0])

	duplicates = duplicateCaseInsensitiveNames([]string{"prettySDK", "pretty_sdk", "prettysdk", "pretty-sdk"})
	require.Len(t, duplicates, 1)
	assert.Equal(t, "prettySDK", duplicates[0])
}

func TestValidateUniqueCaseInsensitiveNamesRejectsIdentifierVariants(t *testing.T) {
	t.Parallel()

	err := validateUniqueCaseInsensitiveNames("option set \"Build\"", "parameter", []string{"a_name", "AName", "aName", "a-name"})
	errortest.AssertError(t, err, commonerrors.ErrConflict)
	errortest.AssertErrorDescription(t, err, `option set "Build" defines duplicate parameter names [a_name]`)

	err = validateUniqueCaseInsensitiveNames("option set \"Build\"", "parameter", []string{"prettySDK", "pretty_sdk", "prettysdk", "pretty-sdk"})
	errortest.AssertError(t, err, commonerrors.ErrConflict)
	errortest.AssertErrorDescription(t, err, `option set "Build" defines duplicate parameter names [prettySDK]`)
}
