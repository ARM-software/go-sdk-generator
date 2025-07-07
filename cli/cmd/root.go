package cmd

import (
	"context"
	"embed"
	"os"

	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/semver"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	app             = "SDK_GENERATOR"
	rootGenerateArg = "overwrite"
	rootInputArg    = "input"
	rootOutputArg   = "output"
)

var (
	//go:embed module.properties
	embeddedFS   embed.FS
	viperSession = viper.New()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-sdk-generator",
	Short: "CLI for generating SDK/CLI go tools from a specification following [schema](https://github.com/Arm-Debug/go-sdk-generator/blob/main/schema.yml)",
	Long:  "CLI for generating SDK/CLI go tools from a specification following [schema](https://github.com/Arm-Debug/go-sdk-generator/blob/main/schema.yml)",
}

func Execute() {
	ctx := context.Background()

	cliVersion, err := semver.DetermineEmbeddedVersion(ctx, &embeddedFS, nil)
	if err != nil {
		// This is for completeness but it is impossible at runtime since we need a valid module.properties in this location to release the package
		cliVersion = "unknown"
	}

	if err := fang.Execute(ctx, rootCmd, fang.WithVersion(cliVersion)); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool(rootGenerateArg, false, "regenerate all files")
	rootCmd.PersistentFlags().String(rootInputArg, "", "path to tool definition file")
	rootCmd.PersistentFlags().String(rootOutputArg, "", "destination of the generated code")

	_ = configUtils.BindFlagToEnv(viperSession, app, "input", rootCmd.PersistentFlags().Lookup(rootInputArg))
	_ = configUtils.BindFlagToEnv(viperSession, app, "output", rootCmd.PersistentFlags().Lookup(rootOutputArg))
	_ = configUtils.BindFlagToEnv(viperSession, app, "overwrite", rootCmd.PersistentFlags().Lookup(rootGenerateArg))
}
