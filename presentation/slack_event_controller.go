package presentation

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kakudo415/survey-bot/usecase"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type SlackEventController struct {
	eventHandler  *usecase.SlackEventHandler
	signingSecret string
}

func NewSlackEventController(eventHandler *usecase.SlackEventHandler, signingSecret string) *SlackEventController {
	return &SlackEventController{
		eventHandler:  eventHandler,
		signingSecret: signingSecret,
	}
}

func (c *SlackEventController) HandleSlackEvent(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, c.signingSecret)
	if err != nil {
		http.Error(w, "Failed to create verifier", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.TeeReader(r.Body, &verifier))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := verifier.Ensure(); err != nil {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}

	if event.Type == slackevents.URLVerification {
		r := event.Data.(*slackevents.EventsAPIURLVerificationEvent)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(r.Challenge))
		return
	}

	if event.Type == slackevents.CallbackEvent {
		innerEvent := event.InnerEvent
		if err := c.eventHandler.HandleEvent(r.Context(), innerEvent); err != nil {
			fmt.Printf("Failed to handle event: %v\n", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

