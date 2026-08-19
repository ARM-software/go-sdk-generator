/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// normaliseKeyNames reduces manifest-defined names to a stable comparison key
// so generated-name uniqueness checks can detect duplicates irrespective of casing.
func normaliseKeyNames(name string) string {
	trimmed := strings.TrimSpace(name)
	if reflection.IsEmpty(trimmed) {
		return ""
	}
	return strings.ToLower(toPascalCase(trimmed))
}

func duplicateCaseInsensitiveNames(names []string) []string {
	seen := make(map[string]string, len(names))
	duplicateKeys := mapset.NewSet[string]()
	collection.ForEach(names, func(name string) {
		key := normaliseKeyNames(name)
		if reflection.IsEmpty(key) {
			return
		}
		if _, found := seen[key]; found {
			duplicateKeys.Add(key)
			return
		}
		seen[key] = strings.TrimSpace(name)
	})
	result := collection.Map(duplicateKeys.ToSlice(), func(key string) string {
		return seen[key]
	})
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i]) < strings.ToLower(result[j])
	})
	return result
}

func validateUniqueCaseInsensitiveNames(scope, concept string, names []string) error {
	duplicates := duplicateCaseInsensitiveNames(names)
	if reflection.IsEmpty(duplicates) {
		return nil
	}
	return commonerrors.Newf(commonerrors.ErrConflict, "%s defines duplicate %s names [%s]; names must be unique irrespective of casing", scope, concept, strings.Join(duplicates, ", "))
}

func isBoolean(langType string) bool {
	return strings.EqualFold(strings.TrimSpace(langType), "bool")
}

func usesFilesystemValidationAlias(langType string) bool {
	switch strings.TrimSpace(langType) {
	case "filepath", "existingFilepath":
		return true
	default:
		return false
	}
}

func usesJSONValidationAlias(langType string) bool {
	return strings.TrimSpace(langType) == "validJSON"
}

func effectiveParameterEnum(enum []string, constraints *ParameterConstraints) []string {
	if hasEnum(enum) {
		return enum
	}
	if constraints == nil {
		return nil
	}
	return constraints.Enum
}

func effectiveParameterPattern(pattern *string, constraints *ParameterConstraints) *string {
	if reflection.IsNotEmpty(pattern) {
		return pattern
	}
	if constraints == nil || reflection.IsEmpty(constraints.Pattern) {
		return nil
	}
	return constraints.Pattern
}

func effectiveParameterRequired(required bool, constraints *ParameterConstraints) bool {
	return required || (constraints != nil && constraints.Required)
}

func usesFilesystemConstraintRule(constraints *ParameterConstraints) bool {
	return constraints != nil && reflection.IsNotEmpty(constraints.Extension)
}

func usesFmtConstraintRule(constraints *ParameterConstraints) bool {
	return constraints != nil && constraints.UniqueItems
}
