package rpc

import (
	"sync"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/retry"
)

var (
	once sync.Once

	UserServiceAddr         string
	DisputeServiceAddr      string
	WorkflowServiceAddr     string
	AIServiceAddr           string
	NotificationServiceAddr string
)

func InitRPCClients() {
	once.Do(func() {
		UserServiceAddr = mustParseServiceAddr(config.GlobalConfig.Services.UserService)
		DisputeServiceAddr = mustParseServiceAddr(config.GlobalConfig.Services.DisputeService)
		WorkflowServiceAddr = mustParseServiceAddr(config.GlobalConfig.Services.WorkflowService)
		AIServiceAddr = mustParseServiceAddr(config.GlobalConfig.Services.AIService)
		NotificationServiceAddr = mustParseServiceAddr(config.GlobalConfig.Services.NotificationService)

		initUserClient()
		initDisputeClient()
		initWorkflowClient()
		initAIClient()
		initNotificationClient()
		logger.Info("All RPC clients initialized")
	})
}

func getClientOptions() []client.Option {
	return []client.Option{
		client.WithRPCTimeout(3000),
		client.WithConnectTimeout(1000),
		client.WithRetry(retry.NewFailurePolicy()),
	}
}

func mustParseServiceAddr(addr string) string {
	if addr == "" {
		panic("service address is empty")
	}
	return addr
}
