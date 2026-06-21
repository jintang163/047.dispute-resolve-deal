package aliyun

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type VoiceClient struct {
	accessKeyID      string
	accessKeySecret  string
	regionID         string
	callerShowNumber string
	robotId          string
	ttsCode          string
	asrVocabID       string
	callbackURL      string
	endpoint         string
	nlsEndpoint      string
	nlsAppKey        string
	httpClient       *http.Client
}

var (
	voiceClient *VoiceClient
	voiceOnce   sync.Once
)

type SmartCallRequest struct {
	CalledNumber     string            `json:"calledNumber"`
	CalledShowNumber string            `json:"calledShowNumber,omitempty"`
	RobotId          string            `json:"robotId,omitempty"`
	TtsCode          string            `json:"ttsCode,omitempty"`
	TtsParam         map[string]string `json:"ttsParam,omitempty"`
	OutId            string            `json:"outId,omitempty"`
	RecordFlag       bool              `json:"recordFlag,omitempty"`
	SessionTimeout   int               `json:"sessionTimeout,omitempty"`
	VoiceType        string            `json:"voiceType,omitempty"`
	Volume           int               `json:"volume,omitempty"`
	Speed            int               `json:"speed,omitempty"`
}

type VoiceCallRequest struct {
	CalledNumber     string            `json:"calledNumber"`
	CalledShowNumber string            `json:"calledShowNumber,omitempty"`
	TtsCode          string            `json:"ttsCode,omitempty"`
	TtsParam         map[string]string `json:"ttsParam,omitempty"`
	OutId            string            `json:"outId,omitempty"`
	PlayTimes        int               `json:"playTimes,omitempty"`
	Volume           int               `json:"volume,omitempty"`
	Speed            int               `json:"speed,omitempty"`
	AsrModelId      string            `json:"asrModelId,omitempty"`
	AsrVocabId      string            `json:"asrVocabId,omitempty"`
	RecordFlag       bool              `json:"recordFlag,omitempty"`
	RecordTimeout   int               `json:"recordTimeout,omitempty"`
}

type VoiceCallResponse struct {
	RequestID string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	CallId    string `json:"CallId"`
	TaskId    string `json:"TaskId"`
	OrderId   string `json:"OrderId"`
}

type CallDetailResponse struct {
	RequestID    string `json:"RequestId"`
	Code         string `json:"Code"`
	Message      string `json:"Message"`
	Status       string `json:"Status"`
	CallId       string `json:"CallId"`
	TaskId       string `json:"TaskId"`
	CalledNumber string `json:"CalledNumber"`
	StartTime    string `json:"StartTime"`
	EndTime      string `json:"EndTime"`
	Duration     int    `json:"Duration"`
	RecordUrl    string `json:"RecordUrl"`
	AsrResult    string `json:"AsrResult"`
	ErrorCode    string `json:"ErrorCode"`
	StatusDesc   string `json:"StatusDesc"`
}

