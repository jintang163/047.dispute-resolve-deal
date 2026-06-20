package ai

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type SentimentResult struct {
	Emotion          string                 `json:"emotion"`
	EmotionLabel     string                 `json:"emotionLabel"`
	SentimentScore   float64                `json:"sentimentScore"`
	Confidence       float64                `json:"confidence"`
	PositiveKeywords []string               `json:"positiveKeywords"`
	NegativeKeywords []string               `json:"negativeKeywords"`
	KeyPoints        []SentimentKeyPoint    `json:"keyPoints"`
	Satisfaction     int                    `json:"satisfaction"`
	Performance      int                    `json:"performance"`
	Summary          string                 `json:"summary"`
	RawResponse      string                 `json:"rawResponse,omitempty"`
}

type SatisfactionSentimentResult struct {
	Emotion              string   `json:"emotion"`
	EmotionLabel         string   `json:"emotionLabel"`
	SentimentScore       float64  `json:"sentimentScore"`
	Confidence           float64  `json:"confidence"`
	PositiveKeywords     []string `json:"positiveKeywords"`
	NegativeKeywords     []string `json:"negativeKeywords"`
	SatisfactionEstimate int      `json:"satisfactionEstimate"`
	Summary              string   `json:"summary"`
	IssueType            string   `json:"issueType"`
	IssueDescription     string   `json:"issueDescription"`
	ImprovementSuggestion string  `json:"improvementSuggestion"`
	PrioritySuggestion   int      `json:"prioritySuggestion"`
}

type SentimentKeyPoint struct {
	Content string `json:"content"`
	Sentiment string `json:"sentiment"`
	Score   float64 `json:"score"`
}

type SentimentAnalyzer struct {
	client *DeepSeekClient
}

var sentimentAnalyzer *SentimentAnalyzer

func NewSentimentAnalyzer() *SentimentAnalyzer {
	return &SentimentAnalyzer{
		client: GetDeepSeekClient(),
	}
}

func InitSentimentAnalyzer() {
	sentimentAnalyzer = NewSentimentAnalyzer()
	logger.Info("Sentiment analyzer initialized")
}

func GetSentimentAnalyzer() *SentimentAnalyzer {
	if sentimentAnalyzer == nil {
		InitSentimentAnalyzer()
	}
	return sentimentAnalyzer
}

func (a *SentimentAnalyzer) AnalyzeText(text string) (*SentimentResult, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text for sentiment analysis")
	}

	systemPrompt := `你是专业的情感分析专家，负责对客户回访对话进行多维度的情绪和满意度分析。

分析维度：
1. 整体情绪分类：positive(正面)、neutral(中性)、negative(负面)
2. 情绪评分：-1.0到1.0之间的浮点数，-1表示非常负面，0表示中性，1表示非常正面
3. 置信度：0到1之间的浮点数，表示分析结果的可靠程度
4. 关键词提取：提取正面和负面的关键词
5. 关键点分析：提取对话中的关键内容点及其情感倾向
6. 满意度评分：1-5分，表示用户对服务的整体满意度
7. 履约评分：1-5分，表示用户对协议履行情况的评价
8. 摘要总结：用简短的语言总结用户的主要反馈

请严格按照以下JSON格式返回结果：
{
  "emotion": "positive|neutral|negative",
  "sentimentScore": 0.85,
  "confidence": 0.92,
  "positiveKeywords": ["满意", "专业", "高效"],
  "negativeKeywords": ["等待时间长"],
  "keyPoints": [
    {"content": "调解员很专业", "sentiment": "positive", "score": 0.8},
    {"content": "协议已全部履行", "sentiment": "neutral", "score": 0}
  ],
  "satisfaction": 5,
  "performance": 5,
  "summary": "用户对调解服务整体满意，调解员专业且高效，协议已全部履行"
}

注意：
- 只返回JSON，不要有任何额外的文字说明
- 确保JSON格式正确，使用双引号
- satisfaction和performance必须是1-5的整数
- sentimentScore必须在-1到1之间
- 分析要客观准确，基于文本内容，不要主观臆断`

	messages := []ChatMessage{
		{
			Role:    "user",
			Content: fmt.Sprintf("请对以下回访对话文本进行情感分析：\n\n%s", text),
		},
	}

	logger.Debug("Sentiment analysis request",
		zap.Int("textLength", len(text)),
	)

	result, err := a.client.ChatCompletion(messages, systemPrompt)
	if err != nil {
		logger.Error("Sentiment analysis API call failed",
			logger.Error(err),
		)
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	logger.Debug("Sentiment analysis raw response",
		zap.String("result", result),
	)

	cleanResult := cleanJSONResponse(result)

	var sentimentResult SentimentResult
	if err := json.Unmarshal([]byte(cleanResult), &sentimentResult); err != nil {
		logger.Warn("Failed to parse sentiment analysis result as JSON, trying fallback",
			zap.String("result", result),
			logger.Error(err),
		)
		sentimentResult = a.parseFallback(result)
	}

	sentimentResult.EmotionLabel = mapEmotionLabel(sentimentResult.Emotion)
	sentimentResult.RawResponse = result

	logger.Info("Sentiment analysis completed",
		zap.String("emotion", sentimentResult.Emotion),
		zap.Float64("sentimentScore", sentimentResult.SentimentScore),
		zap.Int("satisfaction", sentimentResult.Satisfaction),
		zap.Int("performance", sentimentResult.Performance),
	)

	return &sentimentResult, nil
}

