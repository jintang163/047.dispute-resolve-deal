package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/vector"
	"go.uber.org/zap"
)

const (
	LegalSystemPrompt = `你是一名专业的中国法律顾问，擅长解答民事纠纷相关问题。
回答请引用具体法条，给出清晰的法律建议。
请严格按照以下JSON格式输出答案：
{
  "answer": "详细的法律解答内容",
  "relatedArticles": [
    {"lawName": "法律名称", "articleNo": "条款号", "content": "条款内容", "similarity": 0.95}
  ],
  "references": ["引用来源1", "引用来源2"],
  "keywords": ["关键词1", "关键词2"]
}`

	KeywordExtractionPrompt = `请从以下法律问题中提取最重要的3-5个中文关键词，仅返回关键词，用逗号分隔：
问题：%s
关键词：`
)

func GetLegalAdvice(question string, history []ChatMessage) (*AIAnswer, error) {
	if strings.TrimSpace(question) == "" {
		return nil, fmt.Errorf("问题不能为空")
	}

	logger.Info("Start legal advice",
		zap.String("question", question),
		zap.Int("historyCount", len(history)),
	)

	client := GetDeepSeekClient()

	keywords, err := extractKeywords(client, question)
	if err != nil {
		logger.Warn("Extract keywords failed, use empty keywords", logger.Error(err))
		keywords = []string{}
	}

	logger.Debug("Extracted keywords", zap.Strings("keywords", keywords))

	queryVector, err := GetTextEmbedding(question)
	if err != nil {
		logger.Warn("Get question embedding failed", logger.Error(err))
	}

	var relatedArticles []LawReference
	var vectorResults []*vector.SearchResult
	if queryVector != nil {
		vectorResults, err = vector.SearchVectors(queryVector, 5, "")
		if err != nil {
			logger.Warn("Search vectors failed", logger.Error(err))
		}
	}

	if len(vectorResults) > 0 {
		relatedArticles = make([]LawReference, 0, len(vectorResults))
		for _, r := range vectorResults {
			relatedArticles = append(relatedArticles, LawReference{
				LawName:    extractLawName(r.Content),
				ArticleNo:  "",
				Content:    r.Content,
				Similarity: float64(r.Score),
			})
		}
	}

	context := buildContext(question, keywords, relatedArticles)

	messages := buildMessages(question, history, context)

	rawAnswer, err := client.ChatCompletion(messages, "")
	if err != nil {
		logger.Error("Get AI answer failed", logger.Error(err))
		return fallbackAnswer(question, relatedArticles), nil
	}

	answer, err := parseAIAnswer(rawAnswer, relatedArticles)
	if err != nil {
		logger.Warn("Parse AI answer failed, use raw text", logger.Error(err))
		answer = &AIAnswer{
			Answer:          rawAnswer,
			RelatedArticles: relatedArticles,
			Keywords:        keywords,
		}
	}

	logger.Info("Legal advice completed",
		zap.Int("relatedArticles", len(answer.RelatedArticles)),
		zap.Int("keywords", len(answer.Keywords)),
	)

	return answer, nil
}

func extractKeywords(client *DeepSeekClient, question string) ([]string, error) {
	prompt := fmt.Sprintf(KeywordExtractionPrompt, question)

	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}

	result, err := client.ChatCompletion(messages, "你是一个专业的关键词提取助手。")
	if err != nil {
		return nil, err
	}

	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "关键词：")
	result = strings.TrimPrefix(result, "关键词:")
	parts := strings.Split(result, ",")
	keywords := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			keywords = append(keywords, p)
		}
	}

	if len(keywords) == 0 {
		keywords = []string{"纠纷"}
	}

	return keywords, nil
}

func GetTextEmbedding(text string) ([]float32, error) {
	return GetEmbedding(text)
}

func buildContext(question string, keywords []string, articles []LawReference) string {
	var sb strings.Builder

	sb.WriteString("【用户问题】\n")
	sb.WriteString(question)
	sb.WriteString("\n\n")

	if len(keywords) > 0 {
		sb.WriteString("【关键词】\n")
		sb.WriteString(strings.Join(keywords, ", "))
		sb.WriteString("\n\n")
	}

	if len(articles) > 0 {
		sb.WriteString("【相关法律条文】\n")
		for i, art := range articles {
			sb.WriteString(fmt.Sprintf("%d. 《%s》%s\n%s\n\n",
				i+1, art.LawName, art.ArticleNo, art.Content))
		}
	}

	return sb.String()
}

func buildMessages(question string, history []ChatMessage, context string) []ChatMessage {
	messages := make([]ChatMessage, 0)

	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: LegalSystemPrompt,
	})

	for _, h := range history {
		messages = append(messages, h)
	}

	userMessage := context + "\n\n请根据以上信息回答用户问题：" + question
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

func parseAIAnswer(raw string, articles []LawReference) (*AIAnswer, error) {
	raw = strings.TrimSpace(raw)

	if strings.Contains(raw, "```json") {
		start := strings.Index(raw, "```json")
		end := strings.Index(raw[start:], "```")
		if end > 0 {
			raw = raw[start+7 : start+end]
		}
	}

	if strings.Contains(raw, "```") {
		start := strings.Index(raw, "```")
		end := strings.Index(raw[start+3:], "```")
		if end > 0 {
			raw = raw[start+3 : start+3+end]
		}
	}

	raw = strings.TrimSpace(raw)

	var answer AIAnswer
	if err := json.Unmarshal([]byte(raw), &answer); err != nil {
		return nil, fmt.Errorf("parse json failed: %w", err)
	}

	if answer.RelatedArticles == nil {
		answer.RelatedArticles = articles
	}
	if answer.Keywords == nil {
		answer.Keywords = []string{}
	}
	if answer.References == nil {
		answer.References = []string{}
	}

	return &answer, nil
}

func fallbackAnswer(question string, articles []LawReference) *AIAnswer {
	return &AIAnswer{
		Answer:          "根据您的问题，建议您收集相关证据材料，先尝试与对方协商解决。协商不成的，可以向人民调解委员会申请调解，或者直接向人民法院提起诉讼。如需更详细的法律建议，建议携带相关材料到当地司法所或律师事务所现场咨询。",
		RelatedArticles: articles,
		References:      []string{},
		Keywords:        []string{"纠纷", "法律建议"},
	}
}

func extractLawName(content string) string {
	lawNames := []string{
		"民法典", "民事诉讼法", "刑法", "劳动合同法", "公司法",
		"合同法", "婚姻法", "继承法", "物权法", "侵权责任法",
		"物业管理条例", "消费者权益保护法", "道路交通安全法",
	}
	for _, name := range lawNames {
		if strings.Contains(content, name) {
			return "中华人民共和国" + name
		}
	}
	if idx := strings.Index(content, "《"); idx >= 0 {
		if endIdx := strings.Index(content[idx:], "》"); endIdx > 0 {
			return content[idx+1 : idx+endIdx]
		}
	}
	return "相关法律"
}
