/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/cli/cmd/config"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
)

func TestOverwriteFlagsLoadLists(t *testing.T) {
	flags := generateCmd.Flags()
	flags.VisitAll(func(flag *pflag.Flag) {
		changed, value := flag.Changed, flag.Value.String()
		if slice, ok := flag.Value.(pflag.SliceValue); ok {
			values := append([]string{}, slice.GetSlice()...)
			t.Cleanup(func() { require.NoError(t, slice.Replace(values)); flag.Changed = changed })
		} else {
			t.Cleanup(func() { require.NoError(t, flag.Value.Set(value)); flag.Changed = changed })
		}
	})
	require.Equal(t, "stringSlice", flags.Lookup("exclude-groups").Value.Type())
	require.Equal(t, "stringSlice", flags.Lookup("exclude-files").Value.Type())
	require.Equal(t, "stringSlice", flags.Lookup("only-groups").Value.Type())
	require.Equal(t, "stringSlice", flags.Lookup("only-files").Value.Type())
	require.NoError(t, flags.Parse([]string{"--input=generate.yml", "--overwrite", "--exclude-groups=project,agent,cli-implementation", "--exclude-files=README.md", "--exclude-files=.github/workflows/ci.yml", "--only-groups=sdk,cli", "--only-files=generator/interfaces.go", "--only-files=cli/cmd/root.go"}))
	var cfg config.GenerateConfiguration
	require.NoError(t, configUtils.LoadUsing(&cfg,
		configUtils.WithEnvVarPrefix(config.App),
		configUtils.WithDefaultConfiguration(config.DefaultGenerateConfiguration()),
		configUtils.WithViperSession(viperSession)))
	require.True(t, cfg.Overwrite.Enabled)
	require.Equal(t, []string{"project", "agent", "cli-implementation"}, cfg.Overwrite.ExcludeGroups)
	require.Equal(t, []string{"README.md", ".github/workflows/ci.yml"}, cfg.Overwrite.ExcludeFiles)
	require.Equal(t, []string{"sdk", "cli"}, cfg.Overwrite.OnlyGroups)
	require.Equal(t, []string{"generator/interfaces.go", "cli/cmd/root.go"}, cfg.Overwrite.OnlyFiles)
	require.False(t, flags.Lookup("overwrite").Hidden)
	require.Equal(t, "a", flags.Lookup("overwrite").Shorthand)
	require.NoError(t, cfg.Validate())
}
