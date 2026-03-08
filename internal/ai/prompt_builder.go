package ai

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// BuildPrompt monta o prompt para a IA conforme seção 10 do PRD.
func BuildPrompt(cf domain.ChangedFunction, sourceSnippet string) string {
	var b strings.Builder
	b.WriteString("Identify missing test scenarios for this function.\n")
	b.WriteString("Return a list of test cases that should exist.\n\n")
	b.WriteString("Function: ")
	b.WriteString(cf.FuncName)
	b.WriteString("\nFile: ")
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
