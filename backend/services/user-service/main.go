package main

import (
	"net"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/logger"
	user "github.com/dispute-resolve/user-service/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:  "../../config/config.yaml",
		ServiceName: "user-service",
		EnableRedis: true,
		LogLevel:    "info",
	})
	defer initResult.Stop()

	port := bootstrap.GetServicePort("user-service")
	svr := user.NewServer(new(UserServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "user-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("User service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("User service stopped", logger.Error(err))
	}
}
