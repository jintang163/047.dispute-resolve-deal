package main

import (
	"log"
	"net"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	notification "github.com/dispute-resolve/notification-service/kitex_gen/notification/notificationservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.LoadConfig("../../config/config.yaml"); err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitLogger("notification-service")
	database.InitDB()

	port := config.GlobalConfig.ServicePorts.Notification
	svr := notification.NewServer(new(NotificationServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "notification-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Notification service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Notification service stopped", logger.Error(err))
	}
}
