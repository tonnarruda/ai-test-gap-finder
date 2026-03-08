package analyzer

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePatch_AddedLines(t *testing.T) {
	patch := `@@ -1,3 +1,5 @@
 func Foo() {
+   x := 1
+   return x
 }
`
	hunks, err := ParsePatch(patch)
	require.NoError(t, err)
	require.Len(t, hunks, 1)
	assert.Equal(t, 1, hunks[0].StartLine)
	assert.Equal(t, 5, hunks[0].EndLine)
	assert.Contains(t, hunks[0].Content, "x := 1")
}

func TestParsePatch_MultipleHunks(t *testing.T) {
	patch := `@@ -10,2 +10,3 @@
 old
+new line
@@ -20,1 +21,2 @@
 other
+another
`
	hunks, err := ParsePatch(patch)
	require.NoError(t, err)
	require.Len(t, hunks, 2)
	assert.Equal(t, 10, hunks[0].StartLine)
	assert.Equal(t, 12, hunks[0].EndLine)
	assert.Equal(t, 21, hunks[1].StartLine)
	assert.Equal(t, 22, hunks[1].EndLine)
}

func TestParsePatch_Empty(t *testing.T) {
	hunks, err := ParsePatch("")
	require.NoError(t, err)
	assert.Empty(t, hunks)
}

func TestParsePatch_InvalidHeader(t *testing.T) {
	_, err := ParsePatch("@@ invalid @@")
	require.Error(t, err)
}

func TestChangedLinesFromFiles(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "a.go", Patch: "@@ -1,2 +1,4 @@\n+a\n+b"},
		{Filename: "b.go", Patch: ""},
	}
	out := ChangedLinesFromFiles(files)
	require.Len(t, out, 1)
	assert.Equal(t, "a.go", out[0].Filename)
	assert.NotEmpty(t, out[0].Hunks)
}
