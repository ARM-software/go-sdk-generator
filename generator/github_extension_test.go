/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestGitHubRepositoryHelpers(t *testing.T) {
	spec := SDKSpecification{
		PackageName: "verify-cmsis-example-sdk",
		MainModule:  "verify",
		Interface: Interface{
			Namespace: "github.com/Arm-Debug/",
		},
		RepositoryUrl: field.ToOptionalString("https://github.com/Arm-Examples/verify-cmsis-example-sdk"),
	}

	assert.Equal(t, "github.com/Arm-Debug", importPrefix(&spec))
	assert.Equal(t, "github.com/Arm-Debug/verify-cmsis-example-sdk", moduleRoot(&spec))
	assert.Equal(t, "github.com/Arm-Debug/verify-cmsis-example-sdk/verify", sdkModuleRoot(&spec))
	assert.Equal(t, "github.com/Arm-Debug/verify-cmsis-example-sdk/cli", cliModuleRoot(&spec))
	assert.True(t, isHostedOnGitHub(&spec))
	assert.Equal(t, "Arm-Examples/verify-cmsis-example-sdk", repositoryLocationOnGitHub(&spec))
	assert.Equal(t, "github.com/Arm-Examples/verify-cmsis-example-sdk", repositoryImportPathOnGitHub(&spec))
	assert.Equal(t, "https://arm-examples.github.io/verify-cmsis-example-sdk", githubPagesURL(&spec))
	assert.Equal(t, "Arm-Examples/verify-cmsis-example-sdk/actions", projectOwnGitHubActionsLocation(&spec))
	assert.Equal(t, "git config --global url.\"https://${{ secrets.GIT_SECRET }}:x-oauth-basic@github.com/Arm-Examples\".insteadOf \"https://github.com/Arm-Examples\"", githubAuthenticationForGoPrivateDependencies(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/issues", githubIssuesURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/pulls", githubPullsURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/discussions", githubDiscussionsURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/blob/main/schema.yml", linkToFileOnGitHub(&spec, "schema.yml"))

	spec.CommandLineCommands = &CommandLineCommands{Package: "tools"}
	assert.Equal(t, "github.com/Arm-Debug/verify-cmsis-example-sdk/tools", cliModuleRoot(&spec))
}

func TestLinkToManifest(t *testing.T) {
	githubSpec := SDKSpecification{
		PackageName:   "verify-cmsis-example-sdk",
		RepositoryUrl: field.ToOptionalString("https://github.com/Arm-Examples/verify-cmsis-example-sdk"),
	}
	nonGitHubSpec := SDKSpecification{
		PackageName:   "verify-cmsis-example-sdk",
		RepositoryUrl: field.ToOptionalString("https://gitlab.example.com/arm/verify-cmsis-example-sdk"),
	}

	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/blob/main/verify.yml", linkToManifest(&githubSpec, "verify"))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/blob/main/verify.x.yml", linkToManifest(&githubSpec, "verify.x.yml"))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/blob/main/.manifest", linkToManifest(&githubSpec, ".manifest"))
	assert.Empty(t, linkToManifest(&nonGitHubSpec, "verify"))
}

func TestNormaliseManifestFileName(t *testing.T) {
	assert.Equal(t, "verify.yml", normaliseManifestFileName("verify"))
	assert.Equal(t, "verify.json", normaliseManifestFileName("verify.json"))
	assert.Equal(t, "verify.yaml", normaliseManifestFileName("verify.yaml"))
	assert.Equal(t, "verify.x.yml", normaliseManifestFileName("verify.x.yml"))
	assert.Equal(t, ".manifest", normaliseManifestFileName(".manifest"))
	assert.Equal(t, ".manifest.yaml", normaliseManifestFileName(".manifest.yaml"))
	assert.Empty(t, normaliseManifestFileName("   "))
}