func (a *SentimentAnalyzer) AnalyzeCallback(transcript string, caseInfo map[string]interface{}) (*SentimentResult, error) {
	context := ""
	if caseInfo != nil {
		if title, ok := caseInfo["title"].(string); ok {
			context += fmt.Sprintf("案件标题：%s\n", title)
		}
		if mediationResult, ok := caseInfo["mediationResult"].(string); ok {
			context += fmt.Sprintf("调解结果：%s\n", mediationResult)
		}
	}

	fullText := context + "\n回访对话内容：\n" + transcript

	return a.AnalyzeText(fullText)
}

func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}
	
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}
	
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start != -1 && end != -1 && end > start {
		response = response[start : end+1]
	}
	
	return response
}

func (a *SentimentAnalyzer) parseFallback(text string) SentimentResult {
	result := SentimentResult{
		Emotion:        "neutral",
		SentimentScore: 0,
		Confidence:     0.5,
		Satisfaction:   3,
		Performance:    3,
		Summary:        "解析失败，使用默认值",
	}

	lowerText := strings.ToLower(text)

	positiveWords := []string{"满意", "感谢", "很好", "不错", "专业", "高效", "满意", "好评", "点赞", "顺利", "成功", "履行", "完成"}
	negativeWords := []string{"不满意", "投诉", "生气", "愤怒", "失望", "差", "糟糕", "慢", "问题", "未履行", "拖延", "拒绝", "纠纷"}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		if strings.Contains(lowerText, word) {
			positiveCount++
			result.PositiveKeywords = append(result.PositiveKeywords, word)
		}
	}

	for _, word := range negativeWords {
		if strings.Contains(lowerText, word) {
			negativeCount++
			result.NegativeKeywords = append(result.NegativeKeywords, word)
		}
	}

	if positiveCount > negativeCount {
		result.Emotion = "positive"
		result.SentimentScore = float64(positiveCount) / float64(positiveCount+negativeCount+1)
		result.Satisfaction = min(5, 4+positiveCount)
		result.Performance = min(5, 4+positiveCount)
	} else if negativeCount > positiveCount {
		result.Emotion = "negative"
		result.SentimentScore = -float64(negativeCount) / float64(positiveCount+negativeCount+1)
		result.Satisfaction = max(1, 3-negativeCount)
		result.Performance = max(1, 3-negativeCount)
	}

	result.EmotionLabel = mapEmotionLabel(result.Emotion)
	result.Confidence = 0.6

	return result
}

