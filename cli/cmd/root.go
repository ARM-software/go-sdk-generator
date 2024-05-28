package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Arm-Debug/go-sdk-generator/generator"
	"github.com/Arm-Debug/go-sdk-generator/generator/config"

	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	app       = "sdk_generator"
	inputArg  = "input"
	outputArg = "output"
)

var (
	viperSession   = viper.New()
	generateConfig config.GenerateConfig
)

var rootCmd = &cobra.Command{
	Use: "sdk-generator",
	RunE: func(_ *cobra.Command, _ []string) (err error) {
		ctx := context.Background()

		logger, err := logs.NewStdLogger("SDK Generator")
		if err != nil {
			log.Printf("Could not create logger: %v\n", err.Error())
			return
		}

		err = configUtils.LoadFromViper(viperSession, app, &generateConfig, config.DefaultGenerateConfig())
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to initialise CLI with error: %v\n", err.Error()))
			return
		}

		err = generator.Generate(ctx, &generateConfig)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to generate SDK with error: %v\n", err.Error()))
			return
		}

		return
	},
	SilenceUsage: true, // otherwise 'Usage' is printed after any error
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP(inputArg, "i", "", "path to SDK schema file")
	rootCmd.Flags().StringP(outputArg, "o", "", "destination of the output")

	_ = configUtils.BindFlagToEnv(viperSession, app, "INPUT", rootCmd.Flags().Lookup(inputArg))
	_ = configUtils.BindFlagToEnv(viperSession, app, "OUTPUT", rootCmd.Flags().Lookup(outputArg))
}
