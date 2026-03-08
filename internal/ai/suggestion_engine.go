package ai

import (
	"context"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// SuggestionEngine gera sugestões de cenários de teste via IA.
type SuggestionEngine interface {
	Suggest(ctx context.Context, fn domain.ChangedFunction, source string) ([]string, error)
}

// MockEngine retorna sugestões fixas para testes.
type MockEngine struct{}

// NewMockEngine cria um engine de teste.
func NewMockEngine() *MockEngine {
	return &MockEngine{}
}

// Suggest retorna cenários fixos quando há branches.
func (m *MockEngine) Suggest(ctx context.Context, fn domain.ChangedFunction, source string) ([]string, error) {
	if len(fn.Branches) == 0 {
		return nil, nil
	}
	return []string{"user nil", "empty password", "expired token"}, nil
}

// EnrichGapsWithAI preenche Suggested nos gaps usando a IA quando vazio.
func EnrichGapsWithAI(ctx context.Context, engine SuggestionEngine, gaps []domain.Gap, sourceByFile map[string]string) []domain.Gap {
	if engine == nil {
		return gaps
	}
	out := make([]domain.Gap, len(gaps))
	for i, g := range gaps {
		out[i] = g
		if len(g.Suggested) > 0 {
			continue
		}
		src := sourceByFile[g.File]
		if src == "" {
			continue
		}
		cf := domain.ChangedFunction{
			File: g.File, FuncName: g.Function,
			Branches: scenariosToBranches(g.Scenarios),
		}
		suggestions, err := engine.Suggest(ctx, cf, src)
		if err != nil || len(suggestions) == 0 {
			continue
		}
		out[i].Suggested = SuggestTestNamesFromScenarios(g.Function, suggestions)
	}
	return out
}

func scenariosToBranches(s []string) []domain.BranchCondition {
	b := make([]domain.BranchCondition, len(s))
	for i, x := range s {
		b[i] = domain.BranchCondition{Condition: x, Line: i}
	}
	return b
}

// SuggestTestNamesFromScenarios converte cenários em nomes de teste (TestFunc_Cenario).
func SuggestTestNamesFromScenarios(funcName string, scenarios []string) []string {
	out := make([]string, 0, len(scenarios))
	for _, s := range scenarios {
		parts := strings.Fields(s)
		for i := range parts {
			if len(parts[i]) > 0 {
				parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			}
		}
		suffix := strings.ReplaceAll(strings.Join(parts, ""), " ", "")
		out = append(out, "Test"+funcName+"_"+suffix)
	}
	return out
}
