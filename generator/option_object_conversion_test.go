/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
)

func TestOptionObjectInterfaceArguments(t *testing.T) {
	for _, tc := range []struct {
		name      string
		langType  string
		list      bool
		iterator  bool
		reference bool
		trailing  bool
		wantType  string
		wantCall  string
	}{
		{name: "object", langType: "build.BuildOptions", wantType: "...build.BuildOption", wantCall: "settings..."},
		{name: "list", langType: "build.BuildOptions", list: true, wantType: "...build.BuildOption", wantCall: "settings..."},
		{name: "pointer object", langType: "*build.BuildOptions", wantType: "...build.BuildOption", wantCall: "settings..."},
		{name: "reference object", langType: "build.BuildOptions", reference: true, wantType: "...build.BuildOption", wantCall: "settings..."},
		{name: "iterator", langType: "build.BuildOptions", iterator: true, wantType: "iter.Seq[build.BuildOption]", wantCall: "settings"},
		{name: "non-final object", langType: "build.BuildOptions", trailing: true, wantType: "[]build.BuildOption", wantCall: "settings, tail"},
		{name: "existing functional options", langType: "build.BuildOption", list: true, wantType: "...build.BuildOption", wantCall: "settings..."},
		{name: "unrelated object", langType: "other.BuildOptions", wantType: "other.BuildOptions", wantCall: "settings"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &DataModel{SDK: SDK{Specification: SDKSpecification{
				Name: field.ToOptionalString("Example"), PackageName: "example-sdk", MainModule: "example", Description: field.ToOptionalString("Example."),
				Options: OptionSets{{Name: "Build", Package: "build", Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: "Tags", LangType: "string", List: true}}}}}},
				Interface: Interface{Methods: []InterfaceMethod{{Name: "Run", Args: []CommandArgument{{
					Name: "Settings", LangType: tc.langType, List: tc.list, Iterator: tc.iterator, Reference: tc.reference,
				}}, Returns: []CommandArgument{{Name: "err", LangType: "error"}}}}},
			}}}
			if tc.trailing {
				data.Specification.Interface.Methods[0].Args = append(data.Specification.Interface.Methods[0].Args, CommandArgument{Name: "Tail", LangType: "string"})
			}
			require.NoError(t, ProcessArguments(t.Context(), &data.SDK))
			// Reprocessing must not add another list layer.
			require.NoError(t, ProcessArguments(t.Context(), &data.SDK))
			for _, file := range []string{interfaceFile, implementationFile} {
				tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, file)
				require.NoError(t, err)
				var output strings.Builder
				require.NoError(t, tmpl.ExecuteTemplate(&output, strings.TrimPrefix(file, "templates/"), data))
				assert.Contains(t, output.String(), "settings "+tc.wantType)
				if file == implementationFile {
					assert.Contains(t, output.String(), "imp.Delegate.Run(ctx, "+tc.wantCall+")")
				}
			}
		})
	}
}

func TestOptionObjectCLIStorage(t *testing.T) {
	for _, mode := range []OptionSetEnableMode{OptionSetEnableModeIndependent, OptionSetEnableModeTogether} {
		t.Run(string(mode), func(t *testing.T) {
			sdk := &SDK{Specification: SDKSpecification{
				Options: OptionSets{{Name: "Build", EnableMode: mode, Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: "Tracing", LangType: "bool"}}}}}},
				CommandLineCommands: &CommandLineCommands{NoActions: true, Functions: []CliFunction{{
					Name: "Run", WithSummary: field.ToOptionalBool(false), Configuration: []FunctionConfigFields{{
						Name: "Build", LangType: "options.BuildOptions", List: true, MapstructureTag: "build", Description: "Build settings.",
					}},
				}}},
			}}
			require.NoError(t, ProcessArguments(t.Context(), sdk))
			for _, cfg := range sdk.Specification.CommandLineCommands.Functions[0].Configuration {
				if cfg.Name == "Build" {
					assert.Equal(t, "options.BuildOptions", cliConfigFieldType(cfg))
					require.NotNil(t, cfg.DefaultValue)
					assert.Equal(t, "*options.DefaultBuildOptions()", *cfg.DefaultValue)
					return
				}
			}
			t.Fatal("Build configuration field is missing")
		})
	}
}

func TestOptionObjectConversions(t *testing.T) {
	for _, mode := range []OptionSetEnableMode{OptionSetEnableModeIndependent, OptionSetEnableModeTogether} {
		t.Run(string(mode), func(t *testing.T) {
			runOptionSetValidationFixture(t, OptionSet{
				Name: "Build", EnableMode: mode,
				Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: "Tags", LangType: "string", List: true}}}},
			}, nil, "object-conversions")
		})
	}
}

func TestOptionObjectConversionNamesAreReserved(t *testing.T) {
	for _, name := range []string{optionAsListMethodName, optionAsIteratorMethodName, "as_list", "as_iterator"} {
		optionSet := OptionSet{Name: "Build", Functions: []OptionFunction{{Parameters: []OptionParameter{{Name: name, LangType: "string"}}}}}
		require.Error(t, optionSet.ValidateOptionParameterNames(), name)
	}
}

func TestOptionObjectRejectsAmbiguousShape(t *testing.T) {
	spec := SDKSpecification{
		Options: OptionSets{{Name: "Build"}},
		Interface: Interface{Methods: []InterfaceMethod{{Args: []CommandArgument{{
			Name: "Settings", LangType: "options.BuildOptions", List: true, Iterator: true,
		}}}}},
	}
	require.Error(t, processOptionObjectShapes(&spec))
	spec.Interface.Methods = nil
	spec.CommandLineCommands = &CommandLineCommands{Functions: []CliFunction{{Configuration: []FunctionConfigFields{{
		Name: "Settings", LangType: "options.BuildOptions", List: true, Iterator: true,
	}}}}}
	require.Error(t, processOptionObjectShapes(&spec))
}
