/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// SkillInfo is read from the rendered frontmatter, keeping documentation and
// distribution tied to the same metadata as the installed skill.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"-"`
}

// AdditionalSkillCatalogue documents repository-maintained skills without
// promoting them into the product's distribution set.
func (d *DataModel) AdditionalSkillCatalogue() ([]SkillInfo, error) {
	if d.OutputDirectory == "" || d.Specification.CommandLineCommands == nil {
		return nil, nil
	}
	localFs := filesystem.GetGlobalFileSystem()
	root := filesystem.FilePathJoin(localFs, d.OutputDirectory, ".agents", "skills")
	if !localFs.Exists(root) {
		return nil, nil
	}
	product, err := d.SkillCatalogue()
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(product))
	for _, skill := range product {
		names[skill.Name] = true
	}
	files, err := localFs.LsRecursive(context.Background(), root, false)
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	var additional []SkillInfo
	for _, file := range files {
		if filesystem.FilePathBase(localFs, file) != "SKILL.md" {
			continue
		}
		// Product templates may be regenerated later in this pass. Their old
		// on-disk contents are not additional repository-maintained metadata.
		if names[filesystem.FilePathBase(localFs, filesystem.FilePathDir(localFs, file))] {
			continue
		}
		content, readErr := localFs.ReadFile(file)
		if readErr != nil {
			return nil, readErr
		}
		parts := strings.SplitN(strings.ReplaceAll(string(content), "\r\n", "\n"), "---\n", 3)
		if len(parts) != 3 || parts[0] != "" {
			return nil, fmt.Errorf("%s must start with YAML frontmatter", file)
		}
		var skill SkillInfo
		if err = yaml.Unmarshal([]byte(parts[1]), &skill); err != nil {
			return nil, err
		}
		if names[skill.Name] {
			continue
		}
		relative, relErr := localFs.ConvertToRelativePath(d.OutputDirectory, filesystem.FilePathDir(localFs, file))
		if relErr != nil {
			return nil, relErr
		}
		skill.Path = filesystem.FilePathToSlash(localFs, relative[0])
		skill.Description = strings.ReplaceAll(strings.Join(strings.Fields(skill.Description), " "), "|", "&#124;")
		additional = append(additional, skill)
	}
	return additional, nil
}

var skillNameSeparators = regexp.MustCompile(`[^a-z0-9]+`)

func portableSkillName(name string) string {
	normalised := strings.Trim(skillNameSeparators.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(normalised) <= 64 {
		return normalised
	}
	return fmt.Sprintf("%s-%x", strings.TrimRight(normalised[:55], "-"), sha256.Sum256([]byte(name)))[:64]
}

func skillDescription(kind, binary string) string {
	var description string
	switch kind {
	case "install":
		description = fmt.Sprintf("Install, upgrade, or troubleshoot the %s CLI. Use when setting up the tool or checking its installed version.", binary)
	case "action":
		description = fmt.Sprintf("Use %s in GitHub Actions. Use when writing workflows that install a release, build from source, or invoke the tool's generated actions.", binary)
	default:
		description = fmt.Sprintf("Run and configure the %s CLI. Use when selecting commands, supplying arguments, configuring authentication, or interpreting tool output.", binary)
	}
	return string([]rune(description)[:min(len([]rune(description)), 1024)])
}

// SkillCatalogue renders only skill templates; it never renders documentation.
func (d *DataModel) SkillCatalogue() ([]SkillInfo, error) {
	if d.Specification.CommandLineCommands == nil {
		return nil, nil
	}
	files := []string{installSkillFile, cliSkillFile}
	if d.InstallAction != nil {
		files = append(files, githubActionSkillFile)
	}
	var skills []SkillInfo
	for _, file := range files {
		t, err := template.New("skill").Funcs(templateFuncMap).ParseFS(fs, file)
		if err != nil {
			return nil, err
		}
		var rendered bytes.Buffer
		if err = t.ExecuteTemplate(&rendered, path.Base(file), d); err != nil {
			return nil, err
		}
		parts := strings.SplitN(strings.ReplaceAll(rendered.String(), "\r\n", "\n"), "---\n", 3)
		if len(parts) != 3 || parts[0] != "" {
			return nil, fmt.Errorf("%s must start with YAML frontmatter", file)
		}
		var skill SkillInfo
		if err = yaml.UnmarshalStrict([]byte(parts[1]), &map[string]any{}); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if err = yaml.Unmarshal([]byte(parts[1]), &skill); err != nil {
			return nil, err
		}
		skill.Path = path.Join(".agents/skills", skill.Name)
		// Descriptions are single-line table cells in generated documentation.
		skill.Description = strings.ReplaceAll(strings.Join(strings.Fields(skill.Description), " "), "|", "&#124;")
		skills = append(skills, skill)
	}
	return skills, nil
}
