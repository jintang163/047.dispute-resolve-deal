package main

import (
	"net"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/logger"
	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute/disputeservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:     "../../config/config.yaml",
		ServiceName:    "dispute-service",
		EnableFlowable: true,
		LogLevel:       "info",
	})
	defer initResult.Stop()

	port := bootstrap.GetServicePort("dispute-service")
	svr := dispute.NewServer(new(DisputeServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "dispute-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Dispute service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Dispute service stopped", logger.Error(err))
	}
}
