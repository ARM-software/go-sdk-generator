/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateOverwriteSelectionValidation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		overwrite     bool
		groups, files []string
		onlyGroups, onlyFiles []string
		invalid       bool
	}{
		{name: "no exclusions"},
		{name: "groups require overwrite", groups: []string{"project"}, invalid: true},
		{name: "files require overwrite", files: []string{"README.md"}, invalid: true},
		{name: "all groups", overwrite: true, groups: []string{"ci", "linters", "gomod", "cli", "readme", "project", "agent", "sdk", "cli-implementation"}},
		{name: "sdk requires overwrite", groups: []string{"sdk"}, invalid: true},
		{name: "cli implementation requires overwrite", groups: []string{"cli-implementation"}, invalid: true},
		{name: "each group validated", overwrite: true, groups: []string{"project", "unknown"}, invalid: true},
		{name: "both exclusions", overwrite: true, groups: []string{"agent"}, files: []string{"README.md", ".github/workflows/ci.yml"}},
		{name: "only files require overwrite", onlyFiles: []string{"README.md"}, invalid: true},
		{name: "only groups require overwrite", onlyGroups: []string{"sdk"}, invalid: true},
		{name: "all only groups", overwrite: true, onlyGroups: []string{"ci", "linters", "gomod", "cli", "readme", "project", "agent", "sdk", "cli-implementation"}},
		{name: "each only group validated", overwrite: true, onlyGroups: []string{"sdk", "unknown"}, invalid: true},
		{name: "both only selectors", overwrite: true, onlyGroups: []string{"sdk"}, onlyFiles: []string{"README.md"}},
		{name: "only and exclude may overlap", overwrite: true, groups: []string{"sdk"}, files: []string{"README.md"}, onlyGroups: []string{"sdk"}, onlyFiles: []string{"README.md"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validGenerateConfiguration()
			cfg.Overwrite.Enabled, cfg.Overwrite.ExcludeGroups = tc.overwrite, tc.groups
			cfg.Overwrite.ExcludeFiles = tc.files
			cfg.Overwrite.OnlyGroups, cfg.Overwrite.OnlyFiles = tc.onlyGroups, tc.onlyFiles
			if tc.invalid {
				require.Error(t, cfg.Validate())
			} else {
				require.NoError(t, cfg.Validate())
				require.Equal(t, tc.overwrite, cfg.Overwrite.Enabled)
			}
		})
	}
}
