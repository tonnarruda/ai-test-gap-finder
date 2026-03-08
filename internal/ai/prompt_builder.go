package ai

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// BuildPrompt monta o prompt para a IA: cenários faltantes e sugestão de código de teste.
func BuildPrompt(cf domain.ChangedFunction, sourceSnippet string) string {
	var b strings.Builder
	b.WriteString("Analise esta função Go e faça duas coisas:\n\n")
	b.WriteString("1. Liste os **cenários de teste que faltam** (ex: entrada inválida, nil, erro, edge cases).\n")
	b.WriteString("2. Sugira **código de teste em Go** (testes com testing.T) para cobrir esses cenários.\n\n")
	b.WriteString("Use este formato na resposta:\n")
	b.WriteString("- Primeiro: lista de cenários (bullets com -).\n")
	b.WriteString("- Depois: um bloco de código com ```go contendo um ou mais func TestXxx(t *testing.T).\n\n")
	b.WriteString("Função: ")
	b.WriteString(cf.FuncName)
	b.WriteString("\nArquivo: ")
	b.WriteString(cf.File)
	b.WriteString("\n\n```go\n")
	b.WriteString(strings.TrimSpace(sourceSnippet))
	b.WriteString("\n```\n")
	return b.String()
}

var bulletList = regexp.MustCompile(`^[\s]*[-*]\s*(.+)`)
var numberedList = regexp.MustCompile(`^[\s]*\d+[.)]\s*(.+)`)

// ParseSuggestionsResponse extrai lista de cenários da resposta da IA.
func ParseSuggestionsResponse(response string) []string {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(response))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if m := bulletList.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, strings.TrimSpace(m[1]))
			continue
		}
		if m := numberedList.FindStringSubmatch(line); len(m) == 2 {
			out = append(out, strings.TrimSpace(m[1]))
			continue
		}
	}
	return out
}
