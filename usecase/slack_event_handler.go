package usecase

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/kakudo415/survey-bot/domain"
	"github.com/slack-go/slack/slackevents"
)

type SlackEventHandler struct {
	surveyUseCase   *SurveyUseCase
	slackClient     SlackClient
	targetChannelID string
}

type SlackClient interface {
	PostMessageToThread(ctx context.Context, channel, threadTS string, paper *domain.Paper, summary *domain.Summary) error
	PostEphemeralError(ctx context.Context, channel, user, errorMsg string) error
}


func NewSlackEventHandler(surveyUseCase *SurveyUseCase, slackClient SlackClient, targetChannelID string) *SlackEventHandler {
	return &SlackEventHandler{
		surveyUseCase:   surveyUseCase,
		slackClient:     slackClient,
		targetChannelID: targetChannelID,
	}
}

func (h *SlackEventHandler) HandleEvent(ctx context.Context, event slackevents.EventsAPIInnerEvent) error {
	if event.Type != "message" {
		return nil
	}

	messageEvent, ok := event.Data.(*slackevents.MessageEvent)
	if !ok {
		return nil
	}

	if messageEvent.Channel != h.targetChannelID {
		return nil
	}

	if messageEvent.User == "" {
		return nil
	}

	if messageEvent.ThreadTimeStamp != "" {
		return nil
	}

	urls := h.extractURLs(messageEvent.Text)
	if len(urls) == 0 {
		return nil
	}

	for _, url := range urls {
		if err := h.processURL(ctx, url, messageEvent); err != nil {
			log.Printf("Failed to process URL %s: %v", url, err)
			if err := h.slackClient.PostEphemeralError(ctx, messageEvent.Channel, messageEvent.User, err.Error()); err != nil {
				log.Printf("Failed to post ephemeral error: %v", err)
			}
		}
	}

	return nil
}

func (h *SlackEventHandler) processURL(ctx context.Context, url string, event *slackevents.MessageEvent) error {
	paper, summary, err := h.surveyUseCase.ProcessPaper(ctx, url)
	if err != nil {
		return err
	}

	return h.slackClient.PostMessageToThread(ctx, event.Channel, event.TimeStamp, paper, summary)
}

func (h *SlackEventHandler) extractURLs(text string) []string {
	urlRegex := regexp.MustCompile(`https?://[^\s<>]+`)
	matches := urlRegex.FindAllString(text, -1)
	
	var urls []string
	for _, match := range matches {
		if strings.HasPrefix(match, "<") && strings.Contains(match, "|") {
			parts := strings.Split(strings.Trim(match, "<>"), "|")
			if len(parts) > 0 {
				urls = append(urls, parts[0])
			}
		} else {
			urls = append(urls, strings.Trim(match, "<>"))
		}
	}
	
	return urls
}