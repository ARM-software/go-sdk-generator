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
	cfg := config.DefaultSdkGeneratorManagerConfiguration()
	cfg.Input = filesystem.FilePathJoin(localFs, repoRoot, "testdata", fixture.recipe)
	cfg.Output = outputDir
	cfg.ToolVersion = field.ToOptionalString("0.31.0")

	require.NoError(t, Generate(context.Background(), cfg))

	zipPath := filepath.Join(repoRoot, "testdata", fixture.zip)
	if filesystem.Exists(zipPath) {
		require.NoError(t, filesystem.Rm(zipPath))
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
		{recipe: "enum.yml", zip: "enum.zip"},
		{recipe: "packaged-types.yml", zip: "packaged-types.zip"},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.recipe, func(t *testing.T) {
			updateRegressionFixture(t, repoRoot, fixture)
		})
	}
}
