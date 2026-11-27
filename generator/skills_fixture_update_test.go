/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/golang-utils/utils/field"
)

// Refresh only intentional skill-related outputs without rebuilding fixture
// dependencies or changing UUIDs, Go modules, or unrelated generated sources.
func TestUpdateSkillProjectFixtures(t *testing.T) {
	if os.Getenv("UPDATE_SKILL_PROJECT_FIXTURES") != "1" {
		t.Skip("set UPDATE_SKILL_PROJECT_FIXTURES=1 to refresh skill-related fixture entries")
	}
	archives, err := filepath.Glob("testdata/*.zip")
	require.NoError(t, err)
	for _, archive := range archives {
		t.Run(filepath.Base(archive), func(t *testing.T) {
			manifest := strings.TrimSuffix(archive, ".zip") + ".yml"
			if _, statErr := os.Stat(manifest); os.IsNotExist(statErr) {
				manifest = strings.TrimSuffix(archive, ".zip") + ".json"
			}
			if _, statErr := os.Stat(manifest); os.IsNotExist(statErr) {
				t.Skip("not a generated project fixture")
			}
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input, cfg.Output = manifest, t.TempDir()
			cfg.ToolVersion = field.ToOptionalString("0.31.0")
			require.NoError(t, Generate(t.Context(), nil, cfg))
			root, rootErr := os.OpenRoot(cfg.Output)
			require.NoError(t, rootErr)
			t.Cleanup(func() { require.NoError(t, root.Close()) })
			original, readErr := os.ReadFile(archive)
			require.NoError(t, readErr)
			reader, zipErr := zip.NewReader(bytes.NewReader(original), int64(len(original)))
			require.NoError(t, zipErr)
			updates := map[string][]byte{}
			require.NoError(t, filepath.WalkDir(cfg.Output, func(name string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil || entry.IsDir() {
					return walkErr
				}
				rel, relErr := filepath.Rel(cfg.Output, name)
				if relErr != nil {
					return relErr
				}
				rel = filepath.ToSlash(rel)
				if strings.HasPrefix(rel, ".agents/skills/") || strings.HasPrefix(rel, "actions/validate-skills/") || slices.Contains([]string{
					"README.md", "DOCUMENTATION.md", "agent.md", ".github/copilot-instructions.md", "guides/agent-skills.md",
					".goreleaser.yml", ".gitattributes", ".qlty/qlty.toml", "CONTRIBUTING.md",
					".github/workflows/skills.yml", ".github/workflows/ci.yml", ".github/workflows/licensing.yml",
					".github/workflows/upgrade-generator.yml", ".github/workflows/dependabot.yml", ".github/workflows/release.yml", ".github/workflows/sbom.yml",
				}, rel) {
					updates[rel], relErr = root.ReadFile(rel)
				}
				return relErr
			}))
			var result bytes.Buffer
			writer := zip.NewWriter(&result)
			var changed []string
			for _, entry := range reader.File {
				name := strings.ReplaceAll(entry.Name, `\`, "/")
				data, found := updates[name]
				delete(updates, name)
				if found {
					input, openErr := entry.Open()
					require.NoError(t, openErr)
					old, oldErr := io.ReadAll(input)
					require.NoError(t, input.Close())
					require.NoError(t, oldErr)
					found = !bytes.Equal(bytes.ReplaceAll(old, []byte("\r\n"), []byte("\n")), bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")))
				}
				if !found {
					require.NoError(t, writer.Copy(entry))
					continue
				}
				header := entry.FileHeader
				output, createErr := writer.CreateHeader(&header)
				require.NoError(t, createErr)
				_, writeErr := output.Write(data)
				require.NoError(t, writeErr)
				changed = append(changed, name)
			}
			var added []string
			for name := range updates {
				added = append(added, name)
			}
			slices.Sort(added)
			for _, name := range added {
				output, createErr := writer.Create(name)
				require.NoError(t, createErr)
				_, writeErr := output.Write(updates[name])
				require.NoError(t, writeErr)
				changed = append(changed, name)
			}
			require.NoError(t, writer.Close())
			if sha256.Sum256(original) == sha256.Sum256(result.Bytes()) || len(changed) == 0 {
				return
			}
			slices.Sort(changed)
			t.Logf("Changed fixture entries: %s", strings.Join(changed, ", "))
			require.NoError(t, os.WriteFile(archive, result.Bytes(), 0600))
		})
	}
}
