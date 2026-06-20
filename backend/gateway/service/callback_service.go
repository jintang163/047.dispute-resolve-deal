package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/aliyun"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CallbackService interface {
	CreateCallbackRecord(ctx context.Context, caseID int64) (*model.CallbackRecord, error)
	ScheduleCallback(ctx context.Context, record *model.CallbackRecord) error
	InitiateCall(ctx context.Context, recordID int64) error
	ProcessCallResult(ctx context.Context, callID string) error
	HandleCallback(ctx context.Context, data map[string]interface{}) error
	GetCallbackList(ctx context.Context, caseID int64) ([]model.CallbackRecord, error)
	GetCallbackDetail(ctx context.Context, recordID int64) (*model.CallbackRecord, error)
	RetryCallback(ctx context.Context, recordID int64) error
	CancelCallback(ctx context.Context, recordID int64) error
	ProcessPendingCallbacks(ctx context.Context) error
	ProcessRetryCallbacks(ctx context.Context) error
	CleanExpiredRecordings(ctx context.Context) error
	DownloadAndArchiveRecording(ctx context.Context, recordID int64) error
}

type callbackServiceImpl struct {
	db          *gorm.DB
	minioClient *minio.Client
	voiceClient *aliyun.VoiceClient
	sentimentAnalyzer *ai.SentimentAnalyzer
}

var (
	callbackServiceInstance *callbackServiceImpl
	callbackServiceOnce     sync.Once
)

func InitCallbackService() {
	callbackServiceOnce.Do(func() {
		callbackServiceInstance = &callbackServiceImpl{
			db:          database.GetDB(),
			minioClient: database.GetMinioClient(),
			voiceClient: aliyun.GetVoiceClient(),
			sentimentAnalyzer: ai.GetSentimentAnalyzer(),
		}
		logger.Info("Callback service initialized")
	})
}

func CallbackServiceInst() CallbackService {
	if callbackServiceInstance == nil {
		InitCallbackService()
	}
	return callbackServiceInstance
}

func (s *callbackServiceImpl) CreateCallbackRecord(ctx context.Context, caseID int64) (*model.CallbackRecord, error) {
	var caseData model.DisputeCase
	if err := s.db.Where("id = ?", caseID).First(&caseData).Error; err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	if caseData.Status != constants.CaseStatusClosed {
		return nil, fmt.Errorf("case is not closed")
	}

	closedAt := time.Now()
	if caseData.ClosedAt != nil {
		closedAt = *caseData.ClosedAt
	}

	scheduledTime := CalculateScheduledCallbackTime(closedAt)

	record := &model.CallbackRecord{
		ID:             utils.GenerateID(),
		CaseID:         caseData.ID,
		CaseNo:           caseData.CaseNo,
		CaseTitle:      caseData.Title,
		ApplicantID:     caseData.ApplicantID,
		ApplicantName:  caseData.ApplicantName,
		ApplicantPhone: caseData.ApplicantPhone,
		Status:          model.CallbackStatusPending,
		CallStatus:      model.CallStatusNotCalled,
		RetryCount:      0,
		MaxRetryCount:  3,
		ScheduledTime: &scheduledTime,
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("create callback record failed: %w", err)
	}

	logger.Info("Callback record created",
		zap.Int64("caseId", caseID),
		zap.String("caseNo", caseData.CaseNo),
		zap.Time("scheduledTime", scheduledTime),
	)

	return record, nil
}

func (s *callbackServiceImpl) ScheduleCallback(ctx context.Context, record *model.CallbackRecord) error {
	scheduledTime := CalculateNextValidSlot(time.Now())
	record.ScheduledTime = &scheduledTime

	return s.db.Model(record).Update("scheduled_time", scheduledTime).Error
}

