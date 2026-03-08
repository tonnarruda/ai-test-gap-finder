package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

// OpenAIEngine usa a API da OpenAI para sugestões.
type OpenAIEngine struct {
	apiKey     string
	httpClient *http.Client
}

// NewOpenAIEngine cria um engine com a chave da API.
func NewOpenAIEngine(apiKey string) *OpenAIEngine {
	return &OpenAIEngine{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

type openAIReq struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Suggest chama a OpenAI e retorna cenários + markdown (casos de teste e código).
func (e *OpenAIEngine) Suggest(ctx context.Context, fn domain.ChangedFunction, source string) (*SuggestionResult, error) {
	if len(fn.Branches) == 0 {
		return nil, nil
	}
	prompt := BuildPrompt(fn, source)
	reqBody := openAIReq{
		Model: "gpt-4o-mini",
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}
	data, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api: %s", resp.Status)
	}
	var out openAIResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, nil
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	return &SuggestionResult{
		Scenarios: ParseSuggestionsResponse(content),
		Markdown:  content,
	}, nil
}
