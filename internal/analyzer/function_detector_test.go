package analyzer

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectFunctions_Simple(t *testing.T) {
	src := `package pkg
func ValidateLogin(user *User, password string) error {
	if user == nil {
		return ErrUserNotFound
	}
	if password == "" {
		return ErrInvalidPassword
	}
	return nil
}
`
	funcs, err := DetectFunctions("user.go", src)
	require.NoError(t, err)
	require.Len(t, funcs, 1)
	assert.Equal(t, "ValidateLogin", funcs[0].FuncName)
	assert.Equal(t, "user.go", funcs[0].File)
	require.Len(t, funcs[0].Branches, 2)
	assert.Contains(t, funcs[0].Branches[0].Condition, "nil")
	assert.Contains(t, funcs[0].Branches[1].Condition, `""`)
}

func TestDetectFunctions_MultipleFuncs(t *testing.T) {
	src := `package pkg
func Foo() {}
func Bar(x int) bool {
	if x > 0 { return true }
	return false
}
`
	funcs, err := DetectFunctions("f.go", src)
	require.NoError(t, err)
	require.Len(t, funcs, 2)
	assert.Equal(t, "Foo", funcs[0].FuncName)
	assert.Equal(t, "Bar", funcs[1].FuncName)
	assert.Len(t, funcs[0].Branches, 0)
	assert.Len(t, funcs[1].Branches, 1)
}

func TestDetectFunctions_OnlyInHunkRange(t *testing.T) {
	src := `package pkg
func A() { }
func B() { }
func C() { }
`
	// Range 2-3: apenas funções cuja declaração está na linha 2 ou 3 (A e B).
	funcs, err := DetectFunctionsInRange("f.go", src, 2, 3)
	require.NoError(t, err)
	require.Len(t, funcs, 2)
	assert.Equal(t, "A", funcs[0].FuncName)
	assert.Equal(t, "B", funcs[1].FuncName)
}

func TestDetectFunctions_InvalidGo(t *testing.T) {
	_, err := DetectFunctions("x.go", "package pkg func invalid")
	require.Error(t, err)
}

func TestFilterGoSourceFiles(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "a.go"},
		{Filename: "a_test.go"},
		{Filename: "b.go"},
	}
	out := FilterGoSourceFiles(files)
	require.Len(t, out, 2)
	assert.Equal(t, "a.go", out[0].Filename)
	assert.Equal(t, "b.go", out[1].Filename)
}
