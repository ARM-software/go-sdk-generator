package generator

import (
	"context"
	"strings"
	"testing"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormaliseOptionPackageName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, defaultOptionPackage, normaliseOptionPackageName(""))
	assert.Equal(t, "clientoptions", normaliseOptionPackageName("Client-Options"))
	assert.Equal(t, "runtime", normaliseOptionPackageName("Runtime"))
}

func TestValidateOptionPackageName(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateOptionPackageName("options"))
	require.Error(t, validateOptionPackageName("foo/bar"))
	require.Error(t, validateOptionPackageName("foo\\bar"))
	require.Error(t, validateOptionPackageName("foo.bar"))
}

func TestProcessOptionPackagesNormalisesPackage(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{Options: OptionSets{{
		Name:        "Build",
		Package:     "Client-Options",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Enabled", LangType: "bool", Description: "Enabled."}},
		}},
	}}}}

	require.NoError(t, ProcessOptionPackages(context.Background(), sdk))
	assert.Equal(t, "clientoptions", sdk.Specification.Options[0].Package)
}

func TestProcessOptionPackagesRejectsGeneratedHelperCollisions(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{Options: OptionSets{{
		Name:        "Build",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Timeout", LangType: "string", Description: "Timeout."}},
		}},
	}, {
		Name:        "Run",
		Description: "Run options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{Name: "Timeout", LangType: "string", Description: "Timeout."}},
		}},
	}}}}

	err := ProcessOptionPackages(context.Background(), sdk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `generated option identifier "WithTimeout"`)
}

func TestProcessOptionPackagesRejectsPackageCollisionWithAdditionalTypes(t *testing.T) {
	t.Parallel()
	sdk := &SDK{Specification: SDKSpecification{
		AdditionalTypes: Types{{
			Name:        "BuildMetadata",
			Package:     "options",
			Description: "Build metadata.",
			Fields:      []EntityField{{Name: "ID", LangType: "string", Description: "ID."}},
		}},
		Options: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{
				Parameters: []OptionParameter{{Name: "Enabled", LangType: "bool", Description: "Enabled."}},
			}},
		}},
	}}

	err := ProcessOptionPackages(context.Background(), sdk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `option package "options" conflicts with an additional type package`)
}

func TestEmbeddedOptionsTemplateRendersFunctionalOptions(t *testing.T) {
	t.Parallel()
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/options.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		CurrentOptionPackage: "options",
		CurrentOptionSets: OptionSets{{
			Name:        "Build",
			Description: "Build options.",
			Functions: []OptionFunction{{
				FunctionName: field.ToOptionalString("EnableTracing"),
				Parameters:   []OptionParameter{{Name: "Tracing", LangType: "bool", Description: "Tracing.", DefaultValue: field.ToOptionalString("true")}},
			}, {
				Parameters: []OptionParameter{{
					Name:        "Tags",
					LangType:    "string",
					List:        true,
					Description: "Tags.",
				}},
			}, {
				Parameters: []OptionParameter{{
					Name:         "Mode",
					LangType:     "string",
					Enum:         []string{"fast", "safe"},
					DefaultValue: field.ToOptionalString("fast"),
					Description:  "Mode.",
				}},
			}, {
				Parameters: []OptionParameter{{
					Name:         "Verbose",
					LangType:     "bool",
					Description:  "Verbose.",
					DefaultValue: field.ToOptionalString("false"),
				}},
			}, {
				FunctionName: field.ToOptionalString("WithWindow"),
				Parameters: []OptionParameter{{
					Name:         "Width",
					LangType:     "int",
					Description:  "Width.",
					DefaultValue: field.ToOptionalString("80"),
				}, {
					Name:         "Height",
					LangType:     "int",
					Description:  "Height.",
					DefaultValue: field.ToOptionalString("24"),
				}},
			}},
		}},
		SDK: SDK{Specification: SDKSpecification{PackageName: "example-sdk", MainModule: "example"}},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "options.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "type BuildOptions struct {")
	assert.Contains(t, output, "func DefaultBuildOptions() (opt *BuildOptions)")
	assert.Contains(t, output, "func (opt *BuildOptions) Validate() (err error)")
	assert.Contains(t, output, "func (opt *BuildOptions) Clone() (cloned *BuildOptions)")
	assert.Contains(t, output, "clone.Tags = slices.Clone(opt.Tags)")
	assert.Contains(t, output, "type BuildOption func(*BuildOptions)")
	assert.Contains(t, output, "func WithBuildOptions(options ...BuildOption) (opt *BuildOptions)")
	assert.Contains(t, output, "collection.ForEach(options, func(option BuildOption) {")
	assert.Contains(t, output, "func EnableTracing(tracing bool) BuildOption")
	assert.Contains(t, output, "Tracing: true,")
	assert.NotContains(t, output, "func WithoutTracing() BuildOption")
	assert.Contains(t, output, "func WithVerbose() BuildOption")
	assert.Contains(t, output, "func WithoutVerbose() BuildOption")
	assert.Contains(t, output, "Verbose: false,")
	assert.Contains(t, output, "func WithMode(value BuildOptionsMode) BuildOption")
	assert.Contains(t, output, "Mode: BuildOptionsModeFast,")
	assert.Contains(t, output, "func WithWindow(width int, height int) BuildOption")
	assert.Contains(t, output, "opt.Width = width")
	assert.Contains(t, output, "opt.Height = height")
	assert.Contains(t, output, "Width: 80,")
	assert.Contains(t, output, "Height: 24,")
	assert.NotContains(t, output, "func WithTracing() BuildOption")
}

func TestEmbeddedOptionEnumsTemplateUsesTypedFirstConstantOnly(t *testing.T) {
	t.Parallel()
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/options.enums.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{CurrentOptionPackage: "options", CurrentOptionSets: OptionSets{{
		Name:        "Build",
		Description: "Build options.",
		Functions: []OptionFunction{{
			Parameters: []OptionParameter{{
				Name:        "Mode",
				LangType:    "string",
				Enum:        []string{"fast", "safe"},
				Description: "Mode.",
			}},
		}},
	}}}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "options.enums.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "BuildOptionsModeFast BuildOptionsMode = iota // fast")
	assert.Contains(t, output, "BuildOptionsModeSafe // safe")
	assert.NotContains(t, output, "BuildOptionsModeSafe BuildOptionsMode")
}

func TestOptionFunctionValidateRequiresFunctionNameForMultipleParameters(t *testing.T) {
	t.Parallel()
	optionFunction := &OptionFunction{Parameters: []OptionParameter{{
		Name:        "Width",
		LangType:    "int",
		Description: "Width.",
	}, {
		Name:        "Height",
		LangType:    "int",
		Description: "Height.",
	}}}

	err := optionFunction.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function_name")
}
