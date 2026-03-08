package app

import (
	"context"
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPRClient struct {
	diff        *domain.PRDiff
	fileContent map[string]string
	postErr     error
}

func (m *mockPRClient) GetPRDiff(owner, repo string, prNumber int) (*domain.PRDiff, error) {
	return m.diff, nil
}

func (m *mockPRClient) GetFileContent(owner, repo, ref, path string) (string, error) {
	return m.fileContent[path], nil
}

func (m *mockPRClient) PostPRComment(owner, repo string, prNumber int, body string) error {
	return m.postErr
}

func TestPipeline_Run_NoGoFiles(t *testing.T) {
	p := NewPipelineWithClient(&mockPRClient{diff: &domain.PRDiff{Files: []domain.FileChange{
		{Filename: "readme.md", Patch: ""},
	}}}, nil)
	result, err := p.Run(context.Background(), "o", "r", 1, "sha")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.FilesAnalyzed)
	assert.Empty(t, result.Gaps)
}

func TestPipeline_Run_WithSourceAndPatch(t *testing.T) {
	p := NewPipelineWithClient(&mockPRClient{
		diff: &domain.PRDiff{Files: []domain.FileChange{
			{Filename: "pkg/a.go", Patch: "@@ -1,2 +1,5 @@\n+func Foo() {\n+	if x == nil {}\n+}"},
		}},
		fileContent: map[string]string{
			"pkg/a.go": "package pkg\nfunc Foo() {\n\tif x == nil { return }\n}\n",
		},
	}, nil)
	result, err := p.Run(context.Background(), "o", "r", 1, "sha")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.FilesAnalyzed)
	assert.GreaterOrEqual(t, result.FunctionsCount, 1)
}
