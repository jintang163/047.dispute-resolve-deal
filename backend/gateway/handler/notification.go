package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	common "github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type NotificationListRequest struct {
	common.BaseQuery
	Type    int32  `form:"type"`
	Status  int32  `form:"status"`
	IsRead  *bool  `form:"isRead"`
}

type NotificationSendRequest struct {
	ReceiverIDs []int64 `json:"receiverIds" binding:"required"`
	TemplateID  int64   `json:"templateId" binding:"required"`
	Params      map[string]interface{} `json:"params"`
	NotifyType  string  `json:"notifyType"`
}

func GetMyNotifications(ctx context.Context, c *app.RequestContext) {
	var req NotificationListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.receiver_id = ?", userInfo.UserID).
		Where("nr.deleted_at IS NULL")

	if req.Type > 0 {
		db = db.Where("nt.template_type = ?", req.Type)
	}
	if req.Status > 0 {
		db = db.Where("nr.status = ?", req.Status)
	}
	if req.IsRead != nil {
		if *req.IsRead {
			db = db.Where("nr.read_time IS NOT NULL")
		} else {
			db = db.Where("nr.read_time IS NULL")
		}
	}
	if req.Keyword != "" {
		db = db.Where("nr.title LIKE ? OR nr.content LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("nr.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	var unreadCount int64
	database.GetDB().Table("notification_record").
		Where("receiver_id = ?", userInfo.UserID).
		Where("read_time IS NULL").
		Where("deleted_at IS NULL").
		Count(&unreadCount)

	c.JSON(http.StatusOK, response.PageWithExtra(list, total, req.Page, req.PageSize, map[string]interface{}{
		"unreadCount": unreadCount,
	}))
}

func GetNotificationDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var notification map[string]interface{}
	result := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type, nt.content as template_content").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.id = ?", id).
		Find(&notification)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("通知不存在"))
		return
	}

	if notification["receiver_id"].(int64) != userInfo.UserID {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限查看此通知"))
		return
	}

	if notification["read_time"] == nil {
		database.GetDB().Table("notification_record").
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"read_time": time.Now(),
				"status":    2,
			})
		notification["read_time"] = time.Now()
	}

	c.JSON(http.StatusOK, response.Success(notification))
}

func MarkAllAsRead(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	database.GetDB().Table("notification_record").
		Where("receiver_id = ?", userInfo.UserID).
		Where("read_time IS NULL").
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"read_time": time.Now(),
			"status":    2,
		})

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已全部标记为已读"))
}

func MarkAsRead(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	result := database.GetDB().Table("notification_record").
		Where("id = ? AND receiver_id = ?", id, userInfo.UserID).
		Updates(map[string]interface{}{
			"read_time": time.Now(),
			"status":    2,
		})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("通知不存在或无权限"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "标记成功"))
}

func SendNotification(ctx context.Context, c *app.RequestContext) {
	var req NotificationSendRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限发送通知"))
		return
	}

	var template struct {
		ID            int64  `gorm:"column:id"`
		TemplateName  string `gorm:"column:template_name"`
		TemplateType  int32  `gorm:"column:template_type"`
		TitleTemplate string `gorm:"column:title_template"`
		Content       string `gorm:"column:content"`
		Channel       string `gorm:"column:channel"`
		Status        int32  `gorm:"column:status"`
	}

	result := database.GetDB().Table("notification_template").
		Where("id = ?", req.TemplateID).
		First(&template)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("通知模板不存在"))
		return
	}

	if template.Status != 1 {
		c.JSON(http.StatusBadRequest, response.BadRequest("通知模板已禁用"))
		return
	}

	notifyType := req.NotifyType
	if notifyType == "" {
		notifyType = template.Channel
	}

	title := template.TitleTemplate
	content := template.Content

	if req.Params != nil {
		for k, v := range req.Params {
			placeholder := fmt.Sprintf("{%s}", k)
			strVal := fmt.Sprintf("%v", v)
			title = replaceAll(title, placeholder, strVal)
			content = replaceAll(content, placeholder, strVal)
		}
	}

	now := time.Now()
	var records []map[string]interface{}

	for _, rid := range req.ReceiverIDs {
		var receiver struct {
			RealName string `gorm:"column:real_name"`
			Phone    string `gorm:"column:phone"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name, phone").
			Where("id = ?", rid).
			First(&receiver)

		records = append(records, map[string]interface{}{
			"receiver_id":   rid,
			"receiver_name": receiver.RealName,
			"receiver_phone": receiver.Phone,
			"template_id":   template.ID,
			"template_name": template.TemplateName,
			"template_type": template.TemplateType,
			"title":         title,
			"content":       content,
			"channel":       notifyType,
			"status":        1,
			"send_time":     now,
			"sender_id":     userInfo.UserID,
			"sender_name":   userInfo.RealName,
			"params":        toJSON(req.Params),
		})

		go func(phone, name string) {
			msg := map[string]interface{}{
				"receiverId":   rid,
				"receiverName": name,
				"phone":        phone,
				"title":        title,
				"content":      content,
				"notifyType":   notifyType,
				"templateId":   template.ID,
				"sentBy":       userInfo.RealName,
			}
			mq.SendMessage(constants.MQTopicNotification, msg)
		}(receiver.Phone, receiver.RealName)
	}

	tx := database.GetDB().Begin()
	if len(records) > 0 {
		tx.Table("notification_record").Create(records)
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"count": len(records),
	}, "通知发送成功"))
}

func GetNotificationTemplates(ctx context.Context, c *app.RequestContext) {
	templateType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))

	db := database.GetDB().Table("notification_template").
		Where("status = 1").
		Where("deleted_at IS NULL")

	if templateType > 0 {
		db = db.Where("template_type = ?", templateType)
	}

	var templates []map[string]interface{}
	db.Order("sort_order ASC, id ASC").Find(&templates)

	c.JSON(http.StatusOK, response.Success(templates))
}

func GetUnreadCount(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var unreadCount int64
	database.GetDB().Table("notification_record").
		Where("receiver_id = ?", userInfo.UserID).
		Where("read_time IS NULL").
		Where("deleted_at IS NULL").
		Count(&unreadCount)

	var typeCounts []map[string]interface{}
	database.GetDB().Table("notification_record nr").
		Select("nr.template_type, COUNT(*) as count").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.receiver_id = ?", userInfo.UserID).
		Where("nr.read_time IS NULL").
		Where("nr.deleted_at IS NULL").
		Group("nr.template_type").
		Find(&typeCounts)

	result := map[string]interface{}{
		"total":      unreadCount,
		"typeCounts": typeCounts,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func DeleteNotification(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	result := database.GetDB().Table("notification_record").
		Where("id = ? AND receiver_id = ?", id, userInfo.UserID).
		Update("deleted_at", time.Now())

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("通知不存在或无权限"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除成功"))
}

func BatchDeleteNotifications(ctx context.Context, c *app.RequestContext) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	database.GetDB().Table("notification_record").
		Where("id IN ? AND receiver_id = ?", req.IDs, userInfo.UserID).
		Update("deleted_at", time.Now())

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "批量删除成功"))
}

func replaceAll(s, old, new string) string {
	result := s
	for {
		idx := indexOf(result, old)
		if idx == -1 {
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
