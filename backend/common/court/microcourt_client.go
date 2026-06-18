package court

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"go.uber.org/zap"
)

type MicroCourtClient struct {
	baseURL    string
	appID      string
	appSecret  string
	publicKey  string
	httpClient *http.Client
}

var microCourtClient *MicroCourtClient

func NewMicroCourtClient(courtCfg *model.CourtConfig) *MicroCourtClient {
	return &MicroCourtClient{
		baseURL:   courtCfg.APIEndpoint,
		appID:     courtCfg.APIAppID,
		appSecret: courtCfg.APISecret,
		publicKey: courtCfg.APIPublicKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func InitMicroCourt() {
	cfg := config.GetConfig()
	if cfg.Court.APIEndpoint == "" {
		logger.Warn("MicroCourt API not configured, skipping initialization")
		return
	}
	microCourtClient = &MicroCourtClient{
		baseURL:   cfg.Court.APIEndpoint,
		appID:     cfg.Court.APIAppID,
		appSecret: cfg.Court.APISecret,
		publicKey: cfg.Court.APIPublicKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	logger.Info("MicroCourt client initialized", zap.String("baseURL", microCourtClient.baseURL))
}

func GetMicroCourtClient() *MicroCourtClient {
	if microCourtClient == nil {
		InitMicroCourt()
	}
	return microCourtClient
}

func (c *MicroCourtClient) generateSignature(params map[string]interface{}, timestamp string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signStr strings.Builder
	signStr.WriteString(c.appSecret)
	signStr.WriteString(timestamp)
	for _, k := range keys {
		signStr.WriteString(fmt.Sprintf("%s=%v", k, params[k]))
	}
	signStr.WriteString(c.appSecret)

	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(signStr.String()))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type MicroCourtResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Sign    string      `json:"sign"`
}

func (c *MicroCourtClient) doRequest(method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	var reqBody io.Reader
	var params map[string]interface{}

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)

		if err := json.Unmarshal(jsonData, &params); err != nil {
			params = make(map[string]interface{})
		}
	} else {
		params = make(map[string]interface{})
	}

	signature := c.generateSignature(params, timestamp)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-App-ID", c.appID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	logger.Debug("MicroCourt request",
		zap.String("method", method),
		zap.String("url", url),
		zap.String("appId", c.appID),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("microcourt request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	logger.Debug("MicroCourt response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("microcourt api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var apiResp MicroCourtResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
	}

	if apiResp.Code != 0 && apiResp.Code != 200 {
		return fmt.Errorf("microcourt business error: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	if result != nil && apiResp.Data != nil {
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return fmt.Errorf("marshal data failed: %w", err)
		}
		if err := json.Unmarshal(dataBytes, result); err != nil {
			return fmt.Errorf("unmarshal result failed: %w", err)
		}
	}

	return nil
}

type SubmitConfirmationRequest struct {
	ConfirmNo         string `json:"confirmNo"`
	CaseNo            string `json:"caseNo"`
	CaseTitle         string `json:"caseTitle"`
	ApplicantName     string `json:"applicantName"`
	ApplicantPhone    string `json:"applicantPhone"`
	ApplicantIdCard   string `json:"applicantIdCard"`
	ApplicantAddress  string `json:"applicantAddress"`
	RespondentName    string `json:"respondentName"`
	RespondentPhone   string `json:"respondentPhone"`
	RespondentIdCard  string `json:"respondentIdCard"`
	RespondentAddress string `json:"respondentAddress"`
	AgreementContent  string `json:"agreementContent"`
	ConfirmAmount     string `json:"confirmAmount"`
	PerformanceDeadline string `json:"performanceDeadline"`
	MediationProtocol string `json:"mediationProtocol"`
	EvidenceList      []string `json:"evidenceList"`
}

type SubmitConfirmationResponse struct {
	CourtCaseNo      string `json:"courtCaseNo"`
	CourtAcceptTime  string `json:"courtAcceptTime"`
	EstimatedDays    int    `json:"estimatedDays"`
}

func (c *MicroCourtClient) SubmitConfirmation(req *SubmitConfirmationRequest) (*SubmitConfirmationResponse, error) {
	path := "/api/v1/judicial/confirmation/submit"

	var resp SubmitConfirmationResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Submit confirmation to court failed",
			zap.String("confirmNo", req.ConfirmNo),
			logger.Error(err),
		)
		return nil, err
	}

	logger.Info("Confirmation submitted to court",
		zap.String("confirmNo", req.ConfirmNo),
		zap.String("courtCaseNo", resp.CourtCaseNo),
	)

	return &resp, nil
}

type QueryConfirmationStatusRequest struct {
	ConfirmNo   string `json:"confirmNo"`
	CourtCaseNo string `json:"courtCaseNo"`
}

type QueryConfirmationStatusResponse struct {
	ConfirmNo       string `json:"confirmNo"`
	CourtCaseNo     string `json:"courtCaseNo"`
	Status          int    `json:"status"`
	StatusName      string `json:"statusName"`
	ReviewOpinion   string `json:"reviewOpinion"`
	ConfirmDate     string `json:"confirmDate"`
	ConfirmCourt    string `json:"confirmCourt"`
	ConfirmJudge    string `json:"confirmJudge"`
	ConfirmDocumentNo string `json:"confirmDocumentNo"`
	DocumentUrl     string `json:"documentUrl"`
	SealStatus      int    `json:"sealStatus"`
	SealTime        string `json:"sealTime"`
}

