package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/common/vector"
	ai_kitex "github.com/dispute-resolve/ai-service/kitex_gen/ai"
	"go.uber.org/zap"
)

type AIServiceImpl struct{}

func (s *AIServiceImpl) AIConsult(ctx context.Context, req *ai_kitex.AIConsultRequest) (resp *ai_kitex.AIConsultResponse, err error) {
	resp = &ai_kitex.AIConsultResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.Question) == "" {
		resp.Code = 400
		resp.Message = "问题不能为空"
		return resp, nil
	}

	startTime := time.Now()

	answer, err := ai.GetLegalAdvice(req.Question, nil)
	if err != nil {
		logger.Error("Get legal advice failed",
			zap.String("question", req.Question),
			logger.Error(err),
		)
		resp.Code = 500
		resp.Message = "获取法律建议失败"
		return resp, nil
	}

	elapsed := time.Since(startTime).Milliseconds()

	relatedArticles := make([]*ai_kitex.LawArticle, 0)
	articleIDs := make([]string, 0)

	if len(answer.RelatedArticles) > 0 {
		for _, ref := range answer.RelatedArticles {
			lawArticle := convertLawReferenceToKitex(ref)
			relatedArticles = append(relatedArticles, lawArticle)
			if lawArticle.Id > 0 {
				articleIDs = append(articleIDs, fmt.Sprintf("%d", lawArticle.Id))
			}
		}
	}

	if len(relatedArticles) == 0 {
		similarResp, searchErr := s.SearchSimilarLawArticles(ctx, &ai_kitex.SearchSimilarRequest{
			Question: req.Question,
			TopK:     5,
		})
		if searchErr != nil {
			logger.Warn("Search similar articles fallback failed", logger.Error(searchErr))
		}
		if similarResp != nil && len(similarResp.Articles) > 0 {
			relatedArticles = similarResp.Articles
		}
	}

	resp.Answer = answer.Answer
	resp.RelatedArticles = relatedArticles

	consult := &model.AIConsultRecord{
		UserID:       req.UserId,
		Question:     req.Question,
		Answer:       answer.Answer,
		ArticleIDs:   strings.Join(articleIDs, ","),
		TokenUsage:   0,
		ResponseTime: int(elapsed),
	}
	database.GetDB().Create(consult)

	return resp, nil
}

func (s *AIServiceImpl) GetLawArticles(ctx context.Context, req *ai_kitex.GetLawArticlesRequest) (resp *ai_kitex.GetLawArticlesResponse, err error) {
	resp = &ai_kitex.GetLawArticlesResponse{Code: 0, Message: "success"}

	var articles []model.LawArticle
	var total int64

	db := database.GetDB().Model(&model.LawArticle{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ? OR keywords LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}

	db.Count(&total)

	offset := int((req.Page - 1) * req.PageSize)
	db.Offset(offset).Limit(int(req.PageSize)).Order("id DESC").Find(&articles)

	resp.Total = total
	resp.Articles = make([]*ai_kitex.LawArticle, len(articles))
	for i, a := range articles {
		resp.Articles[i] = convertModelLawArticleToKitex(&a)
	}

	return resp, nil
}

func (s *AIServiceImpl) CreateLawArticle(ctx context.Context, req *ai_kitex.CreateLawArticleRequest) (resp *ai_kitex.CreateLawArticleResponse, err error) {
	resp = &ai_kitex.CreateLawArticleResponse{Code: 0, Message: "success"}

	article := &model.LawArticle{
		Title:     req.Article.Title,
		Content:   req.Article.Content,
		Category:  req.Article.Category,
		LawName:   req.Article.LawName,
		ArticleNo: req.Article.ArticleNo,
		Keywords:  req.Article.Keywords,
		Status:    1,
	}

	result := database.GetDB().Create(article)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "创建法条失败"
		logger.Error("Create law article error", logger.Error(result.Error))
		return resp, nil
	}

	resp.Id = article.ID
	return resp, nil
}

