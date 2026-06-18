package main

import (
	"log"
	"net"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.LoadConfig("../../config/config.yaml"); err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitLogger("ai-service")
	database.InitDB()

	port := config.GlobalConfig.ServicePorts.AI
	svr := ai.NewServer(new(AIServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "ai-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("AI service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("AI service stopped", logger.Error(err))
	}
}
