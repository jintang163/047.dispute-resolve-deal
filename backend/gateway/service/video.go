package service

import (
	"context"
	"time"
)

type VideoService interface {
	CreateVideoRoom(ctx context.Context, caseID int64, title string, startTime time.Time, duration int32, participantIDs []int64, creatorID int64) (map[string]interface{}, error)
	GetVideoRoomList(ctx context.Context, caseID int64, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error)
	GetVideoRoomDetail(ctx context.Context, roomID int64, userID int64) (map[string]interface{}, error)
	JoinVideoRoom(ctx context.Context, roomID int64, userID int64) (string, error)
	GetVideoRoomToken(ctx context.Context, roomID int64, userID int64) (string, error)
	EndVideoRoom(ctx context.Context, roomID int64, userID int64, recordURL string) error
	CancelVideoRoom(ctx context.Context, roomID int64, userID int64, reason string) error
	SendVideoVerifyCode(ctx context.Context, roomID int64, mobile string) error
	GenerateTRTCUserSig(ctx context.Context, userID string, roomID int64) (map[string]interface{}, error)
	StartCloudRecord(ctx context.Context, roomID int64, userID string, caseID int64, caseNo string) (string, error)
	StopCloudRecord(ctx context.Context, roomID int64, taskID string) error
	GetRecordSegments(ctx context.Context, roomID int64) ([]map[string]interface{}, error)
	EnqueueMediation(ctx context.Context, caseID int64, caseNo string, mediatorID int64, mediatorName string, partyName string, partyPhone string, partyUserID int64, priority int) (int, error)
	GetQueuePosition(ctx context.Context, caseID int64, userID int64) (int, error)
	GetQueueList(ctx context.Context) ([]map[string]interface{}, error)
	LeaveQueue(ctx context.Context, caseID int64, userID int64) error
	GenerateMeetingMinutes(ctx context.Context, roomID int64, caseID int64, transcript string, durationMinutes int) (map[string]interface{}, error)
	GetMeetingMinutes(ctx context.Context, roomID int64) (map[string]interface{}, error)
	ApproveMeetingMinutes(ctx context.Context, minutesID int64, userID int64) error
	UpdateScreenShareUser(ctx context.Context, roomID int64, userID int64) error
	UpdateVirtualBackground(ctx context.Context, roomID int64, enabled bool) error
	UpdateBeautyFilter(ctx context.Context, roomID int64, enabled bool) error
}
