package config

import (
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	validation "github.com/go-ozzo/ozzo-validation"
)

type GenerateConfig struct {
	Input  string `mapstructure:"input"`
	Output string `mapstructure:"output"`
}

func DefaultGenerateConfig() *GenerateConfig {
	return &GenerateConfig{
		Input:  "",
		Output: "",
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
	)
}
