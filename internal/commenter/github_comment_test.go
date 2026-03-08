package commenter

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestFormatComment(t *testing.T) {
	result := domain.AnalysisResult{
		FilesAnalyzed:  4,
		FunctionsCount: 7,
		Gaps: []domain.Gap{
			{
				File:      "user_service.go",
				Function:  "ValidateLogin",
				Scenarios: []string{"empty password", "user not found", "expired session"},
				Suggested: []string{"TestValidateLogin_EmptyPassword", "TestValidateLogin_UserNotFound"},
			},
		},
	}
	body := FormatComment(result)
	assert.Contains(t, body, "AI Test Gap Finder")
	assert.Contains(t, body, "Files changed: 4")
	assert.Contains(t, body, "Functions analyzed: 7")
	assert.Contains(t, body, "ValidateLogin")
	assert.Contains(t, body, "empty password")
	assert.Contains(t, body, "user not found")
	assert.Contains(t, body, "TestValidateLogin_EmptyPassword")
	assert.Contains(t, body, "Recommendation")
}

func TestFormatComment_NoGaps(t *testing.T) {
	result := domain.AnalysisResult{
		FilesAnalyzed:  2,
		FunctionsCount: 3,
		Gaps:          nil,
	}
	body := FormatComment(result)
	assert.Contains(t, body, "AI Test Gap Finder")
	assert.Contains(t, body, "No potential test gaps detected")
}
