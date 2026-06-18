package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

const (
	SummarySystemPrompt = `你是一名专业的人民调解文书助理。请根据以下案件信息和调解记录，生成一份结构化的调解摘要。
请严格按照以下JSON格式输出：
{
  "summary": "完整的调解摘要文本，包含以下四个部分：1.案件基本情况（案件编号、当事人信息、案由）；2.双方诉求与争议焦点；3.调解过程（调解时间、地点、调解员、主要协商过程）；4.达成协议或调解结果。语言要正式、规范、简洁。"
}`

	RiskAssessmentPrompt = `你是一名专业的法律风险评估专家。请根据以下案件信息，进行风险评估。
请严格按照以下JSON格式输出：
{
  "riskLevel": "低/中/高",
  "riskScore": 0.0-1.0之间的数值,
  "riskFactors": ["风险因素1", "风险因素2", "风险因素3"],
  "suggestions": ["建议1", "建议2", "建议3"]
}
评估维度：
1. 证据充分性：原告方证据是否完整、确凿
2. 法律依据明确性：是否有明确的法条支持
3. 对方态度：对方是否愿意配合调解/诉讼
4. 执行难度：判决后执行的可能性
5. 时间成本：预计处理周期
6. 经济成本：诉讼费用、律师费用等`
)

func GenerateMediationSummary(caseInfo map[string]interface{}, mediationContent string) (string, error) {
	if strings.TrimSpace(mediationContent) == "" {
		return "", fmt.Errorf("调解内容不能为空")
	}

	logger.Info("Start generate mediation summary",
		zap.Int("caseInfoKeys", len(caseInfo)),
		zap.Int("contentLength", len(mediationContent)),
	)

	client := GetDeepSeekClient()

	userMessage := buildMediationPrompt(caseInfo, mediationContent)

	messages := []ChatMessage{
		{Role: "user", Content: userMessage},
	}

	rawResp, err := client.ChatCompletion(messages, SummarySystemPrompt)
	if err != nil {
		logger.Error("Generate mediation summary failed", logger.Error(err))
		return fallbackMediationSummary(caseInfo, mediationContent), nil
	}

	summary, err := parseMediationSummary(rawResp)
	if err != nil {
		logger.Warn("Parse mediation summary failed, use raw text", logger.Error(err))
		summary = rawResp
	}

	logger.Info("Mediation summary generated", zap.Int("length", len(summary)))
	return summary, nil
}

func buildMediationPrompt(caseInfo map[string]interface{}, mediationContent string) string {
	var sb strings.Builder

	sb.WriteString("【案件信息】\n")
	for k, v := range caseInfo {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	sb.WriteString("\n")

	sb.WriteString("【调解记录】\n")
	sb.WriteString(mediationContent)
	sb.WriteString("\n\n")

	sb.WriteString("请根据以上信息，生成一份专业、规范的调解摘要。")

	return sb.String()
}

func parseMediationSummary(raw string) (string, error) {
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

	var result struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return "", fmt.Errorf("parse json failed: %w", err)
	}

	if result.Summary == "" {
		return "", fmt.Errorf("empty summary")
	}

	return result.Summary, nil
}

func fallbackMediationSummary(caseInfo map[string]interface{}, content string) string {
	var sb strings.Builder

	sb.WriteString("调解摘要\n\n")

	sb.WriteString("一、案件基本情况\n")
	if caseNo, ok := caseInfo["caseNo"].(string); ok {
		sb.WriteString(fmt.Sprintf("案件编号：%s\n", caseNo))
	}
	if title, ok := caseInfo["title"].(string); ok {
		sb.WriteString(fmt.Sprintf("案件名称：%s\n", title))
	}
	if applicant, ok := caseInfo["applicantName"].(string); ok {
		sb.WriteString(fmt.Sprintf("申请人：%s\n", applicant))
	}
	if respondent, ok := caseInfo["respondentName"].(string); ok {
		sb.WriteString(fmt.Sprintf("被申请人：%s\n", respondent))
	}
	sb.WriteString("\n")

	sb.WriteString("二、调解内容\n")
	if len(content) > 500 {
		sb.WriteString(content[:500] + "...")
	} else {
		sb.WriteString(content)
	}
	sb.WriteString("\n\n")

	sb.WriteString("三、调解结果\n")
	sb.WriteString("双方已就争议事项进行了调解协商。")

	return sb.String()
}

func GenerateRiskAssessment(caseInfo map[string]interface{}) (*RiskResult, error) {
	if len(caseInfo) == 0 {
		return nil, fmt.Errorf("案件信息不能为空")
	}

	logger.Info("Start generate risk assessment",
		zap.Int("caseInfoKeys", len(caseInfo)),
	)

	client := GetDeepSeekClient()

	userMessage := buildRiskAssessmentPrompt(caseInfo)

	messages := []ChatMessage{
		{Role: "user", Content: userMessage},
	}

	rawResp, err := client.ChatCompletion(messages, RiskAssessmentPrompt)
	if err != nil {
		logger.Error("Generate risk assessment failed", logger.Error(err))
		return fallbackRiskAssessment(caseInfo), nil
	}

	result, err := parseRiskAssessment(rawResp)
	if err != nil {
		logger.Warn("Parse risk assessment failed, use fallback", logger.Error(err))
		return fallbackRiskAssessment(caseInfo), nil
	}

	logger.Info("Risk assessment completed",
		zap.String("riskLevel", result.RiskLevel),
		zap.Float64("riskScore", result.RiskScore),
	)

	return result, nil
}

func buildRiskAssessmentPrompt(caseInfo map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("【案件信息】\n")
	for k, v := range caseInfo {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	sb.WriteString("\n")

	sb.WriteString("请根据以上案件信息，进行全面的法律风险评估。")

	return sb.String()
}

func parseRiskAssessment(raw string) (*RiskResult, error) {
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

	var result RiskResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse json failed: %w", err)
	}

	if result.RiskFactors == nil {
		result.RiskFactors = []string{}
	}
	if result.Suggestions == nil {
		result.Suggestions = []string{}
	}

	return &result, nil
}

func fallbackRiskAssessment(caseInfo map[string]interface{}) *RiskResult {
	return &RiskResult{
		RiskLevel: "中",
		RiskScore: 0.5,
		RiskFactors: []string{
			"证据完整性待核实",
			"法律适用存在一定争议空间",
			"对方配合态度不确定",
		},
		Suggestions: []string{
			"建议尽快收集和固定相关证据材料",
			"先尝试协商或人民调解方式解决",
			"如协商不成，可咨询专业律师后决定是否起诉",
			"注意诉讼时效问题，及时主张权利",
		},
	}
}
