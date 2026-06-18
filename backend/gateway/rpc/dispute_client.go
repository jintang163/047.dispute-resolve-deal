package rpc

import (
	"github.com/dispute-resolve/common/logger"
	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute/disputeservice"

	"github.com/cloudwego/kitex/client"
)

var DisputeClient dispute.Client

func initDisputeClient() {
	var err error
	DisputeClient, err = dispute.NewClient(
		"dispute-service",
		append(getClientOptions(),
			client.WithHostPorts(DisputeServiceAddr),
		)...,
	)
	if err != nil {
		logger.Fatal("Failed to create dispute RPC client", logger.Error(err))
	}
	logger.Info("Dispute RPC client initialized", logger.String("addr", DisputeServiceAddr))
}
