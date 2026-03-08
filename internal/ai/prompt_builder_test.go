package ai

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPrompt_SingleFunc(t *testing.T) {
	cf := domain.ChangedFunction{
		File:     "user.go",
		FuncName: "ValidateLogin",
		Branches: []domain.BranchCondition{
			{Condition: "user == nil", Line: 2},
			{Condition: "password == \"\"", Line: 5},
		},
	}
	src := `func ValidateLogin(user *User, password string) error {
	if user == nil { return ErrUserNotFound }
	if password == "" { return ErrInvalidPassword }
	return nil
}`
	prompt := BuildPrompt(cf, src)
	assert.Contains(t, prompt, "ValidateLogin")
	assert.Contains(t, prompt, "cenários de teste que faltam")
	assert.Contains(t, prompt, "código de teste em Go")
	assert.Contains(t, prompt, src)
}

func TestBuildPrompt_EmptyBranches(t *testing.T) {
	cf := domain.ChangedFunction{File: "a.go", FuncName: "Foo", Branches: nil}
	prompt := BuildPrompt(cf, "func Foo() {}")
	assert.Contains(t, prompt, "Foo")
	assert.Contains(t, prompt, "cenários de teste que faltam")
}

func TestParseSuggestionsResponse_List(t *testing.T) {
	resp := `
- empty password
- user not found
- expired token
`
	out := ParseSuggestionsResponse(resp)
	assert.Len(t, out, 3)
	assert.Equal(t, "empty password", out[0])
	assert.Equal(t, "user not found", out[1])
	assert.Equal(t, "expired token", out[2])
}

func TestParseSuggestionsResponse_Empty(t *testing.T) {
	assert.Empty(t, ParseSuggestionsResponse(""))
	assert.Empty(t, ParseSuggestionsResponse("No suggestions."))
}

func TestParseSuggestionsResponse_Numbered(t *testing.T) {
	resp := "1. invalid input\n2. nil pointer\n"
	out := ParseSuggestionsResponse(resp)
	require.Len(t, out, 2)
	assert.Equal(t, "invalid input", out[0])
	assert.Equal(t, "nil pointer", out[1])
}
