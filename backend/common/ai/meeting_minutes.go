package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

const (
	MeetingMinutesSystemPrompt = `你是一名专业的视频调解会议纪要生成助手。请根据以下视频调解的转录文本和案件信息，生成一份结构完整的会议纪要。

请严格按照以下JSON格式输出：
{
  "meetingTitle": "会议标题",
  "meetingTime": "会议时间",
  "participants": ["参与人1", "参与人2", "参与人3"],
  "duration": "会议时长",
  "summary": "会议概要（100字以内）",
  "keyPoints": ["要点1", "要点2", "要点3"],
  "disputeFocus": ["争议焦点1", "争议焦点2"],
  "mediationProcess": "调解过程描述",
  "evidenceDiscussed": ["讨论的证据1", "讨论的证据2"],
  "agreement": "达成协议内容（如未达成协议则写"未达成协议"）",
  "nextSteps": ["下一步1", "下一步2"],
  "riskPoints": ["风险提示1", "风险提示2"],
  "emotionalState": "各方情绪状态描述",
  "mediatorAdvice": "调解员建议"
}

要求：
1. 客观、准确地记录会议内容
2. 重点标注争议焦点和达成的共识
3. 对各方情绪状态做适当记录
4. 明确下一步行动项和责任人
5. 标注可能存在的风险点`
)

type MeetingMinutesResult struct {
	MeetingTitle     string   `json:"meetingTitle"`
	MeetingTime      string   `json:"meetingTime"`
	Participants     []string `json:"participants"`
	Duration         string   `json:"duration"`
	Summary          string   `json:"summary"`
	KeyPoints        []string `json:"keyPoints"`
	DisputeFocus     []string `json:"disputeFocus"`
	MediationProcess string   `json:"mediationProcess"`
	EvidenceDiscussed []string `json:"evidenceDiscussed"`
	Agreement        string   `json:"agreement"`
	NextSteps        []string `json:"nextSteps"`
	RiskPoints       []string `json:"riskPoints"`
	EmotionalState   string   `json:"emotionalState"`
	MediatorAdvice   string   `json:"mediatorAdvice"`
}

func GenerateMeetingMinutes(caseInfo map[string]interface{}, transcript string, duration time.Duration) (*MeetingMinutesResult, error) {
	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("转录文本不能为空")
	}

	logger.Info("Start generate meeting minutes",
		zap.Int("caseInfoKeys", len(caseInfo)),
		zap.Int("transcriptLength", len(transcript)),
	)

	client := GetDeepSeekClient()

	userMessage := buildMeetingMinutesPrompt(caseInfo, transcript, duration)

	messages := []ChatMessage{
		{Role: "user", Content: userMessage},
	}

	rawResp, err := client.ChatCompletion(messages, MeetingMinutesSystemPrompt)
	if err != nil {
		logger.Error("Generate meeting minutes failed", logger.Error(err))
		return fallbackMeetingMinutes(caseInfo, transcript, duration), nil
	}

	result, err := parseMeetingMinutes(rawResp)
	if err != nil {
		logger.Warn("Parse meeting minutes failed, use raw text", logger.Error(err))
		return fallbackMeetingMinutes(caseInfo, transcript, duration), nil
	}

	logger.Info("Meeting minutes generated", zap.Int("keyPointsCount", len(result.KeyPoints)))
	return result, nil
}

func buildMeetingMinutesPrompt(caseInfo map[string]interface{}, transcript string, duration time.Duration) string {
	var sb strings.Builder

	sb.WriteString("【案件信息】\n")
	for k, v := range caseInfo {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	sb.WriteString("\n")

	sb.WriteString("【会议时长】\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", duration.String()))

	sb.WriteString("【视频转录文本】\n")
	transcriptRunes := []rune(transcript)
	if len(transcriptRunes) > 12000 {
		sb.WriteString(string(transcriptRunes[:12000]))
		sb.WriteString("\n...（文本已截断）")
	} else {
		sb.WriteString(transcript)
	}
	sb.WriteString("\n\n")

	sb.WriteString("请根据以上信息，生成一份专业的视频调解会议纪要。")

	return sb.String()
}

func parseMeetingMinutes(raw string) (*MeetingMinutesResult, error) {
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

	var result MeetingMinutesResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse json failed: %w", err)
	}

	if result.KeyPoints == nil {
		result.KeyPoints = []string{}
	}
	if result.DisputeFocus == nil {
		result.DisputeFocus = []string{}
	}
	if result.EvidenceDiscussed == nil {
		result.EvidenceDiscussed = []string{}
	}
	if result.NextSteps == nil {
		result.NextSteps = []string{}
	}
	if result.RiskPoints == nil {
		result.RiskPoints = []string{}
	}
	if result.Participants == nil {
		result.Participants = []string{}
	}

	return &result, nil
}

func fallbackMeetingMinutes(caseInfo map[string]interface{}, transcript string, duration time.Duration) *MeetingMinutesResult {
	title := "视频调解会议纪要"
	if t, ok := caseInfo["title"].(string); ok && t != "" {
		title = t + " - 视频调解会议纪要"
	}

	transcriptPreview := ""
	transcriptRunes := []rune(transcript)
	if len(transcriptRunes) > 500 {
		transcriptPreview = string(transcriptRunes[:500]) + "..."
	} else {
		transcriptPreview = transcript
	}

	return &MeetingMinutesResult{
		MeetingTitle:     title,
		MeetingTime:      time.Now().Format("2006-01-02 15:04:05"),
		Participants:     []string{},
		Duration:         duration.String(),
		Summary:          "本次视频调解会议已完成，具体内容待人工确认。",
		KeyPoints:        []string{"调解过程已记录"},
		DisputeFocus:     []string{"待人工标注"},
		MediationProcess: transcriptPreview,
		EvidenceDiscussed: []string{},
		Agreement:        "待人工确认",
		NextSteps:        []string{"请调解员确认会议纪要内容"},
		RiskPoints:       []string{},
		EmotionalState:   "待确认",
		MediatorAdvice:   "待确认",
	}
}
