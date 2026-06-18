package rpc

import (
	"github.com/dispute-resolve/common/logger"
	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification/notificationservice"

	"github.com/cloudwego/kitex/client"
)

var NotificationClient notification.Client

func initNotificationClient() {
	var err error
	NotificationClient, err = notification.NewClient(
		"notification-service",
		append(getClientOptions(),
			client.WithHostPorts(NotificationServiceAddr),
		)...,
	)
	if err != nil {
		logger.Fatal("Failed to create notification RPC client", logger.Error(err))
	}
	logger.Info("Notification RPC client initialized", logger.String("addr", NotificationServiceAddr))
}
