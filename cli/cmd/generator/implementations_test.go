/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	richlog "github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/cli/cmd/config"
	rootgenerator "github.com/ARM-software/go-sdk-generator/generator"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/annotations"
	"github.com/ARM-software/golang-utils/utils/logs/summary"
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

		logger, loggerErr := logs.NewPlainStringLogger()
		require.NoError(t, loggerErr)
		annotationLogger, loggerErr := resolveAnnotationLogger(logger, config.GenerateAnnotationNone.String())
		require.NoError(t, loggerErr)

		err = AgentWithAnnotations(context.Background(), annotationLogger, &config.AgentConfiguration{Init: true})
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
		logger, loggerErr := logs.NewPlainStringLogger()
		require.NoError(t, loggerErr)
		annotationLogger, loggerErr := resolveAnnotationLogger(logger, config.GenerateAnnotationNone.String())
		require.NoError(t, loggerErr)

		err := AgentWithAnnotations(context.Background(), annotationLogger, &config.AgentConfiguration{})
		require.Error(t, err)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		logger, loggerErr := logs.NewPlainStringLogger()
		require.NoError(t, loggerErr)
		annotationLogger, loggerErr := resolveAnnotationLogger(logger, config.GenerateAnnotationNone.String())
		require.NoError(t, loggerErr)

		err := AgentWithAnnotations(ctx, annotationLogger, &config.AgentConfiguration{Init: true})
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})
}

func TestGenerateDefaultWrapperUsesSummaryBinding(t *testing.T) {
	oldGenerateImplFunc := GenerateImplFunc
	oldGenerateWithSummaryAndAnnotationsImplFunc := GenerateWithSummaryAndAnnotationsImplFunc
	oldGenerateWithSummaryImplFunc := GenerateWithSummaryImplFunc
	oldCustomLoggerDefinition := CustomLoggerDefinition
	t.Cleanup(func() {
		GenerateImplFunc = oldGenerateImplFunc
		GenerateWithSummaryAndAnnotationsImplFunc = oldGenerateWithSummaryAndAnnotationsImplFunc
		GenerateWithSummaryImplFunc = oldGenerateWithSummaryImplFunc
		CustomLoggerDefinition = oldCustomLoggerDefinition
	})

	withSummaryCalled := false
	GenerateWithSummaryAndAnnotationsImplFunc = nil
	GenerateWithSummaryImplFunc = func(ctx context.Context, logger logs.Loggers, summaryLogger summary.ISummaryLogger, cfg *config.GenerateConfiguration) error {
		withSummaryCalled = true
		require.NotNil(t, ctx)
		require.NotNil(t, logger)
		require.NotNil(t, summaryLogger)
		assert.NotNil(t, cfg)
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)
	err = WrapGenerateWithSummaryAndAnnotations(context.Background(), annotationLogger, config.DefaultGenerateConfiguration())
	require.NoError(t, err)
	assert.True(t, withSummaryCalled)
}

func TestGenerateWrapperFallsBackToPlainImplementationWhenSummaryImplementationIsMissing(t *testing.T) {
	oldGenerateImplFunc := GenerateImplFunc
	oldGenerateWithSummaryAndAnnotationsImplFunc := GenerateWithSummaryAndAnnotationsImplFunc
	oldGenerateWithSummaryImplFunc := GenerateWithSummaryImplFunc
	t.Cleanup(func() {
		GenerateImplFunc = oldGenerateImplFunc
		GenerateWithSummaryAndAnnotationsImplFunc = oldGenerateWithSummaryAndAnnotationsImplFunc
		GenerateWithSummaryImplFunc = oldGenerateWithSummaryImplFunc
	})

	GenerateWithSummaryAndAnnotationsImplFunc = nil
	GenerateWithSummaryImplFunc = nil
	plainCalled := false
	GenerateImplFunc = func(ctx context.Context, logger logs.Loggers, cfg *config.GenerateConfiguration) error {
		plainCalled = true
		require.NotNil(t, ctx)
		require.NotNil(t, logger)
		require.NotNil(t, cfg)
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)

	require.NoError(t, WrapGenerateWithSummaryAndAnnotations(context.Background(), annotationLogger, config.DefaultGenerateConfiguration()))
	assert.True(t, plainCalled)
}

