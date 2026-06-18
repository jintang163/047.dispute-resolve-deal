package main

import (
	"net"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/logger"
	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow/workflowservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:     "../../config/config.yaml",
		ServiceName:    "workflow-service",
		EnableFlowable: true,
		EnableRedis:    true,
		LogLevel:       "info",
	})
	defer initResult.Stop()

	port := bootstrap.GetServicePort("workflow-service")
	svr := workflow.NewServer(new(WorkflowServiceImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "workflow-service"}),
		server.WithServiceAddr(&net.TCPAddr{Port: port}),
	)

	logger.Info("Workflow service starting on port", logger.Int("port", port))
	if err := svr.Run(); err != nil {
		logger.Error("Workflow service stopped", logger.Error(err))
	}
}
