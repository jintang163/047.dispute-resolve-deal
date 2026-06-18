package main

import (
	"log"
	"net"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute/disputeservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.LoadConfig("../../config/config.yaml"); err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitLogger("dispute-service")
	database.InitDB()

	port := config.GlobalConfig.ServicePorts.Dispute
	svr := dispute.NewServer(new(DisputeServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "dispute-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Dispute service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Dispute service stopped", logger.Error(err))
	}
}
