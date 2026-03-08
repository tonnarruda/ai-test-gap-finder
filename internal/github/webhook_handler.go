package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
)

type githubPayload struct {
	Action      string `json:"action"`
	Repository  repoPayload `json:"repository"`
	PullRequest prPayload   `json:"pull_request"`
}

type repoPayload struct {
	Name  string     `json:"name"`
	Owner ownerPayload `json:"owner"`
}

type ownerPayload struct {
	Login string `json:"login"`
}

type prPayload struct {
	Number int    `json:"number"`
	Head   refPayload `json:"head"`
	Base   refPayload `json:"base"`
}

type refPayload struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// ParseWebhookPayload lê o body da requisição e retorna o evento de PR.
func ParseWebhookPayload(req *http.Request) (*domain.PREvent, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or body is nil")
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return ParseWebhookPayloadBody(body)
}

// ParseWebhookPayloadBody parseia o payload já lido (para uso após validação de assinatura).
func ParseWebhookPayloadBody(body []byte) (*domain.PREvent, error) {
	var p githubPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return &domain.PREvent{
		Action: p.Action,
		Repo: domain.Repo{
			Owner: p.Repository.Owner.Login,
			Name:  p.Repository.Name,
		},
		PR: domain.PullRequest{
			Number: p.PullRequest.Number,
			Head: domain.HeadRef{
				Ref: p.PullRequest.Head.Ref,
				SHA: p.PullRequest.Head.SHA,
			},
			Base: domain.BaseRef{
				Ref: p.PullRequest.Base.Ref,
				SHA: p.PullRequest.Base.SHA,
			},
		},
	}, nil
}

// ShouldAnalyzePR retorna true para eventos opened e synchronize (RF01).
func ShouldAnalyzePR(event *domain.PREvent) bool {
	if event == nil {
		return false
	}
	return event.Action == "opened" || event.Action == "synchronize"
}

// ValidateWebhookSignature valida o HMAC do payload com o secret do GitHub App.
func ValidateWebhookSignature(secret, payload []byte, signature string) bool {
	if len(secret) == 0 || len(signature) < 7 {
		return false
	}
	expected := "sha256=" + computeHMAC(secret, payload)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func computeHMAC(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ReadBodyForSignature lê o body sem consumi-lo para validação; use com httptest ou rewind.
func ReadBodyForSignature(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil {
		return nil, errors.New("request or body is nil")
	}
	return io.ReadAll(req.Body)
}
