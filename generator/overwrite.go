/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/ARM-software/go-sdk-generator/generator/options"
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// overwrite belongs to one Generate invocation, never to the manager.
type overwrite struct {
	output   string
	enabled  bool
	excluded overwriteSelection
	only     overwriteSelection
}

// overwriteSelection combines exact basenames, output-relative paths and groups.
// specified records a requested selection even when its groups have no outputs.
type overwriteSelection struct {
	basenames mapset.Set[string]
	paths     mapset.Set[string]
	groups    []string
	specified bool
}

func newOverwrite(output string, enabled bool, opts ...options.OverwriteOption) (*overwrite, error) {
	values := options.WithOverwriteOptions(opts...)
	if err := values.Validate(); err != nil {
		return nil, err
	}
	excluded, err := newOverwriteSelection(values.ExcludeGroups, values.ExcludeFiles)
	if err != nil {
		return nil, err
	}
	only, err := newOverwriteSelection(values.OnlyGroups, values.OnlyFiles)
	if err != nil {
		return nil, err
	}
	return &overwrite{output: output, enabled: enabled, excluded: excluded, only: only}, nil
}

func newOverwriteSelection(groups, files []string) (overwriteSelection, error) {
	selection := overwriteSelection{
		basenames: mapset.NewSet[string](), paths: mapset.NewSet[string](),
		groups: groups, specified: len(groups) > 0 || len(files) > 0,
	}
	err := collection.ForAll(files, func(entry string) error {
		normalised, basename, err := normaliseOverwriteFile(entry)
		if err != nil {
			return err
		}
		if basename {
			selection.basenames.Add(normalised)
		} else {
			selection.paths.Add(normalised)
		}
		return nil
	})
	return selection, err
}

func normaliseOverwriteFile(entry string) (normalised string, basename bool, err error) {
	localFs := filesystem.GetGlobalFileSystem()
	// Accept Windows separators even when the current filesystem uses slashes.
	normalised = strings.ReplaceAll(entry, `\`, "/")
	// Check Windows paths on every OS as well as POSIX paths. Colons also
	// reject drive-relative paths and NTFS alternate data streams.
	if normalised == "" || strings.HasPrefix(normalised, "/") || strings.ContainsAny(normalised, ":\x00\r\n") || strings.HasSuffix(normalised, "/") {
		return "", false, commonerrors.Newf(commonerrors.ErrInvalid, "unsafe file selection %q: expected a basename or output-relative file path", entry)
	}
	basename = !strings.Contains(normalised, "/")
	normalised = filesystem.FilePathFromSlash(localFs, normalised)
	normalised = filesystem.FilePathToSlash(localFs, filesystem.FilePathClean(localFs, normalised))
	if normalised == "." || normalised == ".." || strings.HasPrefix(normalised, "../") {
		return "", false, commonerrors.Newf(commonerrors.ErrInvalid, "file selection %q escapes the generated SDK or does not name a file", entry)
	}
	return normalised, basename, nil
}

func (m *overwrite) shouldExclude(output string) bool {
	if m == nil || !m.enabled || (!m.excluded.specified && !m.only.specified) {
		return false
	}
	localFs := filesystem.GetGlobalFileSystem()
	relative, err := filesystem.FilePathRel(localFs, m.output, output)
	if err != nil {
		return m.only.specified
	}
	relative = filesystem.FilePathClean(localFs, relative)
	if filesystem.FilePathIsAbs(localFs, relative) {
		return m.only.specified
	}
	basename := filesystem.FilePathBase(localFs, relative)
	relative = filesystem.FilePathToSlash(localFs, relative)
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return m.only.specified
	}
	return m.excluded.contains(relative, basename) || (m.only.specified && !m.only.contains(relative, basename))
}

func (m *overwrite) shouldInclude(output string) bool {
	return !m.shouldExclude(output)
}

func (m *overwrite) addGeneratedGroups(groups map[string][]string) {
	m.excluded.addGeneratedGroups(groups)
	m.only.addGeneratedGroups(groups)
}

func (s *overwriteSelection) contains(relative, basename string) bool {
	return s.paths.Contains(relative) || s.basenames.Contains(basename)
}

// addGeneratedGroups adds exact output paths without changing whether a
// selection was requested; an unmatched only-group must still restrict output.
func (s *overwriteSelection) addGeneratedGroups(groups map[string][]string) {
	collection.ForEach(s.groups, func(group string) {
		s.paths.Append(groups[group]...)
	})
}
