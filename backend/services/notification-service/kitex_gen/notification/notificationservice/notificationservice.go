package notificationservice

import (
	"context"

	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type NotificationService interface {
	GetMyNotifications(ctx context.Context, request *notification.GetNotificationsRequest) (r *notification.GetNotificationsResponse, err error)
	GetNotificationDetail(ctx context.Context, id int64, userId int64) (r *notification.GetNotificationsResponse, err error)
	MarkAsRead(ctx context.Context, request *notification.MarkAsReadRequest) (r *notification.MarkAsReadResponse, err error)
	MarkAllAsRead(ctx context.Context, request *notification.MarkAllAsReadRequest) (r *notification.MarkAllAsReadResponse, err error)
	SendNotification(ctx context.Context, request *notification.SendNotificationRequest) (r *notification.SendNotificationResponse, err error)
	GetNotificationTemplates(ctx context.Context, request *notification.GetTemplatesRequest) (r *notification.GetTemplatesResponse, err error)
	GetUnreadCount(ctx context.Context, request *notification.GetUnreadCountRequest) (r *notification.GetUnreadCountResponse, err error)
	DeleteNotification(ctx context.Context, request *notification.DeleteNotificationRequest) (r *notification.DeleteNotificationResponse, err error)
	BatchDeleteNotifications(ctx context.Context, request *notification.BatchDeleteRequest) (r *notification.BatchDeleteResponse, err error)
	SendNotificationByMQ(ctx context.Context, request *notification.SendByMQRequest) (r *notification.SendByMQResponse, err error)
}

type Server interface {
	RegisterService(svc *serviceinfo.ServiceInfo)
}

type Client interface {
	NotificationService
}
