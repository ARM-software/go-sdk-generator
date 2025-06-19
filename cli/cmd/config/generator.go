package config

import (
	configUtils "github.com/ARM-software/golang-utils/utils/config"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// GenerateConfiguration defines the configuration for Generate command
type GenerateConfiguration struct {
	// path to tool definition file
	Input string `mapstructure:"input"`
	// destination of the generated code
	Output string `mapstructure:"output"`
	// regenerate all files
	Generate bool `mapstructure:"overwrite"`
	// targeted go version for the tool
	TargetedGo string `mapstructure:"go_version"`

	// ToolVersion contains the version of the CLI tool
	ToolVersion *string
}

func (cfg *GenerateConfiguration) Validate() error {
	err := configUtils.ValidateEmbedded(cfg)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(cfg,
		validation.Field(&cfg.Input, validation.Required),
		validation.Field(&cfg.Output, validation.Required),
		validation.Field(&cfg.Generate),
		validation.Field(&cfg.TargetedGo, is.Semver),
	)
}

func DefaultGenerateConfiguration() *GenerateConfiguration {
	return &GenerateConfiguration{}
}
