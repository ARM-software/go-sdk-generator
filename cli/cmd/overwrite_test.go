/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package cmd

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/go-sdk-generator/cli/cmd/config"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
)

func overwriteTestViperSession(t *testing.T) *viper.Viper {
	t.Helper()
	v := viper.New()
	for _, binding := range []struct {
		key  string
		flag *pflag.Flag
	}{
		{key: "annotation", flag: rootCmd.PersistentFlags().Lookup(rootAnnotationArg)},
		{key: "logger", flag: rootCmd.PersistentFlags().Lookup(rootLoggerArg)},
		{key: "logger_file", flag: rootCmd.PersistentFlags().Lookup(rootLoggerFileArg)},
		{key: "show_config", flag: rootCmd.PersistentFlags().Lookup(rootShowConfigArg)},
		{key: "input", flag: generateCmd.Flags().Lookup(generateInputArg)},
		{key: "output", flag: generateCmd.Flags().Lookup(generateOutputArg)},
		{key: "go_version", flag: generateCmd.Flags().Lookup(generateTargetedGoArg)},
		{key: "overwrite_enabled", flag: generateCmd.Flags().Lookup(generateOverwriteEnabledArg)},
		{key: "overwrite_exclude_files", flag: generateCmd.Flags().Lookup(generateOverwriteExcludeFilesArg)},
		{key: "overwrite_exclude_groups", flag: generateCmd.Flags().Lookup(generateOverwriteExcludeGroupsArg)},
		{key: "overwrite_only_files", flag: generateCmd.Flags().Lookup(generateOverwriteOnlyFilesArg)},
		{key: "overwrite_only_groups", flag: generateCmd.Flags().Lookup(generateOverwriteOnlyGroupsArg)},
		{key: "summary", flag: generateCmd.Flags().Lookup(generateSummaryArg)},
		{key: "summary_file", flag: generateCmd.Flags().Lookup(generateSummaryFileArg)},
	} {
		require.NoError(t, configUtils.BindFlagToEnv(v, config.App, binding.key, binding.flag))
	}
	return v
}

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
	v := overwriteTestViperSession(t)
	require.NoError(t, configUtils.LoadUsing(&cfg,
		configUtils.WithEnvVarPrefix(config.App),
		configUtils.WithDefaultConfiguration(config.DefaultGenerateConfiguration()),
		configUtils.WithViperSession(v)))
	require.True(t, cfg.Overwrite.Enabled)
	require.Equal(t, []string{"project", "agent", "cli-implementation"}, cfg.Overwrite.ExcludeGroups)
	require.Equal(t, []string{"README.md", ".github/workflows/ci.yml"}, cfg.Overwrite.ExcludeFiles)
	require.Equal(t, []string{"sdk", "cli"}, cfg.Overwrite.OnlyGroups)
	require.Equal(t, []string{"generator/interfaces.go", "cli/cmd/root.go"}, cfg.Overwrite.OnlyFiles)
	require.False(t, flags.Lookup("overwrite").Hidden)
	require.Equal(t, "a", flags.Lookup("overwrite").Shorthand)
	require.NoError(t, cfg.Validate())
}

func TestOverwriteSelectionFlagsImplyOverwriteEnabled(t *testing.T) {
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
	require.NoError(t, flags.Set("input", "generate.yml"))
	require.NoError(t, flags.Set("exclude-groups", "ci,linters"))
	require.NoError(t, flags.Set("only-groups", "sdk"))
	var cfg config.GenerateConfiguration
	v := overwriteTestViperSession(t)
	require.NoError(t, configUtils.LoadUsing(&cfg,
		configUtils.WithEnvVarPrefix(config.App),
		configUtils.WithDefaultConfiguration(config.DefaultGenerateConfiguration()),
		configUtils.WithViperSession(v)))
	require.True(t, cfg.Overwrite.Enabled)
	require.Equal(t, []string{"ci", "linters"}, cfg.Overwrite.ExcludeGroups)
	require.Equal(t, []string{"sdk"}, cfg.Overwrite.OnlyGroups)
	require.NoError(t, cfg.Validate())
}
