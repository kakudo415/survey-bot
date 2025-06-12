package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kakudo415/survey-bot/domain"
)

type OpenAISummaryService struct {
	apiKey string
	client *http.Client
}

func NewOpenAISummaryService(apiKey string) *OpenAISummaryService {
	return &OpenAISummaryService{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type choice struct {
	Message message `json:"message"`
}

func (s *OpenAISummaryService) GenerateSummary(ctx context.Context, paper *domain.Paper) (*domain.Summary, error) {
	prompt := s.buildPrompt(paper)
	
	reqBody := openAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []message{
			{
				Role:    "system",
				Content: "あなたは学術論文の要約を作成する専門家です。論文の内容を日本語で分かりやすく要約してください。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	summaryContent := apiResp.Choices[0].Message.Content
	return domain.NewSummary(summaryContent), nil
}

func (s *OpenAISummaryService) buildPrompt(paper *domain.Paper) string {
	var prompt bytes.Buffer
	
	prompt.WriteString("以下の学術論文を日本語で要約してください。\n\n")
	prompt.WriteString(fmt.Sprintf("タイトル: %s\n", paper.Title))
	
	if len(paper.Authors) > 0 {
		prompt.WriteString(fmt.Sprintf("著者: %s\n", paper.Authors[0]))
		if len(paper.Authors) > 1 {
			prompt.WriteString(" 他")
		}
		prompt.WriteString("\n")
	}
	
	if !paper.PublishedAt.IsZero() {
		prompt.WriteString(fmt.Sprintf("発行年: %d\n", paper.PublishedAt.Year()))
	}
	
	prompt.WriteString("\n論文内容:\n")
	prompt.WriteString(paper.Content)
	prompt.WriteString("\n\n")
	
	prompt.WriteString("要約には以下の要素を含めてください:\n")
	prompt.WriteString("1. 研究の目的・背景\n")
	prompt.WriteString("2. 提案手法・アプローチ\n")
	prompt.WriteString("3. 主要な結果・発見\n")
	prompt.WriteString("4. 研究の意義・貢献\n")
	prompt.WriteString("\n分かりやすい日本語で、専門用語には適切な説明を加えて要約してください。")
	
	return prompt.String()
}