package github

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tonnarruda/ai-test-gap-finder/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWebhookPayload_Opened(t *testing.T) {
	payload := `{
		"action": "opened",
		"repository": {"name": "repo1", "owner": {"login": "owner1"}},
		"pull_request": {
			"number": 42,
			"head": {"ref": "feature", "sha": "abc123"},
			"base": {"ref": "main", "sha": "def456"}
		}
	}`
	body := bytes.NewBufferString(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")

	event, err := ParseWebhookPayload(req)
	require.NoError(t, err)
	assert.Equal(t, "opened", event.Action)
	assert.Equal(t, "owner1", event.Repo.Owner)
	assert.Equal(t, "repo1", event.Repo.Name)
	assert.Equal(t, 42, event.PR.Number)
	assert.Equal(t, "feature", event.PR.Head.Ref)
	assert.Equal(t, "abc123", event.PR.Head.SHA)
}

func TestParseWebhookPayload_Synchronize(t *testing.T) {
	payload := `{
		"action": "synchronize",
		"repository": {"name": "my-repo", "owner": {"login": "org"}},
		"pull_request": {
			"number": 1,
			"head": {"ref": "fix/bug", "sha": "sha1"},
			"base": {"ref": "main", "sha": "sha0"}
		}
	}`
	body := bytes.NewBufferString(payload)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")

	event, err := ParseWebhookPayload(req)
	require.NoError(t, err)
	assert.Equal(t, "synchronize", event.Action)
	assert.Equal(t, "org", event.Repo.Owner)
	assert.Equal(t, "my-repo", event.Repo.Name)
	assert.Equal(t, 1, event.PR.Number)
}

func TestParseWebhookPayload_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString("{ invalid")
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")

	_, err := ParseWebhookPayload(req)
	require.Error(t, err)
}

func TestParseWebhookPayload_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	_, err := ParseWebhookPayload(req)
	require.Error(t, err)
}

func TestShouldAnalyzePR_Opened(t *testing.T) {
	event := &domain.PREvent{Action: "opened"}
	assert.True(t, ShouldAnalyzePR(event))
}

func TestShouldAnalyzePR_Synchronize(t *testing.T) {
	event := &domain.PREvent{Action: "synchronize"}
	assert.True(t, ShouldAnalyzePR(event))
}

func TestShouldAnalyzePR_Other(t *testing.T) {
	for _, action := range []string{"closed", "labeled", "reopened"} {
		event := &domain.PREvent{Action: action}
		assert.False(t, ShouldAnalyzePR(event), "action=%s", action)
	}
}

func TestValidateWebhookSignature(t *testing.T) {
	secret := []byte("my-secret")
	payload := []byte(`{"action":"opened"}`)
	sig := computeHMAC(secret, payload)

	valid := ValidateWebhookSignature(secret, payload, "sha256="+sig)
	assert.True(t, valid)
}

func TestValidateWebhookSignature_Invalid(t *testing.T) {
	valid := ValidateWebhookSignature([]byte("secret"), []byte("payload"), "sha256=wrong")
	assert.False(t, valid)
}

func TestValidateWebhookSignature_EmptySignature(t *testing.T) {
	valid := ValidateWebhookSignature([]byte("secret"), []byte("payload"), "")
	assert.False(t, valid)
}

func TestComputeHMAC(t *testing.T) {
	out := computeHMAC([]byte("key"), []byte("data"))
	assert.NotEmpty(t, out)
	assert.Len(t, out, 64)
}

func Test_githubPayload(t *testing.T) {
	var p githubPayload
	err := json.Unmarshal([]byte(`{"action":"opened","repository":{"name":"r","owner":{"login":"o"}},"pull_request":{"number":1,"head":{"ref":"h","sha":"s"},"base":{"ref":"b","sha":"s0"}}}`), &p)
	require.NoError(t, err)
	assert.Equal(t, "opened", p.Action)
	assert.Equal(t, "r", p.Repository.Name)
	assert.Equal(t, "o", p.Repository.Owner.Login)
}