func (s *AIServiceImpl) UpdateLawArticle(ctx context.Context, req *ai_kitex.UpdateLawArticleRequest) (resp *ai_kitex.UpdateLawArticleResponse, err error) {
	resp = &ai_kitex.UpdateLawArticleResponse{Code: 0, Message: "success"}

	updates := map[string]interface{}{
		"title":      req.Article.Title,
		"content":    req.Article.Content,
		"category":   req.Article.Category,
		"law_name":   req.Article.LawName,
		"article_no": req.Article.ArticleNo,
		"keywords":   req.Article.Keywords,
		"status":     req.Article.Status,
	}

	result := database.GetDB().Model(&model.LawArticle{}).Where("id = ?", req.Article.Id).Updates(updates)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "更新法条失败"
		return resp, nil
	}

	return resp, nil
}

func (s *AIServiceImpl) DeleteLawArticle(ctx context.Context, req *ai_kitex.DeleteLawArticleRequest) (resp *ai_kitex.DeleteLawArticleResponse, err error) {
	resp = &ai_kitex.DeleteLawArticleResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.LawArticle{}).Where("id = ?", req.Id).Update("deleted_at", time.Now())
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "删除法条失败"
		return resp, nil
	}

	go func() {
		if err := vector.DeleteByLawID(req.Id); err != nil {
			logger.Warn("Delete vector for law article failed",
				zap.Int64("lawId", req.Id),
				logger.Error(err),
			)
		}
	}()

	return resp, nil
}

func (s *AIServiceImpl) VectorizeLawArticles(ctx context.Context, req *ai_kitex.VectorizeLawArticlesRequest) (resp *ai_kitex.VectorizeLawArticlesResponse, err error) {
	resp = &ai_kitex.VectorizeLawArticlesResponse{Code: 0, Message: "success"}

	var articles []model.LawArticle
	db := database.GetDB().Where("deleted_at IS NULL")
	if len(req.Ids) > 0 {
		db = db.Where("id IN ?", req.Ids)
	} else {
		db = db.Where("vector_id IS NULL OR vector_id = ''")
	}
	db.Find(&articles)

	if len(articles) == 0 {
		resp.ProcessedCount = 0
		return resp, nil
	}

	processedCount := 0
	batchSize := 10

	for i := 0; i < len(articles); i += batchSize {
		end := i + batchSize
		if end > len(articles) {
			end = len(articles)
		}
		batch := articles[i:end]

		texts := make([]string, len(batch))
		for j, a := range batch {
			texts[j] = fmt.Sprintf("%s %s %s %s", a.LawName, a.ArticleNo, a.Title, a.Content)
		}

		embeddings, embedErr := vector.GetBatchEmbeddings(texts)
		if embedErr != nil {
			logger.Error("Batch embeddings failed",
				zap.Int("batchStart", i),
				logger.Error(embedErr),
			)
			continue
		}

		ids := make([]int64, len(batch))
		vectors := make([][]float32, len(batch))
		metadata := make([]map[string]interface{}, len(batch))

		for j, a := range batch {
			if j >= len(embeddings) || len(embeddings[j]) == 0 {
				continue
			}

			vectorID := "vec_" + utils.GenerateIDStr()
			ids[j] = a.ID
			vectors[j] = embeddings[j]
			metadata[j] = map[string]interface{}{
				"law_id":   a.ID,
				"content":  truncateContent(a.Content, 2000),
				"keywords": a.Keywords,
			}

			database.GetDB().Model(&a).Updates(map[string]interface{}{
				"vector_id":     vectorID,
				"vectorized_at": time.Now(),
			})

			processedCount++
			logger.Info("Law article vectorized",
				zap.Int64("id", a.ID),
				zap.String("vectorId", vectorID),
			)
		}

		validIds := make([]int64, 0)
		validVectors := make([][]float32, 0)
		validMetadata := make([]map[string]interface{}, 0)
		for j := range ids {
			if len(vectors[j]) > 0 {
				validIds = append(validIds, ids[j])
				validVectors = append(validVectors, vectors[j])
				validMetadata = append(validMetadata, metadata[j])
			}
		}

		if len(validIds) > 0 {
			if insertErr := vector.InsertVectors(validIds, validVectors, validMetadata); insertErr != nil {
				logger.Error("Insert vectors to milvus failed",
					zap.Int("count", len(validIds)),
					logger.Error(insertErr),
				)
			}
		}
	}

	resp.ProcessedCount = int32(processedCount)
	return resp, nil
}

