package rpc

import (
	"context"

	"github.com/dispute-resolve/common/logger"
	userModel "github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/gateway/service"
	"github.com/dispute-resolve/user-service/kitex_gen/user"
)

type UserServiceBridge struct{}

func (b *UserServiceBridge) Login(ctx context.Context, username, password string) (*user.LoginResponse, error) {
	if UserClient != nil {
		resp, err := UserClient.Login(ctx, &user.LoginRequest{Username: username, Password: password})
		if err == nil && resp != nil {
			return resp, nil
		}
		logger.Debug("RPC UserClient failed, fallback to local service", logger.Error(err))
	}
	localUser, token, err := service.UserServiceInst().Login(ctx, username, password)
	if err != nil {
		return &user.LoginResponse{Code: 500, Message: err.Error()}, nil
	}
	return &user.LoginResponse{
		Code:    0,
		Message: "success",
		Token:   token,
		User:    convertToRPCUser(localUser),
	}, nil
}

func convertToRPCUser(u *userModel.User) *user.User {
	if u == nil {
		return nil
	}
	return &user.User{
		Id:             u.ID,
		Username:       u.Username,
		RealName:       u.RealName,
		Role:           int32(u.Role),
		Avatar:         u.Avatar,
		Mobile:         u.Phone,
		Email:          u.Email,
		OrganizationId: u.OrganizationID,
		Status:         u.Status,
		Openid:         u.OpenID,
	}
}

var (
	UserBridge = &UserServiceBridge{}
)
