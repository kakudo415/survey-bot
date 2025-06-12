package domain

import (
	"context"
	"time"
)

type PaperID string

type Paper struct {
	ID          PaperID
	Title       string
	Authors     []string
	PublishedAt time.Time
	URL         string
	Content     string
}

func NewPaper(id PaperID, title string, authors []string, publishedAt time.Time, url string, content string) *Paper {
	return &Paper{
		ID:          id,
		Title:       title,
		Authors:     authors,
		PublishedAt: publishedAt,
		URL:         url,
		Content:     content,
	}
}

type PaperRepository interface {
	FindByURL(ctx context.Context, url string) (*Paper, error)
}