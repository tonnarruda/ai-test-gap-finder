package testdetector

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTestFiles_None(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "a.go"},
		{Filename: "b.go"},
	}
	out := FindTestFiles(files)
	assert.Empty(t, out)
}

func TestFindTestFiles_Mixed(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "pkg/user.go"},
		{Filename: "pkg/user_test.go", Patch: "@@ -0,0 +1,10 @@\nfunc TestValidateLogin_Valid(t *testing.T)"},
		{Filename: "pkg/other_test.go"},
	}
	out := FindTestFiles(files)
	require.Len(t, out, 2)
	assert.Equal(t, "pkg/user_test.go", out[0].Filename)
	assert.Equal(t, "pkg/other_test.go", out[1].Filename)
}

func TestFindRelatedTestFuncs_ByPrefix(t *testing.T) {
	src := `package pkg
func TestValidateLogin_Valid(t *testing.T) {}
func TestValidateLogin_Invalid(t *testing.T) {}
func TestOther(t *testing.T) {}
`
	funcs := FindRelatedTestFuncs("user_test.go", src, "ValidateLogin")
	require.Len(t, funcs, 2)
	assert.Contains(t, funcs, "TestValidateLogin_Valid")
	assert.Contains(t, funcs, "TestValidateLogin_Invalid")
	assert.NotContains(t, funcs, "TestOther")
}

func TestFindRelatedTestFuncs_ExactMatch(t *testing.T) {
	src := `package pkg
func TestFoo(t *testing.T) {}
`
	funcs := FindRelatedTestFuncs("foo_test.go", src, "Foo")
	require.Len(t, funcs, 1)
	assert.Equal(t, "TestFoo", funcs[0])
}
