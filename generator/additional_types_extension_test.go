package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			Interface:           Interface{Namespace: "github.com/Arm-Debug/"},
			CommandLineCommands: &CommandLineCommands{
				Package:   "Tools-SDK",
				Functions: []CliFunction{{Name: "Run", Description: "run", Configuration: []FunctionConfigFields{}}},
			},
		},
	}
	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), sdk))
	assert.Equal(t, "toolssdk", sdk.Specification.CommandLineCommands.Package)
}

func TestCommandLineCommandsValidateRejectsNestedPackage(t *testing.T) {
	t.Parallel()
	cmds := &CommandLineCommands{
		Package:   "foo/bar",
		Functions: []CliFunction{{Name: "Run", Description: "run", Configuration: []FunctionConfigFields{}}},
	}
	require.Error(t, cmds.Validate())
}
