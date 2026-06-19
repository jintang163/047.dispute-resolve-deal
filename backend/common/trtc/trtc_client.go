package trtc

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type TRTCClient struct {
	sdkAppID   uint32
	secretKey  string
	adminUser  string
	httpClient *http.Client
}

var trtcClient *TRTCClient

func InitTRTC() {
	cfg := config.GetConfig()
	trtcClient = &TRTCClient{
		sdkAppID:  cfg.TRTC.SdkAppID,
		secretKey: cfg.TRTC.SecretKey,
		adminUser: cfg.TRTC.AdminUserID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
	}
	logger.Info("TRTC client initialized",
		zap.Uint32("sdkAppID", trtcClient.sdkAppID),
		zap.String("adminUser", trtcClient.adminUser),
	)
}

func GetTRTCClient() *TRTCClient {
	if trtcClient == nil {
		InitTRTC()
	}
	return trtcClient
}

func (c *TRTCClient) GenUserSig(userID string, expireSeconds uint32) string {
	content := genSigContent(c.sdkAppID, userID, expireSeconds)
	contentStr, _ := json.Marshal(content)
	serialized := string(contentStr)

	sig := hmacSha256(c.secretKey, serialized)
	sigWithContent := serialized + "." + sig

	return base64UrlEncode([]byte(sigWithContent))
}

func (c *TRTCClient) GetSdkAppID() uint32 {
	return c.sdkAppID
}

func (c *TRTCClient) GenUserSigWithBuffer(userID string, expireSeconds uint32) string {
	return c.GenUserSig(userID, expireSeconds)
}

func (c *TRTCClient) CreateCloudRecord(req *CreateRecordRequest) (*CreateRecordResponse, error) {
	apiURL := "https://trtc.tencentcloudapi.com/"

	body := map[string]interface{}{
		"SdkAppId": c.sdkAppID,
		"RoomId":   req.RoomID,
		"UserId":   req.UserID,
		"UserSig":  c.GenUserSig(req.UserID, 86400),
		"RecordParams": map[string]interface{}{
			"RecordMode":     req.RecordMode,
			"MaxIdleTime":    30,
			"StreamType":     0,
			"OutputFormat":   0,
			"AvMerge":        1,
			"MaxRecordingDuration": req.MaxDuration,
		},
		"StorageParams": map[string]interface{}{
			"CloudStorage": map[string]interface{}{
				"Vendor":   0,
				"Region":   req.StorageRegion,
				"Bucket":   req.StorageBucket,
				"AccessKey": req.StorageAccessKey,
				"SecretKey": req.StorageSecretKey,
				"FileNamePrefix": []string{
					req.StoragePath,
					req.RoomID,
				},
			},
		},
	}

	if req.RecordMode == 2 {
		body["RecordParams"].(map[string]interface{})["MixLayout"] = 1
		body["RecordParams"].(map[string]interface{})["MixLayoutMode"] = 1
	}

	if req.SubscribeStreamUserIds != nil {
		body["SubscribeStreamUserIds"] = req.SubscribeStreamUserIds
	}

	respData, err := c.callTRTCApi("CreateCloudRecording", body)
	if err != nil {
		return nil, fmt.Errorf("create cloud record failed: %w", err)
	}

	var result CreateRecordResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse create record response failed: %w", err)
	}

	return &result, nil
}

func (c *TRTCClient) StopCloudRecord(taskID string, roomID int64) (*StopRecordResponse, error) {
	body := map[string]interface{}{
		"SdkAppId":     c.sdkAppID,
		"TaskId":       taskID,
		"RoomId":       roomID,
	}

	respData, err := c.callTRTCApi("StopCloudRecording", body)
	if err != nil {
		return nil, fmt.Errorf("stop cloud record failed: %w", err)
	}

	var result StopRecordResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse stop record response failed: %w", err)
	}

	return &result, nil
}

func (c *TRTCClient) ModifyCloudRecord(taskID string, roomID int64, req *ModifyRecordRequest) error {
	body := map[string]interface{}{
		"SdkAppId": c.sdkAppID,
		"TaskId":   taskID,
		"RoomId":   roomID,
	}

	if req.SubscribeStreamUserIds != nil {
		body["SubscribeStreamUserIds"] = req.SubscribeStreamUserIds
	}

	_, err := c.callTRTCApi("ModifyCloudRecording", body)
	if err != nil {
		return fmt.Errorf("modify cloud record failed: %w", err)
	}

	return nil
}

