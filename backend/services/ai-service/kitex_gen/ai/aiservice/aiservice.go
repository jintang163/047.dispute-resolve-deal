package aiservice

import (
	"context"

	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type AIService interface {
	AIConsult(ctx context.Context, request *ai.AIConsultRequest) (r *ai.AIConsultResponse, err error)
	GetLawArticles(ctx context.Context, request *ai.GetLawArticlesRequest) (r *ai.GetLawArticlesResponse, err error)
	CreateLawArticle(ctx context.Context, request *ai.CreateLawArticleRequest) (r *ai.CreateLawArticleResponse, err error)
	UpdateLawArticle(ctx context.Context, request *ai.UpdateLawArticleRequest) (r *ai.UpdateLawArticleResponse, err error)
	DeleteLawArticle(ctx context.Context, request *ai.DeleteLawArticleRequest) (r *ai.DeleteLawArticleResponse, err error)
	VectorizeLawArticles(ctx context.Context, request *ai.VectorizeLawArticlesRequest) (r *ai.VectorizeLawArticlesResponse, err error)
	GenerateMediationSummary(ctx context.Context, request *ai.GenerateSummaryRequest) (r *ai.GenerateSummaryResponse, err error)
	SearchSimilarLawArticles(ctx context.Context, request *ai.SearchSimilarRequest) (r *ai.SearchSimilarResponse, err error)
}

type Server interface {
	RegisterService(svc *serviceinfo.ServiceInfo)
}

type Client interface {
	AIService
}
