package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/kakudo415/survey-bot/domain"
	"github.com/slack-go/slack"
)

type SlackClient struct {
	client *slack.Client
}

func NewSlackClient(botToken string) *SlackClient {
	return &SlackClient{
		client: slack.New(botToken),
	}
}


func (s *SlackClient) PostMessageToThread(ctx context.Context, channel, threadTS string, paper *domain.Paper, summary *domain.Summary) error {
	message := s.formatSummaryMessage(paper, summary)
	
	_, _, err := s.client.PostMessageContext(ctx, channel, slack.MsgOptionText(message, false), slack.MsgOptionTS(threadTS))
	return err
}

func (s *SlackClient) PostEphemeralError(ctx context.Context, channel, user, errorMsg string) error {
	message := fmt.Sprintf("❌ エラーが発生しました: %s", errorMsg)
	_, err := s.client.PostEphemeralContext(ctx, channel, user, slack.MsgOptionText(message, false))
	return err
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