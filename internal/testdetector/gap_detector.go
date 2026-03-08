package testdetector

import (
	"regexp"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// DetectGaps identifica funções com branches sem teste correspondente (RF05).
// testFuncs mapeia nome da função (ex: ValidateLogin) para lista de funções de teste (ex: TestValidateLogin_Valid).
// Ignora a função "main" (entrypoint não é alvo de teste unitário).
func DetectGaps(funcs []domain.ChangedFunction, testFuncs map[string][]string) []domain.Gap {
	var gaps []domain.Gap
	for _, cf := range funcs {
		if cf.FuncName == "main" {
			continue
		}
		tests := testFuncs[cf.FuncName]
		hasEnoughTests := len(tests) >= len(cf.Branches) || len(cf.Branches) == 0
		if hasEnoughTests {
			continue
		}
		scenarios := branchConditionsToScenarios(cf.Branches)
		suggested := SuggestTestNames(cf.FuncName, scenarios)
		gaps = append(gaps, domain.Gap{
			File:      cf.File,
			Function:  cf.FuncName,
			Scenarios: scenarios,
			Suggested: suggested,
		})
	}
	return gaps
}

var scenarioClean = regexp.MustCompile(`[^a-z0-9\s]`)

func branchConditionsToScenarios(branches []domain.BranchCondition) []string {
	seen := make(map[string]bool)
	var out []string
	for _, b := range branches {
		s := scenarioClean.ReplaceAllString(strings.ToLower(b.Condition), " ")
		s = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// SuggestTestNames gera sugestões de nomes de teste (RF06).
func SuggestTestNames(funcName string, scenarios []string) []string {
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
