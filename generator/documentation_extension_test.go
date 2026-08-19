/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
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
	repoRoot, err := filesystem.CurrentDirectory()
	require.NoError(t, err)
	manifestPath := filesystem.FilePathJoin(filesystem.NewStandardFileSystem(), repoRoot, "testdata", "casing-types.yml")

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
					DefaultValue:    field.ToOptionalString("json"),
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

func TestConfigTemplateRetainsPatternOnlyValidation(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.go.tmpl")
	require.NoError(t, err)

	localFs := filesystem.NewStandardFileSystem()
	data, err := filesystem.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "verify.yml"))
	require.NoError(t, err)

	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", data, &sdk))
	require.NoError(t, ProcessExtraImports(context.Background(), &sdk))
	require.NoError(t, ProcessRepositoryAndModuleMetadata(context.Background(), &sdk))
	require.NoError(t, ProcessAdditionalTypePackages(context.Background(), &sdk))
	require.NoError(t, ProcessBindings(context.Background(), &sdk))
	require.NoError(t, ProcessArguments(context.Background(), &sdk))
	require.NoError(t, ProcessConstraintMirrors(context.Background(), &sdk))

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.go.tmpl", &DataModel{SDK: sdk}))
	assert.Contains(t, b.String(), `validation.Field(&cfg.UserAgent, validationUtils.Pattern(regexp.MustCompile(`)
}

func TestConfigTemplateIncludesConstraintValidationAliases(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SuppressRequiredValidationMap: map[string]bool{},
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("ConstraintAliasTest"),
				PackageName: "constraint-alias-test",
				MainModule:  "constraintalias",
				GlobalConfiguration: Configuration{{
					Name:            "Format",
					LangType:        "string",
					MapstructureTag: "format",
					Description:     "Format.",
					Constraints: &ParameterConstraints{
						Enum:        []string{"json", "yaml"},
						Pattern:     field.ToOptionalString(`^(json|yaml)$`),
						Required:    true,
						Nullable:    true,
						NotEmpty:    true,
						Contains:    field.ToOptionalString("so"),
						NotContains: field.ToOptionalString("tmp"),
						Prefix:      field.ToOptionalString("j"),
						Suffix:      field.ToOptionalString("n"),
						MinLength:   field.ToOptional(4),
					},
				}, {
					Name:            "Tags",
					LangType:        "string",
					List:            true,
					MapstructureTag: "tags",
					Description:     "Tags.",
					Constraints: &ParameterConstraints{
						MinItems:    field.ToOptional(1),
						UniqueItems: true,
					},
				}, {
					Name:            "ReportFile",
					LangType:        "string",
					MapstructureTag: "report_file",
					Description:     "Report file.",
					Constraints: &ParameterConstraints{
						Extension: []string{".json", ".yaml"},
						Suffix:    field.ToOptionalString(".json"),
					},
				}},
			},
		},
	}
	require.NoError(t, ProcessConstraintMirrors(context.Background(), &data.SDK))

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "validation.Field(&cfg.Format, validationUtils.Required")
	assert.Contains(t, output, "validationUtils.Enum(")
	assert.Contains(t, output, "validationUtils.Pattern(regexp.MustCompile(`^(json|yaml)$`))")
	assert.Contains(t, output, "validationUtils.Nullable(validationUtils.AllOf(")
	assert.Contains(t, output, "validationUtils.NotEmpty()")
	assert.Contains(t, output, `validationUtils.ContainsString("so")`)
	assert.Contains(t, output, `validationUtils.NotContains("tmp")`)
	assert.Contains(t, output, `validationUtils.Prefix("j")`)
	assert.Contains(t, output, `validationUtils.Suffix("n")`)
	assert.Contains(t, output, "validationUtils.MinLength(4)")
	assert.Contains(t, output, "validationUtils.MinItems(1)")
	assert.Contains(t, output, `validationUtils.UniqueItems[any, string]`)
	assert.Contains(t, output, `filesystem.NewOSPathExtensionRule(true, ".json", ".yaml")`)
}

