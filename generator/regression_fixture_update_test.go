package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/require"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

func updateRegressionFixture(t *testing.T, repoRoot string, fixture struct {
	recipe string
	zip    string
}) {
	t.Helper()

	localFs := filesystem.NewStandardFileSystem()
	outputDir := t.TempDir()
	cfg := config.DefaultSDKGeneratorManagerConfiguration()
	cfg.Input = filesystem.FilePathJoin(localFs, repoRoot, "testdata", fixture.recipe)
	cfg.Output = outputDir
	cfg.ToolVersion = field.ToOptionalString("0.31.0")

	require.NoError(t, Generate(context.Background(), cfg))
	buildGeneratedModules(context.Background(), localFs, outputDir, t)

	zipPath := filepath.Join(repoRoot, "testdata", fixture.zip)
	if localFs.Exists(zipPath) {
		require.NoError(t, localFs.Rm(zipPath))
	}
	require.NoError(t, localFs.ZipWithContext(context.Background(), outputDir, zipPath), fmt.Sprintf("failed updating %s", fixture.zip))
}

func TestUpdateRegressionFixtures(t *testing.T) {
	if os.Getenv("UPDATE_GENERATOR_FIXTURES") != "1" {
		t.Skip("set UPDATE_GENERATOR_FIXTURES=1 to rewrite regression fixtures")
	}

	repoRoot, err := os.Getwd()
	require.NoError(t, err)

	fixtures := []struct {
		recipe string
		zip    string
	}{
		{recipe: "workspace.json", zip: "workspace.zip"},
		{recipe: "verify.yml", zip: "verify.zip"},
		{recipe: "agent.yml", zip: "agent.zip"},
		{recipe: "casing-types.yml", zip: "casing-types.zip"},
		{recipe: "list-iterator.yml", zip: "list-iterator.zip"},
		{recipe: "variadic.yml", zip: "variadic.zip"},
		{recipe: "options-variadic.yml", zip: "options-variadic.zip"},
		{recipe: "petshop-purchase.yml", zip: "petshop-purchase.zip"},
		{recipe: "enum.yml", zip: "enum.zip"},
		{recipe: "packaged-types.yml", zip: "packaged-types.zip"},
		{recipe: "cli-package.yml", zip: "cli-package.zip"},
		{recipe: "result-output.yml", zip: "result-output.zip"},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.recipe, func(t *testing.T) {
			updateRegressionFixture(t, repoRoot, fixture)
		})
	}
}
