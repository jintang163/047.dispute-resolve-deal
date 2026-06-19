package trtc

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"go.uber.org/zap"
)

type CloudRecordService struct {
	client *TRTCClient
	cfg    *config.TRTCConfig
}

var cloudRecordService *CloudRecordService

func InitCloudRecordService() {
	cfg := config.GetConfig()
	cloudRecordService = &CloudRecordService{
		client: GetTRTCClient(),
		cfg:    &cfg.TRTC,
	}
	logger.Info("Cloud record service initialized")
}

func GetCloudRecordService() *CloudRecordService {
	if cloudRecordService == nil {
		InitCloudRecordService()
	}
	return cloudRecordService
}

func (s *CloudRecordService) StartRecord(roomID int64, userID string, caseID int64, caseNo string) (string, error) {
	var activeCount int64
	database.GetDB().Table("video_record_segment").
		Where("room_id = ? AND status = ?", roomID, 1).
		Count(&activeCount)

	if activeCount > 0 {
		return "", fmt.Errorf("该房间已有进行中的录制任务")
	}

	minioCfg := config.GetConfig().MinIO

	segmentSec := s.cfg.RecordSegmentSec
	if segmentSec == 0 {
		segmentSec = 600
	}

	req := &CreateRecordRequest{
		RoomID:           roomID,
		UserID:           userID,
		RecordMode:       2,
		MaxDuration:      14400,
		StorageRegion:    "ap-guangzhou",
		StorageBucket:    minioCfg.Bucket,
		StorageAccessKey: minioCfg.AccessKey,
		StorageSecretKey: minioCfg.SecretKey,
		StoragePath:      s.cfg.RecordStoragePath,
	}

	resp, err := s.client.CreateCloudRecord(req)
	if err != nil {
		logger.Error("Start cloud record failed",
			zap.Int64("roomId", roomID),
			logger.Error(err),
		)
		return "", err
	}

	segment := map[string]interface{}{
		"id":             utils.GenerateID(),
		"room_id":        roomID,
		"case_id":        caseID,
		"case_no":        caseNo,
		"task_id":        resp.TaskId,
		"segment_index":  0,
		"segment_sec":    segmentSec,
		"status":         1,
		"start_time":     time.Now(),
		"storage_path":   fmt.Sprintf("%s/%d/segment_0", s.cfg.RecordStoragePath, roomID),
	}

	if err := database.GetDB().Table("video_record_segment").Create(segment).Error; err != nil {
		logger.Error("Save record segment failed", logger.Error(err))
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", roomID).
		Updates(map[string]interface{}{
			"record_task_id": resp.TaskId,
			"record_status":  1,
		})

	return resp.TaskId, nil
}

func (s *CloudRecordService) StopRecord(roomID int64, taskID string) error {
	resp, err := s.client.StopCloudRecord(taskID, roomID)
	if err != nil {
		logger.Error("Stop cloud record failed",
			zap.Int64("roomId", roomID),
			zap.String("taskId", taskID),
			logger.Error(err),
		)
		return err
	}

	database.GetDB().Table("video_record_segment").
		Where("room_id = ? AND task_id = ? AND status = ?", roomID, taskID, 1).
		Updates(map[string]interface{}{
			"status":    2,
			"end_time":  time.Now(),
		})

	database.GetDB().Table("video_room").
		Where("room_id = ?", roomID).
		Updates(map[string]interface{}{
			"record_status": 2,
		})

	logger.Info("Cloud record stopped",
		zap.Int64("roomId", roomID),
		zap.String("taskId", resp.TaskId),
	)

	return nil
}

func (s *CloudRecordService) GetRecordStatus(roomID int64, taskID string) (*DescribeRecordResponse, error) {
	return s.client.DescribeCloudRecord(taskID, roomID)
}

func (s *CloudRecordService) HandleRecordCallback(callbackData map[string]interface{}) error {
	eventType, _ := callbackData["EventType"].(float64)
	roomID, _ := callbackData["RoomId"].(float64)
	taskID, _ := callbackData["TaskId"].(string)

	switch int(eventType) {
	case 1:
		logger.Info("Cloud recording started",
			zap.Int64("roomId", int64(roomID)),
			zap.String("taskId", taskID),
		)

	case 2:
		logger.Info("Cloud recording stopped",
			zap.Int64("roomId", int64(roomID)),
			zap.String("taskId", taskID),
		)
		s.handleRecordStopped(int64(roomID), taskID, callbackData)

	case 3:
		logger.Info("Cloud recording file generated",
			zap.Int64("roomId", int64(roomID)),
			zap.String("taskId", taskID),
		)
		s.handleRecordFileGenerated(int64(roomID), taskID, callbackData)
	}

	return nil
}

func (s *CloudRecordService) handleRecordStopped(roomID int64, taskID string, data map[string]interface{}) {
	database.GetDB().Table("video_record_segment").
		Where("room_id = ? AND task_id = ? AND status = ?", roomID, taskID, 1).
		Updates(map[string]interface{}{
			"status":   2,
			"end_time": time.Now(),
		})

	database.GetDB().Table("video_room").
		Where("room_id = ?", roomID).
		Updates(map[string]interface{}{
			"record_status": 2,
		})
}

func (s *CloudRecordService) handleRecordFileGenerated(roomID int64, taskID string, data map[string]interface{}) {
	fileList, _ := data["FileList"].([]interface{})
	for _, f := range fileList {
		fileInfo, _ := f.(map[string]interface{})
		fileURL, _ := fileInfo["FileUrl"].(string)
		startTimeMs, _ := fileInfo["StartTimeMs"].(float64)
		endTimeMs, _ := fileInfo["EndTimeMs"].(float64)

		var segment map[string]interface{}
		database.GetDB().Table("video_record_segment").
			Where("room_id = ? AND task_id = ? AND status = 2 AND file_url = ''", roomID, taskID).
			First(&segment)

		if segment != nil {
			segID, _ := segment["id"]
			database.GetDB().Table("video_record_segment").
				Where("id = ?", segID).
				Updates(map[string]interface{}{
					"file_url":        fileURL,
					"start_time_ms":   int64(startTimeMs),
					"end_time_ms":     int64(endTimeMs),
					"duration_sec":    (int64(endTimeMs) - int64(startTimeMs)) / 1000,
				})
		} else {
			var segCount int64
			database.GetDB().Table("video_record_segment").
				Where("room_id = ? AND task_id = ?", roomID, taskID).
				Count(&segCount)

			segment := map[string]interface{}{
				"id":             utils.GenerateID(),
				"room_id":        roomID,
				"task_id":        taskID,
				"segment_index":  segCount,
				"segment_sec":    s.cfg.RecordSegmentSec,
				"status":         2,
				"file_url":       fileURL,
				"start_time_ms":  int64(startTimeMs),
				"end_time_ms":    int64(endTimeMs),
				"duration_sec":   (int64(endTimeMs) - int64(startTimeMs)) / 1000,
				"storage_path":   fmt.Sprintf("%s/%d/segment_%d", s.cfg.RecordStoragePath, roomID, segCount),
			}
			database.GetDB().Table("video_record_segment").Create(segment)
		}
	}
}

func (s *CloudRecordService) GetRecordSegments(roomID int64) ([]map[string]interface{}, error) {
	var segments []map[string]interface{}
	result := database.GetDB().Table("video_record_segment").
		Where("room_id = ?", roomID).
		Order("segment_index ASC").
		Find(&segments)

	if result.Error != nil {
		return nil, result.Error
	}
	return segments, nil
}

func (s *CloudRecordService) TriggerAutoSegment(ctx context.Context) error {
	var activeRecords []map[string]interface{}
	database.GetDB().Table("video_record_segment").
		Where("status = ? AND start_time < ?", 1, time.Now().Add(-time.Duration(s.cfg.RecordSegmentSec)*time.Second)).
		Find(&activeRecords)

	for _, record := range activeRecords {
		roomID, _ := record["room_id"].(int64)
		taskID, _ := record["task_id"].(string)
		caseID, _ := record["case_id"].(int64)
		caseNo, _ := record["case_no"].(string)
		segIndex, _ := record["segment_index"].(int64)

		if err := s.StopRecord(roomID, taskID); err != nil {
			logger.Error("Auto segment stop record failed",
				zap.Int64("roomId", roomID),
				logger.Error(err),
			)
			continue
		}

		newTaskID, err := s.StartRecord(roomID, s.cfg.AdminUserID, caseID, caseNo)
		if err != nil {
			logger.Error("Auto segment start new record failed",
				zap.Int64("roomId", roomID),
				logger.Error(err),
			)
			mq.SendAsync("dispute_video_record_error", map[string]interface{}{
				"roomId":   roomID,
				"caseId":   caseID,
				"error":    err.Error(),
				"action":   "auto_segment_restart",
				"timestamp": time.Now(),
			})
			continue
		}

		newSegment := map[string]interface{}{
			"id":            utils.GenerateID(),
			"room_id":       roomID,
			"case_id":       caseID,
			"case_no":       caseNo,
			"task_id":       newTaskID,
			"segment_index": segIndex + 1,
			"segment_sec":   s.cfg.RecordSegmentSec,
			"status":        1,
			"start_time":    time.Now(),
			"storage_path":  fmt.Sprintf("%s/%d/segment_%d", s.cfg.RecordStoragePath, roomID, segIndex+1),
		}
		database.GetDB().Table("video_record_segment").Create(newSegment)

		logger.Info("Auto segment completed",
			zap.Int64("roomId", roomID),
			zap.Int64("segmentIndex", segIndex+1),
			zap.String("newTaskId", newTaskID),
		)
	}

	return nil
}
