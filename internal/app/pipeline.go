package app

import (
	"context"
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
	fileHunks := analyzer.ChangedLinesFromFiles(diff.Files)
	sourceByFile := make(map[string]string)
	for _, f := range sourceFiles {
		content, _ := p.prClient.GetFileContent(owner, repo, headSHA, f.Filename)
		sourceByFile[f.Filename] = content
	}
	var allFuncs []domain.ChangedFunction
	for _, fh := range fileHunks {
		if strings.HasSuffix(fh.Filename, "_test.go") {
			continue
		}
		src := sourceByFile[fh.Filename]
		if src == "" {
			continue
		}
		for _, hunk := range fh.Hunks {
			funcs, err := analyzer.DetectFunctionsInRange(fh.Filename, src, hunk.StartLine, hunk.EndLine)
			if err != nil {
				continue
			}
			allFuncs = append(allFuncs, funcs...)
		}
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
	result := &domain.AnalysisResult{
		FilesAnalyzed:     len(sourceFiles),
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
	body := commenter.FormatComment(*result)
	return p.prClient.PostPRComment(owner, repo, prNumber, body)
}
