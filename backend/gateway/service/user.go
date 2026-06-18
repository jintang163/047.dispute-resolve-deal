package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type UserService interface {
	Login(ctx context.Context, username, password string) (*model.User, string, error)
	KioskLogin(ctx context.Context, deviceNo string) (*model.User, string, error)
	MiniAppLogin(ctx context.Context, openid string) (*model.User, string, error)
	GetUserInfo(ctx context.Context, userID int64) (*model.User, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	GetUserList(ctx context.Context, page, pageSize int, keyword string, role int32, orgID int64) ([]*model.User, int64, error)
	GetUserDetail(ctx context.Context, userID int64) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userID int64) error
	ResetPassword(ctx context.Context, userID int64, newPassword string) error
	GetOrganizationTree(ctx context.Context, orgID int64) ([]*model.Organization, error)
	CreateOrganization(ctx context.Context, org *model.Organization) error
	UpdateOrganization(ctx context.Context, org *model.Organization) error
	DeleteOrganization(ctx context.Context, orgID int64) error
	GetMediatorList(ctx context.Context, orgID int64, specialtyID int64) ([]*model.User, error)
}
