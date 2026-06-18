package main

import (
	"log"
	"net"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	user "github.com/dispute-resolve/user-service/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.LoadConfig("../../config/config.yaml"); err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitLogger("user-service")
	database.InitDB()

	port := config.GlobalConfig.ServicePorts.User
	svr := user.NewServer(new(UserServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "user-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("User service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("User service stopped", logger.Error(err))
	}
}
