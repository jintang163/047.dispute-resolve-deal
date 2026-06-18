package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai"
)

type AIServiceImpl struct{}

func (s *AIServiceImpl) AIConsult(ctx context.Context, req *ai.AIConsultRequest) (resp *ai.AIConsultResponse, err error) {
	resp = &ai.AIConsultResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.Question) == "" {
		resp.Code = 400
		resp.Message = "问题不能为空"
		return resp, nil
	}

	similarResp, err := s.SearchSimilarLawArticles(ctx, &ai.SearchSimilarRequest{
		Question: req.Question,
		TopK:     5,
	})
	if err != nil {
		logger.Error("Search similar articles error", logger.Error(err))
	}

	relatedArticles := make([]*ai.LawArticle, 0)
	if similarResp != nil && len(similarResp.Articles) > 0 {
		relatedArticles = similarResp.Articles
	}

	answer := s.generateAIAnswer(req.Question, relatedArticles)

	consult := &model.AIConsultRecord{
		UserID:       req.UserId,
		Question:     req.Question,
		Answer:       answer,
		ArticleIDs:   s.extractArticleIDs(relatedArticles),
		TokenUsage:   0,
		ResponseTime: 0,
	}
	database.GetDB().Create(consult)

	resp.Answer = answer
	resp.RelatedArticles = relatedArticles

	return resp, nil
}

