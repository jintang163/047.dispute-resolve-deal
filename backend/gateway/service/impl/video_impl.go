package impl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type VideoServiceImpl struct{}

func NewVideoService() service.VideoService {
	return &VideoServiceImpl{}
}

const (
	videoRoomTokenTTL = 86400
	verifyCodeTTL     = 300
)

func (s *VideoServiceImpl) CreateVideoRoom(ctx context.Context, caseID int64, title string, startTime time.Time, duration int32, participantIDs []int64, creatorID int64) (map[string]interface{}, error) {
	var caseInfo map[string]interface{}
	database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseInfo)
	if caseInfo == nil {
		return nil, nil
	}

	roomNo := generateRoomNo()
	token := generateToken()

	room := &model.VideoRoom{
		CaseID:         caseID,
		RoomNo:         roomNo,
		Title:          title,
		StartTime:      startTime,
		Duration:       duration,
		Status:         constants.VideoStatusNotStarted,
		Token:          token,
		CreatorID:      creatorID,
		ParticipantIDs: participantIDs,
	}

	if err := database.GetDB().Create(room).Error; err != nil {
		return nil, err
	}

	s.sendVideoInvitations(room.ID, caseID, participantIDs, title, startTime)

	return map[string]interface{}{
		"roomId":    room.ID,
		"roomNo":    roomNo,
		"title":     title,
		"startTime": startTime,
		"duration":  duration,
		"status":    room.Status,
		"token":     token,
		"createdAt": room.CreatedAt,
	}, nil
}

func (s *VideoServiceImpl) GetVideoRoomList(ctx context.Context, caseID int64, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("video_room vr").
		Select("vr.*, dc.case_no, dc.title as case_title, u.real_name as creator_name").
		Joins("LEFT JOIN dispute_case dc ON vr.case_id = dc.id").
		Joins("LEFT JOIN user u ON vr.creator_id = u.id").
		Where("vr.deleted_at IS NULL")

	if caseID > 0 {
		db = db.Where("vr.case_id = ?", caseID)
	}

	if userID > 0 && role != constants.RoleAdmin {
		db = db.Where("(vr.creator_id = ? OR FIND_IN_SET(?, vr.participant_ids))", userID, userID)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := (page - 1) * pageSize
	db.Order("vr.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int); ok {
			statusMap := map[int]string{
				10: "未开始", 20: "进行中", 30: "已结束", 40: "已取消",
			}
			item["statusText"] = statusMap[status]
		}
	}

	return list, total, nil
}

func (s *VideoServiceImpl) GetVideoRoomDetail(ctx context.Context, roomID int64, userID int64) (map[string]interface{}, error) {
	var room map[string]interface{}
	result := database.GetDB().Table("video_room vr").
		Select("vr.*, dc.case_no, dc.title as case_title, u.real_name as creator_name").
		Joins("LEFT JOIN dispute_case dc ON vr.case_id = dc.id").
		Joins("LEFT JOIN user u ON vr.creator_id = u.id").
		Where("vr.id = ? AND vr.deleted_at IS NULL", roomID).
		First(&room)

	if result.Error != nil {
		return nil, result.Error
	}

	if status, ok := room["status"].(int); ok {
		statusMap := map[int]string{
			10: "未开始", 20: "进行中", 30: "已结束", 40: "已取消",
		}
		room["statusText"] = statusMap[status]
	}

	participantIDs, _ := room["participant_ids"].([]byte)
	if participantIDs != nil {
		var participants []map[string]interface{}
		idStr := string(participantIDs)
		database.GetDB().Table("user").
			Select("id, real_name, avatar, role").
			Where("id IN (?)", idStr).
			Find(&participants)
		room["participants"] = participants
	}

	canJoin := s.checkUserPermission(room, userID)
	room["canJoin"] = canJoin

	return room, nil
}

func (s *VideoServiceImpl) JoinVideoRoom(ctx context.Context, roomID int64, userID int64) (string, error) {
	var room model.VideoRoom
	database.GetDB().Where("id = ? AND deleted_at IS NULL", roomID).First(&room)
	if room.ID == 0 {
		return "", nil
	}

	if room.Status == constants.VideoStatusEnded || room.Status == constants.VideoStatusCancelled {
		return "", nil
	}

	if !s.checkUserPermissionFromModel(&room, userID) {
		return "", nil
	}

	if room.Status == constants.VideoStatusNotStarted {
		database.GetDB().Model(&room).Update("status", constants.VideoStatusRunning)
	}

	token := generateToken()

	key := fmt.Sprintf("video:room:%d:token:%d", roomID, userID)
	database.GetRedisClient().Set(ctx, key, token, videoRoomTokenTTL*time.Second)

	database.GetDB().Table("video_participant").
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Update("join_time", time.Now())

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":      "video_join",
		"roomId":    roomID,
		"userId":    userID,
		"caseId":    room.CaseID,
		"timestamp": time.Now(),
	})

	return token, nil
}

