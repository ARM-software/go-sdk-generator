package generator

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteEmbeddedSchemaJSON(t *testing.T) {
	var output bytes.Buffer

	err := WriteEmbeddedSchemaJSON(&output)
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &schema))
	assert.Equal(t, "SDK", schema["title"])
}