func (s *AIServiceImpl) GenerateMediationSummary(ctx context.Context, req *ai_kitex.GenerateSummaryRequest) (resp *ai_kitex.GenerateSummaryResponse, err error) {
	resp = &ai_kitex.GenerateSummaryResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.MediationContent) == "" {
		resp.Code = 400
		resp.Message = "调解内容不能为空"
		return resp, nil
	}

	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", req.CaseId).First(&caseData)

	caseInfo := map[string]interface{}{
		"caseId":         req.CaseId,
		"caseNo":         caseData.CaseNo,
		"title":          caseData.Title,
		"typeName":       caseData.TypeName,
		"applicantName":  caseData.ApplicantName,
		"applicantPhone": caseData.ApplicantPhone,
		"respondentName": caseData.RespondentName,
		"respondentPhone": caseData.RespondentPhone,
		"description":    caseData.Description,
		"mediatorName":   caseData.MediatorName,
		"level":          caseData.Level,
	}

	summary, summaryErr := ai.GenerateMediationSummary(caseInfo, req.MediationContent)
	if summaryErr != nil {
		logger.Error("Generate mediation summary failed",
			zap.Int64("caseId", req.CaseId),
			logger.Error(summaryErr),
		)
		resp.Code = 500
		resp.Message = "生成调解摘要失败"
		return resp, nil
	}

	record := &model.AIAssistRecord{
		CaseID:     req.CaseId,
		AssistType: 1,
		Input:      req.MediationContent,
		Output:     summary,
		TokenUsage: 0,
	}
	database.GetDB().Create(record)

	resp.Summary = summary
	return resp, nil
}

func (s *AIServiceImpl) SearchSimilarLawArticles(ctx context.Context, req *ai_kitex.SearchSimilarRequest) (resp *ai_kitex.SearchSimilarResponse, err error) {
	resp = &ai_kitex.SearchSimilarResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.Question) == "" {
		resp.Code = 400
		resp.Message = "问题不能为空"
		return resp, nil
	}

	topK := int(req.TopK)
	if topK <= 0 {
		topK = 10
	}

	queryVector, embedErr := vector.GetEmbedding(req.Question)
	if embedErr != nil {
		logger.Warn("Get query embedding failed, fallback to keyword search",
			logger.Error(embedErr),
		)
		return s.keywordSearch(req.Question, topK)
	}

	searchResults, searchErr := vector.SearchVectors(queryVector, topK, "")
	if searchErr != nil {
		logger.Warn("Milvus search failed, fallback to keyword search",
			logger.Error(searchErr),
		)
		return s.keywordSearch(req.Question, topK)
	}

	resp.Articles = make([]*ai_kitex.LawArticle, 0)
	resp.Scores = make([]float64, 0)

	lawIDs := make([]int64, 0)
	scoreMap := make(map[int64]float64)
	contentMap := make(map[int64]string)

	for _, r := range searchResults {
		if r.LawID > 0 {
			lawIDs = append(lawIDs, r.LawID)
			scoreMap[r.LawID] = float64(r.Score)
			contentMap[r.LawID] = r.Content
		}
	}

	if len(lawIDs) > 0 {
		var lawArticles []model.LawArticle
		database.GetDB().Where("id IN ?", lawIDs).Find(&lawArticles)

		lawArticleMap := make(map[int64]*model.LawArticle)
		for i := range lawArticles {
			lawArticleMap[lawArticles[i].ID] = &lawArticles[i]
		}

		for _, lawID := range lawIDs {
			if article, ok := lawArticleMap[lawID]; ok {
				kitexArticle := convertModelLawArticleToKitex(article)
				if content, hasContent := contentMap[lawID]; hasContent && kitexArticle.Content == "" {
					kitexArticle.Content = content
				}
				resp.Articles = append(resp.Articles, kitexArticle)
				if score, hasScore := scoreMap[lawID]; hasScore {
					resp.Scores = append(resp.Scores, score)
				}
			}
		}
	}

	if len(resp.Articles) == 0 {
		logger.Info("Vector search returned no results, fallback to keyword search")
		return s.keywordSearch(req.Question, topK)
	}

	logger.Info("Similar law articles search completed",
		zap.Int("count", len(resp.Articles)),
		zap.String("method", "vector"),
	)

	return resp, nil
}

