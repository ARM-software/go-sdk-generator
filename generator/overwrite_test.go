/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/go-sdk-generator/generator/options"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	jsonutils "github.com/ARM-software/golang-utils/utils/serialisation/json"
)

func TestOverwriteExclusionMatcher(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	root := t.TempDir()
	m, err := newOverwrite(root, true, options.WithExcludeFiles([]string{"README.md", `.github\workflows\ci.yml`, "./root.txt", "foo/../normalised.txt"}))
	require.NoError(t, err)
	for _, file := range []string{"README.md", "nested/README.md", ".github/workflows/ci.yml", "root.txt", "normalised.txt"} {
		output := filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file))
		require.True(t, m.shouldExclude(output), file)
		require.False(t, m.shouldInclude(output), file)
	}
	for _, file := range []string{"README.md.bak", "readme.md", "myREADME.md", "nested/root.txt", "ci.yml", "other/.github/workflows/ci.yml", "../README.md"} {
		output := filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file))
		require.False(t, m.shouldExclude(output), file)
		require.True(t, m.shouldInclude(output), file)
	}
	m.enabled = false
	readme := filesystem.FilePathJoin(localFs, root, "README.md")
	require.False(t, m.shouldExclude(readme))
	require.True(t, m.shouldInclude(readme))
	require.False(t, (*overwrite)(nil).shouldExclude(readme))
	require.True(t, (*overwrite)(nil).shouldInclude(readme))
	for _, file := range []string{"", ".", "..", "../secret", `..\secret`, "a/../../secret", "/tmp/file", `C:\file`, "C:file", `\\server\share\file`, "file:stream", "bad\x00file", "folder/"} {
		t.Run(file, func(t *testing.T) {
			_, err := newOverwrite(root, true, options.WithExcludeFiles([]string{file}))
			require.Error(t, err)
		})
	}
	_, err = newOverwrite(root, true, options.WithExcludeGroups([]string{"project", "invalid"}))
	require.Error(t, err)
}

func TestGenerateRejectsSelectorsWithoutOverwrite(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		option options.OverwriteOption
	}{
		{name: "exclude files", option: options.WithExcludeFiles([]string{"root.go"})},
		{name: "exclude groups", option: options.WithExcludeGroups([]string{"cli"})},
		{name: "only files", option: options.WithOnlyFiles([]string{"README.md"})},
		{name: "only groups", option: options.WithOnlyGroups([]string{"readme"})},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			localFs := filesystem.GetGlobalFileSystem()
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input = filesystem.FilePathJoin(localFs, "testdata", "cli-package.yml")
			cfg.Output = t.TempDir()
			cfg.Overwrite = false

			customOutput := filesystem.FilePathJoin(localFs, cfg.Output, "tools", "cmd", "root.go")
			sentinel := []byte("customised contents must survive invalid selections\n")
			require.NoError(t, localFs.MkDir(filesystem.FilePathDir(localFs, customOutput)))
			require.NoError(t, localFs.WriteFile(customOutput, sentinel, 0600))

			err := Generate(t.Context(), nil, cfg, test.option)

			assert.Error(t, err, "selectors must be rejected when the generator's overwrite setting is disabled")
			content, readErr := localFs.ReadFile(customOutput)
			require.NoError(t, readErr)
			assert.True(t, bytes.Equal(sentinel, content), "invalid selectors must not overwrite customised files")

			files, listErr := localFs.LsRecursive(t.Context(), cfg.Output, false)
			require.NoError(t, listErr)
			assert.Equal(t, 1, len(files), "invalid selectors must not create generated files")
			assert.Contains(t, files, customOutput)
		})
	}
}

