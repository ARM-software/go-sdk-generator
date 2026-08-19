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
)

func TestNormaliseAdditionalTypePackageName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, defaultAdditionalTypePackage, normaliseAdditionalTypePackageName(""))
	assert.Equal(t, "models", normaliseAdditionalTypePackageName("Models"))
	assert.Equal(t, "sdkmodels", normaliseAdditionalTypePackageName("sdk-models"))
}

func TestValidateAdditionalTypePackageName(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateAdditionalTypePackageName("models"))
	require.Error(t, validateAdditionalTypePackageName("foo/bar"))
	require.Error(t, validateAdditionalTypePackageName("foo\\bar"))
	require.Error(t, validateAdditionalTypePackageName("foo.bar"))
}

func TestNormaliseCLIPackageName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, defaultCLIPackage, normaliseCLIPackageName(""))
	assert.Equal(t, "tools", normaliseCLIPackageName("Tools"))
	assert.Equal(t, "toolssdk", normaliseCLIPackageName("tools-sdk"))
}

func TestValidateCLIPackageName(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateCLIPackageName("tools"))
	require.Error(t, validateCLIPackageName("foo/bar"))
	require.Error(t, validateCLIPackageName("foo\\bar"))
	require.Error(t, validateCLIPackageName("foo.bar"))
}

func TestProcessAdditionalTypePackages(t *testing.T) {
	t.Parallel()
	sdk := &SDK{
		Specification: SDKSpecification{
			GlobalConfiguration: Configuration{
				{
					Name:            "Seed",
					LangType:        "types.DirectoryEntry",
					MapstructureTag: "seed",
					Description:     "Seed entry.",
				},
			},
			Interface: Interface{
				Methods: []InterfaceMethod{
					{
						Name:        "List",
						Description: "List entries.",
						Returns:     []CommandArgument{{Name: "page", LangType: "types.Page", Description: "Page."}},
					},
				},
			},
			AdditionalTypes: Types{
				{
					Name:        "DirectoryEntry",
					Package:     "Models",
					Description: "Directory entry.",
					Fields:      []EntityField{{Name: "Name", LangType: "string", Description: "Name."}},
				},
				{
					Name:        "Page",
					Package:     "Results",
					Description: "Page.",
					Fields: []EntityField{
						{Name: "Entries", LangType: "[]types.DirectoryEntry", Description: "Entries."},
						{Name: "Current", LangType: "types.Page", Description: "Current page."},
					},
				},
			},
		},
	}

	require.NoError(t, ProcessAdditionalTypePackages(context.Background(), sdk))
	assert.Equal(t, "models", sdk.Specification.AdditionalTypes[0].Package)
	assert.Equal(t, "results", sdk.Specification.AdditionalTypes[1].Package)
	assert.Equal(t, "models.DirectoryEntry", sdk.Specification.GlobalConfiguration[0].LangType)
	assert.Equal(t, "results.Page", sdk.Specification.Interface.Methods[0].Returns[0].LangType)
	assert.Equal(t, "[]models.DirectoryEntry", sdk.Specification.AdditionalTypes[1].Fields[0].LangType)
	assert.Equal(t, "Page", sdk.Specification.AdditionalTypes[1].Fields[1].LangType)
}

func TestProcessRepositoryAndModuleMetadataNormalisesCLIPackage(t *testing.T) {
	t.Parallel()
	sdk := &SDK{
		Specification: SDKSpecification{
			PackageName:         "pkg",
			MainModule:          "mod",
			GlobalConfiguration: Configuration{},
			Interface:           Interface{Namespace: "github.com/ARM-software/"},
			CommandLineCommands: &CommandLineCommands{
				Package:   "Tools-SDK",
				Functions: []CliFunction{{Name: "Run", Description: "run", Configuration: []FunctionConfigFields{}}},
			},
		},
	}
	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk))
	assert.Equal(t, "toolssdk", sdk.Specification.CommandLineCommands.Package)
}

func TestSuppressedRequiredValidationMapIncludesBoolAndEnum(t *testing.T) {
	t.Parallel()
	spec := &SDKSpecification{
		Name: field.ToOptionalString("Example"),
		GlobalConfiguration: Configuration{
			{Name: "Enabled", LangType: "bool", Required: true, Description: "enabled", MapstructureTag: "enabled"},
			{Name: "Mode", LangType: "string", Required: true, Enum: []string{"fast", "safe"}, DefaultValue: field.ToOptionalString("fast"), Description: "mode", MapstructureTag: "mode"},
		},
		CommandLineCommands: &CommandLineCommands{
			Functions: []CliFunction{{
				Name:        "Run",
				Description: "run",
				Configuration: []FunctionConfigFields{
					{Name: "Verbose", LangType: "bool", Required: true, Description: "verbose", MapstructureTag: "verbose"},
				},
			}},
		},
		AdditionalTypes: Types{{
			Name:        "Resource",
			Description: "resource",
			Fields:      []EntityField{{Name: "Active", LangType: "bool", Required: true, Description: "active"}},
		}},
	}

	suppressed := spec.SuppressedRequiredValidationMap()
	assert.True(t, suppressesRequiredValidation(suppressed, globalConfigurationValidationScope, "", "Enabled"))
	assert.True(t, suppressesRequiredValidation(suppressed, globalConfigurationValidationScope, "", "Mode"))
	assert.True(t, suppressesRequiredValidation(suppressed, cliConfigurationValidationScope, "Run", "Verbose"))
	assert.True(t, suppressesRequiredValidation(suppressed, additionalTypeValidationScope, "Resource", "Active"))
}

func TestCommandLineCommandsValidateRejectsNestedPackage(t *testing.T) {
	t.Parallel()
	cmds := &CommandLineCommands{
		Package:   "foo/bar",
		Functions: []CliFunction{{Name: "Run", Description: "run", Configuration: []FunctionConfigFields{}}},
	}
	require.Error(t, cmds.Validate())
}

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
