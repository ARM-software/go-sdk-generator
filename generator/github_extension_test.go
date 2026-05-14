package generator

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "Arm-Examples/verify-cmsis-example-sdk/actions", projectOwnGitHubActionsLocation(&spec))
	assert.Equal(t, "git config --global url.\"https://${{ secrets.GIT_SECRET }}:x-oauth-basic@github.com/Arm-Examples\".insteadOf \"https://github.com/Arm-Examples\"", githubAuthenticationForGoPrivateDependencies(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/issues", githubIssuesURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/pulls", githubPullsURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/discussions", githubDiscussionsURL(&spec))
	assert.Equal(t, "https://github.com/Arm-Examples/verify-cmsis-example-sdk/blob/main/schema.yml", linkToFileOnGitHub(&spec, "schema.yml"))
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
	assert.Empty(t, linkToManifest(&nonGitHubSpec, "verify"))
}
