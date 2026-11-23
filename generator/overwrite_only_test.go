/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/go-sdk-generator/generator/options"
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestOverwriteOnlySelection(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	root := t.TempDir()
	outputs := []string{
		"README.md", "nested/README.md", "README.md.bak", ".github/workflows/ci.yml",
		"example/example.go", "example/interfaces.go", "example/config/example.go", "example/go.mod",
	}
	for _, tc := range []struct {
		name string
		opts []options.OverwriteOption
		included []string
	}{
		{name: "no selection", included: outputs},
		{name: "empty selectors", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{}), options.WithOnlyGroups(nil)}, included: outputs},
		{name: "exact basename", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"README.md"})}, included: []string{"README.md", "nested/README.md"}},
		{name: "normalised relative paths", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{`.github\workflows\ci.yml`, "./README.md", "folder/../example/interfaces.go"})}, included: []string{".github/workflows/ci.yml", "README.md", "example/interfaces.go"}},
		{name: "group", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"sdk"})}, included: []string{"example/example.go", "example/interfaces.go", "example/config/example.go"}},
		{name: "file and group union", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"README.md"}), options.WithOnlyGroups([]string{"sdk"})}, included: []string{"README.md", "nested/README.md", "example/example.go", "example/interfaces.go", "example/config/example.go"}},
		{name: "excluded file wins over group", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"sdk"}), options.WithExcludeFiles([]string{"interfaces.go"})}, included: []string{"example/example.go", "example/config/example.go"}},
		{name: "excluded group wins over file", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"README.md"}), options.WithExcludeGroups([]string{"readme"})}, included: []string{"nested/README.md"}},
		{name: "excluded file wins over file", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"README.md"}), options.WithExcludeFiles([]string{"README.md"})}},
		{name: "excluded group wins over group", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"sdk"}), options.WithExcludeGroups([]string{"sdk"})}},
		{name: "unmatched file", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"unmatched.txt"})}},
		{name: "group without outputs", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"cli-implementation"})}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules, err := newOverwrite(root, true, tc.opts...)
			require.NoError(t, err)
			rules.addGeneratedGroups(resolveGeneratedFileGroups(&SDKSpecification{MainModule: "example"}, nil))
			included := mapset.NewSet(tc.included...)
			for _, file := range outputs {
				output := filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file))
				require.Equal(t, included.Contains(file), rules.shouldInclude(output), file)
				require.Equal(t, !included.Contains(file), rules.shouldExclude(output), file)
			}
		})
	}
}

func TestOverwriteOnlyFilesRejectUnsafePaths(t *testing.T) {
	for _, file := range []string{"", ".", "../secret", `..\secret`, "/tmp/file", `C:\file`, `\\server\share\file`} {
		t.Run(file, func(t *testing.T) {
			_, err := newOverwrite(t.TempDir(), true, options.WithOnlyFiles([]string{file}))
			require.Error(t, err)
		})
	}
	_, err := newOverwrite(t.TempDir(), true, options.WithOnlyGroups([]string{"sdk", "unknown"}))
	require.Error(t, err)
}

func TestOverwriteOnlyOutputs(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	for _, tc := range []struct {
		name string
		opts []options.OverwriteOption
		files []string
	}{
		{name: "readme only", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"README.md"})}, files: []string{"README.md"}},
		{name: "implementation only", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"cli-implementation"})}, files: []string{"cli/cmd/generator/implementations.go"}},
		{name: "union", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"cli-implementation"}), options.WithOnlyFiles([]string{"README.md"})}, files: []string{"cli/cmd/generator/implementations.go", "README.md"}},
		{name: "exclusions win", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"cli-implementation"}), options.WithOnlyFiles([]string{"README.md"}), options.WithExcludeFiles([]string{"implementations.go"})}, files: []string{"README.md"}},
		{name: "unmatched selection", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"unmatched.txt"})}},
		{name: "sdk only", opts: []options.OverwriteOption{options.WithOnlyGroups([]string{"sdk"})}, files: []string{"generator/generator.go", "generator/interfaces.go", "generator/config/generator.go", "generator/options/generator.go", "generator/options/enums.go"}},
		{name: "bookkeeping only", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{".gitattributes", ".qlty/qlty.toml"})}, files: []string{".gitattributes", ".qlty/qlty.toml"}},
		{name: "protected output only", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"module.properties"})}, files: []string{"cli/cmd/module.properties"}},
		{name: "action outputs only", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{"actions/generate/action.yml", "actions/_shared/install-and-run.sh"})}, files: []string{"actions/generate/action.yml", "actions/_shared/install-and-run.sh"}},
		{name: "skill without embedded copy", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{".agents/skills/run-go-sdk-generator/SKILL.md"})}, files: []string{".agents/skills/run-go-sdk-generator/SKILL.md"}},
		{name: "skill with embedded copy", opts: []options.OverwriteOption{options.WithOnlyFiles([]string{".agents/skills/run-go-sdk-generator/SKILL.md", "generator/skills/run-go-sdk-generator/SKILL.md"})}, files: []string{".agents/skills/run-go-sdk-generator/SKILL.md", "generator/skills/run-go-sdk-generator/SKILL.md"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input = filesystem.FilePathJoin(localFs, "..", "generate.yml")
			cfg.Output, cfg.Overwrite = t.TempDir(), true
			require.NoError(t, Generate(t.Context(), nil, cfg, tc.opts...))
			files, err := localFs.LsRecursive(t.Context(), cfg.Output, false)
			require.NoError(t, err)
			relative := collection.Map(files, func(file string) string {
				name, err := filesystem.FilePathRel(localFs, cfg.Output, file)
				require.NoError(t, err)
				return filesystem.FilePathToSlash(localFs, name)
			})
			require.ElementsMatch(t, tc.files, relative, "only the selected files should be written")
		})
	}
}
