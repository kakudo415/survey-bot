package usecase

import (
	"context"
	"fmt"

	"github.com/kakudo415/survey-bot/domain"
)

type SurveyUseCase struct {
	paperRepo   domain.PaperRepository
	summaryService domain.SummaryService
}

func NewSurveyUseCase(paperRepo domain.PaperRepository, summaryService domain.SummaryService) *SurveyUseCase {
	return &SurveyUseCase{
		paperRepo:   paperRepo,
		summaryService: summaryService,
	}
}

func (uc *SurveyUseCase) ProcessPaper(ctx context.Context, paperURL string) (*domain.Paper, *domain.Summary, error) {
	paper, err := uc.paperRepo.FindByURL(ctx, paperURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch paper: %w", err)
	}

	summary, err := uc.summaryService.GenerateSummary(ctx, paper)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	return paper, summary, nil
}