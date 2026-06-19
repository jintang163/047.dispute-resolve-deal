package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/trtc"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type VideoRoomCreateRequest struct {
	CaseID         int64    `json:"caseId" binding:"required"`
	Title          string   `json:"title" binding:"required"`
	ScheduledTime  string   `json:"scheduledTime" binding:"required"`
	ParticipantIDs []int64  `json:"participantIds" binding:"required"`
	Password       string   `json:"password"`
	Duration       int32    `json:"duration"`
	VirtualBG      bool     `json:"virtualBg"`
	Beauty         bool     `json:"beauty"`
}

type VideoRoomJoinRequest struct {
	RoomID   string `json:"roomId" binding:"required"`
	Password string `json:"password"`
}

type TRTCUserSigRequest struct {
	RoomID int64  `json:"roomId" binding:"required"`
	UserID string `json:"userId" binding:"required"`
}

type CloudRecordRequest struct {
	RoomID int64  `json:"roomId" binding:"required"`
	UserID string `json:"userId"`
}

type QueueEnqueueRequest struct {
	CaseID      int64  `json:"caseId" binding:"required"`
	CaseNo      string `json:"caseNo"`
	PartyName   string `json:"partyName" binding:"required"`
	PartyPhone  string `json:"partyPhone" binding:"required"`
	PartyUserID int64  `json:"partyUserId" binding:"required"`
	Priority    int    `json:"priority"`
}

type MeetingMinutesRequest struct {
	RoomID           int64  `json:"roomId" binding:"required"`
	CaseID           int64  `json:"caseId" binding:"required"`
	Transcript       string `json:"transcript" binding:"required"`
	DurationMinutes  int    `json:"durationMinutes"`
}

type ScreenShareRequest struct {
	RoomID int64 `json:"roomId" binding:"required"`
	UserID int64 `json:"userId" binding:"required"`
}

type VirtualBGRequest struct {
	RoomID  int64 `json:"roomId" binding:"required"`
	Enabled bool  `json:"enabled"`
}

type BeautyFilterRequest struct {
	RoomID  int64 `json:"roomId" binding:"required"`
	Enabled bool  `json:"enabled"`
}

