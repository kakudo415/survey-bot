package usecase

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/kakudo415/survey-bot/domain"
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

type SlackEvent struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge,omitempty"`
	Event     *Event `json:"event,omitempty"`
}

type Event struct {
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	User      string `json:"user"`
	Text      string `json:"text"`
	Timestamp string `json:"ts"`
	ThreadTS  string `json:"thread_ts,omitempty"`
}

func NewSlackEventHandler(surveyUseCase *SurveyUseCase, slackClient SlackClient, targetChannelID string) *SlackEventHandler {
	return &SlackEventHandler{
		surveyUseCase:   surveyUseCase,
		slackClient:     slackClient,
		targetChannelID: targetChannelID,
	}
}

func (h *SlackEventHandler) HandleEvent(ctx context.Context, event SlackEvent) error {
	// URL verification challenge
	if event.Type == "url_verification" {
		return nil // challenge response is handled in controller
	}

	// メッセージイベントでない場合はスキップ
	if event.Event == nil || event.Event.Type != "message" {
		return nil
	}

	// 対象チャンネル以外はスキップ
	if event.Event.Channel != h.targetChannelID {
		return nil
	}

	// Bot自身のメッセージはスキップ
	if event.Event.User == "" {
		return nil
	}

	// スレッドメッセージはスキップ（元メッセージのみ処理）
	if event.Event.ThreadTS != "" {
		return nil
	}

	// URLを抽出
	urls := h.extractURLs(event.Event.Text)
	if len(urls) == 0 {
		return nil
	}

	// 各URLを処理
	for _, url := range urls {
		if err := h.processURL(ctx, url, event.Event); err != nil {
			log.Printf("Failed to process URL %s: %v", url, err)
			// エラーをEphemeral messageで通知
			if err := h.slackClient.PostEphemeralError(ctx, event.Event.Channel, event.Event.User, err.Error()); err != nil {
				log.Printf("Failed to post ephemeral error: %v", err)
			}
		}
	}

	return nil
}

func (h *SlackEventHandler) processURL(ctx context.Context, url string, event *Event) error {
	paper, summary, err := h.surveyUseCase.ProcessPaper(ctx, url)
	if err != nil {
		return err
	}

	// スレッドに要約を投稿
	return h.slackClient.PostMessageToThread(ctx, event.Channel, event.Timestamp, paper, summary)
}

func (h *SlackEventHandler) extractURLs(text string) []string {
	// HTTP/HTTPSのURLを抽出
	urlRegex := regexp.MustCompile(`https?://[^\s<>]+`)
	matches := urlRegex.FindAllString(text, -1)
	
	var urls []string
	for _, match := range matches {
		// Slackのリンク形式 <url|text> から URL部分を抽出
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