/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertRootCallCloudConfiguration verifies that APIKey and Host configuration are present and at root level in all functions
func assertRootCallCloudConfiguration(t *testing.T, functions []CliFunction) {
	t.Helper()
	for _, fct := range functions {
		var hasAPIKey, hasHost bool
		for _, cfg := range fct.Configuration {
			if cfg.Name == "APIKey" {
				hasAPIKey = true
				assert.True(t, cfg.IsRoot, "APIKey should be root")
			}
			if cfg.Name == "Host" {
				hasHost = true
				assert.True(t, cfg.IsRoot, "Host should be root")
			}

		}
		assert.True(t, hasAPIKey, "should have APIKey config")
		assert.True(t, hasHost, "should have Host config")
	}
}
