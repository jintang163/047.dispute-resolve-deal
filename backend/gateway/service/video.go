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
}