func mapEmotionLabel(emotion string) string {
	switch emotion {
	case "positive":
		return "正面"
	case "negative":
		return "负面"
	default:
		return "中性"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ExtractScoreFromText(text string) int {
	for i := 5; i >= 1; i-- {
		if strings.Contains(text, strconv.Itoa(i)+"分") || 
		   strings.Contains(text, "打"+strconv.Itoa(i)+"分") ||
		   strings.Contains(text, strconv.Itoa(i)+"颗星") {
			return i
		}
	}
	
	scoreMap := map[string]int{
		"非常满意": 5,
		"很满意":   5,
		"满意":    4,
		"还可以":   3,
		"一般":    3,
		"不满意":   2,
		"很不满意": 1,
		"非常不满意": 1,
	}
	
	for key, score := range scoreMap {
		if strings.Contains(text, key) {
			return score
		}
	}
	
	return 0
}

func ExtractPerformanceFromText(text string) int {
	perfMap := map[string]int{
		"全部履行": 5,
		"已履行":  5,
		"完全履行": 5,
		"大部分履行": 4,
		"部分履行": 3,
		"正在履行": 3,
		"还没履行": 2,
		"未履行":  1,
		"没有履行": 1,
		"拒绝履行": 1,
	}
	
	for key, score := range perfMap {
		if strings.Contains(text, key) {
			return score
		}
	}
	
	return 0
}

func (a *SentimentAnalyzer) AnalyzeSatisfactionComment(comment string, caseInfo map[string]interface{}) (*SatisfactionSentimentResult, error) {
	if comment == "" {
		return nil, fmt.Errorf("empty satisfaction comment")
	}

	contextStr := ""
	if caseInfo != nil {
		if title, ok := caseInfo["title"].(string); ok {
			contextStr += fmt.Sprintf("案件标题：%s\n", title)
		}
		if mediatorName, ok := caseInfo["mediatorName"].(string); ok {
			contextStr += fmt.Sprintf("调解员：%s\n", mediatorName)
		}
		if mediationResult, ok := caseInfo["mediationResult"].(string); ok {
			contextStr += fmt.Sprintf("调解结果：%s\n", mediationResult)
		}
		if score, ok := caseInfo["satisfactionScore"].(int); ok && score > 0 {
			contextStr += fmt.Sprintf("满意度评分：%d星\n", score)
		}
	}

	systemPrompt := `你是专业的满意度评价情感分析专家，负责对群众填写的调解服务满意度评语进行深度分析。

分析维度：
1. 整体情感分类：positive(正面)、neutral(中性)、negative(负面)
2. 情感评分：-1.0到1.0之间的浮点数
3. 置信度：0到1之间的浮点数
4. 关键词提取：分别提取正面和负面关键词
5. 满意度估算：1-5分，根据评语内容推断满意度
6. 摘要总结：用简短语言总结群众反馈
7. 问题类型（仅负面评价需要填写）：
   - attitude: 态度问题（调解员态度差、不耐烦等）
   - efficiency: 效率问题（处理慢、等待时间长等）
   - professional: 专业性问题（不专业、法律知识不足等）
   - result: 结果不满意（调解结果不公平等）
   - process: 流程问题（流程复杂、不透明等）
   - other: 其他
8. 问题描述（仅负面评价需要填写）：客观描述群众反映的核心问题，50字以内
9. 改进建议（仅负面评价需要填写）：针对问题给出具体改进措施，100字以内
10. 优先级建议（仅负面评价需要填写）：1-高优先级(涉及投诉/违规)，2-中优先级(服务不满意)，3-低优先级(轻微不满)

对于正面或中性评价，issueType、issueDescription、improvementSuggestion填空字符串，prioritySuggestion填0。

请严格按照以下JSON格式返回结果：
{
  "emotion": "positive|neutral|negative",
  "sentimentScore": 0.85,
  "confidence": 0.92,
  "positiveKeywords": ["专业", "耐心"],
  "negativeKeywords": ["等待时间长"],
  "satisfactionEstimate": 4,
  "summary": "群众对调解员专业水平认可，但认为等待时间较长",
  "issueType": "efficiency",
  "issueDescription": "群众反映调解等待时间过长，多次催促才安排调解",
  "improvementSuggestion": "建议优化调解排期流程，设置合理的响应时限，对紧急案件优先处理",
  "prioritySuggestion": 2
}

注意：
- 只返回JSON，不要有任何额外的文字说明
- sentimentScore必须在-1到1之间
- satisfactionEstimate必须是1-5的整数
- 负面评价必须准确识别问题类型并给出改进建议
- 分析要客观准确，基于评语内容，不要主观臆断`

	userContent := fmt.Sprintf("请对以下满意度评语进行情感分析：\n\n%s\n群众评语：%s", contextStr, comment)

	messages := []ChatMessage{
		{
			Role:    "user",
			Content: userContent,
		},
	}

	logger.Debug("Satisfaction sentiment analysis request",
		zap.Int("commentLength", len(comment)),
	)

	result, err := a.client.ChatCompletion(messages, systemPrompt)
	if err != nil {
		logger.Error("Satisfaction sentiment analysis API call failed",
			logger.Error(err),
		)
		return nil, fmt.Errorf("satisfaction sentiment analysis failed: %w", err)
	}

	cleanResult := cleanJSONResponse(result)

	var sentimentResult SatisfactionSentimentResult
	if err := json.Unmarshal([]byte(cleanResult), &sentimentResult); err != nil {
		logger.Warn("Failed to parse satisfaction sentiment result, using fallback",
			zap.String("result", result),
			logger.Error(err),
		)
		sentimentResult = a.parseSatisfactionFallback(comment, result)
	}

	sentimentResult.EmotionLabel = mapEmotionLabel(sentimentResult.Emotion)

	logger.Info("Satisfaction sentiment analysis completed",
		zap.String("emotion", sentimentResult.Emotion),
		zap.Float64("sentimentScore", sentimentResult.SentimentScore),
		zap.Int("satisfactionEstimate", sentimentResult.SatisfactionEstimate),
		zap.String("issueType", sentimentResult.IssueType),
	)

	return &sentimentResult, nil
}

func (a *SentimentAnalyzer) parseSatisfactionFallback(comment string, rawResponse string) SatisfactionSentimentResult {
	result := SatisfactionSentimentResult{
		Emotion:              "neutral",
		SentimentScore:       0,
		Confidence:           0.5,
		SatisfactionEstimate: 3,
		Summary:              "解析失败，使用默认值",
	}

	lowerText := strings.ToLower(comment)

	positiveWords := []string{"满意", "感谢", "很好", "不错", "专业", "高效", "好评", "点赞", "顺利", "成功", "耐心", "认真"}
	negativeWords := []string{"不满意", "投诉", "生气", "愤怒", "失望", "差", "糟糕", "慢", "问题", "拖延", "态度差", "不专业", "不公平", "推诿", "敷衍"}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		if strings.Contains(lowerText, word) {
			positiveCount++
			result.PositiveKeywords = append(result.PositiveKeywords, word)
		}
	}
	for _, word := range negativeWords {
		if strings.Contains(lowerText, word) {
			negativeCount++
			result.NegativeKeywords = append(result.NegativeKeywords, word)
		}
	}

	if positiveCount > negativeCount {
		result.Emotion = "positive"
		result.SentimentScore = float64(positiveCount) / float64(positiveCount+negativeCount+1)
		result.SatisfactionEstimate = min(5, 4)
	} else if negativeCount > positiveCount {
		result.Emotion = "negative"
		result.SentimentScore = -float64(negativeCount) / float64(positiveCount+negativeCount+1)
		result.SatisfactionEstimate = max(1, 2)

		issueTypeMap := map[string]string{
			"态度差": "attitude", "不耐烦": "attitude", "推诿": "attitude", "敷衍": "attitude",
			"慢": "efficiency", "等待": "efficiency", "拖延": "efficiency",
			"不专业": "professional", "不懂": "professional",
			"不公平": "result", "不满意": "result",
		}
		for word, it := range issueTypeMap {
			if strings.Contains(lowerText, word) {
				result.IssueType = it
				break
			}
		}
		if result.IssueType == "" {
			result.IssueType = "other"
		}
		result.IssueDescription = comment
		if len(result.IssueDescription) > 50 {
			result.IssueDescription = result.IssueDescription[:50]
		}
		result.ImprovementSuggestion = "建议加强调解员培训，提升服务质量"
		result.PrioritySuggestion = 2
	}

	return result
}
