package aliyun

import (
	"bytes"
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

type TingwuClient struct {
	accessKeyID     string
	accessKeySecret string
	regionID        string
	appKey          string
	callbackURL     string
	endpoint        string
	httpClient      *http.Client
}

var (
	tingwuClient *TingwuClient
	tingwuOnce   sync.Once
)

type SubmitTaskRequest struct {
	FileUrl            string `json:"fileUrl"`
	FileName           string `json:"fileName"`
	Format             string `json:"format"`
	TaskType           string `json:"taskType"`
	EnableDiarization  bool   `json:"enableDiarization"`
	SpeakerCount       int    `json:"speakerCount"`
	EnablePunctuation  bool   `json:"enablePunctuation"`
	EnableITN          bool   `json:"enableITN"`
	CustomVocabularyId string `json:"customVocabularyId,omitempty"`
	CallbackUrl        string `json:"callbackUrl,omitempty"`
}

type SubmitTaskResponse struct {
	RequestId string `json:"RequestId"`
	TaskId    string `json:"TaskId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
}

type TingwuSentence struct {
	Text      string `json:"Text"`
	BeginTime int    `json:"BeginTime"`
	EndTime   int    `json:"EndTime"`
	SpeakerId int    `json:"SpeakerId"`
}

type TaskResultResponse struct {
	RequestId      string           `json:"RequestId"`
	Code           string           `json:"Code"`
	Message        string           `json:"Message"`
	TaskId         string           `json:"TaskId"`
	Status         string           `json:"Status"`
	TranscriptText string           `json:"TranscriptText"`
	Duration       int              `json:"Duration"`
	WordCount      int              `json:"WordCount"`
	Sentences      []TingwuSentence `json:"Sentences"`
}

type CancelTaskResponse struct {
	RequestId string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	Success   bool   `json:"Success"`
}

func NewTingwuClient() *TingwuClient {
	cfg := config.GetConfig()
	endpoint := cfg.AliyunTingwu.Endpoint
	if endpoint == "" {
		endpoint = "tingwu.cn-hangzhou.aliyuncs.com"
	}
	return &TingwuClient{
		accessKeyID:     cfg.AliyunTingwu.AccessKeyID,
		accessKeySecret: cfg.AliyunTingwu.AccessKeySecret,
		regionID:        cfg.AliyunTingwu.RegionID,
		appKey:          cfg.AliyunTingwu.AppKey,
		callbackURL:     cfg.AliyunTingwu.CallbackURL,
		endpoint:        endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func InitTingwuClient() {
	tingwuOnce.Do(func() {
		tingwuClient = NewTingwuClient()
		logger.Info("Aliyun Tingwu client initialized",
			zap.String("endpoint", tingwuClient.endpoint),
			zap.String("appKey", tingwuClient.appKey),
		)
	})
}

func GetTingwuClient() *TingwuClient {
	if tingwuClient == nil {
		InitTingwuClient()
	}
	return tingwuClient
}

func (c *TingwuClient) SubmitTask(fileUrl, fileName, format string, enableDiarization bool, speakerCount int, customVocabularyId, callbackUrl string) (string, string, error) {
	req := &SubmitTaskRequest{
		FileUrl:           fileUrl,
		FileName:          fileName,
		Format:            format,
		TaskType:          "transcription",
		EnableDiarization: enableDiarization,
		SpeakerCount:      speakerCount,
		EnablePunctuation: true,
		EnableITN:         true,
		CustomVocabularyId: customVocabularyId,
	}

	if callbackUrl != "" {
		req.CallbackUrl = callbackUrl
	} else if c.callbackURL != "" {
		req.CallbackUrl = c.callbackURL
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("marshal request failed: %w", err)
	}

	apiPath := "/openapi/tingwu/2023-09-30/tasks"
	result, err := c.doRequest("POST", apiPath, body)
	if err != nil {
		return "", "", err
	}

	var resp SubmitTaskResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	if resp.Code != "" && resp.Code != "200" && resp.Code != "OK" {
		logger.Error("Aliyun Tingwu SubmitTask error",
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return "", "", fmt.Errorf("tingwu submit task error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return resp.TaskId, resp.RequestId, nil
}

func (c *TingwuClient) GetTaskResult(taskId string) (string, string, int, int, []TingwuSentence, error) {
	apiPath := fmt.Sprintf("/openapi/tingwu/2023-09-30/tasks/%s", taskId)

	result, err := c.doRequest("GET", apiPath, nil)
	if err != nil {
		return "", "", 0, 0, nil, err
	}

	var resp TaskResultResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", "", 0, 0, nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if resp.Code != "" && resp.Code != "200" && resp.Code != "OK" {
		logger.Error("Aliyun Tingwu GetTaskResult error",
			zap.String("taskId", taskId),
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return "", "", 0, 0, nil, fmt.Errorf("tingwu get task result error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return resp.Status, resp.TranscriptText, resp.Duration, resp.WordCount, resp.Sentences, nil
}

func (c *TingwuClient) CancelTask(taskId string) (bool, error) {
	apiPath := fmt.Sprintf("/openapi/tingwu/2023-09-30/tasks/%s", taskId)

	result, err := c.doRequest("DELETE", apiPath, nil)
	if err != nil {
		return false, err
	}

	var resp CancelTaskResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return false, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if resp.Code != "" && resp.Code != "200" && resp.Code != "OK" {
		logger.Error("Aliyun Tingwu CancelTask error",
			zap.String("taskId", taskId),
			zap.String("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return false, fmt.Errorf("tingwu cancel task error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return resp.Success, nil
}

func (c *TingwuClient) doRequest(method, apiPath string, body []byte) ([]byte, error) {
	params := c.buildCommonParams()
	params["Action"] = apiPathToAction(apiPath, method)

	signature := c.generateSignature(params, method)
	params["Signature"] = signature

	query := c.buildQuery(params)
	reqURL := fmt.Sprintf("https://%s%s?%s", c.endpoint, apiPath, query)

	logger.Debug("Aliyun Tingwu API request",
		zap.String("method", method),
		zap.String("path", apiPath),
	)

	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, reqURL, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, reqURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.appKey != "" {
		req.Header.Set("X-App-Key", c.appKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Aliyun Tingwu API request failed",
			zap.String("method", method),
			zap.String("path", apiPath),
			logger.Error(err),
		)
		return nil, fmt.Errorf("tingwu api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	logger.Debug("Aliyun Tingwu API response",
		zap.String("method", method),
		zap.String("path", apiPath),
		zap.String("body", string(respBody)),
	)

	return respBody, nil
}

func apiPathToAction(apiPath, method string) string {
	switch {
	case strings.Contains(apiPath, "/tasks") && method == "POST":
		return "SubmitTask"
	case strings.Contains(apiPath, "/tasks") && method == "GET":
		return "GetTaskResult"
	case strings.Contains(apiPath, "/tasks") && method == "DELETE":
		return "CancelTask"
	default:
		return ""
	}
}

func (c *TingwuClient) buildCommonParams() map[string]string {
	return map[string]string{
		"Version":          "2023-09-30",
		"Format":           "JSON",
		"AccessKeyId":      c.accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"RegionId":         c.regionID,
		"AppKey":           c.appKey,
	}
}

func (c *TingwuClient) generateSignature(params map[string]string, method string) string {
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

func (c *TingwuClient) percentEncode(s string) string {
	s = url.QueryEscape(s)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	s = strings.Replace(s, "%7E", "~", -1)
	return s
}

func (c *TingwuClient) buildQuery(params map[string]string) string {
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
