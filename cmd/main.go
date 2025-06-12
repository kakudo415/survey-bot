package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kakudo415/survey-bot/infrastructure"
	"github.com/kakudo415/survey-bot/presentation"
	"github.com/kakudo415/survey-bot/usecase"
)

func main() {
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackBotToken == "" {
		log.Fatal("SLACK_BOT_TOKEN environment variable is required")
	}

	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if slackSigningSecret == "" {
		log.Fatal("SLACK_SIGNING_SECRET environment variable is required")
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	openaiModel := os.Getenv("OPENAI_MODEL")
	if openaiModel == "" {
		openaiModel = "gpt-3.5-turbo"
	}

	targetChannelID := os.Getenv("TARGET_CHANNEL_ID")
	if targetChannelID == "" {
		log.Fatal("TARGET_CHANNEL_ID environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	paperRepo := infrastructure.NewIEEEPaperRepository()
	summaryService := infrastructure.NewOpenAISummaryService(openaiAPIKey, openaiModel)
	slackClient := infrastructure.NewSlackClient(slackBotToken)

	surveyUseCase := usecase.NewSurveyUseCase(paperRepo, summaryService)
	eventHandler := usecase.NewSlackEventHandler(surveyUseCase, slackClient, targetChannelID)
	controller := presentation.NewSlackEventController(eventHandler, slackSigningSecret)

	http.HandleFunc("/slack/events", controller.HandleSlackEvent)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}