func TestWrapGenerateWithAnnotationsFallsBackToPlainImplementation(t *testing.T) {
	oldGenerateImplFunc := GenerateImplFunc
	t.Cleanup(func() {
		GenerateImplFunc = oldGenerateImplFunc
	})

	plainCalled := false
	GenerateImplFunc = func(ctx context.Context, logger logs.Loggers, cfg *config.GenerateConfiguration) error {
		plainCalled = true
		require.NotNil(t, ctx)
		require.NotNil(t, logger)
		require.NotNil(t, cfg)
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)

	require.NoError(t, WrapGenerateWithAnnotations(context.Background(), annotationLogger, config.DefaultGenerateConfiguration()))
	assert.True(t, plainCalled)
}

func TestWrapGenerateWithSummaryAndAnnotationsPrefersAdvancedImplementation(t *testing.T) {
	oldGenerateWithSummaryAndAnnotationsImplFunc := GenerateWithSummaryAndAnnotationsImplFunc
	oldGenerateWithSummaryImplFunc := GenerateWithSummaryImplFunc
	t.Cleanup(func() {
		GenerateWithSummaryAndAnnotationsImplFunc = oldGenerateWithSummaryAndAnnotationsImplFunc
		GenerateWithSummaryImplFunc = oldGenerateWithSummaryImplFunc
	})

	advancedCalled := false
	simpleCalled := false
	GenerateWithSummaryAndAnnotationsImplFunc = func(ctx context.Context, logger annotations.IAnnotationLogger, summaryLogger summary.ISummaryLogger, cfg *config.GenerateConfiguration) error {
		advancedCalled = true
		require.NotNil(t, logger)
		require.NotNil(t, summaryLogger)
		return nil
	}
	GenerateWithSummaryImplFunc = func(ctx context.Context, logger logs.Loggers, summaryLogger summary.ISummaryLogger, cfg *config.GenerateConfiguration) error {
		simpleCalled = true
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)
	cfg := config.DefaultGenerateConfiguration()
	cfg.ShowConfig = true
	require.NoError(t, WrapGenerateWithSummaryAndAnnotations(context.Background(), annotationLogger, cfg))
	assert.True(t, advancedCalled)
	assert.False(t, simpleCalled)
	assert.Contains(t, baseLogger.GetLogContent(), "command configuration:")
}

func TestWrapSchemaWithAnnotationsDelegatesToSimpleImplementation(t *testing.T) {
	oldSchemaWithAnnotationsImplFunc := SchemaWithAnnotationsImplFunc
	oldSchemaImplFunc := SchemaImplFunc
	t.Cleanup(func() {
		SchemaWithAnnotationsImplFunc = oldSchemaWithAnnotationsImplFunc
		SchemaImplFunc = oldSchemaImplFunc
	})

	SchemaWithAnnotationsImplFunc = nil
	simpleCalled := false
	SchemaImplFunc = func(ctx context.Context, logger logs.Loggers, cfg *config.SchemaConfiguration) error {
		simpleCalled = true
		require.NotNil(t, logger)
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)
	cfg := config.DefaultSchemaConfiguration()
	cfg.ShowConfig = true
	require.NoError(t, WrapSchemaWithAnnotations(context.Background(), annotationLogger, cfg))
	assert.True(t, simpleCalled)
	assert.Contains(t, baseLogger.GetLogContent(), "command configuration:")
}

func TestWrapAgentWithAnnotationsShowsConfigBeforeAdvancedImplementation(t *testing.T) {
	oldAgentWithAnnotationsImplFunc := AgentWithAnnotationsImplFunc
	t.Cleanup(func() {
		AgentWithAnnotationsImplFunc = oldAgentWithAnnotationsImplFunc
	})

	advancedCalled := false
	AgentWithAnnotationsImplFunc = func(ctx context.Context, logger annotations.IAnnotationLogger, cfg *config.AgentConfiguration) error {
		advancedCalled = true
		require.NotNil(t, ctx)
		require.NotNil(t, logger)
		require.NotNil(t, cfg)
		return nil
	}

	baseLogger, err := logs.NewPlainStringLogger()
	require.NoError(t, err)
	annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationNone.String())
	require.NoError(t, err)

	cfg := config.DefaultAgentConfiguration()
	cfg.ShowConfig = true

	require.NoError(t, WrapAgentWithAnnotations(context.Background(), annotationLogger, cfg))
	assert.True(t, advancedCalled)
	assert.Contains(t, baseLogger.GetLogContent(), "command configuration:")
}

