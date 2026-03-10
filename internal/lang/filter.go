package lang

import (
	"path/filepath"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

var codeExtensions = map[string]bool{
	".go": true, ".py": true, ".rb": true, ".java": true, ".ts": true, ".js": true,
	".cs": true, ".php": true, ".rs": true, ".kt": true, ".swift": true, ".dart": true,
	".vb": true, ".c": true, ".cpp": true, ".h": true,
}

type testPattern struct {
	suffix   string
	prefix   string
	contains string
}

var testPatterns = []testPattern{
	{suffix: "_test.go"},
	{prefix: "test_", suffix: ".py"},
	{suffix: "_test.py"},
	{suffix: "_spec.rb"},
	{suffix: "Test.java"},
	{suffix: ".test.ts"},
	{suffix: ".spec.ts"},
	{suffix: ".test.js"},
	{suffix: ".spec.js"},
	{suffix: "Test.kt"},
	{suffix: "_test.rs"},
	{suffix: "Tests.swift"},
	{suffix: "_test.dart"},
	{suffix: "Test.cs"},
	{suffix: "Tests.cs"},
	{suffix: ".test.php"},
	{suffix: "_test.php"},
	{suffix: "_spec.php"},
}

// IsCodeFile retorna true se o path tem extensão de arquivo de código suportada.
func IsCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return codeExtensions[ext]
}

// IsTestFile retorna true se o path corresponde a padrões de arquivo de teste conhecidos.
func IsTestFile(path string) bool {
	base := filepath.Base(path)
	lower := strings.ToLower(base)
	for _, p := range testPatterns {
		if p.suffix != "" && p.prefix != "" {
			if strings.HasPrefix(lower, p.prefix) && strings.HasSuffix(lower, strings.ToLower(p.suffix)) {
				return true
			}
			continue
		}
		if p.suffix != "" && strings.HasSuffix(lower, strings.ToLower(p.suffix)) {
			return true
		}
		if p.prefix != "" && strings.HasPrefix(lower, p.prefix) {
			return true
		}
		if p.contains != "" && strings.Contains(lower, p.contains) {
			return true
		}
	}
	return false
}

// FilterCodeFiles retorna apenas arquivos de código que não são arquivos de teste.
func FilterCodeFiles(files []domain.FileChange) []domain.FileChange {
	if files == nil {
		return nil
	}
	var out []domain.FileChange
	for _, f := range files {
		if !IsCodeFile(f.Filename) {
			continue
		}
		if IsTestFile(f.Filename) {
			continue
		}
		out = append(out, f)
	}
	return out
}

// FileUnitName retorna o nome da unidade (arquivo sem extensão) para uso como identificador.
func FileUnitName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}
