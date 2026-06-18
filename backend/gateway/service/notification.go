package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type NotificationService interface {
	GetMyNotifications(ctx context.Context, userID int64, page, pageSize int, typ int32, status int32, isRead *bool, keyword string) ([]map[string]interface{}, int64, int64, error)
	GetNotificationDetail(ctx context.Context, id int64, userID int64) (map[string]interface{}, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) (int64, error)
	SendNotification(ctx context.Context, receiverIDs []int64, templateID int64, params map[string]interface{}, notifyType string, senderID int64) error
	GetNotificationTemplates(ctx context.Context, templateType int32) ([]*model.NotificationTemplate, error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	DeleteNotification(ctx context.Context, id int64, userID int64) error
	BatchDeleteNotifications(ctx context.Context, ids []int64, userID int64) (int64, error)
	SendNotificationByMQ(ctx context.Context, templateCode string, receiverIDs []int64, params map[string]interface{}) error
}