func (s *AIServiceImpl) keywordSearch(question string, limit int) (*ai_kitex.SearchSimilarResponse, error) {
	resp := &ai_kitex.SearchSimilarResponse{Code: 0, Message: "success"}

	keywords := extractKeywords(question)

	var articles []model.LawArticle
	db := database.GetDB().Model(&model.LawArticle{}).Where("status = 1 AND deleted_at IS NULL")

	orConditions := make([]string, 0)
	params := make([]interface{}, 0)
	for _, kw := range keywords {
		orConditions = append(orConditions, "keywords LIKE ?")
		params = append(params, "%"+kw+"%")
	}
	orConditions = append(orConditions, "title LIKE ?")
	params = append(params, "%"+question+"%")
	orConditions = append(orConditions, "content LIKE ?")
	params = append(params, "%"+question+"%")

	db = db.Where(strings.Join(orConditions, " OR "), params...)
	db.Limit(limit).Find(&articles)

	resp.Articles = make([]*ai_kitex.LawArticle, len(articles))
	resp.Scores = make([]float64, len(articles))
	for i, a := range articles {
		resp.Articles[i] = convertModelLawArticleToKitex(&a)
		resp.Scores[i] = calculateKeywordScore(question, keywords, &a)
	}

	logger.Info("Similar law articles search completed",
		zap.Int("count", len(resp.Articles)),
		zap.String("method", "keyword"),
	)

	return resp, nil
}

func convertLawReferenceToKitex(ref ai.LawReference) *ai_kitex.LawArticle {
	var lawID int64
	parsedID, err := strconv.ParseInt(strings.TrimSpace(ref.ArticleNo), 10, 64)
	if err == nil {
		lawID = parsedID
	}

	return &ai_kitex.LawArticle{
		Id:        lawID,
		Title:     fmt.Sprintf("%s %s", ref.LawName, ref.ArticleNo),
		Content:   ref.Content,
		LawName:   ref.LawName,
		ArticleNo: ref.ArticleNo,
		Keywords:  "",
		Status:    1,
		CreatedAt: "",
	}
}

func convertModelLawArticleToKitex(a *model.LawArticle) *ai_kitex.LawArticle {
	createdAt := ""
	if !a.CreatedAt.IsZero() {
		createdAt = a.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return &ai_kitex.LawArticle{
		Id:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		Category:  a.Category,
		LawName:   a.LawName,
		ArticleNo: a.ArticleNo,
		Keywords:  a.Keywords,
		VectorId:  a.VectorID,
		Status:    int32(a.Status),
		CreatedAt: createdAt,
	}
}

func extractKeywords(question string) []string {
	keywordList := []string{"合同", "侵权", "违约", "赔偿", "离婚", "借贷", "劳动", "物业", "交通", "医疗", "房产", "继承", "消费", "租赁", "担保", "保险", "知识产权", "土地", "建设工程", "合伙", "公司", "担保", "票据", "海商", "纠纷"}
	result := make([]string, 0)
	for _, kw := range keywordList {
		if strings.Contains(question, kw) {
			result = append(result, kw)
		}
	}
	if len(result) == 0 {
		result = append(result, "纠纷")
	}
	return result
}

func calculateKeywordScore(question string, keywords []string, article *model.LawArticle) float64 {
	score := 0.0
	total := len(keywords) + 2

	for _, kw := range keywords {
		if strings.Contains(article.Keywords, kw) {
			score += 1.0
		}
		if strings.Contains(article.Title, kw) {
			score += 0.5
		}
		if strings.Contains(article.Content, kw) {
			score += 0.3
		}
	}

	for _, kw := range keywords {
		if strings.Contains(question, kw) {
			score += 0.2
		}
	}

	if score > 0 {
		return score / float64(total)
	}
	return 0.1
}

func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen])
}
