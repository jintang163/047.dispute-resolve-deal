package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification"
)

type NotificationServiceImpl struct{}

func (s *NotificationServiceImpl) GetMyNotifications(ctx context.Context, req *notification.GetNotificationsRequest) (resp *notification.GetNotificationsResponse, err error) {
	resp = &notification.GetNotificationsResponse{Code: 0, Message: "success"}

	db := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.receiver_id = ? AND nr.deleted_at IS NULL", req.UserId)

	if req.Type > 0 {
		db = db.Where("nt.template_type = ?", req.Type)
	}
	if req.Status > 0 {
		db = db.Where("nr.status = ?", req.Status)
	}
	if req.IsRead {
		db = db.Where("nr.read_time IS NOT NULL")
	} else if req.Type > 0 || req.Status > 0 || req.Keyword != "" {
		db = db.Where("nr.read_time IS NULL")
	}
	if req.Keyword != "" {
		db = db.Where("nr.title LIKE ? OR nr.content LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []model.NotificationRecord
	offset := int((req.Page - 1) * req.PageSize)
	db.Order("nr.created_at DESC").
		Offset(offset).Limit(int(req.PageSize)).
		Find(&list)

	resp.Total = total
	resp.Records = make([]*notification.NotificationRecord, len(list))
	for i, n := range list {
		resp.Records[i] = &notification.NotificationRecord{
			Id:           n.ID,
			ReceiverId:   n.ReceiverID,
			ReceiverName: n.ReceiverName,
			TemplateId:   n.TemplateID,
			TemplateName: n.TemplateName,
			TemplateType: int32(n.TemplateType),
			Title:        n.Title,
			Content:      n.Content,
			Channel:      int32(n.Channel),
			Status:       int32(n.Status),
			CreatedAt:    n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if n.ReadTime != nil {
			resp.Records[i].ReadTime = n.ReadTime.Format("2006-01-02 15:04:05")
		}
	}

	var unreadCount int64
	database.GetDB().Table("notification_record").
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", req.UserId).
		Count(&unreadCount)
	resp.UnreadCount = unreadCount

	return resp, nil
}

func (s *NotificationServiceImpl) GetNotificationDetail(ctx context.Context, id int64, userID int64) (resp *notification.GetNotificationsResponse, err error) {
	resp = &notification.GetNotificationsResponse{Code: 0, Message: "success"}

	var notification model.NotificationRecord
	result := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type, nt.content as template_content").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.id = ? AND nr.receiver_id = ?", id, userID).
		First(&notification)

	if result.Error != nil {
		resp.Code = 404
		resp.Message = "通知不存在"
		return resp, nil
	}

	if notification.ReadTime == nil {
		database.GetDB().Model(&model.NotificationRecord{}).
			Where("id = ?", id).
			Update("read_time", time.Now())
	}

	resp.Records = []*notification.NotificationRecord{{
		Id:           notification.ID,
		ReceiverId:   notification.ReceiverID,
		ReceiverName: notification.ReceiverName,
		TemplateId:   notification.TemplateID,
		TemplateName: notification.TemplateName,
		TemplateType: int32(notification.TemplateType),
		Title:        notification.Title,
		Content:      notification.Content,
		Channel:      int32(notification.Channel),
		Status:       int32(notification.Status),
		CreatedAt:    notification.CreatedAt.Format("2006-01-02 15:04:05"),
	}}

	return resp, nil
}

func (s *NotificationServiceImpl) MarkAsRead(ctx context.Context, req *notification.MarkAsReadRequest) (resp *notification.MarkAsReadResponse, err error) {
	resp = &notification.MarkAsReadResponse{Code: 0, Message: "success"}

	err = database.GetDB().Model(&model.NotificationRecord{}).
		Where("id = ? AND receiver_id = ?", req.Id, req.UserId).
		Update("read_time", time.Now()).Error
	if err != nil {
		resp.Code = 500
		resp.Message = "标记已读失败"
		return resp, nil
	}

	return resp, nil
}

func (s *NotificationServiceImpl) MarkAllAsRead(ctx context.Context, req *notification.MarkAllAsReadRequest) (resp *notification.MarkAllAsReadResponse, err error) {
	resp = &notification.MarkAllAsReadResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.NotificationRecord{}).
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", req.UserId).
		Update("read_time", time.Now())

	if result.Error != nil {
		resp.Code = 500
		resp.Message = "全部标记已读失败"
		return resp, nil
	}

	resp.MarkedCount = result.RowsAffected
	return resp, nil
}

func (s *NotificationServiceImpl) SendNotification(ctx context.Context, req *notification.SendNotificationRequest) (resp *notification.SendNotificationResponse, err error) {
	resp = &notification.SendNotificationResponse{Code: 0, Message: "success"}

	var template model.NotificationTemplate
	database.GetDB().Where("id = ?", req.TemplateId).First(&template)

	if template.ID == 0 {
		resp.Code = 404
		resp.Message = "通知模板不存在"
		return resp, nil
	}

	paramsMap := make(map[string]string)
	if req.Params != nil {
		paramsMap = req.Params
	}

	successCount := 0
	for _, receiverID := range req.ReceiverIds {
		var receiver model.User
		database.GetDB().Select("real_name, mobile").Where("id = ?", receiverID).First(&receiver)

		title := renderTemplate(template.TitleTemplate, paramsMap)
		content := renderTemplate(template.ContentTemplate, paramsMap)

		record := &model.NotificationRecord{
			ReceiverID:   receiverID,
			ReceiverName: receiver.RealName,
			TemplateID:   template.ID,
			TemplateName: template.TemplateName,
			TemplateType: template.TemplateType,
			Title:        title,
			Content:      content,
			Channel:      template.Channel,
			Status:       1,
			CreatedBy:    req.SenderId,
		}

		if err := database.GetDB().Create(record).Error; err != nil {
			logger.Error("Create notification error", logger.Error(err))
			continue
		}

		mq.SendAsync("dispute_notification", map[string]interface{}{
			"notificationId": record.ID,
			"receiverId":     receiverID,
			"receiverName":   receiver.RealName,
			"mobile":         receiver.Phone,
			"title":          title,
			"content":        content,
			"channel":        template.Channel,
		})

		successCount++
	}

	resp.SuccessCount = int32(successCount)
	return resp, nil
}

func (s *NotificationServiceImpl) GetNotificationTemplates(ctx context.Context, req *notification.GetTemplatesRequest) (resp *notification.GetTemplatesResponse, err error) {
	resp = &notification.GetTemplatesResponse{Code: 0, Message: "success"}

	var templates []model.NotificationTemplate
	db := database.GetDB().Where("status = 1 AND deleted_at IS NULL")
	if req.TemplateType > 0 {
		db = db.Where("template_type = ?", req.TemplateType)
	}
	result := db.Order("created_at DESC").Find(&templates)

	if result.Error != nil {
		resp.Code = 500
		resp.Message = "查询模板失败"
		return resp, nil
	}

	resp.Templates = make([]*notification.NotificationTemplate, len(templates))
	for i, t := range templates {
		resp.Templates[i] = &notification.NotificationTemplate{
			Id:             t.ID,
			TemplateName:   t.TemplateName,
			TemplateCode:   t.TemplateCode,
			TemplateType:   int32(t.TemplateType),
			TitleTemplate:  t.TitleTemplate,
			ContentTemplate: t.ContentTemplate,
			Channel:        int32(t.Channel),
			Status:         int32(t.Status),
			CreatedAt:      t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *NotificationServiceImpl) GetUnreadCount(ctx context.Context, req *notification.GetUnreadCountRequest) (resp *notification.GetUnreadCountResponse, err error) {
	resp = &notification.GetUnreadCountResponse{Code: 0, Message: "success"}

	var count int64
	result := database.GetDB().Table("notification_record").
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", req.UserId).
		Count(&count)

	if result.Error != nil {
		resp.Code = 500
		resp.Message = "查询未读数量失败"
		return resp, nil
	}

	resp.Count = count
	return resp, nil
}

func (s *NotificationServiceImpl) DeleteNotification(ctx context.Context, req *notification.DeleteNotificationRequest) (resp *notification.DeleteNotificationResponse, err error) {
	resp = &notification.DeleteNotificationResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.NotificationRecord{}).
		Where("id = ? AND receiver_id = ?", req.Id, req.UserId).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		resp.Code = 500
		resp.Message = "删除通知失败"
		return resp, nil
	}

	return resp, nil
}

func (s *NotificationServiceImpl) BatchDeleteNotifications(ctx context.Context, req *notification.BatchDeleteRequest) (resp *notification.BatchDeleteResponse, err error) {
	resp = &notification.BatchDeleteResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.NotificationRecord{}).
		Where("id IN ? AND receiver_id = ?", req.Ids, req.UserId).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		resp.Code = 500
		resp.Message = "批量删除通知失败"
		return resp, nil
	}

	resp.DeletedCount = result.RowsAffected
	return resp, nil
}

func (s *NotificationServiceImpl) SendNotificationByMQ(ctx context.Context, req *notification.SendByMQRequest) (resp *notification.SendByMQResponse, err error) {
	resp = &notification.SendByMQResponse{Code: 0, Message: "success"}

	paramsMap := make(map[string]string)
	if req.Params != nil {
		paramsMap = req.Params
	}

	mq.SendAsync("dispute_notification_task", map[string]interface{}{
		"templateCode": req.TemplateCode,
		"receiverIds":  req.ReceiverIds,
		"params":       paramsMap,
		"sendTime":     time.Now(),
	})

	return resp, nil
}

func renderTemplate(template string, params map[string]string) string {
	result := template
	for k, v := range params {
		placeholder := "{{" + k + "}}"
		result = replaceAll(result, placeholder, v)
	}
	return result
}

func replaceAll(s, old, new string) string {
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