func CreateVideoRoom(ctx context.Context, c *app.RequestContext) {
	var req VideoRoomCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo     string `gorm:"column:case_no"`
		Title      string `gorm:"column:title"`
		Status     int32  `gorm:"column:status"`
		MediatorID int64  `gorm:"column:mediator_id"`
	}

	result := database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).
		First(&caseData)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	if caseData.Status >= constants.CaseStatusClosed {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件已结案，无法创建视频调解"))
		return
	}

	if caseData.MediatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有案件调解员或领导可以创建视频调解"))
		return
	}

	scheduledTime, err := time.Parse("2006-01-02 15:04:05", req.ScheduledTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("时间格式错误，请使用 yyyy-MM-dd HH:mm:ss"))
		return
	}

	if scheduledTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, response.BadRequest("预约时间不能早于当前时间"))
		return
	}

	duration := req.Duration
	if duration == 0 {
		duration = 60
	}

	roomID := fmt.Sprintf("VR%s", utils.GenerateIDStr())
	roomPwd := req.Password
	if roomPwd == "" {
		roomPwd = utils.GenerateRandomString(6)
	}

	trtcRoomID := int32(time.Now().UnixNano() % 100000000)

	virtualBG := 0
	if req.VirtualBG {
		virtualBG = 1
	}
	beauty := 0
	if req.Beauty {
		beauty = 1
	}

	tx := database.GetDB().Begin()

	room := map[string]interface{}{
		"id":                utils.GenerateID(),
		"room_id":           roomID,
		"case_id":           req.CaseID,
		"case_no":           caseData.CaseNo,
		"title":             req.Title,
		"scheduled_time":    scheduledTime,
		"end_time":          scheduledTime.Add(time.Duration(duration) * time.Minute),
		"password":          roomPwd,
		"duration":          duration,
		"status":            constants.VideoStatusNotStarted,
		"creator_id":        userInfo.UserID,
		"creator_name":      userInfo.RealName,
		"organization_id":   userInfo.OrganizationID,
		"trtc_room_id":      trtcRoomID,
		"virtual_bg_enabled": virtualBG,
		"beauty_enabled":    beauty,
	}

	if err := tx.Table("video_room").Create(room).Error; err != nil {
		tx.Rollback()
		logger.Error("Create video room failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建视频房间失败"))
		return
	}

	var participants []map[string]interface{}
	var participantNames []string

	for _, pid := range req.ParticipantIDs {
		var user struct {
			RealName string `gorm:"column:real_name"`
			Phone    string `gorm:"column:phone"`
			Role     int32  `gorm:"column:role"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name, phone, role").
			Where("id = ?", pid).
			First(&user)

		participants = append(participants, map[string]interface{}{
			"room_id":     roomID,
			"user_id":     pid,
			"user_name":   user.RealName,
			"user_phone":  user.Phone,
			"user_role":   user.Role,
			"is_creator":  pid == userInfo.UserID,
			"join_status": 10,
		})

		participantNames = append(participantNames, user.RealName)

		go func(phone, name string) {
			msg := map[string]interface{}{
				"caseId":        req.CaseID,
				"caseNo":        caseData.CaseNo,
				"caseTitle":     caseData.Title,
				"roomId":        roomID,
				"roomTitle":     req.Title,
				"password":      roomPwd,
				"scheduledTime": req.ScheduledTime,
				"duration":      duration,
				"inviteBy":      userInfo.RealName,
				"participant":   name,
				"phone":         phone,
			}
			mq.SendMessage(constants.MQTopicNotification, msg)
		}(user.Phone, user.RealName)
	}

	if len(participants) > 0 {
		tx.Table("video_participant").Create(participants)
	}

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":         caseData.CaseNo,
		"operation_type":   "VIDEO_CREATE",
		"operation_detail": fmt.Sprintf("创建视频调解: %s，预约时间: %s，参与人: %s",
			req.Title, req.ScheduledTime, strings.Join(participantNames, "、")),
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"roomId":        roomID,
		"password":      roomPwd,
		"scheduledTime": req.ScheduledTime,
		"duration":      duration,
		"trtcRoomId":    trtcRoomID,
	}, "视频房间创建成功"))
}

func GetVideoRoomList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)

	db := database.GetDB().Table("video_room vr").
		Select("vr.*, dc.title as case_title, dc.case_no").
		Joins("LEFT JOIN dispute_case dc ON vr.case_id = dc.id").
		Where("vr.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Joins("INNER JOIN video_participant vp ON vr.room_id = vp.room_id").
			Where("vp.user_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("vr.organization_id = ?", userInfo.OrganizationID)
	}

	if status > 0 {
		db = db.Where("vr.status = ?", status)
	}
	if caseID > 0 {
		db = db.Where("vr.case_id = ?", caseID)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("vr.scheduled_time DESC").
		Limit(50).
		Find(&list)

	statusMap := map[int]string{
		constants.VideoStatusNotStarted: "未开始",
		constants.VideoStatusRunning:    "进行中",
		constants.VideoStatusEnded:      "已结束",
		constants.VideoStatusCancelled:  "已取消",
	}

	recordStatusMap := map[int]string{
		constants.VideoRecordStatusIdle:     "未录制",
		constants.VideoRecordStatusRecording: "录制中",
		constants.VideoRecordStatusStopped:  "已结束",
		constants.VideoRecordStatusFailed:   "失败",
	}

	for _, item := range list {
		if s, ok := item["status"].(int); ok {
			item["statusName"] = statusMap[s]
		}
		if rs, ok := item["record_status"].(int); ok {
			item["recordStatusName"] = recordStatusMap[rs]
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetVideoRoomDetail(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")
	userInfo := middleware.GetUserInfo(c)

	var room map[string]interface{}
	result := database.GetDB().Table("video_room vr").
		Select("vr.*, dc.title as case_title, dc.case_no").
		Joins("LEFT JOIN dispute_case dc ON vr.case_id = dc.id").
		Where("vr.room_id = ?", roomID).
		Find(&room)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("视频房间不存在"))
		return
	}

	if userInfo.Role == constants.RoleMediator {
		var count int64
		database.GetDB().Table("video_participant").
			Where("room_id = ? AND user_id = ?", roomID, userInfo.UserID).
			Count(&count)
		if count == 0 {
			c.JSON(http.StatusForbidden, response.Forbidden("您不是该视频会议的参与人"))
			return
		}
	}

	var participants []map[string]interface{}
	database.GetDB().Table("video_participant").
		Where("room_id = ?", roomID).
		Order("join_time DESC").
		Find(&participants)

	room["participants"] = participants

	roomIDInt, _ := strconv.ParseInt(roomID, 10, 64)
	segments, _ := trtc.GetCloudRecordService().GetRecordSegments(roomIDInt)
	room["recordSegments"] = segments

	queueSvc := trtc.GetVideoQueueService()
	queueList, _ := queueSvc.GetQueueList(ctx)
	room["queueInfo"] = queueList

	c.JSON(http.StatusOK, response.Success(room))
}

func JoinVideoRoom(ctx context.Context, c *app.RequestContext) {
	var req VideoRoomJoinRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var room struct {
		CaseID        int64     `gorm:"column:case_id"`
		CaseNo        string    `gorm:"column:case_no"`
		Password      string    `gorm:"column:password"`
		Status        int32     `gorm:"column:status"`
		ScheduledTime time.Time `gorm:"column:scheduled_time"`
		EndTime       time.Time `gorm:"column:end_time"`
		TRTCRoomID    int32     `gorm:"column:trtc_room_id"`
	}

	result := database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		First(&room)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("视频房间不存在"))
		return
	}

	var participant struct {
		ID         int64 `gorm:"column:id"`
		JoinStatus int32 `gorm:"column:join_status"`
	}

	database.GetDB().Table("video_participant").
		Select("id, join_status").
		Where("room_id = ? AND user_id = ?", req.RoomID, userInfo.UserID).
		First(&participant)

	if participant.ID == 0 {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该视频会议的参与人"))
		return
	}

	if room.Password != "" && room.Password != req.Password {
		c.JSON(http.StatusBadRequest, response.BadRequest("会议密码错误"))
		return
	}

	if room.Status == constants.VideoStatusCancelled {
		c.JSON(http.StatusBadRequest, response.BadRequest("该视频会议已取消"))
		return
	}

	if room.Status == constants.VideoStatusEnded {
		c.JSON(http.StatusBadRequest, response.BadRequest("该视频会议已结束"))
		return
	}

	now := time.Now()
	if now.Before(room.ScheduledTime.Add(-10 * time.Minute)) {
		c.JSON(http.StatusBadRequest, response.BadRequest("会议开始前10分钟才可加入"))
		return
	}

	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"join_status": 20,
		"join_time":   now,
	}
	tx.Table("video_participant").
		Where("id = ?", participant.ID).
		Updates(updates)

	if room.Status == constants.VideoStatusNotStarted {
		tx.Table("video_room").
			Where("room_id = ?", req.RoomID).
			Updates(map[string]interface{}{
				"status":            constants.VideoStatusRunning,
				"actual_start_time": now,
			})

		queueSvc := trtc.GetVideoQueueService()
		queueSvc.IncrementActiveRoomCount(ctx)
	}

	tx.Commit()

	trtcClient := trtc.GetTRTCClient()
	userSig := trtcClient.GenUserSig(strconv.FormatInt(userInfo.UserID, 10), 86400)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"roomId":      req.RoomID,
		"trtcRoomId":  room.TRTCRoomID,
		"userSig":     userSig,
		"sdkAppId":    trtcClient.GetSdkAppID(),
		"userId":      strconv.FormatInt(userInfo.UserID, 10),
		"caseId":      room.CaseID,
		"caseNo":      room.CaseNo,
		"scheduledTime": room.ScheduledTime.Format("2006-01-02 15:04:05"),
	}))
}

func EndVideoRoom(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")
	userInfo := middleware.GetUserInfo(c)

	var room struct {
		ID           int64  `gorm:"column:id"`
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		Status       int32  `gorm:"column:status"`
		CreatorID    int64  `gorm:"column:creator_id"`
		RecordTaskID string `gorm:"column:record_task_id"`
		RecordStatus int32  `gorm:"column:record_status"`
	}

	result := database.GetDB().Table("video_room").
		Where("room_id = ?", roomID).
		First(&room)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("视频房间不存在"))
		return
	}

	if room.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有会议创建者或领导可以结束会议"))
		return
	}

	if room.Status != constants.VideoStatusRunning {
		c.JSON(http.StatusBadRequest, response.BadRequest("会议未在进行中"))
		return
	}

	tx := database.GetDB().Begin()

	if room.RecordStatus == constants.VideoRecordStatusRecording && room.RecordTaskID != "" {
		roomIDInt, _ := strconv.ParseInt(roomID, 10, 64)
		recordSvc := trtc.GetCloudRecordService()
		if err := recordSvc.StopRecord(roomIDInt, room.RecordTaskID); err != nil {
			logger.Error("Stop cloud record on end room failed", logger.Error(err))
		}
	}

	now := time.Now()
	tx.Table("video_room").
		Where("id = ?", room.ID).
		Updates(map[string]interface{}{
			"status":           constants.VideoStatusEnded,
			"actual_end_time":  now,
		})

	tx.Table("video_participant").
		Where("room_id = ? AND join_status = 20", roomID).
		Updates(map[string]interface{}{
			"join_status": 30,
			"leave_time":  now,
		})

	history := map[string]interface{}{
		"case_id":          room.CaseID,
		"case_no":         room.CaseNo,
		"operation_type":   "VIDEO_END",
		"operation_detail": fmt.Sprintf("结束视频调解，房间号: %s", roomID),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	queueSvc := trtc.GetVideoQueueService()
	queueSvc.DecrementActiveRoomCount(ctx)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "会议已结束"))
}

func CancelVideoRoom(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")
	userInfo := middleware.GetUserInfo(c)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var room struct {
		ID        int64  `gorm:"column:id"`
		CaseID    int64  `gorm:"column:case_id"`
		CaseNo    string `gorm:"column:case_no"`
		Status    int32  `gorm:"column:status"`
		CreatorID int64  `gorm:"column:creator_id"`
		Title     string `gorm:"column:title"`
	}

	result := database.GetDB().Table("video_room").
		Where("room_id = ?", roomID).
		First(&room)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("视频房间不存在"))
		return
	}

	if room.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有会议创建者或领导可以取消会议"))
		return
	}

	if room.Status != constants.VideoStatusNotStarted {
		c.JSON(http.StatusBadRequest, response.BadRequest("只能取消未开始的会议"))
		return
	}

	tx := database.GetDB().Begin()

	tx.Table("video_room").
		Where("id = ?", room.ID).
		Updates(map[string]interface{}{
			"status":        constants.VideoStatusCancelled,
			"cancel_reason": req.Reason,
			"cancel_time":   time.Now(),
			"cancel_by":     userInfo.UserID,
		})

	var participants []map[string]interface{}
	database.GetDB().Table("video_participant").
		Select("user_id, user_name, user_phone").
		Where("room_id = ?", roomID).
		Find(&participants)

	for _, p := range participants {
		go func(phone, name string) {
			msg := map[string]interface{}{
				"caseId":      room.CaseID,
				"caseNo":      room.CaseNo,
				"roomId":      roomID,
				"roomTitle":   room.Title,
				"cancelReason": req.Reason,
				"cancelBy":    userInfo.RealName,
				"participant": name,
				"phone":       phone,
			}
			mq.SendMessage(constants.MQTopicNotification, msg)
		}(p["user_phone"].(string), p["user_name"].(string))
	}

	history := map[string]interface{}{
		"case_id":          room.CaseID,
		"case_no":         room.CaseNo,
		"operation_type":   "VIDEO_CANCEL",
		"operation_detail": fmt.Sprintf("取消视频调解: %s，原因: %s", room.Title, req.Reason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "会议已取消"))
}

func GetVideoRoomToken(ctx context.Context, c *app.RequestContext) {
	roomID := c.Query("roomId")
	userInfo := middleware.GetUserInfo(c)

	var participant struct {
		ID int64 `gorm:"column:id"`
	}

	database.GetDB().Table("video_participant").
		Where("room_id = ? AND user_id = ?", roomID, userInfo.UserID).
		First(&participant)

	if participant.ID == 0 {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该视频会议的参与人"))
		return
	}

	token := fmt.Sprintf("vt_%s_%d_%s", roomID, userInfo.UserID, utils.GenerateRandomString(16))

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"token": token,
	}))
}

func GetTRTCUserSig(ctx context.Context, c *app.RequestContext) {
	var req TRTCUserSigRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var room struct {
		TRTCRoomID int32 `gorm:"column:trtc_room_id"`
		Status     int32 `gorm:"column:status"`
	}
	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		First(&room)

	if room.Status != constants.VideoStatusRunning && room.Status != constants.VideoStatusNotStarted {
		c.JSON(http.StatusBadRequest, response.BadRequest("会议未在进行中"))
		return
	}

	trtcClient := trtc.GetTRTCClient()
	userSig := trtcClient.GenUserSig(req.UserID, 86400)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"sdkAppId":   trtcClient.GetSdkAppID(),
		"userSig":    userSig,
		"userId":     req.UserID,
		"trtcRoomId": room.TRTCRoomID,
	}))

	_ = userInfo
}

func StartVideoRecord(ctx context.Context, c *app.RequestContext) {
	var req CloudRecordRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var room struct {
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		RecordStatus int32  `gorm:"column:record_status"`
		CreatorID    int64  `gorm:"column:creator_id"`
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		First(&room)

	if room.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有会议创建者可以启动录制"))
		return
	}

	if room.RecordStatus == constants.VideoRecordStatusRecording {
		c.JSON(http.StatusBadRequest, response.BadRequest("已在录制中"))
		return
	}

	userID := req.UserID
	if userID == "" {
		userID = "administrator"
	}

	recordSvc := trtc.GetCloudRecordService()
	taskID, err := recordSvc.StartRecord(req.RoomID, userID, room.CaseID, room.CaseNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("启动录制失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"taskId": taskID,
	}, "录制已启动"))
}

func StopVideoRecord(ctx context.Context, c *app.RequestContext) {
	var req CloudRecordRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var room struct {
		RecordTaskID string `gorm:"column:record_task_id"`
		RecordStatus int32  `gorm:"column:record_status"`
		CreatorID    int64  `gorm:"column:creator_id"`
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		First(&room)

	if room.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有会议创建者可以停止录制"))
		return
	}

	if room.RecordStatus != constants.VideoRecordStatusRecording {
		c.JSON(http.StatusBadRequest, response.BadRequest("未在录制中"))
		return
	}

	if room.RecordTaskID == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("录制任务ID为空"))
		return
	}

	recordSvc := trtc.GetCloudRecordService()
	if err := recordSvc.StopRecord(req.RoomID, room.RecordTaskID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("停止录制失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "录制已停止"))
}

func GetVideoRecordSegments(ctx context.Context, c *app.RequestContext) {
	roomIDStr := c.Param("roomId")
	roomID, _ := strconv.ParseInt(roomIDStr, 10, 64)

	recordSvc := trtc.GetCloudRecordService()
	segments, err := recordSvc.GetRecordSegments(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("获取录制分段失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(segments))
}

func EnqueueVideoMediation(ctx context.Context, c *app.RequestContext) {
	var req QueueEnqueueRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		MediatorID   int64  `gorm:"column:mediator_id"`
		MediatorName string `gorm:"column:mediator_name"`
	}

	database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).
		First(&caseData)

	priority := req.Priority
	if priority == 0 {
		priority = 3
	}

	queueSvc := trtc.GetVideoQueueService()
	position, err := queueSvc.Enqueue(ctx, &trtc.QueueItem{
		CaseID:       req.CaseID,
		CaseNo:       req.CaseNo,
		CaseTitle:    "",
		MediatorID:   caseData.MediatorID,
		MediatorName: caseData.MediatorName,
		PartyName:    req.PartyName,
		PartyPhone:   req.PartyPhone,
		PartyUserID:  req.PartyUserID,
		Priority:     priority,
		EnqueueTime:  time.Now().Unix(),
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	_ = userInfo

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"position": position,
		"caseId":   req.CaseID,
	}, "已进入排队"))
}

func GetVideoQueuePosition(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	queueSvc := trtc.GetVideoQueueService()
	position, err := queueSvc.GetPosition(ctx, caseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("获取排队位置失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"position": position,
	}))
}

func GetVideoQueueList(ctx context.Context, c *app.RequestContext) {
	queueSvc := trtc.GetVideoQueueService()
	items, err := queueSvc.GetQueueList(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("获取排队列表失败"))
		return
	}

	result := make([]map[string]interface{}, 0, len(items))
	for i, item := range items {
		m := map[string]interface{}{
			"position":     i + 1,
			"caseId":       item.CaseID,
			"caseNo":       item.CaseNo,
			"mediatorName": item.MediatorName,
			"partyName":    item.PartyName,
			"priority":     item.Priority,
			"enqueueTime":  item.EnqueueTime,
		}
		result = append(result, m)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func LeaveVideoQueue(ctx context.Context, c *app.RequestContext) {
	var req struct {
		CaseID int64 `json:"caseId" binding:"required"`
		UserID int64 `json:"userId" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	queueSvc := trtc.GetVideoQueueService()
	if err := queueSvc.RemoveFromQueue(ctx, req.CaseID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("退出排队失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已退出排队"))
}

func GenerateVideoMinutes(ctx context.Context, c *app.RequestContext) {
	var req MeetingMinutesRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseInfo map[string]interface{}
	database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).
		First(&caseInfo)

	if caseInfo == nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	durationMinutes := req.DurationMinutes
	if durationMinutes == 0 {
		durationMinutes = 30
	}
	duration := time.Duration(durationMinutes) * time.Minute

	startTime := time.Now()
	result, err := ai.GenerateMeetingMinutes(caseInfo, req.Transcript, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("生成会议纪要失败: "+err.Error()))
		return
	}
	costTime := int(time.Since(startTime).Milliseconds())

	participants, _ := json.Marshal(result.Participants)
	keyPoints, _ := json.Marshal(result.KeyPoints)
	disputeFocus, _ := json.Marshal(result.DisputeFocus)
	evidenceDiscussed, _ := json.Marshal(result.EvidenceDiscussed)
	nextSteps, _ := json.Marshal(result.NextSteps)
	riskPoints, _ := json.Marshal(result.RiskPoints)

	minutes := map[string]interface{}{
		"id":                utils.GenerateID(),
		"room_id":           req.RoomID,
		"case_id":           req.CaseID,
		"case_no":           caseInfo["case_no"],
		"meeting_title":     result.MeetingTitle,
		"meeting_time":      time.Now(),
		"duration":          result.Duration,
		"participants":      string(participants),
		"summary":           result.Summary,
		"key_points":        string(keyPoints),
		"dispute_focus":     string(disputeFocus),
		"mediation_process": result.MediationProcess,
		"evidence_discussed": string(evidenceDiscussed),
		"agreement":         result.Agreement,
		"next_steps":        string(nextSteps),
		"risk_points":       string(riskPoints),
		"emotional_state":   result.EmotionalState,
		"mediator_advice":   result.MediatorAdvice,
		"transcript":        req.Transcript,
		"ai_model":          "deepseek",
		"cost_time":         costTime,
		"status":            constants.VideoMinutesStatusGenerated,
	}

	if err := database.GetDB().Table("video_meeting_minutes").Create(minutes).Error; err != nil {
		logger.Error("Save meeting minutes failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("保存会议纪要失败"))
		return
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		Update("has_meeting_minutes", 1)

	_ = userInfo

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"meetingTitle": result.MeetingTitle,
		"summary":      result.Summary,
		"keyPoints":    result.KeyPoints,
		"agreement":    result.Agreement,
		"costTime":     costTime,
	}, "会议纪要已生成"))
}

