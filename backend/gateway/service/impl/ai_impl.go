package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type AIServiceImpl struct{}

func NewAIService() service.AIService {
	return &AIServiceImpl{}
}

func (s *AIServiceImpl) AIConsult(ctx context.Context, question string, userID int64) (string, []*model.LawArticle, error) {
	articles, scores, err := s.SearchSimilarLawArticles(ctx, question, 5)
	if err != nil {
		logger.Error("Search similar law articles error", logger.Error(err))
	}

	prompt := buildConsultPrompt(question, articles)
	answer, err := callDeepSeek(prompt)
	if err != nil {
		logger.Error("Call DeepSeek error", logger.Error(err))
		return "", nil, err
	}

	consultRecord := &model.AIConsultRecord{
		UserID:      userID,
		Question:    question,
		Answer:      answer,
		RelatedArticles: getArticleIDs(articles),
	}
	database.GetDB().Create(consultRecord)

	return answer, articles, nil
}

func (s *AIServiceImpl) GetLawArticles(ctx context.Context, page, pageSize int, keyword string, category string) ([]*model.LawArticle, int64, error) {
	var articles []*model.LawArticle
	var total int64

	db := database.GetDB().Model(&model.LawArticle{}).Where("status = 1 AND deleted_at IS NULL")

	if keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ? OR keywords LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&articles)

	return articles, total, nil
}

func (s *AIServiceImpl) CreateLawArticle(ctx context.Context, article *model.LawArticle) error {
	return database.GetDB().Create(article).Error
}

func (s *AIServiceImpl) UpdateLawArticle(ctx context.Context, article *model.LawArticle) error {
	return database.GetDB().Model(article).Omit("created_at", "vector_id").Updates(article).Error
}

func (s *AIServiceImpl) DeleteLawArticle(ctx context.Context, id int64) error {
	return database.GetDB().Model(&model.LawArticle{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (s *AIServiceImpl) VectorizeLawArticles(ctx context.Context, ids []int64) (int, error) {
	var articles []model.LawArticle
	db := database.GetDB().Where("status = 1 AND deleted_at IS NULL")
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}
	db.Find(&articles)

	processed := 0
	for _, article := range articles {
		vectorID, err := insertToMilvus(article.Title + "\n" + article.Content)
		if err != nil {
			logger.Error("Insert to Milvus error", logger.Error(err))
			continue
		}

		database.GetDB().Model(&article).Update("vector_id", vectorID)
		processed++
	}

	return processed, nil
}

func (s *AIServiceImpl) GenerateMediationSummary(ctx context.Context, caseID int64, mediationContent string) (string, error) {
	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", caseID).First(&caseData)

	prompt := buildSummaryPrompt(caseData, mediationContent)
	summary, err := callDeepSeek(prompt)
	if err != nil {
		logger.Error("Call DeepSeek error", logger.Error(err))
		return "", err
	}

	return summary, nil
}

func (s *AIServiceImpl) SearchSimilarLawArticles(ctx context.Context, question string, topK int) ([]*model.LawArticle, []float64, error) {
	vectorIDs, scores, err := searchMilvus(question, topK)
	if err != nil {
		logger.Error("Search Milvus error", logger.Error(err))
		return nil, nil, err
	}

	if len(vectorIDs) == 0 {
		return nil, nil, nil
	}

	var articles []*model.LawArticle
	database.GetDB().Where("vector_id IN ?", vectorIDs).Find(&articles)

	articleMap := make(map[string]*model.LawArticle)
	for _, a := range articles {
		articleMap[a.VectorID] = a
	}

	sortedArticles := make([]*model.LawArticle, 0, len(vectorIDs))
	for _, vid := range vectorIDs {
		if article, ok := articleMap[vid]; ok {
			sortedArticles = append(sortedArticles, article)
		}
	}

	return sortedArticles, scores, nil
}

func (s *AIServiceImpl) GetAIConfig(ctx context.Context) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"deepseek": map[string]interface{}{
			"apiKey":    "***",
			"model":     "deepseek-chat",
			"maxTokens": 2000,
			"temperature": 0.7,
		},
		"milvus": map[string]interface{}{
			"host":      "localhost",
			"port":      19530,
			"collection": "law_articles",
			"dimension": 1536,
		},
		"enabled": true,
	}
	return config, nil
}

func (s *AIServiceImpl) UpdateAIConfig(ctx context.Context, config map[string]interface{}) error {
	configJSON, _ := json.Marshal(config)
	cache.Set(ctx, constants.RedisPrefixAIConfig, string(configJSON), 24*time.Hour)
	return nil
}

func buildConsultPrompt(question string, articles []*model.LawArticle) string {
	var articleTexts []string
	for i, a := range articles {
		articleTexts = append(articleTexts,
			fmt.Sprintf("[%d] %s\n%s\n", i+1, a.Title, a.Content))
	}

	return fmt.Sprintf(`你是一位专业的法律咨询助手。请根据以下问题和相关法条，给出专业的法律建议。

问题：%s

相关法条：
%s

请按照以下格式回答：
1. 首先简要分析问题的法律性质
2. 引用相关法条（标注序号）
3. 给出具体的法律建议和解决方案
4. 提示可能的风险和注意事项

请注意：
- 回答要专业、准确、易懂
- 必须引用提供的法条
- 不要编造法条内容
- 如涉及诉讼，告知诉讼程序和时效
`, question, strings.Join(articleTexts, "\n"))
}

func buildSummaryPrompt(caseData model.DisputeCase, mediationContent string) string {
	return fmt.Sprintf(`请根据以下纠纷案件信息和调解记录，生成一份专业的调解摘要。

案件信息：
案件编号：%s
案件标题：%s
纠纷类型：%s
申请人：%s
被申请人：%s
纠纷描述：%s

调解记录：
%s

请生成调解摘要，包含以下内容：
1. 案件基本情况概述
2. 调解过程简述
3. 双方争议焦点
4. 调解结果或进展
5. 后续建议

要求：
- 语言简洁，专业
- 重点突出，逻辑清晰
- 字数控制在300-500字
`, caseData.CaseNo, caseData.Title, caseData.TypeName, caseData.ApplicantName,
		caseData.RespondentName, caseData.Description, mediationContent)
}

func callDeepSeek(prompt string) (string, error) {
	mq.SendAsync(constants.MQTopicAITask, map[string]interface{}{
		"type":   "deepseek_call",
		"prompt": prompt,
		"time":   time.Now(),
	})

	return "这是AI生成的模拟回复。在实际部署中，这里会调用DeepSeek API生成真实的法律建议。\n\n根据您的问题，我们建议您：\n1. 首先收集相关证据\n2. 可以先尝试与对方协商解决\n3. 如协商不成，可向当地人民调解委员会申请调解\n4. 也可以直接向人民法院提起诉讼\n\n相关法条参考：《民法典》相关条款", nil
}

func insertToMilvus(content string) (string, error) {
	return utils.GenerateIDStr(), nil
}

func searchMilvus(question string, topK int) ([]string, []float64, error) {
	return []string{}, []float64{}, nil
}

func getArticleIDs(articles []*model.LawArticle) string {
	ids := make([]string, 0, len(articles))
	for _, a := range articles {
		ids = append(ids, utils.Int64ToString(a.ID))
	}
	return strings.Join(ids, ",")
}
