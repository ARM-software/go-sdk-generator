/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/logs"
)

type sdkGenerator struct {
	cfg *config.SDKGeneratorManagerConfiguration
}

func (g *sdkGenerator) GetConfig() *config.SDKGeneratorManagerConfiguration {
	return g.cfg
}

func (g *sdkGenerator) Generate(ctx context.Context, logger logs.Loggers) error {
	return Generate(ctx, logger, g.GetConfig())
}

func NewSDKGenerator(cfg *config.SDKGeneratorManagerConfiguration) (g ISDKGeneratorManager, err error) {
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
