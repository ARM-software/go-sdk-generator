/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package options

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverwriteOptionsValidateSelection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options OverwriteOptions
		invalid bool
	}{
		{name: "disabled empty"},
		{name: "disabled groups", options: OverwriteOptions{ExcludeGroups: []string{"project"}}, invalid: true},
		{name: "disabled files", options: OverwriteOptions{ExcludeFiles: []string{"README.md"}}, invalid: true},
		{name: "enabled empty", options: OverwriteOptions{Enabled: true}},
		{name: "enabled exclusions", options: OverwriteOptions{Enabled: true, ExcludeGroups: []string{"project", "agent"}, ExcludeFiles: []string{"README.md"}}},
		{name: "invalid group item", options: OverwriteOptions{Enabled: true, ExcludeGroups: []string{"project", "unknown"}}, invalid: true},
		{name: "disabled only groups", options: OverwriteOptions{OnlyGroups: []string{"sdk"}}, invalid: true},
		{name: "disabled only files", options: OverwriteOptions{OnlyFiles: []string{"README.md"}}, invalid: true},
		{name: "enabled only selection", options: OverwriteOptions{Enabled: true, OnlyGroups: []string{"sdk"}, OnlyFiles: []string{"README.md"}}},
		{name: "invalid only group item", options: OverwriteOptions{Enabled: true, OnlyGroups: []string{"sdk", "unknown"}}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.invalid {
				require.Error(t, tc.options.Validate())
			} else {
				require.NoError(t, tc.options.Validate())
			}
		})
	}
	opts := WithOverwriteOptions(WithExcludeGroups([]string{OverwriteOptionsExcludeGroupsProject.String()}), WithExcludeFiles([]string{"README.md"}))
	require.True(t, opts.Enabled)
	require.NoError(t, opts.Validate())
	require.NoError(t, (*OverwriteOptions)(nil).Validate())
}

func TestOverwriteGroupParsing(t *testing.T) {
	group, err := OverwriteOptionsExcludeGroupsString("project")
	require.NoError(t, err)
	require.Equal(t, OverwriteOptionsExcludeGroupsProject, group)
	require.Equal(t, "project", group.String())
	_, err = OverwriteOptionsExcludeGroupsString("unknown")
	require.Error(t, err)
	var decoded OverwriteOptionsExcludeGroups
	require.NoError(t, decoded.UnmarshalJSON([]byte(`"agent"`)))
	require.Equal(t, OverwriteOptionsExcludeGroupsAgent, decoded)
	require.NoError(t, decoded.UnmarshalText([]byte("ci")))
	require.Equal(t, OverwriteOptionsExcludeGroupsCI, decoded)
}

func TestWithOverwriteEnabled(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		opt := WithOverwriteOptions(WithOverwriteEnabled(enabled))
		require.Equal(t, enabled, opt.Enabled)
		require.NoError(t, opt.Validate())
	}
	require.NotPanics(t, func() { WithOverwriteEnabled(true)(nil) })
}
