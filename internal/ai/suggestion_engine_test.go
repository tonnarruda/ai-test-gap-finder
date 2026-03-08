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
	suggestions, err := engine.Suggest(ctx, cf, src)
	require.NoError(t, err)
	assert.NotEmpty(t, suggestions)
}

func TestSuggestionEngine_Suggest_EmptyWhenNoBranches(t *testing.T) {
	engine := NewMockEngine()
	cf := domain.ChangedFunction{File: "a.go", FuncName: "Foo", Branches: nil}
	src := "func Foo() {}"
	ctx := context.Background()
	suggestions, err := engine.Suggest(ctx, cf, src)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
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
}