func TestOverwriteExclusionGroupsUseExactOutputs(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	var sdk SDK
	manifest, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "..", "generate.yml"))
	require.NoError(t, err)
	require.NoError(t, unmarshalManifestContent(".yml", manifest, &sdk))
	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), &sdk))
	require.NoError(t, ProcessBindings(context.Background(), &sdk))
	require.NoError(t, ProcessArguments(context.Background(), &sdk))
	actions, err := BuildGeneratedActions(&sdk)
	require.NoError(t, err)
	root := t.TempDir()
	for group, files := range map[string][]string{
		"ci":                 {".github/workflows/ci.yml", "actions/generate/action.yml", "actions/_shared/install-and-run.sh", "Dockerfile.generator"},
		"linters":            {".qlty/qlty.toml", ".golangci.yaml"},
		"gomod":              {"go.work", "generator/go.mod", "cli/go.mod"},
		"cli":                {"cli/cmd/root.go", "cli/cmd/generator/wrappers.go"},
		"cli-implementation": {"cli/cmd/generator/implementations.go"},
		"sdk":                {"generator/generator.go", "generator/interfaces.go", "generator/config/generator.go", "generator/options/generator.go", "generator/options/enums.go"},
		"readme":             {"README.md"},
		"project":            {"pyproject.toml", ".github/CODEOWNERS", ".github/dependabot.yml", ".gitattributes"},
		"agent":              {"agent.md", ".github/copilot-instructions.md", ".agents/skills/run-go-sdk-generator/SKILL.md", "generator/skills/run-go-sdk-generator/SKILL.md"},
	} {
		t.Run(group, func(t *testing.T) {
			m, err := newOverwrite(root, true, options.WithExcludeGroups([]string{group}))
			require.NoError(t, err)
			m.addGeneratedGroups(resolveGeneratedFileGroups(&sdk.Specification, actions))
			for _, file := range files {
				require.True(t, m.shouldExclude(filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file))), file)
				require.False(t, m.shouldExclude(filesystem.FilePathJoin(localFs, root, "other", filesystem.FilePathFromSlash(localFs, file))), file)
				require.False(t, m.shouldExclude(filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file+".bak"))), file)
			}
			for _, file := range []string{".github/workflows/custom.yml", "cli/custom.go", ".agents/skills/custom/SKILL.md", "custom/go.mod"} {
				require.False(t, m.shouldExclude(filesystem.FilePathJoin(localFs, root, filesystem.FilePathFromSlash(localFs, file))), file)
			}
		})
	}
}

func TestResolveGeneratedFileGroups(t *testing.T) {
	first := resolveGeneratedFileGroups(&SDKSpecification{
		MainModule:          "first",
		CommandLineCommands: &CommandLineCommands{Package: "tools", Binary: field.ToOptionalString("first-tool")},
	}, []GeneratedAction{{DirectoryName: "publish"}})
	require.Equal(t, []string{"first/go.mod", "go.work", "tools/go.mod"}, first["gomod"])
	require.Contains(t, first["cli"], "tools/main.go")
	require.NotContains(t, first["cli"], "tools/cmd/first/implementations.go")
	require.Equal(t, []string{"tools/cmd/first/implementations.go"}, first["cli-implementation"])
	require.Contains(t, first["ci"], ".github/workflows/ci.yml")
	require.Contains(t, first["ci"], "Dockerfile.first")
	require.Contains(t, first["ci"], "actions/publish/action.yml")
	require.Contains(t, first["ci"], "actions/_shared/install-and-run.sh")
	require.Contains(t, first["agent"], ".agents/skills/run-first-tool/SKILL.md")
	require.Contains(t, first["agent"], ".agents/skills/using-first-tool-in-a-github-action/SKILL.md")
	require.Len(t, first["readme"], 1)
	first["readme"][0] = "changed"

	second := resolveGeneratedFileGroups(&SDKSpecification{MainModule: "second"}, nil)
	require.Equal(t, []string{"second/go.mod"}, second["gomod"])
	require.Empty(t, second["cli"])
	require.Empty(t, second["cli-implementation"])
	require.Empty(t, second["agent"])
	require.Contains(t, second["ci"], "Dockerfile.second")
	require.NotContains(t, second["ci"], "actions/publish/action.yml")
	require.Equal(t, []string{"README.md"}, second["readme"])
	require.Contains(t, second["project"], "module.properties")
	require.NotContains(t, second["project"], "tools/cmd/module.properties")
}