func GetVideoMeetingMinutes(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")

	var minutes map[string]interface{}
	database.GetDB().Table("video_meeting_minutes").
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		First(&minutes)

	if minutes == nil {
		c.JSON(http.StatusNotFound, response.NotFound("会议纪要不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(minutes))
}

func ApproveVideoMinutes(ctx context.Context, c *app.RequestContext) {
	minutesID := c.Param("minutesId")
	userInfo := middleware.GetUserInfo(c)

	database.GetDB().Table("video_meeting_minutes").
		Where("id = ?", minutesID).
		Updates(map[string]interface{}{
			"is_approved": 1,
			"approved_by": userInfo.UserID,
			"approved_at": time.Now(),
			"status":      constants.VideoMinutesStatusApproved,
		})

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "会议纪要已审核"))
}

func UpdateScreenShare(ctx context.Context, c *app.RequestContext) {
	var req ScreenShareRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		Update("screen_share_user_id", req.UserID)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "屏幕共享状态已更新"))
}

func UpdateVirtualBackground(ctx context.Context, c *app.RequestContext) {
	var req VirtualBGRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	enabled := 0
	if req.Enabled {
		enabled = 1
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		Update("virtual_bg_enabled", enabled)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "虚拟背景状态已更新"))
}

func UpdateBeautyFilter(ctx context.Context, c *app.RequestContext) {
	var req BeautyFilterRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	enabled := 0
	if req.Enabled {
		enabled = 1
	}

	database.GetDB().Table("video_room").
		Where("room_id = ?", req.RoomID).
		Update("beauty_enabled", enabled)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "美颜状态已更新"))
}

func RecordCallback(ctx context.Context, c *app.RequestContext) {
	var callbackData map[string]interface{}
	if err := c.BindAndValidate(&callbackData); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	recordSvc := trtc.GetCloudRecordService()
	if err := recordSvc.HandleRecordCallback(callbackData); err != nil {
		logger.Error("Handle record callback failed", logger.Error(err))
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code": 0,
	})
}

func GetVideoQueueStatus(ctx context.Context, c *app.RequestContext) {
	rdb := database.GetRedisClient()
	queueLen, _ := rdb.ZCard(ctx, "video:mediation:queue").Result()
	activeRooms, _ := rdb.Get(ctx, "video:active_rooms:count").Int()

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"queueLength":  queueLen,
		"activeRooms":  activeRooms,
		"estimatedWait": queueLen * 10,
	}))
}
