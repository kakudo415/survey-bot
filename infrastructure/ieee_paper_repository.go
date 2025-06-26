package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kakudo415/survey-bot/domain"
)

type IEEEPaperRepository struct {
	client *http.Client
}

func NewIEEEPaperRepository() *IEEEPaperRepository {
	return &IEEEPaperRepository{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *IEEEPaperRepository) FindByURL(ctx context.Context, paperURL string) (*domain.Paper, error) {
	if !r.isIEEEURL(paperURL) {
		return nil, fmt.Errorf("not an IEEE Xplore URL: %s", paperURL)
	}

	paperID, err := r.extractPaperID(paperURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract paper ID: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", paperURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paper: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch paper: status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	title := r.extractTitle(doc)
	authors := r.extractAuthors(doc)
	publishedAt := r.extractPublishedAt(doc)
	content := r.extractContent(doc)

	paper := domain.NewPaper(
		domain.PaperID(paperID),
		title,
		authors,
		publishedAt,
		paperURL,
		content,
	)

	return paper, nil
}

func (r *IEEEPaperRepository) isIEEEURL(paperURL string) bool {
	u, err := url.Parse(paperURL)
	if err != nil {
		return false
	}
	return strings.Contains(u.Host, "ieeexplore.ieee.org")
}

func (r *IEEEPaperRepository) extractPaperID(paperURL string) (string, error) {
	re := regexp.MustCompile(`document/(\d+)`)
	matches := re.FindStringSubmatch(paperURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("paper ID not found in URL")
	}
	return matches[1], nil
}

func (r *IEEEPaperRepository) extractTitle(doc *goquery.Document) string {
	title := doc.Find("h1.document-title").Text()
	return strings.TrimSpace(title)
}

func (r *IEEEPaperRepository) extractAuthors(doc *goquery.Document) []string {
	var authors []string
	doc.Find(".authors-info .author").Each(func(i int, s *goquery.Selection) {
		author := strings.TrimSpace(s.Text())
		if author != "" {
			authors = append(authors, author)
		}
	})
	return authors
}

func (r *IEEEPaperRepository) extractPublishedAt(doc *goquery.Document) time.Time {
	dateText := doc.Find(".doc-abstract-pubdate").Text()
	dateText = strings.TrimSpace(strings.Replace(dateText, "Date of Publication:", "", 1))
	
	formats := []string{
		"02 January 2006",
		"January 2006",
		"2006",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateText); err == nil {
			return t
		}
	}
	
	return time.Time{}
}

func (r *IEEEPaperRepository) extractContent(doc *goquery.Document) string {
	var content strings.Builder
	
	abstract := doc.Find(".abstract-text").Text()
	if abstract != "" {
		content.WriteString("Abstract: ")
		content.WriteString(strings.TrimSpace(abstract))
		content.WriteString("\n\n")
	}
	
	keywords := doc.Find(".doc-keywords").Text()
	if keywords != "" {
		content.WriteString("Keywords: ")
		content.WriteString(strings.TrimSpace(keywords))
		content.WriteString("\n\n")
	}
	
	return content.String()
}