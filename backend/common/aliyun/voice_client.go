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
