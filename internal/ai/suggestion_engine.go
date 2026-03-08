package ai

import (
	"context"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// SuggestionResult contém cenários e texto markdown (casos de teste + código) da IA.
type SuggestionResult struct {
	Scenarios []string // lista de cenários
	Markdown  string   // resposta completa em markdown (cenários + bloco ```go)
}

// SuggestionEngine gera sugestões de cenários e código de teste via IA.
type SuggestionEngine interface {
	Suggest(ctx context.Context, fn domain.ChangedFunction, source string) (*SuggestionResult, error)
}

// MockEngine retorna sugestões fixas para testes.
type MockEngine struct{}

// NewMockEngine cria um engine de teste.
func NewMockEngine() *MockEngine {
	return &MockEngine{}
}

// Suggest retorna cenários fixos e texto de exemplo quando há branches.
func (m *MockEngine) Suggest(ctx context.Context, fn domain.ChangedFunction, source string) (*SuggestionResult, error) {
	if len(fn.Branches) == 0 {
		return nil, nil
	}
	scenarios := []string{"user nil", "empty password", "expired token"}
	md := "**Cenários sugeridos:**\n- user nil\n- empty password\n- expired token\n\n**Exemplo de teste:**\n```go\nfunc Test" + fn.FuncName + "_InvalidInput(t *testing.T) {\n\t// TODO: testar cenários acima\n}\n```"
	return &SuggestionResult{Scenarios: scenarios, Markdown: md}, nil
}

// EnrichGapsWithAI preenche Suggested e AISuggestions nos gaps usando a IA.
func EnrichGapsWithAI(ctx context.Context, engine SuggestionEngine, gaps []domain.Gap, sourceByFile map[string]string) []domain.Gap {
	if engine == nil {
		return gaps
	}
	out := make([]domain.Gap, len(gaps))
	for i, g := range gaps {
		out[i] = g
		src := sourceByFile[g.File]
		if src == "" {
			continue
		}
		cf := domain.ChangedFunction{
			File: g.File, FuncName: g.Function,
			Branches: scenariosToBranches(g.Scenarios),
		}
		result, err := engine.Suggest(ctx, cf, src)
		if err != nil || result == nil {
			continue
		}
		if len(result.Scenarios) > 0 && len(g.Suggested) == 0 {
			out[i].Suggested = SuggestTestNamesFromScenarios(g.Function, result.Scenarios)
		}
		if result.Markdown != "" {
			out[i].AISuggestions = result.Markdown
		}
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