func (s *callbackServiceImpl) InitiateCall(ctx context.Context, recordID int64) error {
	var record model.CallbackRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	if record.Status == model.CallbackStatusCalling {
		return fmt.Errorf("callback is already in progress")
	}

	if record.Status != model.CallbackStatusPending {
		return fmt.Errorf("callback is not in pending status")
	}

	now := time.Now()

	var template model.CallbackTemplate
	if err := s.db.Where("status = 1 AND is_default = 1").Order("id ASC").First(&template).Error; err != nil {
		logger.Warn("No default callback template found, using config defaults")
		template = model.CallbackTemplate{}
	}

	ttsParam := map[string]string{
		"caseNo":   record.CaseNo,
		"name":     record.ApplicantName,
	}

	callReq := &aliyun.SmartCallRequest{
		CalledNumber:   record.ApplicantPhone,
		OutId:         fmt.Sprintf("%d", record.ID),
		RecordFlag:    true,
		SessionTimeout: 120,
		TtsParam:      ttsParam,
	}

	if template.ID > 0 {
		callReq.TtsCode = template.Code
		if template.VoiceType != "" {
			callReq.VoiceType = template.VoiceType
		}
		if template.Speed != 0 {
			callReq.Speed = template.Speed
		}
		if template.Volume != 0 {
			callReq.Volume = template.Volume
		}
	}

	record.Status = model.CallbackStatusCalling
	record.CallStatus = model.CallStatusRinging
	record.CallTime = &now

	if err := s.db.Save(&record).Error; err != nil {
		return fmt.Errorf("update callback record failed: %w", err)
	}

	callResp, err := s.voiceClient.SmartCall(callReq)
	if err != nil {
		record.CallStatus = model.CallStatusFailed
		s.handleCallFailure(&record)
		return fmt.Errorf("initiate smart call failed: %w", err)
	}

	record.TaskID = callResp.TaskId
	record.CallID = callResp.CallId

	if err := s.db.Save(&record).Error; err != nil {
		return fmt.Errorf("update callback record with call info failed: %w", err)
	}

	logger.Info("Callback smart call initiated",
		zap.Int64("recordId", recordID),
		zap.String("callId", callResp.CallId),
		zap.String("taskId", callResp.TaskId),
	)

	return nil
}

