package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/summary"
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

func TestGenerateDefaultWrapperUsesSummaryBinding(t *testing.T) {
	oldGenerateImplFunc := GenerateImplFunc
	oldGenerateWithSummaryImplFunc := GenerateWithSummaryImplFunc
	t.Cleanup(func() {
		GenerateImplFunc = oldGenerateImplFunc
		GenerateWithSummaryImplFunc = oldGenerateWithSummaryImplFunc
	})

	withSummaryCalled := false
	GenerateWithSummaryImplFunc = func(ctx context.Context, logger logs.Loggers, summaryLogger summary.ISummaryLogger, cfg *config.GenerateConfiguration) error {
		withSummaryCalled = true
		require.NotNil(t, ctx)
		assert.Nil(t, logger)
		require.NotNil(t, summaryLogger)
		assert.NotNil(t, cfg)
		return nil
	}

	err := WrapGenerate(context.Background(), nil, config.DefaultGenerateConfiguration(), BaseGenerateImpl)
	require.NoError(t, err)
	assert.True(t, withSummaryCalled)
}

func TestSummaryLoggerFactory(t *testing.T) {
	t.Run("defaults to noop summary logger", func(t *testing.T) {
		logger := SummaryLoggerFactory("", "", nil)
		require.NotNil(t, logger)
		assert.NoError(t, logger.WriteString("ignored"))
	})

	t.Run("uses file logger when possible", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "summary.md")
		logger := SummaryLoggerFactory(config.GenerateSummaryFile, path, nil)
		require.NotNil(t, logger)
		require.NoError(t, logger.WriteString("hello"))
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(data), "hello")
	})

	t.Run("uses provided logger for output summaries", func(t *testing.T) {
		baseLogger, err := summary.NewInMemorySummaryLogger("test")
		require.NoError(t, err)
		logger := SummaryLoggerFactory(config.GenerateSummaryOutput, "", baseLogger)
		require.NotNil(t, logger)
		require.NoError(t, logger.WriteString("hello"))
		assert.Contains(t, baseLogger.GetSummary(), "hello")
	})
}
