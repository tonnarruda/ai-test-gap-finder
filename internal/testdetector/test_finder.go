package testdetector

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// FindTestFiles retorna apenas arquivos *_test.go da lista (RF04).
func FindTestFiles(files []domain.FileChange) []domain.FileChange {
	var out []domain.FileChange
	for _, f := range files {
		if strings.HasSuffix(f.Filename, "_test.go") {
			out = append(out, f)
		}
	}
	return out
}

// FindRelatedTestFuncs extrai nomes de funções de teste relacionadas à função (RF04).
// Relação: Test<FuncName> ou Test<FuncName>_*.
func FindRelatedTestFuncs(filename, src, funcName string) []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		return nil
	}
	prefix := "Test" + funcName
	var names []string
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			return true
		}
		name := fn.Name.Name
		if name == prefix || strings.HasPrefix(name, prefix+"_") {
			names = append(names, name)
		}
		return true
	})
	return names
}
