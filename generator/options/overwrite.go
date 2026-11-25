/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package options

// WithOverwriteEnabled controls overwrite for a single Generate invocation.
// Options are applied in order. Selection helpers enable overwrite themselves;
// disabling it afterwards is invalid if any selectors remain populated.
func WithOverwriteEnabled(enabled bool) OverwriteOption {
	return func(opt *OverwriteOptions) {
		if opt != nil {
			opt.Enabled = enabled
		}
	}
}
