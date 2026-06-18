package main

import (
	"net"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/logger"
	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification/notificationservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:  "../../config/config.yaml",
		ServiceName: "notification-service",
		EnableRedis: true,
		LogLevel:    "info",
	})
	defer initResult.Stop()

	port := bootstrap.GetServicePort("notification-service")
	svr := notification.NewServer(new(NotificationServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "notification-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Notification service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Notification service stopped", logger.Error(err))
	}
}
