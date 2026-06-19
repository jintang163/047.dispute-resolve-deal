package fadada

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
)

type FaDaDaClient struct {
	domain      string
	appID       string
	appSecret   string
	customerID  string
	notifyURL   string
	autoSeal    bool
	crossPage   bool
	httpClient  *http.Client
}

var defaultClient *FaDaDaClient

func InitClient() {
	cfg := config.GetConfig()
	if cfg == nil {
		return
	}
	defaultClient = &FaDaDaClient{
		domain:     cfg.FaDaDa.APIDomain,
		appID:      cfg.FaDaDa.AppID,
		appSecret:  cfg.FaDaDa.AppSecret,
		customerID: cfg.FaDaDa.CustomerID,
		notifyURL:  cfg.FaDaDa.NotifyURL,
		autoSeal:   cfg.FaDaDa.AutoSeal,
		crossPage:  cfg.FaDaDa.CrossPageSeal,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func GetClient() *FaDaDaClient {
	if defaultClient == nil {
		InitClient()
	}
	return defaultClient
}

type CreateSignFlowReq struct {
	DocTitle    string `json:"docTitle"`
	DocURL      string `json:"docUrl"`
	SignerIDs   []FaDaDaSigner `json:"signers"`
	ExpireHours int    `json:"expireHours"`
	NotifyURL   string `json:"notifyUrl"`
	CrossPageSeal bool  `json:"crossPageSeal"`
}

type FaDaDaSigner struct {
	CustomerID  string `json:"customerId"`
	CustomerName string `json:"customerName"`
	SignOrder   int    `json:"signOrder"`
	SignType    string `json:"signType"`
	AutoSign    bool   `json:"autoSign"`
}

type FaDaDaResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

type SignFlowResult struct {
	FlowID      string `json:"flowId"`
	ShortURL    string `json:"shortUrl"`
	QRCodeURL   string `json:"qrcodeUrl"`
}

type SignProgressResult struct {
	FlowID       string              `json:"flowId"`
	Status       string              `json:"status"`
	SignedCount  int                 `json:"signedCount"`
	TotalCount   int                 `json:"totalCount"`
	Signers      []SignerProgress    `json:"signers"`
}

type SignerProgress struct {
	CustomerID  string `json:"customerId"`
	CustomerName string `json:"customerName"`
	Status      string `json:"status"`
	SignedAt    string `json:"signedAt"`
	SignURL     string `json:"signUrl"`
}

func (c *FaDaDaClient) CreateSignFlow(req *CreateSignFlowReq) (*SignFlowResult, error) {
	if c.crossPage {
		req.CrossPageSeal = true
	}
	if req.NotifyURL == "" {
		req.NotifyURL = c.notifyURL
	}

	params := map[string]string{
		"app_id":         c.appID,
		"timestamp":      fmt.Sprintf("%d", time.Now().Unix()),
		"v":              "2.0",
		"customer_id":    c.customerID,
		"doc_title":      req.DocTitle,
		"notify_url":     req.NotifyURL,
		"expire_time":    fmt.Sprintf("%d", req.ExpireHours),
	}

	signersJSON, _ := json.Marshal(req.SignerIDs)
	params["signers"] = string(signersJSON)

	if req.CrossPageSeal {
		params["cross_page_seal"] = "1"
	}

	params["sign"] = c.generateSign(params)

	signersStr := string(signersJSON)
	formData := url.Values{}
	for k, v := range params {
		if k != "signers" {
			formData.Set(k, v)
		}
	}
	formData.Set("signers", signersStr)

	resp, err := c.httpClient.PostForm(
		c.domain+"/api/signFlow/create",
		formData,
	)
	if err != nil {
		logger.Error("FaDaDa create sign flow failed", logger.Error(err))
		return nil, fmt.Errorf("法大大创建签署流程失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result FaDaDaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("法大大响应解析失败: %v", err)
	}

	if result.Code != "0" && result.Code != "1" {
		return nil, fmt.Errorf("法大大接口错误: %s", result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var signResult SignFlowResult
	json.Unmarshal(dataBytes, &signResult)

	return &signResult, nil
}

func (c *FaDaDaClient) GetSignProgress(flowID string) (*SignProgressResult, error) {
	params := map[string]string{
		"app_id":    c.appID,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"v":         "2.0",
		"flow_id":   flowID,
	}
	params["sign"] = c.generateSign(params)

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := c.httpClient.PostForm(
		c.domain+"/api/signFlow/detail",
		formData,
	)
	if err != nil {
		return nil, fmt.Errorf("法大大查询签署进度失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result FaDaDaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("法大大响应解析失败: %v", err)
	}

	if result.Code != "0" && result.Code != "1" {
		return nil, fmt.Errorf("法大大接口错误: %s", result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var progress SignProgressResult
	json.Unmarshal(dataBytes, &progress)

	return &progress, nil
}

func (c *FaDaDaClient) GetSignedDocument(flowID string) (string, error) {
	params := map[string]string{
		"app_id":    c.appID,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"v":         "2.0",
		"flow_id":   flowID,
	}
	params["sign"] = c.generateSign(params)

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := c.httpClient.PostForm(
		c.domain+"/api/signFlow/download",
		formData,
	)
	if err != nil {
		return "", fmt.Errorf("法大大下载签署文件失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result FaDaDaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("法大大响应解析失败: %v", err)
	}

	if result.Code != "0" && result.Code != "1" {
		return "", fmt.Errorf("法大大接口错误: %s", result.Message)
	}

	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("法大大响应数据格式错误")
	}

	docURL, _ := dataMap["downloadUrl"].(string)
	return docURL, nil
}

func (c *FaDaDaClient) RevokeSignFlow(flowID string, reason string) error {
	params := map[string]string{
		"app_id":    c.appID,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"v":         "2.0",
		"flow_id":   flowID,
		"reason":    reason,
	}
	params["sign"] = c.generateSign(params)

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := c.httpClient.PostForm(
		c.domain+"/api/signFlow/revoke",
		formData,
	)
	if err != nil {
		return fmt.Errorf("法大大撤销签署流程失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result FaDaDaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("法大大响应解析失败: %v", err)
	}

	if result.Code != "0" && result.Code != "1" {
		return fmt.Errorf("法大大接口错误: %s", result.Message)
	}

	return nil
}

func (c *FaDaDaClient) VerifyCallback(timestamp, flowID, status, sign string) bool {
	params := map[string]string{
		"app_id":    c.appID,
		"timestamp": timestamp,
		"flow_id":   flowID,
		"status":    status,
	}
	expectedSign := c.generateSign(params)
	return strings.EqualFold(expectedSign, sign)
}

func (c *FaDaDaClient) generateSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params[k])
		buf.WriteString("&")
	}

	raw := strings.TrimSuffix(buf.String(), "&") + c.appSecret

	hash := md5.Sum([]byte(raw))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