func (c *TRTCClient) DescribeCloudRecord(taskID string, roomID int64) (*DescribeRecordResponse, error) {
	body := map[string]interface{}{
		"SdkAppId": c.sdkAppID,
		"TaskId":   taskID,
		"RoomId":   roomID,
	}

	respData, err := c.callTRTCApi("DescribeCloudRecording", body)
	if err != nil {
		return nil, fmt.Errorf("describe cloud record failed: %w", err)
	}

	var result DescribeRecordResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("parse describe record response failed: %w", err)
	}

	return &result, nil
}

func (c *TRTCClient) callTRTCApi(action string, payload map[string]interface{}) (json.RawMessage, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://trtc.tencentcloudapi.com/", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	timestamp := time.Now().Unix()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", "2022-06-01")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", "ap-guangzhou")

	logger.Debug("TRTC API call",
		zap.String("action", action),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("trtc api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("TRTC API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("trtc api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Response json.RawMessage `json:"Response"`
		Error    *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse api response failed: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("trtc api error: code=%s, message=%s", apiResp.Error.Code, apiResp.Error.Message)
	}

	return apiResp.Response, nil
}

type CreateRecordRequest struct {
	RoomID               int64    `json:"roomId"`
	UserID               string   `json:"userId"`
	RecordMode           int      `json:"recordMode"`
	MaxDuration          int      `json:"maxDuration"`
	StorageRegion        string   `json:"storageRegion"`
	StorageBucket        string   `json:"storageBucket"`
	StorageAccessKey     string   `json:"storageAccessKey"`
	StorageSecretKey     string   `json:"storageSecretKey"`
	StoragePath          string   `json:"storagePath"`
	SubscribeStreamUserIds *struct {
		SubscribeAudioUserIds []string `json:"SubscribeAudioUserIds,omitempty"`
		SubscribeVideoUserIds []string `json:"SubscribeVideoUserIds,omitempty"`
		UnSubscribeAudioUserIds []string `json:"UnSubscribeAudioUserIds,omitempty"`
		UnSubscribeVideoUserIds []string `json:"UnSubscribeVideoUserIds,omitempty"`
	} `json:"subscribeStreamUserIds,omitempty"`
}

type CreateRecordResponse struct {
	TaskId string `json:"TaskId"`
	RequestId string `json:"RequestId"`
}

type StopRecordResponse struct {
	TaskId string `json:"TaskId"`
	RequestId string `json:"RequestId"`
}

type ModifyRecordRequest struct {
	SubscribeStreamUserIds *struct {
		SubscribeAudioUserIds []string `json:"SubscribeAudioUserIds,omitempty"`
		SubscribeVideoUserIds []string `json:"SubscribeVideoUserIds,omitempty"`
		UnSubscribeAudioUserIds []string `json:"UnSubscribeAudioUserIds,omitempty"`
		UnSubscribeVideoUserIds []string `json:"UnSubscribeVideoUserIds,omitempty"`
	} `json:"subscribeStreamUserIds,omitempty"`
}

type DescribeRecordResponse struct {
	TaskId        string `json:"TaskId"`
	Status        int    `json:"Status"`
	RequestId     string `json:"RequestId"`
	StorageFileList []struct {
		FileUrl     string `json:"FileUrl"`
		UserId      string `json:"UserId"`
		RecordFormat string `json:"RecordFormat"`
		TrackType   string `json:"TrackType"`
		StartTimeMs int64  `json:"StartTimeMs"`
		EndTimeMs   int64  `json:"EndTimeMs"`
	} `json:"StorageFileList,omitempty"`
}

func genSigContent(sdkAppID uint32, userID string, expireSeconds uint32) map[string]interface{} {
	currTime := uint32(time.Now().Unix())
	return map[string]interface{}{
		"TLS.ver":          "2.0",
		"TLS.identifier":   userID,
		"TLS.sdkappid":     sdkAppID,
		"TLS.expire":       expireSeconds,
		"TLS.time":         currTime,
		"TLS.sig":          "",
		"TLS.toast":        "",
		"TLS.cloud":        "",
	}
}

func hmacSha256(key string, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64UrlEncode(h.Sum(nil))
}

func base64UrlEncode(data []byte) string {
	result := base64.StdEncoding.EncodeToString(data)
	result = replaceChars(result)
	return result
}

func replaceChars(s string) string {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '+':
			buf = append(buf, '*')
		case '/':
			buf = append(buf, '-')
		case '=':
			buf = append(buf, '_')
		default:
			buf = append(buf, s[i])
		}
	}
	return string(buf)
}

func uint32ToBytes(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return buf
}
