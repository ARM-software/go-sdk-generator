package generator

import (
	"context"
	"testing"
	"time"

	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/require"
)

func TestDeprecationNoticeValidateRejectsMissingIssueDate(t *testing.T) {
	t.Parallel()

	notice := &DeprecationNotice{
		RemovalDate: field.ToOptionalTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
	}

	require.Error(t, notice.Validate())
}

func TestDeprecationNoticeValidateRejectsRemovalBeforeIssue(t *testing.T) {
	t.Parallel()

	notice := &DeprecationNotice{
		IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
		RemovalDate: field.ToOptionalTime(time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC)),
	}

	require.Error(t, notice.Validate())
}

func TestFunctionConfigFieldsValidateAcceptsValidDeprecationNotice(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "LegacyToken",
		LangType:        "string",
		MapstructureTag: "legacy_token",
		Description:     "Legacy token.",
		DeprecationNotice: &DeprecationNotice{
			IssueDate:   time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC),
			RemovalDate: field.ToOptionalTime(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)),
			Description: field.ToOptionalString("Prefer the scoped token path."),
			Replacement: field.ToOptionalString("ScopedToken"),
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestProcessConstraintMirrorsPreservesExistingPattern(t *testing.T) {
	t.Parallel()

	localFs := filesystem.NewStandardFileSystem()
	data, err := filesystem.ReadFile(filesystem.FilePathJoin(localFs, "testdata", "verify.yml"))
	require.NoError(t, err)

	var sdk SDK
	require.NoError(t, unmarshalManifestContent(".yml", data, &sdk))
	require.GreaterOrEqual(t, len(sdk.Specification.GlobalConfiguration), 3)
	require.NotNil(t, sdk.Specification.GlobalConfiguration[2].Pattern)
	assertPattern := *sdk.Specification.GlobalConfiguration[2].Pattern
	require.Equal(t, `.+?[/\s][\d.]+`, assertPattern)

	require.NoError(t, ProcessConstraintMirrors(context.Background(), &sdk))
	require.NotNil(t, sdk.Specification.GlobalConfiguration[2].Pattern)
	require.Equal(t, `.+?[/\s][\d.]+`, *sdk.Specification.GlobalConfiguration[2].Pattern)
}

func TestConfigurationOptionValidateRejectsStringLengthConstraintOnNonString(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Count",
		LangType:        "int",
		MapstructureTag: "count",
		Description:     "Count.",
		Constraints: &ParameterConstraints{
			MinLength: field.ToOptional(1),
		},
	}

	require.Error(t, cfg.Validate())
}

func TestFunctionConfigFieldsValidateRejectsExactlyOneOnScalar(t *testing.T) {
	t.Parallel()

	cfg := &FunctionConfigFields{
		Name:            "Mode",
		LangType:        "string",
		MapstructureTag: "mode",
		Description:     "Mode.",
		Constraints: &ParameterConstraints{
			ExactlyOne: true,
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateRejectsConflictingConstraintEnumAlias(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Format",
		LangType:        "string",
		MapstructureTag: "format",
		Description:     "Format.",
		Enum:            []string{"json", "yaml"},
		Constraints: &ParameterConstraints{
			Enum: []string{"xml"},
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateAcceptsConstraintValidationAliases(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Format",
		LangType:        "string",
		MapstructureTag: "format",
		Description:     "Format.",
		DefaultValue:    field.ToOptionalString("json"),
		Constraints: &ParameterConstraints{
			Enum:      []string{"json", "yaml"},
			Pattern:   field.ToOptionalString(`^(json|yaml)$`),
			Required:  true,
			MinLength: field.ToOptional(4),
			MaxLength: field.ToOptional(4),
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestConfigurationOptionValidateRejectsSubstringConstraintsOnNonString(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Count",
		LangType:        "int",
		MapstructureTag: "count",
		Description:     "Count.",
		Constraints: &ParameterConstraints{
			Contains: field.ToOptionalString("1"),
		},
	}

	require.Error(t, cfg.Validate())
}

func TestConfigurationOptionValidateAcceptsNullableStringConstraints(t *testing.T) {
	t.Parallel()

	cfg := &ConfigurationOption{
		Name:            "Name",
		LangType:        "string",
		Reference:       true,
		MapstructureTag: "name",
		Description:     "Name.",
		Constraints: &ParameterConstraints{
			Nullable:    true,
			NotEmpty:    true,
			Contains:    field.ToOptionalString("sdk"),
			NotContains: field.ToOptionalString("tmp"),
			Prefix:      field.ToOptionalString("arm-"),
			Suffix:      field.ToOptionalString("-gen"),
		},
	}

	require.NoError(t, cfg.Validate())
}
