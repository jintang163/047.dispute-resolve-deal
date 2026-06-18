package userservice

import (
	"context"

	user "github.com/dispute-resolve/user-service/kitex_gen/user"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type UserService interface {
	Login(ctx context.Context, request *user.LoginRequest) (r *user.LoginResponse, err error)
	KioskLogin(ctx context.Context, deviceNo string) (r *user.LoginResponse, err error)
	MiniAppLogin(ctx context.Context, openid string) (r *user.LoginResponse, err error)
	GetUserInfo(ctx context.Context, request *user.GetUserRequest) (r *user.GetUserResponse, err error)
	ChangePassword(ctx context.Context, request *user.ChangePasswordRequest) (r *user.ChangePasswordResponse, err error)
	GetUserList(ctx context.Context, request *user.GetUserListRequest) (r *user.GetUserListResponse, err error)
	GetUserDetail(ctx context.Context, request *user.GetUserRequest) (r *user.GetUserResponse, err error)
	CreateUser(ctx context.Context, request *user.CreateUserRequest) (r *user.CreateUserResponse, err error)
	UpdateUser(ctx context.Context, request *user.UpdateUserRequest) (r *user.UpdateUserResponse, err error)
	DeleteUser(ctx context.Context, request *user.DeleteUserRequest) (r *user.DeleteUserResponse, err error)
	ResetPassword(ctx context.Context, request *user.ResetPasswordRequest) (r *user.ResetPasswordResponse, err error)
	GetOrganizationTree(ctx context.Context, request *user.GetOrganizationTreeRequest) (r *user.GetOrganizationTreeResponse, err error)
	GetMediatorList(ctx context.Context, request *user.GetMediatorListRequest) (r *user.GetMediatorListResponse, err error)
}

type Server interface {
	RegisterService(svc *serviceinfo.ServiceInfo)
}

type Client interface {
	UserService
}
