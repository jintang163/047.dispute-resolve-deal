package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
)

type BlockchainClient struct {
	apiEndpoint   string
	appCode       string
	appKey        string
	appSecret     string
	chainName     string
	contractAddr  string
	certTplID     string
	qrCodeBaseURL string
	httpClient    *http.Client
}

var defaultClient *BlockchainClient

func InitClient() {
	cfg := config.GetConfig()
	if cfg == nil {
		return
	}
	defaultClient = &BlockchainClient{
		apiEndpoint:   cfg.Blockchain.APIEndpoint,
		appCode:       cfg.Blockchain.AppCode,
		appKey:        cfg.Blockchain.AppKey,
		appSecret:     cfg.Blockchain.AppSecret,
		chainName:     cfg.Blockchain.ChainName,
		contractAddr:  cfg.Blockchain.ContractAddr,
		certTplID:     cfg.Blockchain.CertTemplateID,
		qrCodeBaseURL: cfg.Blockchain.QRCodeBaseURL,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

func GetClient() *BlockchainClient {
	if defaultClient == nil {
		InitClient()
	}
	return defaultClient
}

type BCResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type StoreEvidenceReq struct {
	EvidenceID   string `json:"evidenceId"`
	EvidenceType string `json:"evidenceType"`
	EvidenceHash string `json:"evidenceHash"`
	EvidenceName string `json:"evidenceName"`
	Description  string `json:"description"`
	Metadata     string `json:"metadata"`
	ChainName    string `json:"chainName"`
}

type StoreEvidenceResult struct {
	TxID        string `json:"txId"`
	BlockHeight int64  `json:"blockHeight"`
	Timestamp   string `json:"timestamp"`
	CertNo      string `json:"certNo"`
}

type VerifyResult struct {
	Valid       bool   `json:"valid"`
	TxID        string `json:"txId"`
	BlockHeight int64  `json:"blockHeight"`
	Timestamp   string `json:"timestamp"`
	EvidenceHash string `json:"evidenceHash"`
	CertNo      string `json:"certNo"`
}

type CertificateResult struct {
	CertNo       string `json:"certNo"`
	CertURL      string `json:"certUrl"`
	QRCodeURL    string `json:"qrcodeUrl"`
	VerifyURL    string `json:"verifyUrl"`
	TxID         string `json:"txId"`
	BlockHeight  int64  `json:"blockHeight"`
	OnChainTime  string `json:"onChainTime"`
}

func ComputeSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *BlockchainClient) StoreEvidence(req *StoreEvidenceReq) (*StoreEvidenceResult, error) {
	if req.ChainName == "" {
		req.ChainName = c.chainName
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", c.apiEndpoint+"/api/v1/evidence/store", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建存证请求失败: %v", err)
	}

	c.setAuthHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Blockchain store evidence failed", logger.Error(err))
		return nil, fmt.Errorf("区块链存证失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result BCResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("区块链响应解析失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("区块链存证错误: %s", result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var storeResult StoreEvidenceResult
	json.Unmarshal(dataBytes, &storeResult)

	return &storeResult, nil
}

func (c *BlockchainClient) VerifyEvidence(certNo string) (*VerifyResult, error) {
	url := fmt.Sprintf("%s/api/v1/evidence/verify?certNo=%s", c.apiEndpoint, certNo)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建验证请求失败: %v", err)
	}

	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("区块链验证失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result BCResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("区块链响应解析失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("区块链验证错误: %s", result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var verifyResult VerifyResult
	json.Unmarshal(dataBytes, &verifyResult)

	return &verifyResult, nil
}

func (c *BlockchainClient) GetCertificate(certNo string) (*CertificateResult, error) {
	url := fmt.Sprintf("%s/api/v1/evidence/certificate?certNo=%s&templateId=%s", c.apiEndpoint, certNo, c.certTplID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建证书请求失败: %v", err)
	}

	c.setAuthHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("获取存证证书失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result BCResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("区块链响应解析失败: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("获取证书错误: %s", result.Message)
	}

	dataBytes, _ := json.Marshal(result.Data)
	var certResult CertificateResult
	json.Unmarshal(dataBytes, &certResult)

	if certResult.VerifyURL == "" {
		certResult.VerifyURL = fmt.Sprintf("%s/%s", c.qrCodeBaseURL, certNo)
	}

	return &certResult, nil
}

func (c *BlockchainClient) StorePDFHash(pdfData []byte, evidenceID, evidenceName, metadata string) (*StoreEvidenceResult, error) {
	pdfHash := ComputeSHA256(pdfData)

	req := &StoreEvidenceReq{
		EvidenceID:   evidenceID,
		EvidenceType: "mediation_protocol",
		EvidenceHash: pdfHash,
		EvidenceName: evidenceName,
		Description:  fmt.Sprintf("调解协议书PDF上链存证: %s", evidenceName),
		Metadata:     metadata,
	}

	return c.StoreEvidence(req)
}

func (c *BlockchainClient) setAuthHeaders(req *http.Request) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signStr := c.appKey + timestamp + c.appSecret
	sign := ComputeSHA256([]byte(signStr))

	req.Header.Set("X-App-Code", c.appCode)
	req.Header.Set("X-App-Key", c.appKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Sign", sign)
}