func TestResolveGeneratedFileGroupsSDK(t *testing.T) {
	spec := &SDKSpecification{
		MainModule:          "example",
		Interface:           Interface{Methods: []InterfaceMethod{{Name: "Run", Args: []CommandArgument{{Name: "Mode", LangType: "string", Enum: []string{"fast", "safe"}}}}}},
		GlobalConfiguration: Configuration{{Name: "Mode", LangType: "string", Enum: []string{"fast", "safe"}}},
		AdditionalTypes:     Types{{Name: "Model", Package: "models", Fields: []EntityField{{Name: "Mode", LangType: "string", Enum: []string{"fast", "safe"}}}}},
		Options:             OptionSets{{Name: "Build", Package: "customoptions", Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: "Mode", LangType: "string", Enum: []string{"fast", "safe"}}}}}}},
	}
	groups := resolveGeneratedFileGroups(spec, nil)
	require.ElementsMatch(t, []string{
		"example/example.go", "example/interfaces.go", "example/enums.go",
		"example/config/example.go", "example/config/enums.go",
		"example/models/example.go", "example/models/enums.go",
		"example/customoptions/example.go", "example/customoptions/enums.go",
	}, groups["sdk"])
	require.NotContains(t, groups["sdk"], "example/go.mod")
	require.NotContains(t, groups["sdk"], "example/models/custom.go")
	require.NotContains(t, groups["sdk"], "example/types.gen.go")

	groups = resolveGeneratedFileGroups(&SDKSpecification{MainModule: "minimal"}, nil)
	require.ElementsMatch(t, []string{"minimal/minimal.go", "minimal/interfaces.go", "minimal/config/minimal.go"}, groups["sdk"])
}

func TestOverwriteCLIImplementationGroups(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	implementation := filesystem.FilePathJoin(localFs, "tools", "cmd", "clipackage", "implementations.go")
	wrapper := filesystem.FilePathJoin(localFs, "tools", "cmd", "clipackage", "wrappers.go")
	for _, tc := range []struct {
		group       string
		preserved   string
		regenerated string
	}{
		{group: "cli-implementation", preserved: implementation, regenerated: wrapper},
		{group: "cli", preserved: wrapper, regenerated: implementation},
	} {
		t.Run(tc.group, func(t *testing.T) {
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input = filesystem.FilePathJoin(localFs, "testdata", "cli-package.yml")
			cfg.Output, cfg.Overwrite = t.TempDir(), true
			sentinel := []byte("customised contents\n")
			for _, file := range []string{implementation, wrapper} {
				output := filesystem.FilePathJoin(localFs, cfg.Output, file)
				require.NoError(t, localFs.MkDir(filesystem.FilePathDir(localFs, output)))
				require.NoError(t, localFs.WriteFile(output, sentinel, 0600))
			}
			require.NoError(t, Generate(t.Context(), nil, cfg, options.WithExcludeGroups([]string{tc.group})))
			preserved, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, cfg.Output, tc.preserved))
			require.NoError(t, err)
			require.Equal(t, sentinel, preserved)
			regenerated, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, cfg.Output, tc.regenerated))
			require.NoError(t, err)
			require.NotEqual(t, sentinel, regenerated)
		})
	}
}