func TestDetermineCommandConfigurationEnvironmentVariablesForPrintingStringifiesPointers(t *testing.T) {
	cfg := config.DefaultGenerateConfiguration()
	cfg.ToolVersion = field.ToOptionalString("0.47.0")

	envVars, err := DetermineCommandConfigurationEnvironmentVariablesForPrinting(context.Background(), config.App, cfg)
	require.NoError(t, err)

	toolVersionVar := ""
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar.String(), "SDK_GENERATOR_TOOLVERSION=") {
			toolVersionVar = envVar.String()
			break
		}
	}

	require.NotEmpty(t, toolVersionVar)
	assert.Regexp(t, regexp.MustCompile(`^SDK_GENERATOR_TOOLVERSION=0x[0-9a-fA-F]+ -> 0\.47\.0$`), toolVersionVar)
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

func TestResolveRuntimeLoggerAndAnnotations(t *testing.T) {
	oldCustomLoggerDefinition := CustomLoggerDefinition
	t.Cleanup(func() {
		CustomLoggerDefinition = oldCustomLoggerDefinition
	})

	t.Run("custom logger requires registration", func(t *testing.T) {
		CustomLoggerDefinition = nil
		_, _, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerCustom.String(), "", config.GenerateAnnotationNone.String(), "Generate")
		errortest.AssertError(t, err, commonerrors.ErrNoLogger)
	})

	t.Run("custom logger uses the registered implementation", func(t *testing.T) {
		stringLogger, err := logs.NewPlainStringLogger()
		require.NoError(t, err)
		CustomLoggerDefinition = stringLogger

		logger, annotationLogger, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerCustom.String(), "", config.GenerateAnnotationNone.String(), "Generate")
		require.NoError(t, err)
		require.NotNil(t, annotationLogger)
		logger.Log("hello")
		assert.Contains(t, stringLogger.GetLogContent(), "hello")
	})

	t.Run("github annotations use the resolved logger", func(t *testing.T) {
		stringLogger, err := logs.NewPlainStringLogger()
		require.NoError(t, err)
		CustomLoggerDefinition = stringLogger

		_, annotationLogger, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerCustom.String(), "", config.GenerateAnnotationGithub.String(), "Generate")
		require.NoError(t, err)
		require.NoError(t, annotationLogger.WriteError("broken"))
		assert.Contains(t, stringLogger.GetLogContent(), "::error::broken")
	})

	t.Run("github annotations fall back to stdout with rich logging", func(t *testing.T) {
		var output bytes.Buffer
		baseLogger := &RichLoggerAdapter{
			logger: richlog.New(
				richlog.WithOutput(&output),
				richlog.WithPrefix("Generate"),
				richlog.WithTimestamp(),
			),
		}

		originalStdout := os.Stdout
		reader, writer, err := os.Pipe()
		require.NoError(t, err)
		t.Cleanup(func() {
			os.Stdout = originalStdout
			_ = reader.Close()
			_ = writer.Close()
		})
		os.Stdout = writer

		annotationLogger, err := resolveAnnotationLogger(baseLogger, config.GenerateAnnotationGithub.String())
		require.NoError(t, err)
		require.NoError(t, annotationLogger.WriteError("broken"))
		require.NoError(t, writer.Close())
		stdout, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Contains(t, output.String(), "the chosen logger does not provide support for annotations, stdout will be used for annotations")
		assert.True(t, strings.HasPrefix(strings.TrimSpace(string(stdout)), "::error::broken"), string(stdout))
	})

	t.Run("none annotations fall back to plain logger output", func(t *testing.T) {
		stringLogger, err := logs.NewPlainStringLogger()
		require.NoError(t, err)
		CustomLoggerDefinition = stringLogger

		_, annotationLogger, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerCustom.String(), "", config.GenerateAnnotationNone.String(), "Generate")
		require.NoError(t, err)
		require.NoError(t, annotationLogger.WriteError("ignored"))
		assert.Contains(t, stringLogger.GetLogContent(), "ignored")
	})

	t.Run("rich logger resolves successfully", func(t *testing.T) {
		logger, annotationLogger, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerRich.String(), "", config.GenerateAnnotationNone.String(), "Generate")
		require.NoError(t, err)
		require.NotNil(t, logger)
		require.NotNil(t, annotationLogger)
		require.NoError(t, logger.Check())
	})

	t.Run("file logger writes to the requested path", func(t *testing.T) {
		logPath := filepath.Join(t.TempDir(), "tool.log")
		logger, _, err := ResolveRuntimeLoggerAndAnnotations(config.GenerateLoggerFile.String(), logPath, config.GenerateAnnotationNone.String(), "Generate")
		require.NoError(t, err)
		logger.Log("hello")
		data, readErr := os.ReadFile(logPath)
		require.NoError(t, readErr)
		assert.Contains(t, string(data), "hello")
	})
}
