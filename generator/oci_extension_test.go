/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
)

func TestOCIHelpers(t *testing.T) {
	spec := SDKSpecification{
		PackageName:   "verify-cmsis-example-sdk",
		MainModule:    "verify",
		Licence:       "Apache-2.0",
		Description:   field.ToOptionalString("Verify examples from a published container image."),
		RepositoryUrl: field.ToOptionalString("https://github.com/Arm-Examples/verify-cmsis-example-sdk"),
		CommandLineCommands: &CommandLineCommands{
			Binary: field.ToOptionalString("verify-cmsis-example-sdk"),
		},
		DistributeOci: &DistributeOci{
			Description: "OCI image for CI and reproducible verification runs.",
			Labels: OciMapStringString{
				"com.example.label": "true",
			},
			Attributes: OciMapStringString{
				"com.example.attribute": "present",
			},
		},
	}

	assert.True(t, publishesOCIToGHCR(&spec))
	assert.Equal(t, "ghcr.io/arm-examples/verify-cmsis-example-sdk", ghcrImageName(&spec))
	assert.Equal(t, "/etc/verify-cmsis-example-sdk/.agents/skills", toolSkillsInstallPath(&spec))
	assert.Equal(t, "/etc/verify-cmsis-example-sdk/LICENSE", toolLicenceInstallPath(&spec))

	labels := ociLabels(&spec)
	assert.Contains(t, labels, OciMetadata{Name: "com.example.label", Value: "true"})
	assert.Contains(t, labels, OciMetadata{Name: "org.opencontainers.image.source", Value: "https://github.com/Arm-Examples/verify-cmsis-example-sdk"})

	annotations := ociAnnotations(&spec)
	assert.Contains(t, annotations, OciMetadata{Name: "com.example.attribute", Value: "present"})
	assert.Contains(t, annotations, OciMetadata{Name: "org.opencontainers.image.description", Value: "OCI image for CI and reproducible verification runs."})
	assert.Contains(t, annotations, OciMetadata{Name: "org.opencontainers.image.version", Value: "{{ .Version }}"})
}

func TestValidateDistributeOCI(t *testing.T) {
	t.Run("requires CLI", func(t *testing.T) {
		spec := &SDKSpecification{
			PackageName: "example-sdk",
			MainModule:  "example",
			Interface: Interface{Methods: []InterfaceMethod{{
				Name:        "Ping",
				Description: "Ping.",
			}}},
			DistributeOci: &DistributeOci{Description: "OCI image."},
		}

		err := validateDistributeOCI(spec)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
	})

	t.Run("requires GitHub hosting", func(t *testing.T) {
		spec := &SDKSpecification{
			PackageName:   "example-sdk",
			MainModule:    "example",
			RepositoryUrl: field.ToOptionalString("https://gitlab.example.com/arm/example-sdk"),
			CommandLineCommands: &CommandLineCommands{
				Binary: field.ToOptionalString("example-sdk"),
			},
			DistributeOci: &DistributeOci{Description: "OCI image."},
		}

		err := validateDistributeOCI(spec)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
	})

	t.Run("valid GitHub CLI distribution", func(t *testing.T) {
		spec := &SDKSpecification{
			PackageName:   "example-sdk",
			MainModule:    "example",
			RepositoryUrl: field.ToOptionalString("https://github.com/Arm-Examples/example-sdk"),
			CommandLineCommands: &CommandLineCommands{
				Binary: field.ToOptionalString("example-sdk"),
			},
			DistributeOci: &DistributeOci{
				Description: "OCI image.",
				Labels: OciMapStringString{
					"example": "true",
				},
			},
		}

		require.NoError(t, validateDistributeOCI(spec))
	})
}
