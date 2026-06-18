package notificationservice

import (
	"context"

	clientpkg "github.com/cloudwego/kitex/client"

	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification"
)

type kClient struct {
	c clientpkg.Client
}

func NewClient(destService string, opts ...clientpkg.Option) (Client, error) {
	cli, err := clientpkg.NewClient(destService, opts...)
	if err != nil {
		return nil, err
	}
	return &kClient{c: cli}, nil
}

func (c *kClient) GetMyNotifications(ctx context.Context, request *notification.GetNotificationsRequest) (r *notification.GetNotificationsResponse, err error) {
	r = &notification.GetNotificationsResponse{Code: 500, Message: "RPC not implemented - using local service fallback"}
	return
}

func (c *kClient) GetNotificationDetail(ctx context.Context, id int64, userId int64) (r *notification.GetNotificationsResponse, err error) {
	r = &notification.GetNotificationsResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) MarkAsRead(ctx context.Context, request *notification.MarkAsReadRequest) (r *notification.MarkAsReadResponse, err error) {
	r = &notification.MarkAsReadResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) MarkAllAsRead(ctx context.Context, request *notification.MarkAllAsReadRequest) (r *notification.MarkAllAsReadResponse, err error) {
	r = &notification.MarkAllAsReadResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) SendNotification(ctx context.Context, request *notification.SendNotificationRequest) (r *notification.SendNotificationResponse, err error) {
	r = &notification.SendNotificationResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetNotificationTemplates(ctx context.Context, request *notification.GetTemplatesRequest) (r *notification.GetTemplatesResponse, err error) {
	r = &notification.GetTemplatesResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetUnreadCount(ctx context.Context, request *notification.GetUnreadCountRequest) (r *notification.GetUnreadCountResponse, err error) {
	r = &notification.GetUnreadCountResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) DeleteNotification(ctx context.Context, request *notification.DeleteNotificationRequest) (r *notification.DeleteNotificationResponse, err error) {
	r = &notification.DeleteNotificationResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) BatchDeleteNotifications(ctx context.Context, request *notification.BatchDeleteRequest) (r *notification.BatchDeleteResponse, err error) {
	r = &notification.BatchDeleteResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) SendNotificationByMQ(ctx context.Context, request *notification.SendByMQRequest) (r *notification.SendByMQResponse, err error) {
	r = &notification.SendByMQResponse{Code: 500, Message: "RPC not implemented"}
	return
}
