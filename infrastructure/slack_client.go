package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kakudo415/survey-bot/domain"
)

type SlackClient struct {
	botToken string
	client   *http.Client
}

func NewSlackClient(botToken string) *SlackClient {
	return &SlackClient{
		botToken: botToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type SlackMessage struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	Parse     string `json:"parse,omitempty"`
	LinkNames bool   `json:"link_names,omitempty"`
}

type SlackEphemeralMessage struct {
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

type SlackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (s *SlackClient) PostMessageToThread(ctx context.Context, channel, threadTS string, paper *domain.Paper, summary *domain.Summary) error {
	message := s.formatSummaryMessage(paper, summary)
	
	slackMsg := SlackMessage{
		Channel:   channel,
		Text:      message,
		ThreadTS:  threadTS,
		Parse:     "full",
		LinkNames: true,
	}

	return s.postMessage(ctx, "https://slack.com/api/chat.postMessage", slackMsg)
}

func (s *SlackClient) PostEphemeralError(ctx context.Context, channel, user, errorMsg string) error {
	ephemeralMsg := SlackEphemeralMessage{
		Channel: channel,
		User:    user,
		Text:    fmt.Sprintf("❌ エラーが発生しました: %s", errorMsg),
	}

	jsonData, err := json.Marshal(ephemeralMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal ephemeral message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postEphemeral", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post ephemeral message: %w", err)
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return nil
}

func (s *SlackClient) postMessage(ctx context.Context, url string, message interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return nil
}

func (s *SlackClient) formatSummaryMessage(paper *domain.Paper, summary *domain.Summary) string {
	var msg strings.Builder
	
	msg.WriteString(fmt.Sprintf("📄 **%s**\n", paper.Title))
	
	if len(paper.Authors) > 0 {
		msg.WriteString(fmt.Sprintf("👥 著者: %s", strings.Join(paper.Authors, ", ")))
		msg.WriteString("\n")
	}
	
	if !paper.PublishedAt.IsZero() {
		msg.WriteString(fmt.Sprintf("📅 発行年: %d\n", paper.PublishedAt.Year()))
	}
	
	msg.WriteString(fmt.Sprintf("🔗 URL: %s\n", paper.URL))
	msg.WriteString("\n## 要約\n")
	msg.WriteString(summary.Content)
	
	return msg.String()
}