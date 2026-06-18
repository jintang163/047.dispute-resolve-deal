package main

import (
	"context"

	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	user "github.com/dispute-resolve/user-service/kitex_gen/user"

	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct{}

func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginRequest) (resp *user.LoginResponse, err error) {
	resp = &user.LoginResponse{Code: 0, Message: "success"}

	var userModel model.User
	result := database.GetDB().Where("username = ? AND deleted_at IS NULL", req.Username).First(&userModel)
	if result.Error != nil {
		resp.Code = 400
		resp.Message = "用户名或密码错误"
		return resp, nil
	}

	if userModel.Status != 1 {
		resp.Code = 400
		resp.Message = "账号已被禁用"
		return resp, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.Password))
	if err != nil {
		resp.Code = 400
		resp.Message = "用户名或密码错误"
		return resp, nil
	}

	token, err := utils.GenerateToken(userModel.ID, userModel.Username, userModel.Role, userModel.OrganizationID)
	if err != nil {
		resp.Code = 500
		resp.Message = "生成Token失败"
		return resp, nil
	}

	resp.User = &user.User{
		Id:             userModel.ID,
		Username:       userModel.Username,
		RealName:       userModel.RealName,
		Role:           userModel.Role,
		Avatar:         userModel.Avatar,
		Mobile:         userModel.Mobile,
		Email:          userModel.Email,
		OrganizationId: userModel.OrganizationID,
		Status:         userModel.Status,
		CreatedAt:      userModel.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	resp.Token = token

	return resp, nil
}

func (s *UserServiceImpl) KioskLogin(ctx context.Context, deviceNo string) (resp *user.LoginResponse, err error) {
	resp = &user.LoginResponse{Code: 0, Message: "success"}

	var kiosk model.Kiosk
	result := database.GetDB().Where("device_no = ? AND deleted_at IS NULL", deviceNo).First(&kiosk)
	if result.Error != nil {
		resp.Code = 400
		resp.Message = "终端设备不存在"
		return resp, nil
	}

	if kiosk.Status != 1 {
		resp.Code = 400
		resp.Message = "终端设备已被禁用"
		return resp, nil
	}

	var userModel model.User
	result = database.GetDB().Where("id = ? AND deleted_at IS NULL", kiosk.BindUserID).First(&userModel)
	if result.Error != nil {
		resp.Code = 400
		resp.Message = "终端绑定用户不存在"
		return resp, nil
	}

	token, _ := utils.GenerateToken(userModel.ID, userModel.Username, userModel.Role, userModel.OrganizationID)

	resp.User = &user.User{
		Id:             userModel.ID,
		Username:       userModel.Username,
		RealName:       userModel.RealName,
		Role:           userModel.Role,
		OrganizationId: userModel.OrganizationID,
	}
	resp.Token = token

	return resp, nil
}

func (s *UserServiceImpl) MiniAppLogin(ctx context.Context, openid string) (resp *user.LoginResponse, err error) {
	resp = &user.LoginResponse{Code: 0, Message: "success"}

	var userModel model.User
	result := database.GetDB().Where("openid = ? AND deleted_at IS NULL", openid).First(&userModel)
	if result.Error != nil {
		userModel = model.User{
			Username:       "wx_" + openid[:8],
			RealName:       "微信用户",
			Role:           99,
			OpenID:         openid,
			Status:         1,
			OrganizationID: 1,
		}
		userModel.Password, _ = utils.HashPassword(utils.GenerateRandomString(16))
		database.GetDB().Create(&userModel)
	}

	token, _ := utils.GenerateToken(userModel.ID, userModel.Username, userModel.Role, userModel.OrganizationID)

	resp.User = &user.User{
		Id:             userModel.ID,
		Username:       userModel.Username,
		RealName:       userModel.RealName,
		Role:           userModel.Role,
		Openid:         userModel.OpenID,
		OrganizationId: userModel.OrganizationID,
	}
	resp.Token = token

	return resp, nil
}

