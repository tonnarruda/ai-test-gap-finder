package github

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// PRClient obtém dados de PRs via GitHub API.
type PRClient struct {
	client *github.Client
}

// NewPRClient cria um cliente com optional token para autenticação.
// baseURL vazio usa api.github.com; em testes use server.URL + "/".
func NewPRClient(httpClient *http.Client, token string, baseURL string) *PRClient {
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(context.Background(), ts)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := github.NewClient(httpClient)
	if baseURL != "" {
		u, _ := url.Parse(baseURL)
		client.BaseURL = u
	}
	return &PRClient{client: client}
}

// GetPRDiff retorna arquivos alterados e patches do PR (RF02).
func (c *PRClient) GetPRDiff(owner, repo string, prNumber int) (*domain.PRDiff, error) {
	ctx := context.Background()
	files, _, err := c.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}
	out := &domain.PRDiff{Files: make([]domain.FileChange, 0, len(files))}
	for _, f := range files {
		patch := ""
		if f.Patch != nil {
			patch = *f.Patch
		}
		status := ""
		if f.Status != nil {
			status = *f.Status
		}
		name := ""
		if f.Filename != nil {
			name = *f.Filename
		}
		out.Files = append(out.Files, domain.FileChange{
			Filename: name,
			Patch:    patch,
			Status:   status,
		})
	}
	return out, nil
}

// FilterGoFiles retorna apenas arquivos .go (incluindo _test.go).
func FilterGoFiles(files []domain.FileChange) []domain.FileChange {
	out := make([]domain.FileChange, 0)
	for _, f := range files {
		if len(f.Filename) > 3 && (f.Filename[len(f.Filename)-3:] == ".go") {
			out = append(out, f)
		}
	}
	return out
}

// PostPRComment publica um comentário no PR (RF07).
func (c *PRClient) PostPRComment(owner, repo string, prNumber int, body string) error {
	ctx := context.Background()
	comment := &github.IssueComment{Body: &body}
	_, _, err := c.client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	return err
}

// GetFileContent retorna o conteúdo de um arquivo no ref (ex: SHA do PR head).
func (c *PRClient) GetFileContent(owner, repo, ref, path string) (string, error) {
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{Ref: ref}
	fc, _, _, err := c.client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return "", err
	}
	if fc == nil || fc.Content == nil {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(*fc.Content)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
