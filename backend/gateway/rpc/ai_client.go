package rpc

import (
	"github.com/dispute-resolve/common/logger"
	ai "github.com/dispute-resolve/ai-service/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/client"
)

var AIClient ai.Client

func initAIClient() {
	var err error
	AIClient, err = ai.NewClient(
		"ai-service",
		append(getClientOptions(),
			client.WithHostPorts(AIServiceAddr),
		)...,
	)
	if err != nil {
		logger.Fatal("Failed to create AI RPC client", logger.Error(err))
	}
	logger.Info("AI RPC client initialized", logger.String("addr", AIServiceAddr))
}
