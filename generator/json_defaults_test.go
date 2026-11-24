/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	stdjson "encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	jsonutils "github.com/ARM-software/golang-utils/utils/serialisation/json"
)

type jsonDefaultCase struct {
	Name  string             `json:"name"`
	Value stdjson.RawMessage `json:"value"`
}

func jsonDefaultFixture(t *testing.T, name string) []byte {
	t.Helper()
	localFs := filesystem.GetGlobalFileSystem()
	content, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "json-defaults", name))
	require.NoError(t, err)
	return content
}

func jsonDefaultCases(t *testing.T) []jsonDefaultCase {
	t.Helper()
	var cases []jsonDefaultCase
	require.NoError(t, jsonutils.Unmarshal(jsonDefaultFixture(t, "cases.json"), &cases))
	require.Len(t, cases, 24)
	for _, tc := range cases {
		require.True(t, stdjson.Valid(tc.Value), "case=%s invalid fixture JSON=%q", tc.Name, tc.Value)
	}
	return cases
}

// checkJSONDefaultValue checks lexical preservation separately from the decoded
// JSON value: replacing an escape with an equivalent character is not enough.
func checkJSONDefaultValue(t *testing.T, original, recovered, diagnostic string) {
	t.Helper()
	assert.True(t, stdjson.Valid([]byte(recovered)), "[JSON validity] %s recovered=%q", diagnostic, recovered)
	assert.Equal(t, original, recovered, "[exact preservation] %s", diagnostic)
	var want, got any
	require.NoError(t, jsonutils.Unmarshal([]byte(original), &want), "invalid fixture: %s", diagnostic)
	if assert.NoError(t, jsonutils.Unmarshal([]byte(recovered), &got), "[semantic equivalence: decode] %s recovered=%q", diagnostic, recovered) {
		assert.Equal(t, want, got, "[semantic equivalence] %s recovered=%q", diagnostic, recovered)
	}
}

func TestJSONDefaultRoundTrip(t *testing.T) {
	cases := jsonDefaultCases(t)
	for _, langType := range []string{"json", "validJSON"} {
		for _, tc := range cases {
			t.Run(langType+"/"+tc.Name, func(t *testing.T) {
				value := string(tc.Value)
				expression := defaultFieldValue(nil, "", "json-default-regression", "0.0.0", "JSONDefaults", "Payload", langType, nil, value, false)
				require.True(t, strings.HasPrefix(expression, `"`) || strings.HasPrefix(expression, "`"), "expected a Go string literal: %q", expression)
				recovered, err := strconv.Unquote(expression)
				require.NoError(t, err, "invalid Go string literal: %q", expression)
				checkJSONDefaultValue(t, value, recovered, t.Name())
			})
		}
	}
}

func jsonDefaultManifest(t *testing.T, langType, value string) []byte {
	t.Helper()
	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", jsonDefaultFixture(t, "manifest.yml"), &sdk))
	spec := &sdk.Specification
	spec.GlobalConfiguration[0].LangType = langType
	spec.AdditionalTypes[0].Fields[0].LangType = langType
	spec.Options[0].Functions[0].Parameters[0].LangType = langType
	spec.CommandLineCommands.Functions[0].Configuration[0].LangType = langType
	spec.GlobalConfiguration[0].DefaultValue = field.ToOptionalString(value)
	spec.AdditionalTypes[0].Fields[0].DefaultValue = field.ToOptionalString(value)
	spec.Options[0].Functions[0].Parameters[0].DefaultValue = field.ToOptionalString(value)
	spec.CommandLineCommands.Functions[0].Configuration[0].DefaultValue = field.ToOptionalString(value)
	content, err := jsonutils.Marshal(sdk)
	require.NoError(t, err)
	return content
}

func TestJSONDefaultManifestValues(t *testing.T) {
	for _, langType := range []string{"json", "validJSON"} {
		for _, tc := range jsonDefaultCases(t) {
			t.Run(langType+"/"+tc.Name, func(t *testing.T) {
				manifest := jsonDefaultManifest(t, langType, string(tc.Value))
				require.NoError(t, validateManifestContent(t.Context(), ".json", manifest), "case=%s type=%s input=%q: the manifest must accept this JSON text", tc.Name, langType, tc.Value)
				var decoded SDK
				require.NoError(t, unmarshalManifestContent(".json", manifest, &decoded))
				require.Equal(t, string(tc.Value), *decoded.Specification.GlobalConfiguration[0].DefaultValue, "manifest decoding must preserve the raw JSON text")
				require.Equal(t, string(tc.Value), *decoded.Specification.AdditionalTypes[0].Fields[0].DefaultValue)
				require.Equal(t, string(tc.Value), *decoded.Specification.Options[0].Functions[0].Parameters[0].DefaultValue)
				require.Equal(t, string(tc.Value), *decoded.Specification.CommandLineCommands.Functions[0].Configuration[0].DefaultValue)
			})
		}
	}
}

func TestJSONDefaultManifestRequiresText(t *testing.T) {
	var manifest map[string]any
	require.NoError(t, jsonutils.Unmarshal(jsonDefaultManifest(t, "json", `{}`), &manifest))
	spec := manifest["specification"].(map[string]any)
	configuration := spec["global_configuration"].([]any)
	configuration[0].(map[string]any)["default_value"] = map[string]any{"name": "test"}
	content, err := jsonutils.Marshal(manifest)
	require.NoError(t, err)
	require.Error(t, validateManifestContent(t.Context(), ".json", content), "default_value is JSON text stored in a manifest string, not a native object")
}