func TestOverwriteExclusionsRegression(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	var fixture struct {
		Manifest    string   `json:"manifest"`
		Groups      []string `json:"exclude_groups"`
		Files       []string `json:"exclude_files"`
		OnlyGroups  []string `json:"only_groups"`
		OnlyFiles   []string `json:"only_files"`
		Preserved   []string `json:"preserved"`
		Regenerated []string `json:"regenerated"`
	}
	raw, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "overwrite-exclusions.json"))
	require.NoError(t, err)
	require.NoError(t, jsonutils.Unmarshal(raw, &fixture))
	cfg := config.DefaultSDKGeneratorManagerConfiguration()
	cfg.Input, cfg.Output = fixture.Manifest, t.TempDir()
	require.NoError(t, Generate(context.Background(), nil, cfg))
	sentinel := []byte("handwritten contents must survive\n")
	modified := map[string]time.Time{}
	for _, file := range append(append([]string{}, fixture.Preserved...), fixture.Regenerated...) {
		output := filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file))
		_, err := localFs.Stat(output)
		require.NoError(t, err, "fixture must name an actual generated output: %s", file)
		require.NoError(t, localFs.WriteFile(output, sentinel, 0600))
		past := time.Unix(1000000000, 0)
		require.NoError(t, localFs.Chtimes(output, past, past))
		info, err := localFs.Stat(output)
		require.NoError(t, err)
		modified[file] = info.ModTime()
	}
	cfg.Overwrite = true
	manager, err := NewSDKGenerator(cfg)
	require.NoError(t, err)
	wrapped := NewSDKGeneratorWithDelegate(manager)
	require.NoError(t, wrapped.Generate(context.Background(), nil,
		options.WithExcludeGroups(fixture.Groups), options.WithExcludeFiles(fixture.Files),
		options.WithOnlyGroups(fixture.OnlyGroups), options.WithOnlyFiles(fixture.OnlyFiles)))
	for _, file := range fixture.Preserved {
		output := filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file))
		content, err := localFs.ReadFile(output)
		require.NoError(t, err)
		require.Equal(t, sentinel, content, file)
		info, err := localFs.Stat(output)
		require.NoError(t, err)
		require.Equal(t, modified[file], info.ModTime(), file)
	}
	for _, file := range fixture.Regenerated {
		content, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file)))
		require.NoError(t, err)
		require.NotEqual(t, sentinel, content, file)
	}
	// Exclusions are method-scoped; the same manager's next call has none.
	require.NoError(t, wrapped.Generate(context.Background(), nil))
	for _, file := range []string{"pyproject.toml", "agent.md", "generator/options/generator.go", "actions/generate/action.yml", ".qlty/qlty.toml", ".gitattributes"} {
		content, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file)))
		require.NoError(t, err)
		require.NotEqual(t, sentinel, content, file)
	}
}

func TestOverwriteExclusionsPreventNewProtectedAndDirectOutputs(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	cfg := config.DefaultSDKGeneratorManagerConfiguration()
	cfg.Input, cfg.Output, cfg.Overwrite = "../generate.yml", t.TempDir(), true
	files := []string{"cli/cmd/module.properties", "generator/options/generator.go", "actions/generate/action.yml", ".qlty/qlty.toml", ".gitattributes", "generator/skills/run-go-sdk-generator/SKILL.md"}
	require.NoError(t, Generate(context.Background(), nil, cfg, options.WithExcludeFiles(files)))
	for _, file := range files {
		_, err := localFs.Stat(filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file)))
		require.ErrorIs(t, err, commonerrors.ErrNotFound, file)
	}
}

func TestOverwriteDirectTypeOutputs(t *testing.T) {
	localFs := filesystem.GetGlobalFileSystem()
	files := []string{"enumcollections/types/enumcollections.go", "enumcollections/types/enums.go"}
	for _, tc := range []struct {
		name   string
		option options.OverwriteOption
	}{
		{name: "explicit files", option: options.WithExcludeFiles(files)},
		{name: "sdk group", option: options.WithExcludeGroups([]string{options.OverwriteOptionsExcludeGroupsSDK.String()})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input, cfg.Output = filesystem.FilePathJoin(localFs, "testdata", "enum-collections.yml"), t.TempDir()
			require.NoError(t, Generate(context.Background(), nil, cfg))
			for _, file := range files {
				output := filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file))
				_, err := localFs.Stat(output)
				require.NoError(t, err)
				require.NoError(t, localFs.WriteFile(output, []byte("preserved type"), 0600))
			}
			cfg.Overwrite = true
			require.NoError(t, Generate(context.Background(), nil, cfg, tc.option))
			for _, file := range files {
				content, err := localFs.ReadFile(filesystem.FilePathJoin(localFs, cfg.Output, filesystem.FilePathFromSlash(localFs, file)))
				require.NoError(t, err)
				require.Equal(t, "preserved type", string(content))
			}
		})
	}
}