func (s *UserServiceImpl) GetUserInfo(ctx context.Context, req *user.GetUserRequest) (resp *user.GetUserResponse, err error) {
	resp = &user.GetUserResponse{Code: 0, Message: "success"}

	var userModel model.User
	result := database.GetDB().Select("id, username, real_name, role, avatar, mobile, email, organization_id, created_at").
		Where("id = ? AND deleted_at IS NULL", req.UserId).First(&userModel)
	if result.Error != nil {
		resp.Code = 404
		resp.Message = "用户不存在"
		return resp, nil
	}

	resp.User = &user.User{
		Id:             userModel.ID,
		Username:       userModel.Username,
		RealName:       userModel.RealName,
		Role:           userModel.Role,
		Avatar:         userModel.Avatar,
		Mobile:         userModel.Mobile,
		Email:          userModel.Email,
		OrganizationId: userModel.OrganizationID,
		CreatedAt:      userModel.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

func (s *UserServiceImpl) GetUserList(ctx context.Context, req *user.GetUserListRequest) (resp *user.GetUserListResponse, err error) {
	resp = &user.GetUserListResponse{Code: 0, Message: "success"}

	var users []model.User
	var total int64

	db := database.GetDB().Model(&model.User{}).Where("deleted_at IS NULL")

	if req.Keyword != "" {
		db = db.Where("username LIKE ? OR real_name LIKE ? OR mobile LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Role > 0 {
		db = db.Where("role = ?", req.Role)
	}
	if req.OrganizationId > 0 {
		db = db.Where("organization_id = ?", req.OrganizationId)
	}

	db.Count(&total)
	offset := int((req.Page - 1) * req.PageSize)
	db.Offset(offset).Limit(int(req.PageSize)).Order("created_at DESC").Find(&users)

	resp.Total = total
	resp.Users = make([]*user.User, len(users))
	for i, u := range users {
		resp.Users[i] = &user.User{
			Id:             u.ID,
			Username:       u.Username,
			RealName:       u.RealName,
			Role:           u.Role,
			Avatar:         u.Avatar,
			Mobile:         u.Mobile,
			Email:          u.Email,
			OrganizationId: u.OrganizationID,
			Status:         u.Status,
			CreatedAt:      u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *UserServiceImpl) GetMediatorList(ctx context.Context, req *user.GetMediatorListRequest) (resp *user.GetMediatorListResponse, err error) {
	resp = &user.GetMediatorListResponse{Code: 0, Message: "success"}

	var users []model.User
	db := database.GetDB().Where("role = ? AND status = 1 AND deleted_at IS NULL", 4)
	if req.OrganizationId > 0 {
		db = db.Where("organization_id = ?", req.OrganizationId)
	}
	db.Order("real_name").Find(&users)

	resp.Mediators = make([]*user.User, len(users))
	for i, u := range users {
		resp.Mediators[i] = &user.User{
			Id:       u.ID,
			RealName: u.RealName,
			Mobile:   u.Mobile,
		}
	}

	return resp, nil
}

func (s *UserServiceImpl) GetOrganizationTree(ctx context.Context, req *user.GetOrganizationTreeRequest) (resp *user.GetOrganizationTreeResponse, err error) {
	resp = &user.GetOrganizationTreeResponse{Code: 0, Message: "success"}

	var orgs []model.Organization
	db := database.GetDB().Where("deleted_at IS NULL").Order("level, sort_order")
	if req.OrganizationId > 0 {
		db = db.Where("parent_id = ? OR id = ?", req.OrganizationId, req.OrganizationId)
	}
	db.Find(&orgs)

	resp.Organizations = make([]*user.Organization, len(orgs))
	for i, o := range orgs {
		resp.Organizations[i] = &user.Organization{
			Id:        o.ID,
			Name:      o.Name,
			Code:      o.Code,
			ParentId:  o.ParentID,
			Level:     o.Level,
			SortOrder: o.SortOrder,
			Leader:    o.Leader,
			Contact:   o.Contact,
			Address:   o.Address,
			Longitude: o.Longitude,
			Latitude:  o.Latitude,
			Status:    o.Status,
		}
	}

	return resp, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (resp *user.CreateUserResponse, err error) {
	resp = &user.CreateUserResponse{Code: 0, Message: "success"}

	var count int64
	database.GetDB().Model(&model.User{}).Where("username = ? AND deleted_at IS NULL", req.User.Username).Count(&count)
	if count > 0 {
		resp.Code = 400
		resp.Message = "用户名已存在"
		return resp, nil
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.User.Username+"123456"), bcrypt.DefaultCost)

	userModel := model.User{
		Username:       req.User.Username,
		RealName:       req.User.RealName,
		Password:       string(hashedPassword),
		Role:           req.User.Role,
		Mobile:         req.User.Mobile,
		Email:          req.User.Email,
		OrganizationID: req.User.OrganizationId,
		Status:         1,
	}

	result := database.GetDB().Create(&userModel)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "创建用户失败"
		logger.Error("Create user error", logger.Error(result.Error))
		return resp, nil
	}

	resp.UserId = userModel.ID
	return resp, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (resp *user.UpdateUserResponse, err error) {
	resp = &user.UpdateUserResponse{Code: 0, Message: "success"}

	updates := map[string]interface{}{
		"real_name":       req.User.RealName,
		"role":            req.User.Role,
		"mobile":          req.User.Mobile,
		"email":           req.User.Email,
		"organization_id": req.User.OrganizationId,
		"status":          req.User.Status,
	}

	result := database.GetDB().Model(&model.User{}).Where("id = ?", req.User.Id).Updates(updates)
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "更新用户失败"
		return resp, nil
	}

	return resp, nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (resp *user.DeleteUserResponse, err error) {
	resp = &user.DeleteUserResponse{Code: 0, Message: "success"}

	result := database.GetDB().Model(&model.User{}).Where("id = ?", req.UserId).Update("deleted_at", database.GetDB().NowFunc())
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "删除用户失败"
		return resp, nil
	}

	return resp, nil
}

func (s *UserServiceImpl) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (resp *user.ChangePasswordResponse, err error) {
	resp = &user.ChangePasswordResponse{Code: 0, Message: "success"}

	var userModel model.User
	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", req.UserId).First(&userModel)
	if result.Error != nil {
		resp.Code = 404
		resp.Message = "用户不存在"
		return resp, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.OldPassword))
	if err != nil {
		resp.Code = 400
		resp.Message = "原密码错误"
		return resp, nil
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	database.GetDB().Model(&userModel).Update("password", string(hashedPassword))

	return resp, nil
}

func (s *UserServiceImpl) ResetPassword(ctx context.Context, req *user.ResetPasswordRequest) (resp *user.ResetPasswordResponse, err error) {
	resp = &user.ResetPasswordResponse{Code: 0, Message: "success"}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	result := database.GetDB().Model(&model.User{}).Where("id = ?", req.UserId).Update("password", string(hashedPassword))
	if result.Error != nil {
		resp.Code = 500
		resp.Message = "重置密码失败"
		return resp, nil
	}

	return resp, nil
}

func (s *UserServiceImpl) GetUserDetail(ctx context.Context, req *user.GetUserRequest) (resp *user.GetUserResponse, err error) {
	return s.GetUserInfo(ctx, req)
}
