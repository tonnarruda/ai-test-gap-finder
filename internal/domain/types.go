package domain

// PREvent representa o evento de Pull Request do GitHub (opened ou synchronize).
type PREvent struct {
	Action string
	Repo   Repo
	PR     PullRequest
}

// Repo contém dados do repositório.
type Repo struct {
	Owner string
	Name  string
}

// PullRequest contém dados do PR.
type PullRequest struct {
	Number int
	Head   HeadRef
	Base   BaseRef
}

// HeadRef referência do branch do PR.
type HeadRef struct {
	Ref string
	SHA string
}

// BaseRef referência do branch base.
type BaseRef struct {
	Ref string
	SHA string
}

// PRDiff contém o diff e arquivos alterados do PR.
type PRDiff struct {
	Files   []FileChange
	RawDiff string
}

// FileChange representa um arquivo modificado no PR.
type FileChange struct {
	Filename string
	Patch    string
	Status   string
}

// ChangedFunction representa uma função alterada no código.
type ChangedFunction struct {
	File     string
	FuncName string
	Branches []BranchCondition
}

// BranchCondition representa uma condição/branch na função.
type BranchCondition struct {
	Condition string
	Line      int
}

// TestInfo contém informações sobre testes existentes.
type TestInfo struct {
	TestFile    string
	TestFuncs   []string
	RelatedFunc string
}

// Gap representa uma lacuna de teste detectada.
type Gap struct {
	File       string
	Function   string
	Scenarios  []string
	Suggested  []string
}

// AnalysisResult resultado da análise de test gaps.
type AnalysisResult struct {
	FilesAnalyzed   int
	FunctionsCount  int
	Gaps            []Gap
}
