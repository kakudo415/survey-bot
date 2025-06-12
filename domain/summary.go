package domain

import "context"

type Summary struct {
	Content string
}

func NewSummary(content string) *Summary {
	return &Summary{
		Content: content,
	}
}

type SummaryService interface {
	GenerateSummary(ctx context.Context, paper *Paper) (*Summary, error)
}