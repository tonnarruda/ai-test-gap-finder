package commenter

import (
	"fmt"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// FormatComment gera o corpo do comentário no PR (RF07, seção 13 do PRD).
func FormatComment(result domain.AnalysisResult) string {
	var b strings.Builder
	b.WriteString("## AI Test Gap Finder\n\n")
	b.WriteString("### Analysis Summary\n\n")
	b.WriteString(fmt.Sprintf("- Files changed: %d\n", result.FilesAnalyzed))
	b.WriteString(fmt.Sprintf("- Functions analyzed: %d\n", result.FunctionsCount))
	if len(result.FunctionsAnalyzed) > 0 {
		b.WriteString("\nFunctions:\n")
		for _, f := range result.FunctionsAnalyzed {
			b.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", f.FuncName, f.File))
		}
		b.WriteString("\n")
	}
	if len(result.Gaps) == 0 {
		b.WriteString("✅ No potential test gaps detected.\n")
		return b.String()
	}
	b.WriteString("### ⚠ Potential Missing Tests\n\n")
	for _, g := range result.Gaps {
		b.WriteString(fmt.Sprintf("**%s** (`%s`)\n\n", g.Function, g.File))
		b.WriteString("Missing scenarios:\n")
		for _, s := range g.Scenarios {
			b.WriteString(fmt.Sprintf("- %s\n", s))
		}
		if len(g.Suggested) > 0 {
			b.WriteString("\nSuggested tests:\n")
			for _, t := range g.Suggested {
				b.WriteString(fmt.Sprintf("- `%s`\n", t))
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("---\n\n💡 **Recommendation:** Add test cases for these scenarios to improve coverage and reduce regression risk.\n")
	return b.String()
}
