package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Arm-Debug/go-sdk-generator/cli/cmd/config"
	rootgenerator "github.com/Arm-Debug/go-sdk-generator/generator"
)

func TestSchema(t *testing.T) {
	var output bytes.Buffer

	err := rootgenerator.WriteEmbeddedSchemaJSON(&output)
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &schema))
	assert.Equal(t, "SDK", schema["title"])
}

func TestAgent(t *testing.T) {
	t.Run("init installs repo skills", func(t *testing.T) {
		workingDir := t.TempDir()
		localFs := filesystem.NewStandardFileSystem()

		originalDir, err := filesystem.CurrentDirectory()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(workingDir))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalDir))
		})

		err = Agent(context.Background(), nil, &config.AgentConfiguration{Init: true})
		require.NoError(t, err)

		for _, path := range []string{
			filesystem.FilePathJoin(localFs, workingDir, ".agents", "skills", "run-go-sdk-generator", "SKILL.md"),
			filesystem.FilePathJoin(localFs, workingDir, ".agents", "skills", "using-go-sdk-generator-in-a-github-action", "SKILL.md"),
			filesystem.FilePathJoin(localFs, workingDir, ".agents", "skills", "install-go-sdk-generator", "SKILL.md"),
		} {
			data, readErr := filesystem.ReadFile(path)
			require.NoError(t, readErr)
			assert.NotEmpty(t, data)
		}
	})

	t.Run("init flag required", func(t *testing.T) {
		err := Agent(context.Background(), nil, &config.AgentConfiguration{})
		require.Error(t, err)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Agent(ctx, nil, &config.AgentConfiguration{Init: true})
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
}
