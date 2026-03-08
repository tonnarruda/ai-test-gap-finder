package ai

import (
	"context"
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestionEngine_Suggest_Mock(t *testing.T) {
	engine := NewMockEngine()
	cf := domain.ChangedFunction{
		File: "user.go", FuncName: "ValidateLogin",
		Branches: []domain.BranchCondition{{Condition: "user == nil", Line: 1}},
	}
	src := "func ValidateLogin(user *User) error { if user == nil { return nil }; return nil }"
	ctx := context.Background()
	result, err := engine.Suggest(ctx, cf, src)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Scenarios)
	assert.NotEmpty(t, result.Markdown)
}

func TestSuggestionEngine_Suggest_EmptyWhenNoBranches(t *testing.T) {
	engine := NewMockEngine()
	cf := domain.ChangedFunction{File: "a.go", FuncName: "Foo", Branches: nil}
	src := "func Foo() {}"
	ctx := context.Background()
	result, err := engine.Suggest(ctx, cf, src)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnrichGapsWithAI_Mock(t *testing.T) {
	engine := NewMockEngine()
	gaps := []domain.Gap{
		{File: "u.go", Function: "Validate", Scenarios: []string{"nil"}, Suggested: nil},
	}
	sourceByFile := map[string]string{
		"u.go": "func Validate(x *T) error { if x == nil { return err }; return nil }",
	}
	ctx := context.Background()
	enriched := EnrichGapsWithAI(ctx, engine, gaps, sourceByFile)
	require.Len(t, enriched, 1)
	assert.NotEmpty(t, enriched[0].Suggested)
	assert.NotEmpty(t, enriched[0].AISuggestions)
}