func TestCLIConfigTemplateIncludesSummaryRuntimeValidationRules(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.cli.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SuppressRequiredValidationMap: map[string]bool{},
		SDK: SDK{
			Specification: SDKSpecification{
				PackageName: "summary-test-sdk",
				MainModule:  "summarytest",
				ExtraImports: ExtraImports{
					{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
					{ImportRef: field.ToOptionalString("validationUtils"), ImportUrl: field.ToOptionalString(validationUtilsImportPath)},
				},
				CommandLineCommands: &CommandLineCommands{
					RootEnvironmentVar: field.ToOptionalString("SUMMARY_TEST"),
					Functions: []CliFunction{{
						Name:        "Generate",
						Description: "Generate output",
						WithSummary: field.ToOptionalBool(true),
						Configuration: []FunctionConfigFields{
							{
								Name:            "Mode",
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString("fast"),
								Enum:            []string{"fast", "slow"},
								MapstructureTag: "mode",
								Description:     "mode",
							},
							{
								Name:            loggerArgument,
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString(defaultCLILoggerMode),
								Enum:            []string{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeCustom.String()},
								MapstructureTag: loggerMapstructureTag,
								Description:     "Select the implementation used for normal runtime logs.",
							},
							{
								Name:            loggerFileArg,
								LangType:        string(BuiltInLangTypeFilepath),
								MapstructureTag: loggerFileMapstructureTag,
								Description:     "Path to write normal runtime logs when logger is set to file.",
								Constraints: &ParameterConstraints{
									RequiredWhen: &ParameterValueConstraint{
										Operator:  ParameterValueConstraintOperatorEquals,
										Parameter: loggerArgument,
										Value:     runtimeLoggerModeFile.String(),
									},
									ForbiddenWhen: &ParameterValueConstraint{
										Operator:  ParameterValueConstraintOperatorNotEquals,
										Parameter: loggerArgument,
										Value:     runtimeLoggerModeFile.String(),
									},
								},
							},
							{
								Name:            annotationArgument,
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString(defaultCLIAnnotationMode),
								Enum:            []string{runtimeAnnotationModeGithub.String(), runtimeAnnotationModeNone.String()},
								MapstructureTag: syntheticCLIMapstructureTagAnnotation.String(),
								Description:     "Select the implementation used for structured annotations such as errors, warnings, and notices.",
							},
							{
								Name:            summaryArgument,
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString(SummaryDestinationOutput.String()),
								Enum:            []string{SummaryDestinationFile.String(), SummaryDestinationGithub.String(), SummaryDestinationOutput.String()},
								MapstructureTag: summaryMapstructureTag,
								Description:     "summary destination",
							},
							{
								Name:            summaryFileArg,
								LangType:        "filepath",
								MapstructureTag: summaryFileMapstructureTag,
								Description:     "summary file",
								Constraints: &ParameterConstraints{
									DependentRequired: []string{summaryArgument},
									RequiredWhen: &ParameterValueConstraint{
										Operator:  ParameterValueConstraintOperatorEquals,
										Parameter: summaryArgument,
										Value:     SummaryDestinationFile.String(),
									},
								},
							},
						},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.cli.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "Mode string `mapstructure:\"mode\"`")
	assert.Contains(t, output, "func (cfg *GenerateConfiguration) GetMode() (GenerateMode, error) {")
	assert.Contains(t, output, "value, err := GenerateModeString(cfg.Mode)")
	assert.Contains(t, output, "return 0, commonerrors.UndefinedVariable(\"Mode\")")
	assert.Contains(t, output, "validation.When(reflection.IsNotEmpty(cfg.Mode), validationUtils.Enum(GenerateModeFast.String(), GenerateModeSlow.String()))")
	assert.Contains(t, output, "Mode: GenerateModeFast.String(),")
	assert.Contains(t, output, "validation.Field(&cfg.SummaryFile, validation.When(reflection.IsNotEmpty(cfg.SummaryFile), filesystem.NewOSPathValidationRule(true)))")
	assert.Contains(t, output, "validation.Field(&cfg.Logger, validation.When(reflection.IsNotEmpty(cfg.Logger), validationUtils.Enum(GenerateLoggerConsole.String(), GenerateLoggerRich.String(), GenerateLoggerFile.String(), GenerateLoggerNone.String(), GenerateLoggerCustom.String())))")
	assert.Contains(t, output, "validation.Field(&cfg.Annotation, validation.When(reflection.IsNotEmpty(cfg.Annotation), validationUtils.Enum(GenerateAnnotationGithub.String(), GenerateAnnotationNone.String())))")
	assert.Contains(t, output, "validationUtils.WhenFieldEqualsValue(&cfg.Logger, GenerateLoggerFile.String(), validationUtils.RequiredFieldsBy(&cfg.LoggerFile))")
	assert.Contains(t, output, "validationUtils.WhenFieldNotEqualsValue(&cfg.Logger, GenerateLoggerFile.String(), validationUtils.ForbiddenFieldsBy(&cfg.LoggerFile))")
	assert.NotContains(t, output, "func validateGenerateSummaryConfiguration(cfg *GenerateConfiguration) error {")
	assert.Contains(t, output, "err = validationUtils.NewAllRule(")
	assert.Contains(t, output, "validationUtils.FieldDependencyBy(map[any][]any{&cfg.SummaryFile: []any{&cfg.Summary}})")
	assert.Contains(t, output, "validation.When(reflection.IsNotEmpty(cfg.Summary), validationUtils.Enum(GenerateSummaryFile.String(), GenerateSummaryGithub.String(), GenerateSummaryOutput.String()))")
	assert.Contains(t, output, "validationUtils.WhenFieldEqualsValue(&cfg.Summary, GenerateSummaryFile.String(), validationUtils.RequiredFieldsBy(&cfg.SummaryFile))")
	assert.NotContains(t, output, "validationUtils.WhenFieldEqualsValue(&cfg.Summary, \"file\", validationUtils.RequiredFieldsBy(&cfg.SummaryFile))")
	assert.Contains(t, output, "Logger: GenerateLoggerConsole.String(),")
	assert.Contains(t, output, "Annotation: GenerateAnnotationNone.String(),")
	assert.Contains(t, output, "Summary: GenerateSummaryOutput.String(),")
	assert.Contains(t, output, "func (cfg *GenerateConfiguration) GetSummary() (GenerateSummary, error) {")
	assert.Contains(t, output, "value, err := GenerateSummaryString(cfg.Summary)")
	assert.Equal(t, 1, strings.Count(output, `github.com/ARM-software/golang-utils/utils/filesystem`))
	assert.Less(t, strings.Index(output, "err := configUtils.ValidateEmbedded(cfg)"), strings.Index(output, "err = validation.ValidateStruct(cfg,"))
}

func TestCLIConfigTemplateIncludesResultOutputRuntimeValidationRules(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/config.cli.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SuppressRequiredValidationMap: map[string]bool{},
		SDK: SDK{
			Specification: SDKSpecification{
				PackageName: "result-output-sdk",
				MainModule:  "resultoutput",
				ExtraImports: ExtraImports{
					{ImportRef: field.ToOptionalString("filesystem"), ImportUrl: field.ToOptionalString(filesystemImportPath)},
					{ImportRef: field.ToOptionalString("validationUtils"), ImportUrl: field.ToOptionalString(validationUtilsImportPath)},
				},
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Inspect",
						Description: "Emit command results to the selected destination.",
						WithResult:  true,
						WithSummary: field.ToOptionalBool(false),
						Configuration: []FunctionConfigFields{
							{
								Name:            resultOutputArgument,
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString(resultOutputModeHuman.String()),
								Enum:            []string{resultOutputModeHuman.String(), resultOutputModeJSON.String(), resultOutputModeYAML.String()},
								MapstructureTag: resultOutputMapstructureTag,
								Description:     "output format",
							},
							{
								Name:            resultOutputFileArgument,
								LangType:        string(BuiltInLangTypeFilepath),
								MapstructureTag: resultOutputFileMapstructureTag,
								Description:     "Write machine-readable result output to the selected destination. Use `-` to write to standard output explicitly. This is only valid when output format is json or yaml. When a file path is used, the destination file is created or truncated immediately before writing the result, so an existing file is reset even if the later write fails.",
								Constraints: &ParameterConstraints{
									DependentRequired: []string{resultOutputArgument},
									ForbiddenWhen: &ParameterValueConstraint{
										Operator:  ParameterValueConstraintOperatorEquals,
										Parameter: resultOutputArgument,
										Value:     resultOutputModeHuman.String(),
									},
									ParameterValue: &ParameterValueComparisonConstraint{
										Parameter: loggerFileArg,
										Operator:  ParameterValueComparisonConstraintOperatorNotEquals,
										When:      ParameterValueComparisonConstraintWhenNotEmpty,
									},
								},
							},
							{
								Name:            loggerArgument,
								LangType:        stringLangType,
								DefaultValue:    field.ToOptionalString(defaultCLILoggerMode),
								Enum:            []string{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeCustom.String()},
								MapstructureTag: loggerMapstructureTag,
								Description:     "logger",
								Constraints: &ParameterConstraints{
									RequiredValuesWhen: []ParameterRequiredValuesConstraint{
										{
											Parameter:      resultOutputFileArgument,
											Operator:       ParameterRequiredValuesConstraintOperatorAbsent,
											RequiredValues: []any{runtimeLoggerModeNone.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeCustom.String()},
										},
										{
											Parameter:      resultOutputFileArgument,
											Operator:       ParameterRequiredValuesConstraintOperatorEquals,
											Value:          stdoutArgumentValue,
											RequiredValues: []any{runtimeLoggerModeNone.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeCustom.String()},
										},
									},
								},
							},
							{
								Name:            loggerFileArg,
								LangType:        string(BuiltInLangTypeFilepath),
								MapstructureTag: loggerFileMapstructureTag,
								Description:     "logger file",
							},
						},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "config.cli.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "func (cfg *InspectConfiguration) GetOutputFormat() (InspectOutputFormat, error) {")
	assert.Contains(t, output, "OutputFormat: InspectOutputFormatHuman.String(),")
	assert.Contains(t, output, "validation.When(reflection.IsNotEmpty(cfg.OutputFormat), validationUtils.Enum(InspectOutputFormatHuman.String(), InspectOutputFormatJSON.String(), InspectOutputFormatYAML.String()))")
	assert.Contains(t, output, "validation.Field(&cfg.OutputFile, validation.When(reflection.IsNotEmpty(cfg.OutputFile), filesystem.NewOSPathValidationRule(true)))")
	assert.Contains(t, output, "destination file is created or truncated immediately before writing the result")
	assert.Contains(t, output, "validationUtils.FieldDependencyBy(map[any][]any{&cfg.OutputFile: []any{&cfg.OutputFormat}})")
	assert.Contains(t, output, "validationUtils.WhenFieldEqualsValue(&cfg.OutputFormat, InspectOutputFormatHuman.String(), validationUtils.ForbiddenFieldsBy(&cfg.OutputFile))")
	assert.NotContains(t, output, "validationUtils.WhenFieldEqualsValue(&cfg.OutputFormat, \"human\", validationUtils.ForbiddenFieldsBy(&cfg.OutputFile))")
	assert.Contains(t, output, `validationUtils.WhenFieldAbsent(&cfg.OutputFile, validation.By(func(value any) error { return validation.Validate(cfg.Logger, validationUtils.Enum(`)
	assert.Contains(t, output, `validationUtils.WhenFieldEqualsValue(&cfg.OutputFile, "-", validation.By(func(value any) error { return validation.Validate(cfg.Logger, validationUtils.Enum(`)
	assert.Contains(t, output, `validationUtils.WhenFieldPresent(&cfg.OutputFile, validation.By(func(value any) error {`)
	assert.Contains(t, output, `if !reflect.DeepEqual(any(cfg.OutputFile), any(cfg.LoggerFile)) {`)
	assert.Contains(t, output, `parameter \"OutputFile\" must satisfy the configured value comparison against \"LoggerFile\"`)
}

func TestWrapperTemplateIncludesResultOutputHelpers(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/wrapper.cli.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				PackageName: "result-output-sdk",
				MainModule:  "resultoutput",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Inspect",
						Description: "Emit command results to the selected destination.",
						WithResult:  true,
						WithSummary: field.ToOptionalBool(false),
						Configuration: []FunctionConfigFields{
							{Name: resultOutputArgument, LangType: stringLangType, MapstructureTag: resultOutputMapstructureTag, Enum: []string{resultOutputModeHuman.String(), resultOutputModeJSON.String(), resultOutputModeYAML.String()}, DefaultValue: field.ToOptionalString(resultOutputModeHuman.String()), Description: "output format"},
							{Name: resultOutputFileArgument, LangType: string(BuiltInLangTypeFilepath), MapstructureTag: resultOutputFileMapstructureTag, Description: "Write machine-readable result output to the selected destination. Use `-` to write to standard output explicitly. This is only valid when output format is json or yaml. When a file path is used, the destination file is created or truncated immediately before writing the result, so an existing file is reset even if the later write fails.", Constraints: &ParameterConstraints{ParameterValue: &ParameterValueComparisonConstraint{Parameter: loggerFileArg, Operator: ParameterValueComparisonConstraintOperatorNotEquals, When: ParameterValueComparisonConstraintWhenNotEmpty}}},
							{Name: loggerArgument, LangType: stringLangType, MapstructureTag: loggerMapstructureTag, Enum: []string{runtimeLoggerModeConsole.String(), runtimeLoggerModeRich.String(), runtimeLoggerModeFile.String(), runtimeLoggerModeNone.String(), runtimeLoggerModeCustom.String()}, DefaultValue: field.ToOptionalString(defaultCLILoggerMode), Description: "logger"},
							{Name: loggerFileArg, LangType: string(BuiltInLangTypeFilepath), MapstructureTag: loggerFileMapstructureTag, Description: "logger file"},
							{Name: annotationArgument, LangType: stringLangType, MapstructureTag: syntheticCLIMapstructureTagAnnotation.String(), Enum: []string{runtimeAnnotationModeGithub.String(), runtimeAnnotationModeNone.String()}, DefaultValue: field.ToOptionalString(defaultCLIAnnotationMode), Description: "annotation"},
						},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "wrapper.cli.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, `jsonserialisation "github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell`) //nolint:misspell
	assert.Contains(t, output, `yamlserialisation "github.com/ARM-software/golang-utils/utils/serialization/yaml" //nolint:misspell`) //nolint:misspell
	assert.Contains(t, output, "type IResultWriter interface {")
	assert.Contains(t, output, "Write(result any) error")
	assert.Contains(t, output, "type HumanWriter struct {")
	assert.Contains(t, output, "type JSONWriter struct {")
	assert.Contains(t, output, "type YAMLWriter struct {")
	assert.Contains(t, output, "closeStore *parallelisation.CloserStore")
	assert.Contains(t, output, "defer func() { _ = writer.closeStore.Close() }()")
	assert.Contains(t, output, "func resolveRuntimeLoggerTypeForResultOutput(option any, outputFile string, loggerType string, defaultLoggerType string, fileLoggerType string, customLoggerType string, noneLoggerType string) string {")
	assert.Contains(t, output, "func ResolveRuntimeLoggerAndAnnotationsForResultOutput(option any, outputFile string, loggerType string, loggerFile string, annotationType string, loggerSource string) (logs.Loggers, annotations.IAnnotationLogger, error) {")
	assert.Contains(t, output, "selectedLogger := resolveRuntimeLoggerTypeForResultOutput(option, outputFile, loggerType, config.InspectLoggerConsole.String(), config.InspectLoggerFile.String(), config.InspectLoggerCustom.String(), config.InspectLoggerNone.String())")
	assert.Contains(t, output, "func writeResultFactory(ctx context.Context, output string, outputFile string) (IResultWriter, error) {")
	assert.Contains(t, output, "converted, err := valueUtils.StringConverter.ConvertValue(writer.ctx, result)")
	assert.Contains(t, output, "case config.InspectOutputFormatHuman.String():")
	assert.Contains(t, output, "if err := filesystem.NewOSPathValidationRule(true).Validate(trimmed); err != nil {")
	assert.Contains(t, output, "closeStore = parallelisation.NewCloserStore(false)")
	assert.Contains(t, output, "return &HumanWriter{ctx: ctx, writer: writer, closeStore: closeStore}, nil")
	assert.Contains(t, output, "var InspectWithResultAndAnnotationsImplFunc func(context.Context, annotations.IAnnotationLogger, *config.InspectConfiguration) (any, error) = nil")
	assert.Contains(t, output, "var InspectWithResultImplFunc func(context.Context, logs.Loggers, *config.InspectConfiguration) (any, error) = nil")
	assert.Contains(t, output, "logger.LogError(fmt.Sprintf(\"an error occurred: %v\", resultErr))")
	assert.Contains(t, output, "resultWriter, err := writeResultFactory(ctx, cfg.OutputFormat, cfg.OutputFile)")
	assert.Contains(t, output, "logger.Log(result)")
}

func TestDocumentationTemplateIncludesResultOutputCombinationTable(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/documentation.md.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		Distribution: precomputeDistribution(&SDKSpecification{}),
		SDK: SDK{
			Specification: SDKSpecification{
				Name:        field.ToOptionalString("Result Output SDK"),
				PackageName: "result-output-sdk",
				MainModule:  "resultoutput",
				Interface:   Interface{Methods: []InterfaceMethod{}},
				CommandLineCommands: &CommandLineCommands{
					Binary:             field.ToOptionalString("result-output"),
					RootEnvironmentVar: field.ToOptionalString("RESULT_OUTPUT"),
					Description:        field.ToOptionalString("CLI for testing generated result output support."),
					Functions: []CliFunction{{
						Name:        "Inspect",
						Description: "Emit command results to the selected destination.",
						WithResult:  true,
						Configuration: []FunctionConfigFields{{
							Name:            resultOutputArgument,
							LangType:        stringLangType,
							MapstructureTag: resultOutputMapstructureTag,
							Description:     "output format",
						}},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "documentation.md.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "Result output and logging combinations:")
	assert.Contains(t, output, "| `output_format` | `output_file` | Allowed `logger` values | Behaviour | Allowed |")
	assert.Contains(t, output, "| `human` | empty or `-` | `console`, `rich`, `file`, `none`, `custom` |")
	assert.Contains(t, output, "| `json` or `yaml` | empty or `-` | `console`, `rich`, `file`, `none`, `custom` |")
	assert.Contains(t, output, "`logger=file` still requires `logger_file` in every allowed combination.")
	assert.NotContains(t, output, "Every generated CLI command exposes the shared logging and annotation options.")
}

func TestWrapperTemplateIncludesShowConfigHelper(t *testing.T) {
	tmpl, err := template.New("sdk").Funcs(templateFuncMap).ParseFS(fs, "templates/wrapper.cli.go.tmpl")
	require.NoError(t, err)

	data := &DataModel{
		SDK: SDK{
			Specification: SDKSpecification{
				PackageName: "show-config-sdk",
				MainModule:  "showconfig",
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{{
						Name:        "Generate",
						Description: "Generate output",
						Configuration: []FunctionConfigFields{{
							Name:            showConfigArgument,
							LangType:        string(BuiltInLangTypeBool),
							MapstructureTag: showConfigMapstructureTag,
							Description:     "show config",
						}},
					}},
				},
			},
		},
	}

	var b strings.Builder
	require.NoError(t, tmpl.ExecuteTemplate(&b, "wrapper.cli.go.tmpl", data))
	output := b.String()
	assert.Contains(t, output, "func ShowCommandConfiguration(ctx context.Context, envVarPrefix string, cfg any, logger logs.Loggers, show bool) (err error) {")
	assert.Contains(t, output, "func DetermineCommandConfigurationEnvironmentVariables(ctx context.Context, envVarPrefix string, cfg any) (envVars []environmentUtils.IEnvironmentVariable, err error) {")
	assert.Contains(t, output, "configUtils.DetermineConfigurationEnvironmentVariablesWithContext(ctx, envVarPrefix, serviceCfg, sensitiveDataRedacters()...)")
	assert.Contains(t, output, `logger.Log("command configuration:")`)
	assert.Contains(t, output, "environmentUtils.UniqueEnvironmentVariables(false, envVars...)")
	assert.Contains(t, output, "environmentUtils.SortEnvironmentVariables(envVars)")
	assert.Contains(t, output, "configUtils.CommonFieldTagSecretConverter")
	assert.Contains(t, output, "func redactProtectedString() valueUtils.IValueConverter {")
	assert.Contains(t, output, "func redactLikelyHexHash() valueUtils.IValueConverter {")
	assert.Contains(t, output, "hashing.IsLikelyHexHashString(text)")
	assert.Contains(t, output, "configUtils.SecretConverter.ConvertValue(ctx, text)")
	assert.Contains(t, output, "ShowCommandConfiguration(ctx, config.App, cfg, logger, cfg.ShowConfig)")
}
