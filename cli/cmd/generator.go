// Package cmd defines all the CLI commands for SDK Generator
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/semver"
	"github.com/spf13/cobra"

	"github.com/Arm-Debug/go-sdk-generator/cli/cmd/config"
	"github.com/Arm-Debug/go-sdk-generator/cli/cmd/generator"
)

const (
	generateTargetedGoArg = "go-version"
)

func addInterruptHandlerToContext(ctx context.Context) (cancellableCtx context.Context, cancel context.CancelFunc) {
	cancellableCtx, cancel = context.WithCancel(ctx)
	end := make(chan error, 1)
	go func() {
		// Waits for a user interrupt or SIGTERM or expired/cancelled context
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		for wait := true; wait; {
			select {
			case <-shutdown:
				wait = false
			case <-ctx.Done():
				wait = false
			}
		}

		cancel()
		close(end)
	}()
	return
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate tool",
	Long:  "Generate tool",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx := context.Background()
		logger, err := logs.NewStdLogger("Generate")
		if err != nil {
			err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed initialising the logger")
			return
		}

		var generateCfg config.GenerateConfiguration

		err = configUtils.LoadFromViper(viperSession, app, &generateCfg, config.DefaultGenerateConfiguration())
		if err != nil {
			return
		}

		cancellableCtx, cancel := addInterruptHandlerToContext(ctx)
		defer cancel()

		cliVersion, err := semver.DetermineEmbeddedVersion(cancellableCtx, &embeddedFS, nil)
		if err == nil {
			generateCfg.ToolVersion = field.ToOptionalString(cliVersion)
		}

		err = generator.WrapGenerate(cancellableCtx, logger, &generateCfg, generator.GenerateImplFunc)
		if err != nil {
			cancel()
			return
		}

		return
	},
}

func init() {

	generateCmd.Flags().String(generateTargetedGoArg, "", "targeted go version for the tool")
	_ = configUtils.BindFlagToEnv(viperSession, app, "go_version", generateCmd.Flags().Lookup(generateTargetedGoArg))

	rootCmd.AddCommand(generateCmd)

}
