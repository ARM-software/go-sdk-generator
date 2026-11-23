/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jsonutils "github.com/ARM-software/golang-utils/utils/serialisation/json"
)

func TestWriteEmbeddedSchemaJSON(t *testing.T) {
	var output bytes.Buffer

	err := WriteEmbeddedSchemaJSON(&output)
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, jsonutils.Unmarshal(output.Bytes(), &schema))
	assert.Equal(t, "SDK", schema["title"])
}
