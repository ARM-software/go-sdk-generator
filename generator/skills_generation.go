/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"os"
	"path"
	"strings"

	"github.com/ARM-software/go-sdk-generator/generator/options"
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// addAgentFiles selects agent-facing outputs independently of CLI source files.
func (t *templates) addAgentFiles(spec *SDKSpecification, outputDir string, actions []GeneratedAction, overwrite bool) error {
	if spec.CommandLineCommands == nil {
		return nil
	}
	onGitHub := isHostedOnGitHub(spec)
	if err := t.addSkillQualityFiles(outputDir, onGitHub, actions, overwrite); err != nil {
		return err
	}
	localFs := filesystem.GetGlobalFileSystem()
	t.addStateFile(agentFile, filesystem.FilePathJoin(localFs, outputDir, "agent.md"), true, true)
	if onGitHub {
		t.addStateFile(copilotFile, filesystem.FilePathJoin(localFs, outputDir, ".github", "copilot-instructions.md"), true, true)
	}
	binary := field.OptionalString(spec.CommandLineCommands.Binary, "")
	t.addStateFile(cliSkillFile, filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", runSkillName(binary), "SKILL.md"), true, true)
	t.addStateFile(installSkillFile, filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", installSkillName(binary), "SKILL.md"), overwrite, true)
	if len(actions) > 0 {
		t.addStateFile(githubActionSkillFile, filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills", githubActionSkillName(binary), "SKILL.md"), true, true)
	}
	return nil
}

// Skill quality tooling remains applicable when command actions are disabled.
func (t *templates) addSkillQualityFiles(outputDir string, onGitHub bool, actions []GeneratedAction, overwrite bool) error {
	localFs := filesystem.GetGlobalFileSystem()
	t.addStateFile(skillsGuideFile, filesystem.FilePathJoin(localFs, outputDir, "guides", "agent-skills.md"), overwrite, true)
	t.addStateFile(skillsCheckFile, filesystem.FilePathJoin(localFs, outputDir, "actions", "validate-skills", "validate_skills.py"), overwrite, true)
	t.addStateFile(skillsRequirementsFile, filesystem.FilePathJoin(localFs, outputDir, "actions", "validate-skills", "requirements.txt"), overwrite, true)
	if !onGitHub {
		return nil
	}
	t.addStateFile(skillsWorkflowFile, filesystem.FilePathJoin(localFs, outputDir, ".github", "workflows", "skills.yml"), overwrite, true)
	for _, action := range actions {
		if action.DirectoryName == "validate-skills" {
			return commonerrors.Newf(commonerrors.ErrInvalid, "action directory validate-skills is reserved for skill validation")
		}
	}
	t.addStateFile(skillsActionFile, filesystem.FilePathJoin(localFs, outputDir, "actions", "validate-skills", "action.yml"), overwrite, true)
	return nil
}

// copyGoSDKGeneratorSkillsForEmbedding refreshes the generator's build-time
// mirror after rendering its product skills. See guides/embedded-agent-skills.md.
func copyGoSDKGeneratorSkillsForEmbedding(outputDir string, generated []string, overwriteRules *overwrite) ([]string, error) {
	localFs := filesystem.GetGlobalFileSystem()
	skillsRoot := filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills")
	if !localFs.Exists(skillsRoot) {
		return generated, nil
	}
	skillPaths, err := discoverEmbeddedSkillFiles(localFs, skillsRoot)
	if err != nil {
		return generated, err
	}
	if overwriteRules != nil {
		embeddedPaths := collection.Map(skillPaths, func(skillPath string) string {
			return path.Join("generator/skills", skillPath)
		})
		overwriteRules.addGeneratedGroups(map[string][]string{options.OverwriteOptionsExcludeGroupsAgent.String(): embeddedPaths})
	}
	for _, skillPath := range skillPaths {
		embeddedPath := filesystem.FilePathJoin(localFs, append([]string{outputDir, "generator", "skills"}, strings.Split(skillPath, "/")...)...)
		if overwriteRules.shouldExclude(embeddedPath) {
			generated = append(generated, embeddedPath)
			continue
		}
		sourcePath := filesystem.FilePathJoin(localFs, append([]string{skillsRoot}, strings.Split(skillPath, "/")...)...)
		if !localFs.Exists(sourcePath) {
			continue
		}
		if err = copyEmbeddedSkillFile(localFs, sourcePath, embeddedPath); err != nil {
			return generated, err
		}
		generated = append(generated, embeddedPath)
	}
	return generated, nil
}

// Each registered SKILL.md selects a complete directory, including resources.
func discoverEmbeddedSkillFiles(localFs filesystem.FS, skillsRoot string) ([]string, error) {
	var paths []string
	for _, entry := range embeddedAgentSkillPaths {
		root := filesystem.FilePathJoin(localFs, skillsRoot, path.Dir(entry))
		if !localFs.Exists(root) {
			continue
		}
		err := localFs.Walk(root, func(name string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.Mode().IsRegular() {
				relative, relErr := localFs.ConvertToRelativePath(skillsRoot, name)
				if relErr != nil {
					return relErr
				}
				paths = append(paths, filesystem.FilePathToSlash(localFs, relative[0]))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func copyEmbeddedSkillFile(localFs filesystem.FS, sourcePath, embeddedPath string) error {
	skill, err := localFs.ReadFile(sourcePath)
	if err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not read tool-related skill %q", sourcePath)
	}
	embeddedDir := filesystem.FilePathDir(localFs, embeddedPath)
	if !localFs.Exists(embeddedDir) {
		if err = localFs.MkDirAll(embeddedDir, 0755); err != nil {
			return commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", embeddedDir)
		}
	}
	info, err := localFs.Stat(sourcePath)
	if err != nil {
		return err
	}
	if err = localFs.WriteFile(embeddedPath, skill, info.Mode().Perm()); err != nil {
		return commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not write embedded skill %q", embeddedPath)
	}
	return nil
}
