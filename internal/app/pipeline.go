package app

import (
	"context"
	"log"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/ai"
	"github.com/tonnarruda/ai-test-gap-finder/internal/analyzer"
	"github.com/tonnarruda/ai-test-gap-finder/internal/commenter"
	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/tonnarruda/ai-test-gap-finder/internal/github"
	"github.com/tonnarruda/ai-test-gap-finder/internal/testdetector"
)

// PRClient obtém diff e conteúdo de arquivos e posta comentários.
type PRClient interface {
	GetPRDiff(owner, repo string, prNumber int) (*domain.PRDiff, error)
	GetFileContent(owner, repo, ref, path string) (string, error)
	PostPRComment(owner, repo string, prNumber int, body string) error
}

// Config configura o pipeline.
type Config struct {
	GitHubToken string
	AIEngine    ai.SuggestionEngine
}

// Pipeline orquestra análise de PR e comentário.
type Pipeline struct {
	prClient PRClient
	aiEngine ai.SuggestionEngine
}

// NewPipeline cria o pipeline.
func NewPipeline(prClient *github.PRClient, aiEngine ai.SuggestionEngine) *Pipeline {
	return &Pipeline{prClient: prClient, aiEngine: aiEngine}
}

// NewPipelineWithClient cria o pipeline com um cliente custom (ex: mock para testes).
func NewPipelineWithClient(prClient PRClient, aiEngine ai.SuggestionEngine) *Pipeline {
	return &Pipeline{prClient: prClient, aiEngine: aiEngine}
}

// Run executa análise do PR e retorna o resultado (sem postar).
func (p *Pipeline) Run(ctx context.Context, owner, repo string, prNumber int, headSHA string) (*domain.AnalysisResult, error) {
	diff, err := p.prClient.GetPRDiff(owner, repo, prNumber)
	if err != nil {
		return nil, err
	}
	goFiles := github.FilterGoFiles(diff.Files)
	sourceFiles := analyzer.FilterGoSourceFiles(goFiles)
	if len(sourceFiles) == 0 {
		return &domain.AnalysisResult{FilesAnalyzed: 0, FunctionsCount: 0, Gaps: nil}, nil
	}
	sourceByFile := make(map[string]string)
	for _, f := range sourceFiles {
		content, _ := p.prClient.GetFileContent(owner, repo, headSHA, f.Filename)
		sourceByFile[f.Filename] = content
	}
	var allFuncs []domain.ChangedFunction
	for _, f := range sourceFiles {
		if strings.HasSuffix(f.Filename, "_test.go") {
			continue
		}
		src := sourceByFile[f.Filename]
		if src == "" {
			continue
		}
		funcs, err := analyzer.DetectFunctions(f.Filename, src)
		if err != nil {
			// Não falhamos o pipeline por erro em um arquivo específico.
			continue
		}
		allFuncs = append(allFuncs, funcs...)
	}
	allFuncs = dedupeFunctions(allFuncs)
	testFiles := testdetector.FindTestFiles(diff.Files)
	testFuncs := make(map[string][]string)
	for _, tf := range testFiles {
		content, _ := p.prClient.GetFileContent(owner, repo, headSHA, tf.Filename)
		if content == "" {
			content = tf.Patch
		}
		for _, cf := range allFuncs {
			related := testdetector.FindRelatedTestFuncs(tf.Filename, content, cf.FuncName)
			testFuncs[cf.FuncName] = append(testFuncs[cf.FuncName], related...)
		}
	}
	gaps := testdetector.DetectGaps(allFuncs, testFuncs)
	if p.aiEngine != nil {
		gaps = ai.EnrichGapsWithAI(ctx, p.aiEngine, gaps, sourceByFile)
	}
	filesSet := make(map[string]struct{})
	for _, cf := range allFuncs {
		filesSet[cf.File] = struct{}{}
	}
	result := &domain.AnalysisResult{
		FilesAnalyzed:     len(filesSet),
		FunctionsCount:    len(allFuncs),
		FunctionsAnalyzed: make([]domain.AnalyzedFunction, 0, len(allFuncs)),
		Gaps:              gaps,
	}
	for _, cf := range allFuncs {
		result.FunctionsAnalyzed = append(result.FunctionsAnalyzed, domain.AnalyzedFunction{File: cf.File, FuncName: cf.FuncName})
	}
	return result, nil
}

func dedupeFunctions(funcs []domain.ChangedFunction) []domain.ChangedFunction {
	seen := make(map[string]bool)
	var out []domain.ChangedFunction
	for _, cf := range funcs {
		key := cf.File + "\x00" + cf.FuncName
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, cf)
	}
	return out
}

// RunAndComment executa a análise e posta o comentário no PR.
func (p *Pipeline) RunAndComment(ctx context.Context, owner, repo string, prNumber int, headSHA string) error {
	result, err := p.Run(ctx, owner, repo, prNumber, headSHA)
	if err != nil {
		return err
	}
	LogAnalysisResult(result)
	body := commenter.FormatComment(*result)
	return p.prClient.PostPRComment(owner, repo, prNumber, body)
}

// LogAnalysisResult registra no log arquivos/funções analisadas e sugestões da IA (para Render/logs).
func LogAnalysisResult(result *domain.AnalysisResult) {
	if result == nil {
		return
	}
	log.Printf("[analysis] files changed: %d | functions analyzed: %d", result.FilesAnalyzed, result.FunctionsCount)
	for _, f := range result.FunctionsAnalyzed {
		log.Printf("[analysis] function: %s (file: %s)", f.FuncName, f.File)
	}
	for i, g := range result.Gaps {
		log.Printf("[gap %d] %s (%s) | scenarios: %s", i+1, g.Function, g.File, strings.Join(g.Scenarios, ", "))
		if len(g.Suggested) > 0 {
			log.Printf("[gap %d] suggested test names: %s", i+1, strings.Join(g.Suggested, ", "))
		}
		if g.AISuggestions != "" {
			log.Printf("[gap %d] AI suggestions:\n%s", i+1, g.AISuggestions)
		}
	}
}