func (s *callbackServiceImpl) ProcessCallResult(ctx context.Context, callID string) error {
	var record model.CallbackRecord
	if err := s.db.Where("call_id = ?", callID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	detail, err := s.voiceClient.QueryCallDetailByCallId(callID)
	if err != nil {
		return fmt.Errorf("query call detail failed: %w", err)
	}

	record.CallStatus = int32(aliyun.MapCallStatus(detail.Status))

	if detail.Duration > 0 {
		record.CallDuration = detail.Duration
	}

	if detail.Status == "200000" {
		record.Status = model.CallbackStatusSuccess
		record.TranscriptText = detail.AsrResult

		caseInfo := map[string]interface{}{
			"title": record.CaseTitle,
		}

		if detail.AsrResult != "" {
			sentimentResult, err := s.sentimentAnalyzer.AnalyzeCallback(detail.AsrResult, caseInfo)
			if err != nil {
				logger.Warn("Sentiment analysis failed",
					zap.String("callId", callID),
					logger.Error(err),
				)
			} else {
				sentimentJSON, _ := json.Marshal(sentimentResult)
				record.SentimentResult = string(sentimentJSON)
				record.SentimentScore = sentimentResult.SentimentScore
				record.Emotion = sentimentResult.Emotion
				record.SatisfactionScore = sentimentResult.Satisfaction
				record.PerformanceScore = sentimentResult.Performance

				if record.SatisfactionScore > 0 {
					s.db.Model(&model.DisputeCase{}).
						Where("id = ?", record.CaseID).
						Update("satisfaction_score", record.SatisfactionScore)
				}
			}
		}

		if detail.RecordUrl != "" {
			record.RecordingURL = detail.RecordUrl
			go s.DownloadAndArchiveRecording(ctx, record.ID)
		}
	} else {
		record.Remark = aliyun.MapCallStatusDesc(detail.Status)
		s.handleCallFailure(&record)
	}

	resultData := map[string]interface{}{
		"status":     detail.Status,
		"statusDesc": aliyun.MapCallStatusDesc(detail.Status),
		"duration":   detail.Duration,
		"startTime":  detail.StartTime,
		"endTime":    detail.EndTime,
	}
	resultJSON, _ := json.Marshal(resultData)
	record.ResultData = string(resultJSON)

	if err := s.db.Save(&record).Error; err != nil {
		return fmt.Errorf("update callback record failed: %w", err)
	}

	logger.Info("Callback call result processed",
		zap.String("callId", callID),
		zap.Int("status", int(record.Status)),
		zap.Int("callStatus", int(record.CallStatus)),
	)

	return nil
}

func (s *callbackServiceImpl) HandleCallback(ctx context.Context, data map[string]interface{}) error {
	callID, ok := data["call_id"].(string)
	if !ok {
		return fmt.Errorf("call_id not found in callback data")
	}

	status, _ := data["status"].(string)
	recordURL, _ := data["record_url"].(string)
	asrResult, _ := data["asr_result"].(string)

	logger.Info("Received Aliyun voice callback",
		zap.String("callId", callID),
		zap.String("status", status),
	)

	var record model.CallbackRecord
	if err := s.db.Where("call_id = ?", callID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	record.CallStatus = int32(aliyun.MapCallStatus(status))

	if status == "200000" {
		record.Status = model.CallbackStatusSuccess
		record.TranscriptText = asrResult
		record.RecordingURL = recordURL

		caseInfo := map[string]interface{}{
			"title": record.CaseTitle,
		}

		if asrResult != "" {
			sentimentResult, err := s.sentimentAnalyzer.AnalyzeCallback(asrResult, caseInfo)
			if err != nil {
				logger.Warn("Sentiment analysis failed",
					zap.String("callId", callID),
					logger.Error(err),
				)
			} else {
				sentimentJSON, _ := json.Marshal(sentimentResult)
				record.SentimentResult = string(sentimentJSON)
				record.SentimentScore = sentimentResult.SentimentScore
				record.Emotion = sentimentResult.Emotion
				record.SatisfactionScore = sentimentResult.Satisfaction
				record.PerformanceScore = sentimentResult.Performance

				if record.SatisfactionScore > 0 {
					s.db.Model(&model.DisputeCase{}).
						Where("id = ?", record.CaseID).
						Update("satisfaction_score", record.SatisfactionScore)
				}
			}
		}

		if recordURL != "" {
			go s.DownloadAndArchiveRecording(ctx, record.ID)
		}
	} else {
		record.Remark = aliyun.MapCallStatusDesc(status)
		s.handleCallFailure(&record)
	}

	return s.db.Save(&record).Error
}

func (s *callbackServiceImpl) GetCallbackList(ctx context.Context, caseID int64) ([]model.CallbackRecord, error) {
	var records []model.CallbackRecord
	err := s.db.Where("case_id = ?", caseID).Order("created_at DESC").Find(&records).Error
	return records, err
}

func (s *callbackServiceImpl) GetCallbackDetail(ctx context.Context, recordID int64) (*model.CallbackRecord, error) {
	var record model.CallbackRecord
	err := s.db.Where("id = ?", recordID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *callbackServiceImpl) RetryCallback(ctx context.Context, recordID int64) error {
	var record model.CallbackRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	if record.RetryCount >= record.MaxRetryCount {
		return fmt.Errorf("max retry count reached")
	}

	record.RetryCount++
	nextRetryTime := time.Now().Add(2 * time.Hour)
	record.NextRetryTime = &nextRetryTime
	record.Status = model.CallbackStatusPending
	if err := s.db.Save(&record).Error; err != nil {
		return fmt.Errorf("update retry info failed: %w", err)
	}

	return s.InitiateCall(ctx, recordID)
}

func (s *callbackServiceImpl) CancelCallback(ctx context.Context, recordID int64) error {
	var record model.CallbackRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	if record.CallID != "" {
		s.voiceClient.CancelCall(record.CallID)
	}

	record.Status = model.CallbackStatusCancelled
	record.Remark = "用户取消回访"

	return s.db.Save(&record).Error
}

func (s *callbackServiceImpl) ProcessPendingCallbacks(ctx context.Context) error {
	now := time.Now()

	var records []model.CallbackRecord
	err := s.db.Where(`status = ? AND scheduled_time <= ? AND (next_retry_time IS NULL OR next_retry_time <= ?)`,
		model.CallbackStatusPending, now, now,
	).Order("scheduled_time ASC").Limit(50).Find(&records).Error

	if err != nil {
		return fmt.Errorf("query pending callbacks failed: %w", err)
	}

	logger.Info("Found pending callbacks", zap.Int("count", len(records)))

	successCount := 0
	failedCount := 0

	for _, record := range records {
		if !IsValidCallbackTime(now) {
			s.ScheduleCallback(ctx, &record)
			continue
		}

		err := s.InitiateCall(ctx, record.ID)
		if err != nil {
			logger.Warn("Initiate callback failed",
				zap.Int64("recordId", record.ID),
				logger.Error(err),
			)
			failedCount++
		} else {
			successCount++
		}

		time.Sleep(100 * time.Millisecond)
	}

	logger.Info("Process pending callbacks completed",
		zap.Int("success", successCount),
		zap.Int("failed", failedCount),
	)

	return nil
}

func (s *callbackServiceImpl) ProcessRetryCallbacks(ctx context.Context) error {
	now := time.Now()

	var records []model.CallbackRecord
	err := s.db.Where(`status = ? AND retry_count < max_retry_count AND next_retry_time IS NOT NULL AND next_retry_time <= ?`,
		model.CallbackStatusPending, now,
	).Order("next_retry_time ASC").Limit(50).Find(&records).Error

	if err != nil {
		return fmt.Errorf("query retry callbacks failed: %w", err)
	}

	logger.Info("Found retry callbacks", zap.Int("count", len(records)))

	for _, record := range records {
		if !IsValidCallbackTime(now) {
			nextRetry := CalculateNextValidSlot(now)
			record.NextRetryTime = &nextRetry
			s.db.Save(&record)
			continue
		}

		err := s.InitiateCall(ctx, record.ID)
		if err != nil {
			logger.Warn("Retry callback failed",
				zap.Int64("recordId", record.ID),
				logger.Error(err),
			)
		}
	}

	return nil
}

func (s *callbackServiceImpl) CleanExpiredRecordings(ctx context.Context) error {
	now := time.Now()

	var records []model.CallbackRecord
	err := s.db.Where(`expire_at IS NOT NULL AND expire_at <= ? AND recording_url IS NOT NULL`,
		now,
	).Limit(100).Find(&records).Error

	if err != nil {
		return fmt.Errorf("query expired recordings failed: %w", err)
	}

	logger.Info("Found expired recordings", zap.Int("count", len(records)))

	for _, record := range records {
		if s.minioClient != nil && record.RecordingURL != "" {
			objectName := extractObjectNameFromURL(record.RecordingURL)
			if objectName != "" {
				cfg := config.GetConfig()
				err := s.minioClient.RemoveObject(ctx, cfg.MinIO.Bucket, objectName, minio.RemoveObjectOptions{})
				if err != nil {
					logger.Warn("Failed to remove expired recording",
						zap.String("objectName", objectName),
						logger.Error(err),
					)
				} else {
					logger.Info("Expired recording removed",
						zap.Int64("recordId", record.ID),
						zap.String("objectName", objectName),
					)
				}
			}
		}

		record.RecordingURL = ""
		record.RecordingSize = 0
		s.db.Save(&record)
	}

	return nil
}

func (s *callbackServiceImpl) DownloadAndArchiveRecording(ctx context.Context, recordID int64) error {
	var record model.CallbackRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		return fmt.Errorf("callback record not found: %w", err)
	}

	if record.RecordingURL == "" || s.minioClient == nil {
		return nil
	}

	resp, err := http.Get(record.RecordingURL)
	if err != nil {
		return fmt.Errorf("download recording failed: %w", err)
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "callback-recording-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("save recording to temp file failed: %w", err)
	}

	fileInfo, _ := tempFile.Stat()
	record.RecordingSize = fileInfo.Size()

	objectName := fmt.Sprintf("callback/%d/%d_%s.mp3",
		record.CaseID, record.ID, time.Now().Format("20060102150405"))

	cfg := config.GetConfig()
	_, err = s.minioClient.FPutObject(ctx, cfg.MinIO.Bucket, objectName, tempFile.Name(), minio.PutObjectOptions{
		ContentType: "audio/mpeg",
	})
	if err != nil {
		return fmt.Errorf("upload recording to minio failed: %w", err)
	}

	minioURL := fmt.Sprintf("%s/%s/%s", cfg.MinIO.Endpoint, cfg.MinIO.Bucket, objectName)
	record.RecordingURL = minioURL

	now := time.Now()
	expireAt := now.AddDate(1, 0, 0)
	record.ExpireAt = &expireAt

	logger.Info("Recording archived to MinIO",
		zap.Int64("recordId", recordID),
		zap.String("objectName", objectName),
		zap.Int64("size", fileInfo.Size()),
		zap.Time("expireAt", expireAt),
	)

	return s.db.Save(&record).Error
}

func CalculateScheduledCallbackTime(closedAt time.Time) time.Time {
	targetDate := closedAt.AddDate(0, 0, 7)
	target := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 9, 0, 0, 0, targetDate.Location())
	return adjustToValidSlot(target)
}

func CalculateNextValidSlot(from time.Time) time.Time {
	target := time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), 0, 0, 0, from.Location())
	return adjustToValidSlot(target)
}

