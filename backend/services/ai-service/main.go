package main

import (
	"net"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/logger"
	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:   "../../config/config.yaml",
		ServiceName:  "ai-service",
		EnableAI:     true,
		EnableMilvus: true,
		EnableRedis:  true,
		LogLevel:     "info",
	})
	defer initResult.Stop()

	port := bootstrap.GetServicePort("ai-service")
	svr := ai.NewServer(new(AIServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "ai-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("AI service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("AI service stopped", logger.Error(err))
	}
}
