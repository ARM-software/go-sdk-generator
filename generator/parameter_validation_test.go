package generator

import (
	"testing"
	"time"

	"github.com/ARM-software/golang-utils/utils/field"
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