func (c *MicroCourtClient) QueryConfirmationStatus(confirmNo, courtCaseNo string) (*QueryConfirmationStatusResponse, error) {
	path := "/api/v1/judicial/confirmation/status"

	req := QueryConfirmationStatusRequest{
		ConfirmNo:   confirmNo,
		CourtCaseNo: courtCaseNo,
	}

	var resp QueryConfirmationStatusResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Query confirmation status from court failed",
			zap.String("confirmNo", confirmNo),
			logger.Error(err),
		)
		return nil, err
	}

	logger.Info("Confirmation status queried from court",
		zap.String("confirmNo", confirmNo),
		zap.Int("status", resp.Status),
		zap.String("statusName", resp.StatusName),
	)

	return &resp, nil
}

type GetConfirmationDocumentRequest struct {
	ConfirmNo       string `json:"confirmNo"`
	CourtCaseNo     string `json:"courtCaseNo"`
	DocumentNo      string `json:"documentNo"`
}

type GetConfirmationDocumentResponse struct {
	DocumentNo      string `json:"documentNo"`
	DocumentName    string `json:"documentName"`
	DocumentType    string `json:"documentType"`
	DocumentContent string `json:"documentContent"`
	SealStatus      int    `json:"sealStatus"`
	SealImage       string `json:"sealImage"`
}

func (c *MicroCourtClient) GetConfirmationDocument(confirmNo, courtCaseNo, documentNo string) (*GetConfirmationDocumentResponse, error) {
	path := "/api/v1/judicial/confirmation/document"

	req := GetConfirmationDocumentRequest{
		ConfirmNo:   confirmNo,
		CourtCaseNo: courtCaseNo,
		DocumentNo:  documentNo,
	}

	var resp GetConfirmationDocumentResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Get confirmation document from court failed",
			zap.String("confirmNo", confirmNo),
			zap.String("documentNo", documentNo),
			logger.Error(err),
		)
		return nil, err
	}

	logger.Info("Confirmation document retrieved from court",
		zap.String("confirmNo", confirmNo),
		zap.String("documentNo", documentNo),
	)

	return &resp, nil
}

type SealDocumentRequest struct {
	ConfirmNo       string `json:"confirmNo"`
	CourtCaseNo     string `json:"courtCaseNo"`
	DocumentNo      string `json:"documentNo"`
	SealCertNo      string `json:"sealCertNo"`
	SealPosition    string `json:"sealPosition"`
}

type SealDocumentResponse struct {
	SealStatus      int    `json:"sealStatus"`
	SealTime        string `json:"sealTime"`
	SealedDocumentUrl string `json:"sealedDocumentUrl"`
}

func (c *MicroCourtClient) SealDocument(confirmNo, courtCaseNo, documentNo, sealCertNo string) (*SealDocumentResponse, error) {
	path := "/api/v1/judicial/confirmation/seal"

	req := SealDocumentRequest{
		ConfirmNo:    confirmNo,
		CourtCaseNo:  courtCaseNo,
		DocumentNo:   documentNo,
		SealCertNo:   sealCertNo,
		SealPosition: "bottom-right",
	}

	var resp SealDocumentResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Seal document with court failed",
			zap.String("confirmNo", confirmNo),
			zap.String("documentNo", documentNo),
			logger.Error(err),
		)
		return nil, err
	}

	logger.Info("Document sealed by court",
		zap.String("confirmNo", confirmNo),
		zap.String("documentNo", documentNo),
		zap.String("sealedDocumentUrl", resp.SealedDocumentUrl),
	)

	return &resp, nil
}

type SendReminderRequest struct {
	ConfirmNo       string `json:"confirmNo"`
	ReminderType    int    `json:"reminderType"`
	ReminderContent string `json:"reminderContent"`
	TargetPhone     string `json:"targetPhone"`
}

type SendReminderResponse struct {
	ReminderID string `json:"reminderId"`
	SentTime   string `json:"sentTime"`
}

func (c *MicroCourtClient) SendReminder(confirmNo string, reminderType int, reminderContent, targetPhone string) (*SendReminderResponse, error) {
	path := "/api/v1/judicial/confirmation/reminder"

	req := SendReminderRequest{
		ConfirmNo:       confirmNo,
		ReminderType:    reminderType,
		ReminderContent: reminderContent,
		TargetPhone:     targetPhone,
	}

	var resp SendReminderResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Send reminder via court failed",
			zap.String("confirmNo", confirmNo),
			zap.Int("reminderType", reminderType),
			logger.Error(err),
		)
		return nil, err
	}

	logger.Info("Reminder sent via court",
		zap.String("confirmNo", confirmNo),
		zap.String("reminderId", resp.ReminderID),
	)

	return &resp, nil
}

func ConvertCourtStatusToLocal(courtStatus int) int32 {
	switch courtStatus {
	case 10:
		return model.JudicialStatusSubmitted
	case 20, 21:
		return model.JudicialStatusReviewing
	case 30:
		return model.JudicialStatusConfirmed
	case 40, 41:
		return model.JudicialStatusRejected
	case 50:
		return model.JudicialStatusExpired
	default:
		return model.JudicialStatusSubmitted
	}
}

func GetStatusName(status int32) string {
	switch status {
	case model.JudicialStatusSubmitted:
		return "已提交"
	case model.JudicialStatusReviewing:
		return "审核中"
	case model.JudicialStatusConfirmed:
		return "已确认"
	case model.JudicialStatusRejected:
		return "已驳回"
	case model.JudicialStatusExpired:
		return "已失效"
	default:
		return "未知"
	}
}
