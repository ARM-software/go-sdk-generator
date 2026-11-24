/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	jsonutils "github.com/ARM-software/golang-utils/utils/serialisation/json"
	yamlutils "github.com/ARM-software/golang-utils/utils/serialisation/yaml"
)

func TestDisableDependabotGeneratorWorkflow(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	fixture, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "casing-types.yml"))
	require.NoError(t, err)
	for _, format := range []struct {
		ext     string
		marshal func(any) ([]byte, error)
	}{
		{ext: ".json", marshal: jsonutils.Marshal},
		{ext: ".yml", marshal: yamlutils.Marshal},
	} {
		for _, tc := range []struct {
			name     string
			value    any
			disabled bool
			invalid  bool
		}{
			{name: "omitted"},
			{name: "false", value: false},
			{name: "true", value: true, disabled: true},
			{name: "string rejected", value: "true", invalid: true},
		} {
			t.Run(format.ext+"/"+tc.name, func(t *testing.T) {
				var manifest map[string]any
				require.NoError(t, yamlutils.Unmarshal(fixture, &manifest))
				if tc.value != nil {
					manifest["specification"].(map[string]any)["disable_dependabot_generator_workflow"] = tc.value
				}
				content, err := format.marshal(manifest)
				require.NoError(t, err)
				err = validateManifestContent(t.Context(), format.ext, content)
				if tc.invalid {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				var sdk SDK
				require.NoError(t, unmarshalManifestContent(format.ext, content, &sdk))
				assert.Equal(t, tc.disabled, sdk.Specification.DisableDependabotGeneratorWorkflow)
				data := &DataModel{SDK: sdk, ToolVersion: "0.31.0", TargetedGoVersion: "1.27.3", ManifestFileName: "generate.yml"}
				output := filesystem.FilePathJoin(localFs, t.TempDir(), "upgrade-generator.yml")
				require.NoError(t, populateTemplate(t.Context(), data, upgradeGeneratorWorkflowFile, output))
				generated, err := localFs.ReadFile(output)
				require.NoError(t, err)
				generatedText := strings.ReplaceAll(string(generated), "\r\n", "\n")
				var workflow struct {
					Jobs map[string]struct{ If string }
				}
				require.NoError(t, yamlutils.Unmarshal(generated, &workflow))
				condition := "github.event_name != 'pull_request'"
				if !tc.disabled {
					condition += " || startsWith(github.head_ref, 'dependabot/')"
				}
				assert.Equal(t, condition, workflow.Jobs["upgrade-generator"].If)
				assert.Equal(t, !tc.disabled, strings.Contains(generatedText, "  pull_request:\n"))
				assert.Contains(t, generatedText, "  schedule:\n    - cron: \"0 6 * * 1\"")
				assert.Contains(t, generatedText, "  workflow_dispatch:\n")
			})
		}
	}
}

func TestGeneratorDisablesDependabotGeneratorWorkflow(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	content, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "..", "generate.yml"))
	require.NoError(t, err)
	require.NoError(t, validateManifestContent(t.Context(), ".yml", content))
	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", content, &sdk))
	assert.True(t, sdk.Specification.DisableDependabotGeneratorWorkflow)
	workflow, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "..", ".github", "workflows", "upgrade-generator.yml"))
	require.NoError(t, err)
	assert.NotContains(t, string(workflow), "  pull_request:")
	assert.NotContains(t, string(workflow), "startsWith(github.head_ref, 'dependabot/')")
}
