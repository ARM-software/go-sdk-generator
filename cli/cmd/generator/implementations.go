// Package generator contains the implementation for all the CLI commands.
// Importing the package also performs the automatic binding between the definition and the implementation.
package generator

import (
	"context"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"

	"github.com/Arm-Debug/go-sdk-generator/cli/cmd/config"
	"github.com/Arm-Debug/go-sdk-generator/generator"
	generatorConfig "github.com/Arm-Debug/go-sdk-generator/generator/config"
)

/*
Generate function: Generate tool
*/
func Generate(ctx context.Context, logger logs.Loggers, cfg *config.GenerateConfiguration) (err error) {
	if cfg == nil {
		err = commonerrors.UndefinedVariable("configuration")
		return
	}
	gCfg := generatorConfig.DefaultSdkGeneratorManagerConfiguration()
	gCfg.Input = cfg.Input
	gCfg.Output = cfg.Output
	gCfg.Overwrite = cfg.Generate
	gCfg.TargetedGoVersion = cfg.TargetedGo
	g, err := generator.NewSdkGenerator(gCfg)
	if err != nil {
		return
	}
	err = generator.NewSdkGeneratorWithDelegate(g).Generate(ctx)
	if err != nil {
		logger.LogError(fmt.Sprintf("Failed to generate SDK with error: %v\n", err.Error()))
	}
	return
}

func init() {
	// Note: The following makes the link between the CLI command definitions and their actual implementations.
	// The binding is performed at runtime when package generator is imported.
	GenerateImplFunc = Generate
}
