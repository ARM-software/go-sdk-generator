/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"go/ast"
	"go/parser"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestPatternValidationRuleProducesValidGo(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{pattern: `^(fast|safe)$`, expected: ", validationUtils.Pattern(regexp.MustCompile(`^(fast|safe)$`))"},
		{pattern: `^[a-z]+\.json$`, expected: ", validationUtils.Pattern(regexp.MustCompile(`^[a-z]+\\.json$`))"},
		{pattern: "^[^;&|$`]*$", expected: `, validationUtils.Pattern(regexp.MustCompile("^[^;&|$` + "`" + `]*$"))`},
		{pattern: "^\\d+`\\w+$", expected: ", validationUtils.Pattern(regexp.MustCompile(" + strconv.Quote("^\d+`\\w+$") + "))"},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.pattern, func(t *testing.T) {
			rule := patternValidationRule(field.ToOptionalString(test.pattern), false, false)
			assert.Equal(t, test.expected, rule)

			expr, err := parser.ParseExpr(strings.TrimPrefix(rule, ", "))
			require.NoError(t, err)
			call := expr.(*ast.CallExpr).Args[0].(*ast.CallExpr)
			literal, ok := call.Args[0].(*ast.BasicLit)
			require.True(t, ok)
			value, err := strconv.Unquote(literal.Value)
			require.NoError(t, err)
			assert.Equal(t, test.pattern, value)
		})
	}
}