func (s *AIServiceImpl) GetLawArticles(ctx context.Context, req *ai.GetLawArticlesRequest) (resp *ai.GetLawArticlesResponse, err error) {
	resp = &ai.GetLawArticlesResponse{Code: 0, Message: "success"}

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
	resp.Articles = make([]*ai.LawArticle, len(articles))
	for i, a := range articles {
		resp.Articles[i] = &ai.LawArticle{
			Id:         a.ID,
			Title:      a.Title,
			Content:    a.Content,
			Category:   a.Category,
			LawName:    a.LawName,
			ArticleNo:  a.ArticleNo,
			Keywords:   a.Keywords,
			VectorId:   a.VectorID,
			Status:     int32(a.Status),
			CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *AIServiceImpl) CreateLawArticle(ctx context.Context, req *ai.CreateLawArticleRequest) (resp *ai.CreateLawArticleResponse, err error) {
	resp = &ai.CreateLawArticleResponse{Code: 0, Message: "success"}

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

func (s *AIServiceImpl) UpdateLawArticle(ctx context.Context, req *ai.UpdateLawArticleRequest) (resp *ai.UpdateLawArticleResponse, err error) {
	resp = &ai.UpdateLawArticleResponse{Code: 0, Message: "success"}

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

func (s *AIServiceImpl) DeleteLawArticle(ctx context.Context, req *ai.DeleteLawArticleRequest) (resp *ai.DeleteLawArticleResponse, err error) {
	resp = &ai.DeleteLawArticleResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.LawArticle{}).Where("id = ?", req.Id).Update("deleted_at", time.Now())
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "删除法条失败"
		return resp, nil
	}

	return resp, nil
}

func (s *AIServiceImpl) VectorizeLawArticles(ctx context.Context, req *ai.VectorizeLawArticlesRequest) (resp *ai.VectorizeLawArticlesResponse, err error) {
	resp = &ai.VectorizeLawArticlesResponse{Code: 0, Message: "success"}

	var articles []model.LawArticle
	db := database.GetDB().Where("deleted_at IS NULL")
	if len(req.Ids) > 0 {
		db = db.Where("id IN ?", req.Ids)
	}
	db.Find(&articles)

	processedCount := 0
	for _, article := range articles {
		vectorID := "vec_" + utils.GenerateIDStr()

		database.GetDB().Model(&article).Updates(map[string]interface{}{
			"vector_id": vectorID,
			"vectorized_at": time.Now(),
		})

		processedCount++
		logger.Info("Law article vectorized", logger.Int64("id", article.ID), logger.String("vectorId", vectorID))
	}

	resp.ProcessedCount = int32(processedCount)
	return resp, nil
}

func (s *AIServiceImpl) GenerateMediationSummary(ctx context.Context, req *ai.GenerateSummaryRequest) (resp *ai.GenerateSummaryResponse, err error) {
	resp = &ai.GenerateSummaryResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.MediationContent) == "" {
		resp.Code = 400
		resp.Message = "调解内容不能为空"
		return resp, nil
	}

	summary := s.generateSummary(req.CaseId, req.MediationContent)

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

func (s *AIServiceImpl) SearchSimilarLawArticles(ctx context.Context, req *ai.SearchSimilarRequest) (resp *ai.SearchSimilarResponse, err error) {
	resp = &ai.SearchSimilarResponse{Code: 0, Message: "success"}

	if strings.TrimSpace(req.Question) == "" {
		resp.Code = 400
		resp.Message = "问题不能为空"
		return resp, nil
	}

	keywords := s.extractKeywords(req.Question)

	var articles []model.LawArticle
	db := database.GetDB().Model(&model.LawArticle{}).Where("status = 1 AND deleted_at IS NULL")

	orConditions := make([]string, 0)
	params := make([]interface{}, 0)
	for _, kw := range keywords {
		orConditions = append(orConditions, "keywords LIKE ?")
		params = append(params, "%"+kw+"%")
	}

	if len(orConditions) > 0 {
		db = db.Where(strings.Join(orConditions, " OR "), params...)
	}

	limit := 10
	if req.TopK > 0 {
		limit = int(req.TopK)
	}
	db.Limit(limit).Find(&articles)

	resp.Articles = make([]*ai.LawArticle, len(articles))
	resp.Scores = make([]float64, len(articles))
	for i, a := range articles {
		resp.Articles[i] = &ai.LawArticle{
			Id:         a.ID,
			Title:      a.Title,
			Content:    a.Content,
			Category:   a.Category,
			LawName:    a.LawName,
			ArticleNo:  a.ArticleNo,
			Keywords:   a.Keywords,
			VectorId:   a.VectorID,
			Status:     int32(a.Status),
			CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		resp.Scores[i] = s.calculateSimilarity(req.Question, a.Content)
	}

	return resp, nil
}

func (s *AIServiceImpl) generateAIAnswer(question string, relatedArticles []*ai.LawArticle) string {
	var articleRefs string
	if len(relatedArticles) > 0 {
		articleRefs = "\n\n相关法律依据："
		for i, art := range relatedArticles {
			articleRefs += fmt.Sprintf("\n%d. 《%s》%s：%s", i+1, art.LawName, art.ArticleNo, art.Title)
		}
	}

	prefix := "根据您的问题，以下是相关法律建议：\n\n"

	if strings.Contains(question, "离婚") || strings.Contains(question, "婚姻") {
		return prefix + "关于婚姻家庭纠纷，建议您先尝试协商解决。如果无法协商，可以向人民调解委员会申请调解，或者直接向人民法院提起诉讼。" +
			"根据《民法典》相关规定，夫妻双方自愿离婚的，应当签订书面离婚协议，并亲自到婚姻登记机关申请离婚登记。" +
			"涉及财产分割和子女抚养问题的，应当按照照顾子女、女方和无过错方权益的原则处理。" + articleRefs
	}

	if strings.Contains(question, "借贷") || strings.Contains(question, "借款") || strings.Contains(question, "欠钱") {
		return prefix + "关于民间借贷纠纷，建议您收集相关证据（借条、转账记录、聊天记录等），先与对方协商还款。" +
			"如果协商不成，可以向人民法院提起诉讼。根据《民法典》第六百六十七条规定，借款合同是借款人向贷款人借款，到期返还借款并支付利息的合同。" +
			"注意诉讼时效为三年，从约定的还款期限届满之日起计算。" + articleRefs
	}

	if strings.Contains(question, "物业") || strings.Contains(question, "小区") {
		return prefix + "关于物业纠纷，建议您先与物业公司沟通协商，也可以向业主委员会反映情况。" +
			"如果无法解决，可以向街道办事处或住建部门投诉，或者通过诉讼途径解决。" +
			"根据《物业管理条例》，业主应当按照约定支付物业费，物业公司应当按照物业服务合同的约定提供相应服务。" + articleRefs
	}

	if strings.Contains(question, "劳动") || strings.Contains(question, "工资") || strings.Contains(question, "辞退") {
		return prefix + "关于劳动争议，建议您先与用人单位协商解决。协商不成的，可以向劳动争议仲裁委员会申请仲裁。" +
			"对仲裁裁决不服的，可以向人民法院提起诉讼。根据《劳动合同法》，用人单位应当按照劳动合同约定和国家规定，向劳动者及时足额支付劳动报酬。" +
			"用人单位违法解除劳动合同的，应当支付经济赔偿金。" + articleRefs
	}

	return prefix + "您好！感谢您的咨询。根据您描述的情况，建议您：\n\n" +
		"1. 首先收集和保存好相关证据材料（合同、单据、聊天记录等）\n" +
		"2. 尝试与对方友好协商，争取达成和解\n" +
		"3. 协商不成的，可以向所在地人民调解委员会申请调解\n" +
		"4. 也可以向相关行政主管部门投诉举报\n" +
		"5. 必要时通过诉讼或仲裁途径维护合法权益\n\n" +
		"如需更详细的法律建议，建议您携带相关材料到当地司法所或律师事务所现场咨询。" + articleRefs
}

func (s *AIServiceImpl) generateSummary(caseID int64, content string) string {
	var caseData model.DisputeCase
	database.GetDB().Select("case_no, title, applicant_name, respondent_name").Where("id = ?", caseID).First(&caseData)

	summary := fmt.Sprintf("调解摘要\n\n案件编号：%s\n案件名称：%s\n", caseData.CaseNo, caseData.Title)
	if caseData.ApplicantName != "" {
		summary += fmt.Sprintf("申请人：%s\n", caseData.ApplicantName)
	}
	if caseData.RespondentName != "" {
		summary += fmt.Sprintf("被申请人：%s\n", caseData.RespondentName)
	}
	summary += fmt.Sprintf("\n调解时间：%s\n", time.Now().Format("2006年01月02日 15:04"))
	summary += "\n调解内容：\n" + s.extractKeyPoints(content) + "\n\n"
	summary += "调解结果：双方自愿达成协议，争议事项已妥善解决。"

	return summary
}

func (s *AIServiceImpl) extractKeywords(question string) []string {
	keywords := []string{"合同", "侵权", "违约", "赔偿", "离婚", "借贷", "劳动", "物业", "交通", "医疗", "房产", "继承"}
	result := make([]string, 0)
	for _, kw := range keywords {
		if strings.Contains(question, kw) {
			result = append(result, kw)
		}
	}
	if len(result) == 0 {
		result = append(result, "纠纷")
	}
	return result
}

func (s *AIServiceImpl) extractKeyPoints(content string) string {
	lines := strings.Split(content, "\n")
	points := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) > 5 {
			if len(points) < 5 {
				points = append(points, "• "+line)
			}
		}
	}
	if len(points) == 0 {
		if len(content) > 200 {
			return content[:200] + "..."
		}
		return content
	}
	return strings.Join(points, "\n")
}

func (s *AIServiceImpl) calculateSimilarity(text1, text2 string) float64 {
	k1 := s.extractKeywords(text1)
	k2 := s.extractKeywords(text2)

	intersection := 0
	for _, a := range k1 {
		for _, b := range k2 {
			if a == b {
				intersection++
				break
			}
		}
	}

	union := len(k1) + len(k2) - intersection
	if union == 0 {
		return 0.5
	}
	return float64(intersection) / float64(union)
}

func (s *AIServiceImpl) extractArticleIDs(articles []*ai.LawArticle) string {
	ids := make([]string, len(articles))
	for i, a := range articles {
		ids[i] = fmt.Sprintf("%d", a.Id)
	}
	return strings.Join(ids, ",")
}
