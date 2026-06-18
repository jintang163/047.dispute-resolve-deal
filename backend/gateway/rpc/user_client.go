package rpc

import (
	"github.com/dispute-resolve/common/logger"
	user "github.com/dispute-resolve/user-service/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/client"
)

var UserClient user.Client

func initUserClient() {
	var err error
	UserClient, err = user.NewClient(
		"user-service",
		append(getClientOptions(),
			client.WithHostPorts(UserServiceAddr),
		)...,
	)
	if err != nil {
		logger.Fatal("Failed to create user RPC client", logger.Error(err))
	}
	logger.Info("User RPC client initialized", logger.String("addr", UserServiceAddr))
}
