package acronyms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRules(t *testing.T) {
	rules, err := loadRules(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, rules)
}

func TestApplyKnownAcronyms(t *testing.T) {
	t.Run("replaces configured acronym in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("ApiClient")
		require.NoError(t, err)
		assert.Equal(t, "APIClient", value)
	})

	t.Run("replaces consecutive configured acronyms in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("OpenAiApiKey")
		require.NoError(t, err)
		assert.Equal(t, "OpenAIAPIKey", value)
	})

	t.Run("replaces configured acronym in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("apiClient")
		require.NoError(t, err)
		assert.Equal(t, "apiClient", value)
	})

	t.Run("replaces consecutive configured acronyms in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("openAiApiKey")
		require.NoError(t, err)
		assert.Equal(t, "openAIAPIKey", value)
	})

	t.Run("preserves identifier family words", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("MaddieIdentifier")
		require.NoError(t, err)
		assert.Equal(t, "MaddieIdentifier", value)
	})

	t.Run("preserves idempotent family words", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("idempotentRetry")
		require.NoError(t, err)
		assert.Equal(t, "idempotentRetry", value)
	})

	t.Run("still rewrites explicit id token", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("adrienId")
		require.NoError(t, err)
		assert.Equal(t, "adrienID", value)
	})

	t.Run("preserves already uppercase acronym runs", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("OpenAIAPIKey")
		require.NoError(t, err)
		assert.Equal(t, "OpenAIAPIKey", value)
	})

	t.Run("preserves adjacent uppercase acronym runs", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("AIAPI")
		require.NoError(t, err)
		assert.Equal(t, "AIAPI", value)
	})

	t.Run("preserves already uppercase acronym runs in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("openAIAPIKey")
		require.NoError(t, err)
		assert.Equal(t, "openAIAPIKey", value)
	})

	t.Run("preserves configured ai exception words", func(t *testing.T) {
		testCases := map[string]string{
			"acquaint":      "acquaint",
			"emailAddress":  "emailAddress",
			"maintainState": "maintainState",
			"railwayTrack":  "railwayTrack",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured api exception words", func(t *testing.T) {
		testCases := map[string]string{
			"apiaryLocation":  "apiaryLocation",
			"apicultureGuide": "apicultureGuide",
			"capillaryFlow":   "capillaryFlow",
			"vapidityScore":   "vapidityScore",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("leaves unknown words alone apart from case normalisation", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("sourceName")
		require.NoError(t, err)
		assert.Equal(t, "sourceName", value)
	})
}
