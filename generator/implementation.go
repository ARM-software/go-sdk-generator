package generator

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

type sdkGenerator struct {
	cfg *config.SdkGeneratorManagerConfiguration
}

func (g *sdkGenerator) GetConfig() *config.SdkGeneratorManagerConfiguration {
	return g.cfg
}

func (g *sdkGenerator) Generate(ctx context.Context) error {
	return Generate(ctx, g.GetConfig())
}

func NewSdkGenerator(cfg *config.SdkGeneratorManagerConfiguration) (g ISdkGeneratorManager, err error) {
	if cfg == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}
	err = cfg.Validate()
	if err != nil {
		return
	}
	g = &sdkGenerator{cfg: cfg}
	return
}
