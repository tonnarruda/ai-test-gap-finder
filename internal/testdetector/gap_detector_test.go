package testdetector

import (
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectGaps_NoBranches(t *testing.T) {
	funcs := []domain.ChangedFunction{
		{File: "a.go", FuncName: "Foo", Branches: nil},
	}
	testFuncs := map[string][]string{}
	gaps := DetectGaps(funcs, testFuncs)
	require.Len(t, gaps, 1)
	assert.Equal(t, "a.go", gaps[0].File)
	assert.Equal(t, "Foo", gaps[0].Function)
}

func TestDetectGaps_BranchWithoutTest(t *testing.T) {
	funcs := []domain.ChangedFunction{
		{File: "user.go", FuncName: "ValidateLogin", Branches: []domain.BranchCondition{
			{Condition: "user == nil", Line: 3},
			{Condition: "password == \"\"", Line: 6},
		}},
	}
	testFuncs := map[string][]string{
		"ValidateLogin": {"TestValidateLogin_Valid"},
	}
	gaps := DetectGaps(funcs, testFuncs)
	require.Len(t, gaps, 1)
	assert.Equal(t, "user.go", gaps[0].File)
	assert.Equal(t, "ValidateLogin", gaps[0].Function)
	assert.GreaterOrEqual(t, len(gaps[0].Scenarios), 1)
}

func TestDetectGaps_AllCovered(t *testing.T) {
	funcs := []domain.ChangedFunction{
		{File: "a.go", FuncName: "Bar", Branches: []domain.BranchCondition{
			{Condition: "x > 0", Line: 2},
		}},
	}
	testFuncs := map[string][]string{
		"Bar": {"TestBar_Positive", "TestBar_Zero", "TestBar_Negative"},
	}
	gaps := DetectGaps(funcs, testFuncs)
	assert.Empty(t, gaps)
}

func TestSuggestTestNames(t *testing.T) {
	scenarios := []string{"empty password", "user not found"}
	out := SuggestTestNames("ValidateLogin", scenarios)
	require.Len(t, out, 2)
	assert.Contains(t, out, "TestValidateLogin_EmptyPassword")
	assert.Contains(t, out, "TestValidateLogin_UserNotFound")
}