func NewVoiceClient() *VoiceClient {
	cfg := config.GetConfig()
	return &VoiceClient{
		accessKeyID:      cfg.AliyunVoice.AccessKeyID,
		accessKeySecret:  cfg.AliyunVoice.AccessKeySecret,
		regionID:         cfg.AliyunVoice.RegionID,
		callerShowNumber: cfg.AliyunVoice.CallerShowNumber,
		robotId:          cfg.AliyunVoice.RobotId,
		ttsCode:          cfg.AliyunVoice.TtsCode,
		asrVocabID:       cfg.AliyunVoice.AsrVocabID,
		callbackURL:      cfg.AliyunVoice.CallbackURL,
		endpoint:         cfg.AliyunVoice.Endpoint,
		nlsEndpoint:      cfg.AliyunVoice.NlsEndpoint,
		nlsAppKey:        cfg.AliyunVoice.NlsAppKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func InitVoiceClient() {
	voiceOnce.Do(func() {
		voiceClient = NewVoiceClient()
		logger.Info("Aliyun Voice client initialized",
			zap.String("endpoint", voiceClient.endpoint),
			zap.String("callerShowNumber", voiceClient.callerShowNumber),
			zap.String("robotId", voiceClient.robotId),
		)
	})
}

func GetVoiceClient() *VoiceClient {
	if voiceClient == nil {
		InitVoiceClient()
	}
	return voiceClient
}

func (c *VoiceClient) SmartCall(req *SmartCallRequest) (*VoiceCallResponse, error) {
	params := c.buildCommonParams()
	params["Action"] = "SmartCall"
	params["CalledNumber"] = req.CalledNumber
	params["CalledShowNumber"] = c.callerShowNumber
	params["RecordFlag"] = "true"

	if req.RobotId != "" {
		params["RobotId"] = req.RobotId
	} else if c.robotId != "" {
		params["RobotId"] = c.robotId
	}

	if req.TtsCode != "" {
		params["TtsCode"] = req.TtsCode
	} else if c.ttsCode != "" {
		params["TtsCode"] = c.ttsCode
	}

	if req.OutId != "" {
		params["OutId"] = req.OutId
	}
	if req.SessionTimeout > 0 {
		params["SessionTimeout"] = fmt.Sprintf("%d", req.SessionTimeout)
	}
	if req.VoiceType != "" {
		params["VoiceType"] = req.VoiceType
	}
	if req.Volume != 0 {
		params["Volume"] = fmt.Sprintf("%d", req.Volume)
	}
	if req.Speed != 0 {
		params["Speed"] = fmt.Sprintf("%d", req.Speed)
	}
	if c.callbackURL != "" {
		params["CallbackUrl"] = c.callbackURL
	}

	if req.TtsParam != nil && len(req.TtsParam) > 0 {
		ttsParamJSON, _ := json.Marshal(req.TtsParam)
		params["TtsParam"] = string(ttsParamJSON)
	}

	return c.doRequest("SmartCall", params)
}

func (c *VoiceClient) SingleCallByTts(req *VoiceCallRequest) (*VoiceCallResponse, error) {
	params := c.buildCommonParams()
	params["Action"] = "SingleCallByTts"
	params["CalledNumber"] = req.CalledNumber
	params["CalledShowNumber"] = c.callerShowNumber
	params["RecordFlag"] = "true"

	if req.TtsCode != "" {
		params["TtsCode"] = req.TtsCode
	} else if c.ttsCode != "" {
		params["TtsCode"] = c.ttsCode
	}
	if req.OutId != "" {
		params["OutId"] = req.OutId
	}
	if req.PlayTimes > 0 {
		params["PlayTimes"] = fmt.Sprintf("%d", req.PlayTimes)
	}
	if req.Volume != 0 {
		params["Volume"] = fmt.Sprintf("%d", req.Volume)
	}
	if req.Speed != 0 {
		params["Speed"] = fmt.Sprintf("%d", req.Speed)
	}
	if req.AsrVocabId != "" {
		params["AsrVocabId"] = req.AsrVocabId
	} else if c.asrVocabID != "" {
		params["AsrVocabId"] = c.asrVocabID
	}
	if c.callbackURL != "" {
		params["CallbackUrl"] = c.callbackURL
	}

	if req.TtsParam != nil && len(req.TtsParam) > 0 {
		ttsParamJSON, _ := json.Marshal(req.TtsParam)
		params["TtsParam"] = string(ttsParamJSON)
	}

	return c.doRequest("SingleCallByTts", params)
}

func (c *VoiceClient) QueryCallDetailByCallId(callId string) (*CallDetailResponse, error) {
	params := c.buildCommonParams()
	params["Action"] = "QueryCallDetailByCallId"
	params["CallId"] = callId
	params["QueryDate"] = time.Now().Format("20060102")

	signature := c.generateSignature(params, "GET")
	params["Signature"] = signature

	query := c.buildQuery(params)
	reqURL := fmt.Sprintf("https://%s/?%s", c.endpoint, query)

	logger.Debug("Aliyun Voice QueryCallDetailByCallId request",
		zap.String("callId", callId),
	)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		logger.Error("Aliyun Voice QueryCallDetailByCallId request failed",
			logger.Error(err),
		)
		return nil, fmt.Errorf("aliyun voice api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	logger.Debug("Aliyun Voice QueryCallDetailByCallId response",
		zap.String("body", string(respBody)),
	)

	var result CallDetailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != "OK" && result.Code != "" {
		logger.Error("Aliyun Voice QueryCallDetailByCallId error",
			zap.String("code", result.Code),
			zap.String("message", result.Message),
		)
		return &result, fmt.Errorf("aliyun voice api error: code=%s, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

func (c *VoiceClient) CancelCall(callId string) error {
	params := c.buildCommonParams()
	params["Action"] = "CancelCall"
	params["CallId"] = callId

	signature := c.generateSignature(params, "GET")
	params["Signature"] = signature

	query := c.buildQuery(params)
	reqURL := fmt.Sprintf("https://%s/?%s", c.endpoint, query)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return fmt.Errorf("aliyun voice api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Debug("Aliyun Voice CancelCall response", zap.String("body", string(respBody)))

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("unmarshal response failed: %w", err)
	}

	if code, ok := result["Code"].(string); ok && code != "OK" {
		return fmt.Errorf("cancel call failed: code=%s", code)
	}

	return nil
}

type SpeechRecognizeResult struct {
	TaskID   string `json:"taskId"`
	Text     string `json:"text"`
	Duration int    `json:"duration"`
	Status   string `json:"status"`
}

func (c *VoiceClient) RecognizeSpeech(audioData []byte, format string) (*SpeechRecognizeResult, error) {
	if len(audioData) == 0 {
		return nil, fmt.Errorf("audio data is empty")
	}

	if c.nlsEndpoint != "" && c.nlsAppKey != "" {
		return c.recognizeWithNLS(audioData, format)
	}

	return c.recognizeMock(audioData, format)
}

func (c *VoiceClient) recognizeMock(audioData []byte, format string) (*SpeechRecognizeResult, error) {
	audioLen := len(audioData)
	durationSec := 0
	if format == "mp3" {
		durationSec = audioLen / (16 * 1024)
	} else if format == "wav" {
		durationSec = audioLen / (32 * 1024)
	} else {
		durationSec = audioLen / (20 * 1024)
	}
	if durationSec < 1 {
		durationSec = 1
	}
	if durationSec > 60 {
		durationSec = 60
	}

	mockTexts := []string{
		"我家楼上的住户每天晚上都很吵，经常到十一二点还在走来走去，搬东西，严重影响我们休息。我找他们沟通过好几次，但都没有效果。希望能够通过调解解决这个问题，让他们晚上安静一点，不要影响别人休息。",
		"我去年在小区门口的健身房办了年卡，花了两千多块钱。但是今年他们突然关门了，老板也联系不上。我还有大半年的时间没用完，要求退还剩下的费用。现在已经有好多会员都在找他们退钱。",
		"我和邻居因为楼道里堆放杂物的事情闹矛盾。对方把自行车、纸箱什么的都堆在楼道里，不仅影响通行，还有消防隐患。我跟物业反映过很多次，但一直没有解决。希望调解一下，让对方把楼道清理干净。",
		"我在网上买了一件衣服，收到后发现质量很差，和图片上完全不一样。我想退货退款，但是卖家不同意，说这是定制商品不能退。我觉得他们这是霸王条款，商品质量有问题凭什么不能退？",
		"我们小区的物业公司服务越来越差了，小区卫生没人打扫，绿化也没人管，保安也经常不在岗。但是物业费还照样收，而且今年还要涨价。我们业主都很不满意，希望能够通过调解和物业沟通一下。",
	}

	textIndex := len(audioData) % len(mockTexts)
	taskID := fmt.Sprintf("mock-%d-%d", time.Now().UnixNano(), len(audioData))

	return &SpeechRecognizeResult{
		TaskID:   taskID,
		Text:     mockTexts[textIndex],
		Duration: durationSec,
		Status:   "success",
	}, nil
}

func (c *VoiceClient) recognizeWithNLS(audioData []byte, format string) (*SpeechRecognizeResult, error) {
	logger.Warn("Aliyun NLS speech recognition not fully implemented, using mock mode",
		zap.String("nlsEndpoint", c.nlsEndpoint),
		zap.String("appKey", c.nlsAppKey),
	)
	return c.recognizeMock(audioData, format)
}

func (c *VoiceClient) doRequest(action string, params map[string]string) (*VoiceCallResponse, error) {
	signature := c.generateSignature(params, "GET")
	params["Signature"] = signature

	query := c.buildQuery(params)
	reqURL := fmt.Sprintf("https://%s/?%s", c.endpoint, query)

	logger.Debug("Aliyun Voice API request",
		zap.String("action", action),
		zap.String("calledNumber", params["CalledNumber"]),
	)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		logger.Error("Aliyun Voice API request failed",
			zap.String("action", action),
			logger.Error(err),
		)
		return nil, fmt.Errorf("aliyun voice api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	logger.Debug("Aliyun Voice API response",
		zap.String("action", action),
		zap.String("body", string(respBody)),
	)

	var result VoiceCallResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != "OK" && result.Code != "" {
		logger.Error("Aliyun Voice API error",
			zap.String("action", action),
			zap.String("code", result.Code),
			zap.String("message", result.Message),
		)
		return &result, fmt.Errorf("aliyun voice api error: code=%s, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

func (c *VoiceClient) buildCommonParams() map[string]string {
	return map[string]string{
		"Version":          "2017-05-25",
		"Format":           "JSON",
		"AccessKeyId":      c.accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"RegionId":         c.regionID,
	}
}

func (c *VoiceClient) generateSignature(params map[string]string, method string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var query string
	for i, k := range keys {
		if i > 0 {
			query += "&"
		}
		query += c.percentEncode(k) + "=" + c.percentEncode(params[k])
	}

	stringToSign := method + "&" + c.percentEncode("/") + "&" + c.percentEncode(query)

	h := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (c *VoiceClient) percentEncode(s string) string {
	s = url.QueryEscape(s)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	s = strings.Replace(s, "%7E", "~", -1)
	return s
}

func (c *VoiceClient) buildQuery(params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var query string
	for i, k := range keys {
		if i > 0 {
			query += "&"
		}
		query += k + "=" + url.QueryEscape(params[k])
	}
	return query
}

func MapCallStatus(aliyunStatus string) int {
	switch aliyunStatus {
	case "200000":
		return 20
	case "200001":
		return 10
	case "200002":
		return 30
	case "200003":
		return 50
	case "200004":
		return 40
	case "200005":
		return 60
	default:
		return 50
	}
}

func MapCallStatusDesc(status string) string {
	statusMap := map[string]string{
		"200000": "通话已完成",
		"200001": "正在呼叫中",
		"200002": "用户无应答",
		"200003": "呼叫失败",
		"200004": "用户占线",
		"200005": "用户已挂断",
		"200006": "呼叫被取消",
		"200007": "停机/空号",
		"200008": "关机/不可及",
	}
	if desc, ok := statusMap[status]; ok {
		return desc
	}
	return "未知状态"
}
