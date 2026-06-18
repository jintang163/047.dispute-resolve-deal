package userservice

import (
	"context"

	clientpkg "github.com/cloudwego/kitex/client"

	user "github.com/dispute-resolve/user-service/kitex_gen/user"
)

type kClient struct {
	c clientpkg.Client
}

func NewClient(destService string, opts ...clientpkg.Option) (Client, error) {
	cli, err := clientpkg.NewClient(destService, opts...)
	if err != nil {
		return nil, err
	}
	return &kClient{c: cli}, nil
}

func (c *kClient) Login(ctx context.Context, request *user.LoginRequest) (r *user.LoginResponse, err error) {
	r = &user.LoginResponse{Code: 500, Message: "RPC not implemented - using local service fallback"}
	return
}

func (c *kClient) KioskLogin(ctx context.Context, deviceNo string) (r *user.LoginResponse, err error) {
	r = &user.LoginResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) MiniAppLogin(ctx context.Context, openid string) (r *user.LoginResponse, err error) {
	r = &user.LoginResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetUserInfo(ctx context.Context, request *user.GetUserRequest) (r *user.GetUserResponse, err error) {
	r = &user.GetUserResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) ChangePassword(ctx context.Context, request *user.ChangePasswordRequest) (r *user.ChangePasswordResponse, err error) {
	r = &user.ChangePasswordResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetUserList(ctx context.Context, request *user.GetUserListRequest) (r *user.GetUserListResponse, err error) {
	r = &user.GetUserListResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetUserDetail(ctx context.Context, request *user.GetUserRequest) (r *user.GetUserResponse, err error) {
	r = &user.GetUserResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) CreateUser(ctx context.Context, request *user.CreateUserRequest) (r *user.CreateUserResponse, err error) {
	r = &user.CreateUserResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UpdateUser(ctx context.Context, request *user.UpdateUserRequest) (r *user.UpdateUserResponse, err error) {
	r = &user.UpdateUserResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) DeleteUser(ctx context.Context, request *user.DeleteUserRequest) (r *user.DeleteUserResponse, err error) {
	r = &user.DeleteUserResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) ResetPassword(ctx context.Context, request *user.ResetPasswordRequest) (r *user.ResetPasswordResponse, err error) {
	r = &user.ResetPasswordResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetOrganizationTree(ctx context.Context, request *user.GetOrganizationTreeRequest) (r *user.GetOrganizationTreeResponse, err error) {
	r = &user.GetOrganizationTreeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetMediatorList(ctx context.Context, request *user.GetMediatorListRequest) (r *user.GetMediatorListResponse, err error) {
	r = &user.GetMediatorListResponse{Code: 500, Message: "RPC not implemented"}
	return
}
