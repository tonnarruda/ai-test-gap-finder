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

	b.WriteString("Você é um especialista em testes de software, QA e engenharia de qualidade.\n")
	b.WriteString("Possui profundo conhecimento em testes de unidade, testes de integração, análise de código e identificação de edge cases.\n\n")

	b.WriteString("Analise o código abaixo que faz parte de um Pull Request.\n")
	b.WriteString("Primeiro identifique qual é a linguagem, framework e stack utilizada no código.\n")
	b.WriteString("Baseado nisso, sugira cenários de testes apropriados para essa stack.\n\n")

	b.WriteString("Tarefas:\n\n")

	b.WriteString("1. Identifique a linguagem e o framework utilizados no código.\n")
	b.WriteString("2. Analise a lógica da função/método.\n")
	b.WriteString("3. Identifique todos os caminhos possíveis de execução (condições, branches, erros e retornos).\n")
	b.WriteString("4. Liste cenários de teste que deveriam existir para validar corretamente esse código.\n")
	b.WriteString("5. Considere cenários:\n")
	b.WriteString("   - positivos\n")
	b.WriteString("   - negativos\n")
	b.WriteString("   - edge cases\n")
	b.WriteString("   - entradas inválidas\n")
	b.WriteString("   - valores limites\n")
	b.WriteString("   - tratamento de erros\n")
	b.WriteString("   - dependências externas ou mocks\n\n")

	b.WriteString("Depois disso, sugira exemplos de testes automatizados usando o framework de testes mais comum para essa stack.\n")
	b.WriteString("Os testes devem seguir boas práticas da linguagem identificada.\n\n")

	b.WriteString("Formato da resposta:\n\n")

	b.WriteString("1️⃣ Stack identificada\n")
	b.WriteString("- Linguagem\n")
	b.WriteString("- Framework de testes sugerido\n\n")

	b.WriteString("2️⃣ Cenários de teste recomendados\n")
	b.WriteString("- lista em bullets (-)\n\n")

	b.WriteString("3️⃣ Exemplos de testes automatizados\n")
	b.WriteString("Gerar exemplos de testes usando o framework adequado da stack.\n")
	b.WriteString("Use um bloco de código com a linguagem identificada.\n\n")

	b.WriteString("Arquivo: ")
	b.WriteString(cf.File)
	b.WriteString("\n")

	b.WriteString("Função/Método: ")
	b.WriteString(cf.FuncName)
	b.WriteString("\n\n")

	b.WriteString("Código:\n")
	b.WriteString("```\n")
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
