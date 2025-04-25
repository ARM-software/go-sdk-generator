package config

import (
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type GenerateConfig struct {
	Input             string `mapstructure:"input"`
	Output            string `mapstructure:"output"`
	Overwrite         bool   `mapstructure:"overwrite"`
	TargetedGoVersion string `mapstructure:"go_version"`
}

func DefaultGenerateConfig() *GenerateConfig {
	return &GenerateConfig{
		Input:             "",
		Output:            "",
		Overwrite:         false,
		TargetedGoVersion: "1.24.0",
	}
}

func (cfg *GenerateConfig) Validate() error {
	err := configUtils.ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Input, validation.Required),
		validation.Field(&cfg.Output, validation.Required),
		validation.Field(&cfg.TargetedGoVersion, validation.Required, is.Semver),
	)
}
