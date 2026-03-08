package github

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRClient_GetPRDiff_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/pulls/1/files" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{"filename": "pkg/user.go", "patch": "@@ -1,3 +1,5 @@\n func Foo() {}", "status": "modified"},
				{"filename": "pkg/user_test.go", "patch": "@@ -0,0 +1,2 @@\n", "status": "added"}
			]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	diff, err := client.GetPRDiff("owner", "repo", 1)
	require.NoError(t, err)
	require.Len(t, diff.Files, 2)
	assert.Equal(t, "pkg/user.go", diff.Files[0].Filename)
	assert.Equal(t, "modified", diff.Files[0].Status)
	assert.Contains(t, diff.Files[0].Patch, "func Foo()")
	assert.Equal(t, "pkg/user_test.go", diff.Files[1].Filename)
	assert.Equal(t, "added", diff.Files[1].Status)
}

func TestPRClient_GetPRDiff_EmptyFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	diff, err := client.GetPRDiff("owner", "repo", 99)
	require.NoError(t, err)
	assert.Empty(t, diff.Files)
}

func TestPRClient_GetPRDiff_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	_, err := client.GetPRDiff("owner", "repo", 1)
	require.Error(t, err)
}

func TestPRClient_GetPRDiff_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	_, err := client.GetPRDiff("owner", "repo", 1)
	require.Error(t, err)
}

func TestFilterGoFiles(t *testing.T) {
	files := []domain.FileChange{
		{Filename: "a.go"},
		{Filename: "b_test.go"},
		{Filename: "c.txt"},
		{Filename: "d.go"},
	}
	out := FilterGoFiles(files)
	require.Len(t, out, 3)
	assert.Equal(t, "a.go", out[0].Filename)
	assert.Equal(t, "b_test.go", out[1].Filename)
	assert.Equal(t, "d.go", out[2].Filename)
}

func TestPRClient_PostPRComment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/repos/owner/repo/issues/5/comments" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"body":"test"}`))
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	err := client.PostPRComment("owner", "repo", 5, "Hello")
	require.NoError(t, err)
}

func TestPRClient_PostPRComment_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewPRClient(server.Client(), "", server.URL+"/")
	err := client.PostPRComment("owner", "repo", 1, "x")
	require.Error(t, err)
}
