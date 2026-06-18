package aiservice

import (
	"context"

	clientpkg "github.com/cloudwego/kitex/client"

	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai"
)

type kClient struct {
	c clientpkg.Client
}

func NewClient(destService string, opts ...clientpkg.Option) (Client, error) {
	cli, err := clientpkg.NewClient(destService, opts...)
	if err != nil {
		return nil, err
	}
	return &kClient{c: cli}, nil
}

func (c *kClient) AIConsult(ctx context.Context, request *ai.AIConsultRequest) (r *ai.AIConsultResponse, err error) {
	r = &ai.AIConsultResponse{Code: 500, Message: "RPC not implemented - using local service fallback"}
	return
}

func (c *kClient) GetLawArticles(ctx context.Context, request *ai.GetLawArticlesRequest) (r *ai.GetLawArticlesResponse, err error) {
	r = &ai.GetLawArticlesResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) CreateLawArticle(ctx context.Context, request *ai.CreateLawArticleRequest) (r *ai.CreateLawArticleResponse, err error) {
	r = &ai.CreateLawArticleResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UpdateLawArticle(ctx context.Context, request *ai.UpdateLawArticleRequest) (r *ai.UpdateLawArticleResponse, err error) {
	r = &ai.UpdateLawArticleResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) DeleteLawArticle(ctx context.Context, request *ai.DeleteLawArticleRequest) (r *ai.DeleteLawArticleResponse, err error) {
	r = &ai.DeleteLawArticleResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) VectorizeLawArticles(ctx context.Context, request *ai.VectorizeLawArticlesRequest) (r *ai.VectorizeLawArticlesResponse, err error) {
	r = &ai.VectorizeLawArticlesResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GenerateMediationSummary(ctx context.Context, request *ai.GenerateSummaryRequest) (r *ai.GenerateSummaryResponse, err error) {
	r = &ai.GenerateSummaryResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) SearchSimilarLawArticles(ctx context.Context, request *ai.SearchSimilarRequest) (r *ai.SearchSimilarResponse, err error) {
	r = &ai.SearchSimilarResponse{Code: 500, Message: "RPC not implemented"}
	return
}
