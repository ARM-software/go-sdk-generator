/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/cli/cmd/config"
	"github.com/ARM-software/go-sdk-generator/generator/options"
)

func TestOverwriteOptionsBridge(t *testing.T) {
	cfg := &config.GenerateConfiguration{
		Overwrite: options.OverwriteOptions{
			Enabled: true, ExcludeGroups: []string{"project", "agent"}, ExcludeFiles: []string{"README.md", ".github/workflows/ci.yml"},
			OnlyGroups: []string{"sdk"}, OnlyFiles: []string{"cli/cmd/root.go"},
		},
	}
	opts := options.WithOverwriteOptions(overwriteOptions(cfg)...)
	require.NoError(t, opts.Validate())
	require.True(t, opts.Enabled)
	require.Equal(t, cfg.Overwrite, *opts)
	cfg.Overwrite.Enabled = false
	opts = options.WithOverwriteOptions(overwriteOptions(cfg)...)
	require.Equal(t, cfg.Overwrite, *opts)
	require.Error(t, opts.Validate(), "disabled selectors must reach SDK validation rather than being discarded")

	cfg.Overwrite = options.OverwriteOptions{}
	opts = options.WithOverwriteOptions(overwriteOptions(cfg)...)
	require.False(t, opts.Enabled)
	require.NoError(t, opts.Validate())

	cfg.Overwrite.Enabled = true
	opts = options.WithOverwriteOptions(overwriteOptions(cfg)...)
	require.True(t, opts.Enabled, "plain --overwrite must work without selectors")
	require.NoError(t, opts.Validate())
}
