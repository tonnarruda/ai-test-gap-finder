package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// DetectFunctions analisa o código Go e retorna funções com seus branches (RF03).
func DetectFunctions(filename, src string) ([]domain.ChangedFunction, error) {
	return detectFunctionsInRange(filename, src, 0, 1<<30)
}

// DetectFunctionsInRange retorna apenas funções que tocam o intervalo [startLine, endLine].
func DetectFunctionsInRange(filename, src string, startLine, endLine int) ([]domain.ChangedFunction, error) {
	return detectFunctionsInRange(filename, src, startLine, endLine)
}

func detectFunctionsInRange(filename, src string, startLine, endLine int) ([]domain.ChangedFunction, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var result []domain.ChangedFunction
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}
		pos := fset.Position(fn.Pos())
		posEnd := fset.Position(fn.End())
		// função entra se seu trecho (declaração até fim) sobrepõe o hunk
		if posEnd.Line < startLine || pos.Line > endLine {
			return true
		}
		cf := domain.ChangedFunction{
			File:     filename,
			FuncName: fn.Name.Name,
			Branches: extractBranches(fset, fn.Body),
		}
		result = append(result, cf)
		return true
	})
	return result, nil
}

func extractBranches(fset *token.FileSet, body *ast.BlockStmt) []domain.BranchCondition {
	var branches []domain.BranchCondition
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if stmt, ok := n.(*ast.IfStmt); ok && stmt.Cond != nil {
			line := fset.Position(stmt.Cond.Pos()).Line
			cond := exprString(stmt.Cond)
			branches = append(branches, domain.BranchCondition{Condition: cond, Line: line})
			return true
		}
		switch stmt := n.(type) {
		case *ast.SwitchStmt:
			if stmt.Tag != nil {
				line := fset.Position(stmt.Tag.Pos()).Line
				branches = append(branches, domain.BranchCondition{Condition: "switch " + exprString(stmt.Tag), Line: line})
			}
		case *ast.TypeSwitchStmt:
			line := fset.Position(stmt.Pos()).Line
			branches = append(branches, domain.BranchCondition{Condition: "type switch", Line: line})
		}
		return true
	})
	return branches
}

func exprString(e ast.Expr) string {
	if e == nil {
		return ""
	}
	// simplified string representation
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.BinaryExpr:
		return exprString(x.X) + " " + x.Op.String() + " " + exprString(x.Y)
	case *ast.UnaryExpr:
		return x.Op.String() + exprString(x.X)
	case *ast.CallExpr:
		return exprString(x.Fun) + "(...)"
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + exprString(x.X)
	case *ast.BasicLit:
		return x.Value
	case *ast.ParenExpr:
		return "(" + exprString(x.X) + ")"
	default:
		return "?"
	}
}

// FilterGoSourceFiles retorna apenas arquivos .go que não são _test.go.
func FilterGoSourceFiles(files []domain.FileChange) []domain.FileChange {
	var out []domain.FileChange
	for _, f := range files {
		if !strings.HasSuffix(f.Filename, ".go") {
			continue
		}
		if strings.HasSuffix(f.Filename, "_test.go") {
			continue
		}
		out = append(out, f)
	}
	return out
}
