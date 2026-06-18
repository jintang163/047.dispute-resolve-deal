package main

import (
	"log"
	"net"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow/workflowservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	if err := config.LoadConfig("../../config/config.yaml"); err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitLogger("workflow-service")
	database.InitDB()

	port := config.GlobalConfig.ServicePorts.Workflow
	svr := workflow.NewServer(new(WorkflowServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "workflow-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Workflow service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Workflow service stopped", logger.Error(err))
	}
}
