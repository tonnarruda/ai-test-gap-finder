package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/tonnarruda/ai-test-gap-finder/internal/ai"
	"github.com/tonnarruda/ai-test-gap-finder/internal/app"
	"github.com/tonnarruda/ai-test-gap-finder/internal/github"
)

func main() {
	_ = godotenv.Load()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Print("GITHUB_TOKEN not set; API calls may be rate-limited")
	}
	prClient := github.NewPRClient(nil, token, "")
	var aiEngine ai.SuggestionEngine
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		aiEngine = ai.NewOpenAIEngine(key)
	} else {
		aiEngine = ai.NewMockEngine()
		log.Print("OPENAI_API_KEY not set; using mock AI suggestions")
	}
	pipeline := app.NewPipeline(prClient, aiEngine)

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(webhookSecret) > 0 {
			sig := r.Header.Get("X-Hub-Signature-256")
			if !github.ValidateWebhookSignature([]byte(webhookSecret), body, sig) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		event, err := github.ParseWebhookPayloadBody(body)
		if err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if !github.ShouldAnalyzePR(event) {
			w.WriteHeader(http.StatusOK)
			return
		}
		go func() {
			if err := pipeline.RunAndComment(nil, event.Repo.Owner, event.Repo.Name, event.PR.Number, event.PR.Head.SHA); err != nil {
				log.Printf("pipeline error: %v", err)
			}
		}()
		w.WriteHeader(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