func (s *VideoServiceImpl) GetVideoRoomToken(ctx context.Context, roomID int64, userID int64) (string, error) {
	key := fmt.Sprintf("video:room:%d:token:%d", roomID, userID)
	token, err := database.GetRedisClient().Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *VideoServiceImpl) EndVideoRoom(ctx context.Context, roomID int64, userID int64, recordURL string) error {
	var room model.VideoRoom
	database.GetDB().Where("id = ? AND deleted_at IS NULL", roomID).First(&room)
	if room.ID == 0 {
		return nil
	}

	if room.CreatorID != userID {
		return nil
	}

	updates := map[string]interface{}{
		"status":      constants.VideoStatusEnded,
		"end_time":    time.Now(),
		"record_url":  recordURL,
	}

	err := database.GetDB().Model(&room).Updates(updates).Error
	if err != nil {
		return err
	}

	database.GetDB().Table("video_participant").
		Where("room_id = ?", roomID).
		Update("leave_time", time.Now())

	return nil
}

func (s *VideoServiceImpl) CancelVideoRoom(ctx context.Context, roomID int64, userID int64, reason string) error {
	var room model.VideoRoom
	database.GetDB().Where("id = ? AND deleted_at IS NULL", roomID).First(&room)
	if room.ID == 0 {
		return nil
	}

	if room.CreatorID != userID {
		return nil
	}

	updates := map[string]interface{}{
		"status":       constants.VideoStatusCancelled,
		"cancel_reason": reason,
		"cancel_time":  time.Now(),
	}

	err := database.GetDB().Model(&room).Updates(updates).Error
	if err != nil {
		return err
	}

	participantIDs := room.ParticipantIDs
	if len(participantIDs) > 0 {
		for _, pid := range participantIDs {
			mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
				"type":      "video_cancel",
				"roomId":    roomID,
				"userId":    pid,
				"caseId":    room.CaseID,
				"title":     room.Title,
				"reason":    reason,
				"timestamp": time.Now(),
			})
		}
	}

	return nil
}

func (s *VideoServiceImpl) SendVideoVerifyCode(ctx context.Context, roomID int64, mobile string) error {
	code := generateVerifyCode()

	key := fmt.Sprintf("video:verify:%s:%d", mobile, roomID)
	database.GetRedisClient().Set(ctx, key, code, verifyCodeTTL*time.Second)

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":      "sms_verify",
		"mobile":    mobile,
		"code":      code,
		"roomId":    roomID,
		"expire":    verifyCodeTTL,
		"timestamp": time.Now(),
	})

	logger.Info("Video verify code sent", logger.String("mobile", mobile), logger.String("code", code))

	return nil
}

func (s *VideoServiceImpl) sendVideoInvitations(roomID, caseID int64, participantIDs []int64, title string, startTime time.Time) {
	for _, pid := range participantIDs {
		mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
			"type":      "video_invite",
			"roomId":    roomID,
			"userId":    pid,
			"caseId":    caseID,
			"title":     title,
			"startTime": startTime,
			"timestamp": time.Now(),
		})
	}
}

func (s *VideoServiceImpl) checkUserPermission(room map[string]interface{}, userID int64) bool {
	creatorID, _ := room["creator_id"].(int64)
	if creatorID == userID {
		return true
	}

	participantIDs, _ := room["participant_ids"].([]byte)
	if participantIDs != nil {
		idStr := string(participantIDs)
		return utils.Int64InSlice(userID, utils.ParseInt64Slice(idStr))
	}

	return false
}

func (s *VideoServiceImpl) checkUserPermissionFromModel(room *model.VideoRoom, userID int64) bool {
	if room.CreatorID == userID {
		return true
	}

	for _, pid := range room.ParticipantIDs {
		if pid == userID {
			return true
		}
	}

	return false
}

func generateRoomNo() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("VR%s", hex.EncodeToString(b))
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateVerifyCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d", uint(b[0])<<16|uint(b[1])<<8|uint(b[2])%1000000)
}
