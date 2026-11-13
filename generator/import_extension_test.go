/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestFilterExtraImportsByPath(t *testing.T) {
	t.Parallel()

	extraImports := ExtraImports{
		{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
		{ImportRef: field.ToOptionalString("jsonutils"), ImportUrl: field.ToOptionalString(jsonUtilsImportPath)},
		{ImportRef: field.ToOptionalString("custom"), ImportUrl: field.ToOptionalString("example.com/custom")},
	}

	filtered := filterExtraImportsByPath(extraImports, filesystemImportPath, "")
	assert.Len(t, filtered, 2)
	assert.Equal(t, jsonUtilsImportPath, field.OptionalString(filtered[0].ImportUrl, ""))
	assert.Equal(t, "example.com/custom", field.OptionalString(filtered[1].ImportUrl, ""))
}

func TestConfigExtraImportsKeepsExplicitTemplateImportsWhenUsed(t *testing.T) {
	t.Parallel()

	spec := &SDKSpecification{
		ExtraImports: ExtraImports{
			{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
			{ImportRef: field.ToOptionalString("jsonutils"), ImportUrl: field.ToOptionalString(jsonUtilsImportPath)},
			{ImportRef: field.ToOptionalString("custom"), ImportUrl: field.ToOptionalString("example.com/custom")},
		},
		GlobalConfiguration: Configuration{{
			Name:            "ReportFile",
			LangType:        string(BuiltInLangTypeFilepath),
			MapstructureTag: "report_file",
			Description:     "Report file.",
		}, {
			Name:            "Payload",
			LangType:        string(BuiltInLangTypeValidJSON),
			MapstructureTag: "payload",
			Description:     "Payload.",
		}},
	}

	filtered := configExtraImports(spec)
	assert.Len(t, filtered, 3)
	assert.Equal(t, filesystemImportPath, field.OptionalString(filtered[0].ImportUrl, ""))
	assert.Equal(t, jsonUtilsImportPath, field.OptionalString(filtered[1].ImportUrl, ""))
	assert.Equal(t, "example.com/custom", field.OptionalString(filtered[2].ImportUrl, ""))
}

func TestCLIConfigExtraImportsKeepsExplicitTemplateImportsWhenUsed(t *testing.T) {
	t.Parallel()

	spec := &SDKSpecification{
		ExtraImports: ExtraImports{
			{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
			{ImportRef: field.ToOptionalString("jsonutils"), ImportUrl: field.ToOptionalString(jsonUtilsImportPath)},
			{ImportRef: field.ToOptionalString("custom"), ImportUrl: field.ToOptionalString("example.com/custom")},
			{ImportRef: field.ToOptionalString("validationUtils"), ImportUrl: field.ToOptionalString(validationUtilsImportPath)},
		},
		CommandLineCommands: &CommandLineCommands{Functions: []CliFunction{{
			Name:        "Generate",
			Description: "Generate.",
			Configuration: []FunctionConfigFields{{
				Name:            "SummaryFile",
				LangType:        string(BuiltInLangTypeFilepath),
				MapstructureTag: "summary_file",
				Description:     "Summary file.",
			}, {
				Name:            "Payload",
				LangType:        string(BuiltInLangTypeValidJSON),
				MapstructureTag: "payload",
				Description:     "Payload.",
				Required:        true,
			}},
		}}},
	}

	filtered := cliConfigExtraImports(spec)
	assert.Len(t, filtered, 4)
	assert.Equal(t, filesystemImportPath, field.OptionalString(filtered[0].ImportUrl, ""))
	assert.Equal(t, jsonUtilsImportPath, field.OptionalString(filtered[1].ImportUrl, ""))
	assert.Equal(t, "example.com/custom", field.OptionalString(filtered[2].ImportUrl, ""))
	assert.Equal(t, validationUtilsImportPath, field.OptionalString(filtered[3].ImportUrl, ""))
}

func TestTypesExtraImportsKeepsExplicitTemplateImportsWhenUsed(t *testing.T) {
	t.Parallel()

	spec := &SDKSpecification{
		ExtraImports: ExtraImports{
			{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
			{ImportRef: field.ToOptionalString("jsonutils"), ImportUrl: field.ToOptionalString(jsonUtilsImportPath)},
			{ImportRef: field.ToOptionalString("custom"), ImportUrl: field.ToOptionalString("example.com/custom")},
		},
		AdditionalTypes: Types{{
			Name:        "Artefact",
			Description: "Artefact.",
			Fields: []EntityField{{
				Name:            "Path",
				LangType:        string(BuiltInLangTypeFilepath),
				MapstructureTag: field.ToOptionalString("path"),
				Description:     "Path.",
			}, {
				Name:            "Payload",
				LangType:        string(BuiltInLangTypeValidJSON),
				MapstructureTag: field.ToOptionalString("payload"),
				Description:     "Payload.",
			}},
		}},
	}

	filtered := typesExtraImports(spec)
	assert.Len(t, filtered, 3)
	assert.Equal(t, filesystemImportPath, field.OptionalString(filtered[0].ImportUrl, ""))
	assert.Equal(t, jsonUtilsImportPath, field.OptionalString(filtered[1].ImportUrl, ""))
	assert.Equal(t, "example.com/custom", field.OptionalString(filtered[2].ImportUrl, ""))
}

func TestInterfaceExtraImportsOnlyKeepsReferencedImports(t *testing.T) {
	t.Parallel()

	spec := &SDKSpecification{
		ExtraImports: ExtraImports{
			{ImportRef: field.ToOptionalString("xdg"), ImportUrl: field.ToOptionalString(xdgImportPath)},
			{ImportRef: field.ToOptionalString("jsonutils"), ImportUrl: field.ToOptionalString(jsonUtilsImportPath)},
			{ImportRef: field.ToOptionalString("custom"), ImportUrl: field.ToOptionalString("example.com/custom")},
		},
		Interface: Interface{Methods: []InterfaceMethod{{
			Name:        "Generate",
			Description: "Generate.",
			Args: []CommandArgument{{
				Name:        "Home",
				LangType:    "xdg.RelPath",
				Description: "Home path.",
			}},
		}}},
	}

	filtered := interfaceExtraImports(spec)
	assert.Len(t, filtered, 1)
	assert.Equal(t, xdgImportPath, field.OptionalString(filtered[0].ImportUrl, ""))
}

func TestFormatImportsHonoursLocalPrefix(t *testing.T) {

	localFs := filesystem.NewStandardFileSystem()

	tmpDir := t.TempDir()

	tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")

	tmpFileContent := "package main\n\nimport (\n\t\"github.com/example/project/generator/config\"\n\t\"context\"\n\t\"github.com/stretchr/testify/assert\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = assert.ObjectsAreEqual\n\t_ = config.DefaultSDKGeneratorManagerConfiguration\n}\n"

	err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)

	require.NoError(t, err)

	cleaned, err := formatImports(context.Background(), tmpFile, "github.com/example/project")

	assert.NoError(t, err)

	assert.Equal(t, "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/stretchr/testify/assert\"\n\n\t\"github.com/example/project/generator/config\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = assert.ObjectsAreEqual\n\t_ = config.DefaultSDKGeneratorManagerConfiguration\n}\n", string(cleaned))

}

func TestFormatImportsMergesOrganisationPrefixedImports(t *testing.T) {

	localFs := filesystem.NewStandardFileSystem()

	tmpDir := t.TempDir()

	tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")

	tmpFileContent := "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/charmbracelet/fang\"\n\t\"github.com/spf13/cobra\"\n\t\"github.com/spf13/viper\"\n\n\tconfigUtils \"github.com/ARM-software/golang-utils/utils/config\"\n\t\"github.com/ARM-software/golang-utils/utils/semver\"\n\n\t\"github.com/ARM-software/go-sdk-generator/cli/cmd/config\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = fang.Version(\"1\")\n\t_ = cobra.Command{}\n\t_ = viper.New()\n\t_ = configUtils.DefaultKeyDelim\n\t_ = semver.TrimGoPrefix\n\t_ = config.App\n}\n"

	err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)

	require.NoError(t, err)

	cleaned, err := formatImports(context.Background(), tmpFile, "github.com/ARM-software")

	assert.NoError(t, err)

	assert.Equal(t, "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/charmbracelet/fang\"\n\t\"github.com/spf13/cobra\"\n\t\"github.com/spf13/viper\"\n\n\tconfigUtils \"github.com/ARM-software/golang-utils/utils/config\"\n\t\"github.com/ARM-software/golang-utils/utils/semver\"\n\t\"github.com/ARM-software/go-sdk-generator/cli/cmd/config\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = fang.Version(\"1\")\n\t_ = cobra.Command{}\n\t_ = viper.New()\n\t_ = configUtils.DefaultKeyDelim\n\t_ = semver.TrimGoPrefix\n\t_ = config.App\n}\n", string(cleaned))

}

func TestFormatImportsMergesOtherOrganisationPrefixedImports(t *testing.T) {

	localFs := filesystem.NewStandardFileSystem()

	tmpDir := t.TempDir()

	tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")

	tmpFileContent := "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/charmbracelet/fang\"\n\n\tconfigUtils \"github.com/example/shared/config\"\n\t\"github.com/example/shared/semver\"\n\n\t\"github.com/example/project/cli/cmd/config\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = fang.Version(\"1\")\n\t_ = configUtils.DefaultKeyDelim\n\t_ = semver.TrimGoPrefix\n\t_ = config.App\n}\n"

	err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)

	require.NoError(t, err)

	cleaned, err := formatImports(context.Background(), tmpFile, "github.com/example")

	assert.NoError(t, err)

	assert.Equal(t, "package main\n\nimport (\n\t\"context\"\n\n\t\"github.com/charmbracelet/fang\"\n\n\tconfigUtils \"github.com/example/shared/config\"\n\t\"github.com/example/shared/semver\"\n\t\"github.com/example/project/cli/cmd/config\"\n)\n\nfunc main() {\n\t_ = context.Background()\n\t_ = fang.Version(\"1\")\n\t_ = configUtils.DefaultKeyDelim\n\t_ = semver.TrimGoPrefix\n\t_ = config.App\n}\n", string(cleaned))

}
