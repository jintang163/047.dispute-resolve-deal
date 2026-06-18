package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type AIService interface {
	AIConsult(ctx context.Context, question string, userID int64) (string, []*model.LawArticle, error)
	GetLawArticles(ctx context.Context, page, pageSize int, keyword string, category string) ([]*model.LawArticle, int64, error)
	CreateLawArticle(ctx context.Context, article *model.LawArticle) error
	UpdateLawArticle(ctx context.Context, article *model.LawArticle) error
	DeleteLawArticle(ctx context.Context, id int64) error
	VectorizeLawArticles(ctx context.Context, ids []int64) (int, error)
	GenerateMediationSummary(ctx context.Context, caseID int64, mediationContent string) (string, error)
	SearchSimilarLawArticles(ctx context.Context, question string, topK int) ([]*model.LawArticle, []float64, error)
	GetAIConfig(ctx context.Context) (map[string]interface{}, error)
	UpdateAIConfig(ctx context.Context, config map[string]interface{}) error
}
