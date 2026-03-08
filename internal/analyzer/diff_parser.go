package analyzer

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// Hunk representa um bloco de linhas alteradas em um patch.
type Hunk struct {
	StartLine int
	EndLine   int
	Content   string
}

// FileHunks contém hunks de um arquivo.
type FileHunks struct {
	Filename string
	Hunks    []Hunk
}

var hunkHeader = regexp.MustCompile(`^@@ -(\d+),?\d* \+(\d+),?(\d*) @@`)

// ParsePatch extrai hunks de um patch no formato unified diff (RF02).
func ParsePatch(patch string) ([]Hunk, error) {
	if strings.TrimSpace(patch) == "" {
		return nil, nil
	}
	var hunks []Hunk
	lines := strings.Split(patch, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		if strings.HasPrefix(line, "@@") {
			matches := hunkHeader.FindStringSubmatch(line)
			if len(matches) < 4 {
				return nil, ErrInvalidPatch
			}
			start, _ := strconv.Atoi(matches[2])
			sizeStr := matches[3]
			size := 1
			if sizeStr != "" {
				size, _ = strconv.Atoi(sizeStr)
			}
			end := start + size - 1
			if size == 0 {
				end = start
			}
			var content []string
			i++
			for i < len(lines) && !strings.HasPrefix(lines[i], "@@") {
				content = append(content, lines[i])
				i++
			}
			hunks = append(hunks, Hunk{
				StartLine: start,
				EndLine:   end,
				Content:   strings.Join(content, "\n"),
			})
			continue
		}
		i++
	}
	return hunks, nil
}

// ChangedLinesFromFiles extrai hunks de cada arquivo alterado.
func ChangedLinesFromFiles(files []domain.FileChange) []FileHunks {
	var result []FileHunks
	for _, f := range files {
		if f.Patch == "" {
			continue
		}
		hunks, err := ParsePatch(f.Patch)
		if err != nil {
			continue
		}
		if len(hunks) > 0 {
			result = append(result, FileHunks{Filename: f.Filename, Hunks: hunks})
		}
	}
	return result
}

// ErrInvalidPatch é retornado quando o patch não é válido.
var ErrInvalidPatch = errors.New("invalid patch")
