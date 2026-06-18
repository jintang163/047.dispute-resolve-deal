package rpc

import (
	"github.com/dispute-resolve/common/logger"
	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow/workflowservice"

	"github.com/cloudwego/kitex/client"
)

var WorkflowClient workflow.Client

func initWorkflowClient() {
	var err error
	WorkflowClient, err = workflow.NewClient(
		"workflow-service",
		append(getClientOptions(),
			client.WithHostPorts(WorkflowServiceAddr),
		)...,
	)
	if err != nil {
		logger.Fatal("Failed to create workflow RPC client", logger.Error(err))
	}
	logger.Info("Workflow RPC client initialized", logger.String("addr", WorkflowServiceAddr))
}
