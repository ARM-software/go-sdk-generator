package generator

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedDocumentationTemplate(t *testing.T) {
	content, err := fs.ReadFile(documentationFile)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(content), "documentedAdditionalTypeFieldType"))
}

func TestLowerFirstPreservesLeadingAcronyms(t *testing.T) {
	assert.Equal(t, "document status", lowerFirstDescription("Document status"))
	assert.Equal(t, "AI Portal entities", lowerFirstDescription("AI Portal entities"))
	assert.Equal(t, "URL to the documentation", lowerFirstDescription("URL to the documentation"))
	assert.Equal(t, "HTTP server configuration", lowerFirstDescription("HTTP server configuration"))
	assert.Equal(t, "whether the test configuration is enabled.", lowerFirstDescription("Whether the test configuration is enabled."))
	assert.Equal(t, "published document.", lowerFirstDescription("Published document."))
}

func TestTemplateLowerFirstRendering(t *testing.T) {
	tmpl, err := template.New("test").Funcs(templateFuncMap).Parse(`{{ goComment "" (printf "%s is %s" "Enabled" ("Whether the test configuration is enabled." | lowerFirst)) }}`)
	require.NoError(t, err)

	var b strings.Builder
	require.NoError(t, tmpl.Execute(&b, nil))
	assert.Equal(t, "// Enabled is whether the test configuration is enabled.", b.String())
}

func TestTemplateLowerFirstRenderingFromField(t *testing.T) {
	tmpl, err := template.New("test").Funcs(templateFuncMap).Parse(`{{ goComment "" (printf "%s is %s" .Name (.Description | lowerFirst)) }}`)
	require.NoError(t, err)

	var b strings.Builder
	require.NoError(t, tmpl.Execute(&b, struct {
		Name        string
		Description string
	}{Name: "Enabled", Description: "Whether the test configuration is enabled."}))
	assert.Equal(t, "// Enabled is whether the test configuration is enabled.", b.String())
}

func TestDescriptionWithDeprecation(t *testing.T) {
	notice := &DeprecationNotice{
		IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
		RemovalDate: field.ToOptionalTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Description: field.ToOptionalString("This field is superseded by a nested configuration object"),
		Replacement: field.ToOptionalString("NewField"),
	}

	assert.Equal(t, "Field description.\nDeprecated: Deprecated since 2026-07-24T00:00:00Z. Planned removal date: 2027-01-01T00:00:00Z. Replacement: NewField. This field is superseded by a nested configuration object.", descriptionWithDeprecation("Field description.", notice))
	assert.Equal(t, "use NewField instead; deprecated since 2026-07-24T00:00:00Z; planned removal date 2027-01-01T00:00:00Z; This field is superseded by a nested configuration object", cobraDeprecationMessage(notice))
}

func TestEmbeddedConfigTemplateCommentRendering(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SuppressRequiredValidationMap: map[string]bool{
			validationSuppressionKey(globalConfigurationValidationScope, "", "OutputFormat"): true,
		},
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("PortalEntities"),
				PackageName: "portal-entities-sdk",
				MainModule:  "portalentities",
				GlobalConfiguration: Configuration{{
					Name:            "Enabled",
					LangType:        "bool",
					MapstructureTag: "enabled",
					Description:     "Whether the test configuration is enabled.",
					DeprecationNotice: &DeprecationNotice{
						IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
						Replacement: field.ToOptionalString("FeatureToggle"),
					},
					DefaultValue: field.ToOptionalString("true"),
				}},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.go.tmpl", data))
	assert.Contains(t, b.String(), "// Enabled is whether the test configuration is enabled.")
	assert.Contains(t, b.String(), "// Deprecated: Deprecated since 2026-07-24T00:00:00Z. Replacement: FeatureToggle.")
}

func TestDocumentedAdditionalTypeFieldTypePreservesAcronymCasing(t *testing.T) {
	fieldDef := &EntityField{Name: "EncryptedPayload", LangType: "*encryption.HybridAESRSAEncryptedPayload"}
	assert.Equal(t, "*encryption.HybridAESRSAEncryptedPayload", documentedAdditionalTypeFieldType("AuthenticationPayload", fieldDef))
}

func TestCasingTypesDescriptionsRemainCleanAfterPreprocessing(t *testing.T) {
	repoRoot, err := filepath.Abs(".")
	require.NoError(t, err)
	manifestPath := filepath.Join(repoRoot, "testdata", "casing-types.yml")

	var sdk SDK
	content, err := filesystem.ReadFile(manifestPath)
	require.NoError(t, err)
	require.NoError(t, unmarshalManifestContent(".yml", content, &sdk))
	require.NoError(t, ProcessExtraImports(context.Background(), &sdk))
	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), &sdk))
	require.NoError(t, ProcessAdditionalTypePackages(context.Background(), &sdk))
	require.NoError(t, ProcessBindings(context.Background(), &sdk))
	require.NoError(t, ProcessArguments(context.Background(), &sdk))

	assert.Equal(t, "Whether the test configuration is enabled.", sdk.Specification.GlobalConfiguration[0].Description)
	assert.Equal(t, "Minimal HAL link representation for an affected AI Portal entity.", sdk.Specification.AdditionalTypes[0].Description)
	assert.Equal(t, "AI Portal entities affected by a documentation link update, grouped by resource kind and described as HAL links.", sdk.Specification.AdditionalTypes[1].Description)
}

func TestEmbeddedTypeEnumTemplateUsesTypedFirstConstantOnly(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/types.enums.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		CurrentTypePackage: "results",
		CurrentAdditionalTypes: Types{{
			Name:        "Page",
			Description: "Paged result.",
			Fields: []EntityField{{
				Name:        "Status",
				LangType:    "string",
				Enum:        []string{"draft", "published"},
				Description: "Page status.",
			}},
		}},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "types.enums.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "PageStatusDraft PageStatus = iota // draft")
	assert.Contains(t, output, "PageStatusPublished // published")
	assert.NotContains(t, output, "PageStatusPublished PageStatus")
}

func TestEmbeddedConfigTemplateOmitsRequiredForEnumFields(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SuppressRequiredValidationMap: map[string]bool{
			validationSuppressionKey(globalConfigurationValidationScope, "", "OutputFormat"): true,
		},
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("EnumTest"),
				PackageName: "enum-test-sdk",
				MainModule:  "enumtest",
				GlobalConfiguration: Configuration{{
					Name:            "OutputFormat",
					LangType:        "string",
					Enum:            []string{"json", "yaml"},
					MapstructureTag: "output_format",
					Description:     "Output format.",
					Required:        true,
				}},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "validation.Field(&cfg.OutputFormat, validationUtils.Enum(")
	assert.NotContains(t, output, "validation.Field(&cfg.OutputFormat, validation.Required")
}
