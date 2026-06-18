package impl

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/gateway/service"
)

type NotificationServiceImpl struct{}

func NewNotificationService() service.NotificationService {
	return &NotificationServiceImpl{}
}

func (s *NotificationServiceImpl) GetMyNotifications(ctx context.Context, userID int64, page, pageSize int, typ int32, status int32, isRead *bool, keyword string) ([]map[string]interface{}, int64, int64, error) {
	db := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.receiver_id = ? AND nr.deleted_at IS NULL", userID)

	if typ > 0 {
		db = db.Where("nt.template_type = ?", typ)
	}
	if status > 0 {
		db = db.Where("nr.status = ?", status)
	}
	if isRead != nil {
		if *isRead {
			db = db.Where("nr.read_time IS NOT NULL")
		} else {
			db = db.Where("nr.read_time IS NULL")
		}
	}
	if keyword != "" {
		db = db.Where("nr.title LIKE ? OR nr.content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := (page - 1) * pageSize
	db.Order("nr.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list)

	var unreadCount int64
	database.GetDB().Table("notification_record").
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", userID).
		Count(&unreadCount)

	return list, total, unreadCount, nil
}

func (s *NotificationServiceImpl) GetNotificationDetail(ctx context.Context, id int64, userID int64) (map[string]interface{}, error) {
	var notification map[string]interface{}
	result := database.GetDB().Table("notification_record nr").
		Select("nr.*, nt.template_name, nt.template_type, nt.content as template_content").
		Joins("LEFT JOIN notification_template nt ON nr.template_id = nt.id").
		Where("nr.id = ? AND nr.receiver_id = ?", id, userID).
		Find(&notification)

	if result.Error != nil {
		return nil, result.Error
	}

	if notification["read_time"] == nil {
		database.GetDB().Model(&model.NotificationRecord{}).
			Where("id = ?", id).
			Update("read_time", time.Now())
	}

	return notification, nil
}

func (s *NotificationServiceImpl) MarkAsRead(ctx context.Context, id int64, userID int64) error {
	return database.GetDB().Model(&model.NotificationRecord{}).
		Where("id = ? AND receiver_id = ?", id, userID).
		Update("read_time", time.Now()).Error
}

func (s *NotificationServiceImpl) MarkAllAsRead(ctx context.Context, userID int64) (int64, error) {
	result := database.GetDB().Model(&model.NotificationRecord{}).
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", userID).
		Update("read_time", time.Now())
	return result.RowsAffected, result.Error
}

func (s *NotificationServiceImpl) SendNotification(ctx context.Context, receiverIDs []int64, templateID int64, params map[string]interface{}, notifyType string, senderID int64) error {
	var template model.NotificationTemplate
	database.GetDB().Where("id = ?", templateID).First(&template)

	if template.ID == 0 {
		return nil
	}

	for _, receiverID := range receiverIDs {
		record := &model.NotificationRecord{
			ReceiverID:   receiverID,
			TemplateID:   templateID,
			TemplateName: template.TemplateName,
			TemplateType: template.TemplateType,
			Title:        renderTemplate(template.TitleTemplate, params),
			Content:      renderTemplate(template.ContentTemplate, params),
			Channel:      template.Channel,
			Status:       1,
			CreatedBy:    senderID,
		}

		if err := database.GetDB().Create(record).Error; err != nil {
			logger.Error("Create notification error", logger.Error(err))
			continue
		}

		mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
			"notificationId": record.ID,
			"receiverId":     receiverID,
			"title":          record.Title,
			"content":        record.Content,
			"channel":        template.Channel,
		})
	}

	return nil
}

func (s *NotificationServiceImpl) GetNotificationTemplates(ctx context.Context, templateType int32) ([]*model.NotificationTemplate, error) {
	var templates []*model.NotificationTemplate
	db := database.GetDB().Where("status = 1 AND deleted_at IS NULL")
	if templateType > 0 {
		db = db.Where("template_type = ?", templateType)
	}
	result := db.Order("created_at DESC").Find(&templates)
	return templates, result.Error
}

func (s *NotificationServiceImpl) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	var count int64
	result := database.GetDB().Table("notification_record").
		Where("receiver_id = ? AND read_time IS NULL AND deleted_at IS NULL", userID).
		Count(&count)
	return count, result.Error
}

func (s *NotificationServiceImpl) DeleteNotification(ctx context.Context, id int64, userID int64) error {
	return database.GetDB().Model(&model.NotificationRecord{}).
		Where("id = ? AND receiver_id = ?", id, userID).
		Update("deleted_at", time.Now()).Error
}

func (s *NotificationServiceImpl) BatchDeleteNotifications(ctx context.Context, ids []int64, userID int64) (int64, error) {
	result := database.GetDB().Model(&model.NotificationRecord{}).
		Where("id IN ? AND receiver_id = ?", ids, userID).
		Update("deleted_at", time.Now())
	return result.RowsAffected, result.Error
}

func (s *NotificationServiceImpl) SendNotificationByMQ(ctx context.Context, templateCode string, receiverIDs []int64, params map[string]interface{}) error {
	mq.SendAsync(constants.MQTopicNotificationTask, map[string]interface{}{
		"templateCode": templateCode,
		"receiverIds":  receiverIDs,
		"params":       params,
		"sendTime":     time.Now(),
	})
	return nil
}

func renderTemplate(template string, params map[string]interface{}) string {
	result := template
	for k, v := range params {
		placeholder := "{{" + k + "}}"
		result = replaceAll(result, placeholder, toString(v))
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
	switch val := v.(type) {
	case string:
		return val
	case int, int64, int32:
		return fmtInt(val)
	case float64, float32:
		return fmtFloat(val)
	default:
		return ""
	}
}

func fmtInt(v interface{}) string {
	switch val := v.(type) {
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	case int32:
		return string(rune(val))
	default:
		return ""
	}
}

func fmtFloat(v interface{}) string {
	switch val := v.(type) {
	case float64:
		return string(rune(int64(val)))
	case float32:
		return string(rune(int64(val)))
	default:
		return ""
	}
}