func adjustToValidSlot(t time.Time) time.Time {
	for {
		if t.Weekday() == time.Saturday {
			t = t.AddDate(0, 0, 2)
			t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
			continue
		}
		if t.Weekday() == time.Sunday {
			t = t.AddDate(0, 0, 1)
			t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
			continue
		}
		hour := t.Hour()
		if hour < 9 {
			t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
			continue
		}
		if hour >= 12 && hour < 14 {
			t = time.Date(t.Year(), t.Month(), t.Day(), 14, 0, 0, 0, t.Location())
			continue
		}
		if hour >= 18 {
			t = t.AddDate(0, 0, 1)
			t = time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, t.Location())
			continue
		}
		break
	}
	return t
}

func (s *callbackServiceImpl) handleCallFailure(record *model.CallbackRecord) {
	if record.RetryCount < record.MaxRetryCount {
		record.RetryCount++
		nextRetryTime := time.Now().Add(2 * time.Hour)
		record.NextRetryTime = &nextRetryTime
		record.Status = model.CallbackStatusPending
	} else {
		record.Status = model.CallbackStatusFailed
	}

	if err := s.db.Save(record).Error; err != nil {
		logger.Warn("Failed to save callback failure state",
			zap.Int64("recordId", record.ID),
			logger.Error(err),
		)
	}

	logger.Info("Call failure handled",
		zap.Int64("recordId", record.ID),
		zap.Int("retryCount", record.RetryCount),
		zap.Int("maxRetryCount", record.MaxRetryCount),
		zap.Int32("status", record.Status),
	)
}

func IsValidCallbackTime(t time.Time) bool {
	hour := t.Hour()

	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}

	if hour < 9 || hour >= 18 {
		return false
	}

	if hour >= 12 && hour < 14 {
		return false
	}

	return true
}

func extractObjectNameFromURL(url string) string {
	if url == "" {
		return ""
	}

	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return ""
}

func (s *callbackServiceImpl) CreateCallbackForClosedCase(ctx context.Context, caseID int64) error {
	var count int64
	s.db.Model(&model.CallbackRecord{}).Where("case_id = ?", caseID).Count(&count)
	if count > 0 {
		return fmt.Errorf("callback record already exists for this case")
	}

	_, err := s.CreateCallbackRecord(ctx, caseID)
	return err
}
