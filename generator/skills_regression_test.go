/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// Check skill snapshots without resolving or building generated SDK dependencies.
// UPDATE_CLI_SKILL_FIXTURES=1 updates only CLI skill entries in the archives.
func TestCLISkillRegressionFixtures(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	archives, err := localFs.Glob(filesystem.FilePathJoin(localFs, "testdata", "*.zip"))
	require.NoError(t, err)
	require.NotEmpty(t, archives, "regression fixtures should be discovered")
	for _, archive := range archives {
		t.Run(filesystem.FilePathBase(localFs, archive), func(t *testing.T) {
			original, err := localFs.ReadFile(archive)
			require.NoError(t, err)
			reader, err := zip.NewReader(bytes.NewReader(original), int64(len(original)))
			require.NoError(t, err)
			var skill *zip.File
			for _, entry := range reader.File {
				name := strings.ReplaceAll(entry.Name, `\`, "/")
				if strings.HasPrefix(name, ".agents/skills/run-") && strings.HasSuffix(name, "/SKILL.md") {
					skill = entry
					break
				}
			}
			if skill == nil {
				t.Skip("fixture does not contain a CLI skill")
			}

			manifest := strings.TrimSuffix(archive, ".zip") + ".yml"
			content, err := localFs.ReadFile(manifest)
			if commonerrors.Any(err, commonerrors.ErrNotFound) {
				manifest = strings.TrimSuffix(archive, ".zip") + ".json"
				content, err = localFs.ReadFile(manifest)
			}
			require.NoError(t, err)
			var sdk SDK
			require.NoError(t, unmarshalManifestContent(filesystem.FilePathExt(localFs, manifest), content, &sdk))
			for _, preprocess := range []func(context.Context, *SDK) error{
				ProcessExtraImports, ProcessRepositoryAndModuleMetadata, ProcessAdditionalTypePackages,
				ProcessOptionPackages, ProcessBindings, ProcessArguments, ProcessConstraintMirrors,
			} {
				require.NoError(t, preprocess(t.Context(), &sdk))
			}
			data := &DataModel{SDK: sdk, ToolVersion: "0.31.0", Distribution: precomputeDistribution(&sdk.Specification), OCI: precomputeOCIImageDistribution(&sdk.Specification)}
			tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, cliSkillFile)
			require.NoError(t, err)
			var rendered bytes.Buffer
			require.NoError(t, tmpl.ExecuteTemplate(&rendered, "cli.skill.md.tmpl", data))
			input, err := skill.Open()
			require.NoError(t, err)
			baseline, err := io.ReadAll(input)
			require.NoError(t, input.Close())
			require.NoError(t, err)
			baselineText := strings.ReplaceAll(string(baseline), "\r\n", "\n")
			renderedText := strings.ReplaceAll(rendered.String(), "\r\n", "\n")
			if baselineText == renderedText {
				return
			}
			if os.Getenv("UPDATE_CLI_SKILL_FIXTURES") != "1" {
				require.Equal(t, baselineText, renderedText, "skill snapshot %s", skill.Name)
				return
			}
			var updated bytes.Buffer
			writer := zip.NewWriter(&updated)
			for _, entry := range reader.File {
				if entry != skill {
					require.NoError(t, writer.Copy(entry))
					continue
				}
				header := entry.FileHeader
				output, err := writer.CreateHeader(&header)
				require.NoError(t, err)
				_, err = io.WriteString(output, renderedText)
				require.NoError(t, err)
			}
			require.NoError(t, writer.Close())
			require.NoError(t, localFs.WriteFile(archive, updated.Bytes(), 0600))
		})
	}
}
