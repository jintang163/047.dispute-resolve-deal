package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type VideoRoomCreateRequest struct {
	CaseID        int64    `json:"caseId" binding:"required"`
	Title         string   `json:"title" binding:"required"`
	ScheduledTime string   `json:"scheduledTime" binding:"required"`
	ParticipantIDs []int64 `json:"participantIds" binding:"required"`
	Password      string   `json:"password"`
	Duration      int32    `json:"duration"`
}

type VideoRoomJoinRequest struct {
	RoomID string `json:"roomId" binding:"required"`
	Password string `json:"password"`
}

func CreateVideoRoom(ctx context.Context, c *app.RequestContext) {
	var req VideoRoomCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo  string `gorm:"column:case_no"`
		Title   string `gorm:"column:title"`
		Status  int32  `gorm:"column:status"`
		MediatorID int64 `gorm:"column:mediator_id"`
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

	tx := database.GetDB().Begin()

	room := map[string]interface{}{
		"id":              utils.GenerateID(),
		"room_id":         roomID,
		"case_id":         req.CaseID,
		"case_no":         caseData.CaseNo,
		"title":           req.Title,
		"scheduled_time":  scheduledTime,
		"end_time":        scheduledTime.Add(time.Duration(duration) * time.Minute),
		"password":        roomPwd,
		"duration":        duration,
		"status":          constants.VideoStatusNotStarted,
		"creator_id":      userInfo.UserID,
		"creator_name":    userInfo.RealName,
		"organization_id": userInfo.OrganizationID,
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
		"case_no":          caseData.CaseNo,
		"operation_type":   "VIDEO_CREATE",
		"operation_detail": fmt.Sprintf("创建视频调解: %s，预约时间: %s，参与人: %s", 
			req.Title, req.ScheduledTime, strings.Join(participantNames, "、")),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"roomId":        roomID,
		"password":      roomPwd,
		"scheduledTime": req.ScheduledTime,
		"duration":      duration,
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

	for _, item := range list {
		if s, ok := item["status"].(int); ok {
			item["status_name"] = statusMap[s]
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
		CaseID        int64  `gorm:"column:case_id"`
		CaseNo        string `gorm:"column:case_no"`
		Password      string `gorm:"column:password"`
		Status        int32  `gorm:"column:status"`
		ScheduledTime time.Time `gorm:"column:scheduled_time"`
		EndTime       time.Time `gorm:"column:end_time"`
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
				"status":      constants.VideoStatusRunning,
				"actual_start_time": now,
			})
	}

	tx.Commit()

	token := fmt.Sprintf("vt_%s_%d_%s", req.RoomID, userInfo.UserID, utils.GenerateRandomString(16))

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"roomId":      req.RoomID,
		"token":       token,
		"caseId":      room.CaseID,
		"caseNo":      room.CaseNo,
		"scheduledTime": room.ScheduledTime.Format("2006-01-02 15:04:05"),
	}))
}

func EndVideoRoom(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")
	userInfo := middleware.GetUserInfo(c)

	var room struct {
		ID         int64 `gorm:"column:id"`
		CaseID     int64 `gorm:"column:case_id"`
		CaseNo     string `gorm:"column:case_no"`
		Status     int32 `gorm:"column:status"`
		CreatorID  int64 `gorm:"column:creator_id"`
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

	now := time.Now()
	tx.Table("video_room").
		Where("id = ?", room.ID).
		Updates(map[string]interface{}{
			"status":         constants.VideoStatusEnded,
			"actual_end_time": now,
		})

	tx.Table("video_participant").
		Where("room_id = ? AND join_status = 20", roomID).
		Updates(map[string]interface{}{
			"join_status": 30,
			"leave_time":  now,
		})

	history := map[string]interface{}{
		"case_id":          room.CaseID,
		"case_no":          room.CaseNo,
		"operation_type":   "VIDEO_END",
		"operation_detail": fmt.Sprintf("结束视频调解，房间号: %s", roomID),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

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
		ID         int64  `gorm:"column:id"`
		CaseID     int64  `gorm:"column:case_id"`
		CaseNo     string `gorm:"column:case_no"`
		Status     int32  `gorm:"column:status"`
		CreatorID  int64  `gorm:"column:creator_id"`
		Title      string `gorm:"column:title"`
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
		"case_no":          room.CaseNo,
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
