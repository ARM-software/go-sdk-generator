/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestEmbeddedAgentSkills(t *testing.T) {
	skills, err := EmbeddedAgentSkills()
	require.NoError(t, err)

	assert.Len(t, skills, 3)
	assert.Contains(t, skills, "run-go-sdk-generator/SKILL.md")
	assert.Contains(t, skills, "using-go-sdk-generator-in-a-github-action/SKILL.md")
	assert.Contains(t, skills, "install-go-sdk-generator/SKILL.md")
	assert.Contains(t, string(skills["run-go-sdk-generator/SKILL.md"]), "name: run-go-sdk-generator")
}

func TestInitialiseAgentSkills(t *testing.T) {
	outputDir := t.TempDir()
	localFs := filesystem.NewStandardFileSystem()

	err := InitialiseAgentSkills(context.Background(), outputDir)
	require.NoError(t, err)

	for _, path := range []string{
		"run-go-sdk-generator/SKILL.md",
		"using-go-sdk-generator-in-a-github-action/SKILL.md",
		"install-go-sdk-generator/SKILL.md",
	} {
		data, readErr := filesystem.ReadFile(filesystem.FilePathJoin(localFs, append([]string{outputDir, ".agents", "skills"}, strings.Split(path, "/")...)...))
		require.NoError(t, readErr)
		assert.NotEmpty(t, data)
	}

	data, err := filesystem.ReadFile(filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", "run-go-sdk-generator", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: run-go-sdk-generator")
}

func TestInitialiseAgentSkillsInCurrentDirectory(t *testing.T) {
	workingDir := t.TempDir()
	localFs := filesystem.NewStandardFileSystem()

	originalDir, err := filesystem.CurrentDirectory()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workingDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	err = InitialiseAgentSkillsInCurrentDirectory(context.Background())
	require.NoError(t, err)

	data, err := filesystem.ReadFile(filesystem.FilePathJoin(localFs, workingDir, ".agents", "skills", "run-go-sdk-generator", "SKILL.md"))
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestCopyGoSDKGeneratorSkillsForEmbedding(t *testing.T) {
	outputDir := t.TempDir()
	localFs := filesystem.NewStandardFileSystem()

	for _, skillPath := range []string{
		filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", "run-go-sdk-generator", "SKILL.md"),
		filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", "install-go-sdk-generator", "SKILL.md"),
	} {
		err := filesystem.MkDir(filesystem.FilePathDir(localFs, skillPath))
		require.NoError(t, err)
		err = filesystem.WriteFile(skillPath, []byte(filesystem.FilePathBase(localFs, filesystem.FilePathDir(localFs, skillPath))), 0644)
		require.NoError(t, err)
	}

	generated, err := copyGoSDKGeneratorSkillsForEmbedding(outputDir, nil)
	require.NoError(t, err)
	assert.Contains(t, generated, filesystem.FilePathJoin(localFs, outputDir, "generator", "skills", "run-go-sdk-generator", "SKILL.md"))

	embeddedPath := filesystem.FilePathJoin(localFs, outputDir, "generator", "skills", "run-go-sdk-generator", "SKILL.md")
	embeddedData, err := filesystem.ReadFile(embeddedPath)
	require.NoError(t, err)
	assert.Equal(t, "run-go-sdk-generator", string(embeddedData))

	sourcePath := filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", "run-go-sdk-generator", "SKILL.md")
	sourceData, err := filesystem.ReadFile(sourcePath)
	require.NoError(t, err)
	assert.Equal(t, "run-go-sdk-generator", string(sourceData))
}

func TestInitialiseAgentSkillsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := InitialiseAgentSkills(ctx, t.TempDir())
	require.Error(t, err)
}

func TestInitialiseAgentSkillsInCurrentDirectoryCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := InitialiseAgentSkillsInCurrentDirectory(ctx)
	require.Error(t, err)
}
