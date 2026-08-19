/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import "fmt"

const (
	globalConfigurationValidationScope = "global-configuration"
	cliConfigurationValidationScope    = "cli-configuration"
	additionalTypeValidationScope      = "additional-type"
)

func validationSuppressionKey(scope, ownerName, fieldName string) string {
	return fmt.Sprintf("%s|%s|%s", scope, ownerName, fieldName)
}

func suppressesRequiredValidation(suppressed map[string]bool, scope, ownerName, fieldName string) bool {
	if len(suppressed) == 0 {
		return false
	}
	return suppressed[validationSuppressionKey(scope, ownerName, fieldName)]
}
