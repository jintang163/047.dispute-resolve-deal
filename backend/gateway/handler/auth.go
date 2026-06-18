package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/dispute-resolve/common/auth"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
	UserInfo  *UserInfoResp `json:"userInfo"`
}

type UserInfoResp struct {
	UserID         int64  `json:"userId"`
	Username       string `json:"username"`
	RealName       string `json:"realName"`
	Phone          string `json:"phone"`
	Avatar         string `json:"avatar"`
	Role           int32  `json:"role"`
	RoleName       string `json:"roleName"`
	OrganizationID int64  `json:"organizationId"`
	OrgName        string `json:"orgName"`
	Position       string `json:"position"`
}

func HealthCheck(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"status":  "ok",
		"time":    time.Now().Unix(),
		"service": "dispute-gateway",
	}))
}

func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	var user struct {
		ID             int64  `gorm:"column:id"`
		Username       string `gorm:"column:username"`
		Password       string `gorm:"column:password"`
		RealName       string `gorm:"column:real_name"`
		Phone          string `gorm:"column:phone"`
		Avatar         string `gorm:"column:avatar"`
		Role           int32  `gorm:"column:role"`
		OrganizationID int64  `gorm:"column:organization_id"`
		Position       string `gorm:"column:position"`
		Status         int32  `gorm:"column:status"`
	}

	result := database.GetDB().Table("sys_user").
		Where("username = ? AND deleted_at IS NULL", req.Username).
		First(&user)

	if result.Error != nil {
		logger.Error("User not found", logger.Error(result.Error))
		c.JSON(http.StatusUnauthorized, response.Unauthorized("用户名或密码错误"))
		return
	}

	if user.Status != 1 {
		c.JSON(http.StatusForbidden, response.Forbidden("账号已被禁用"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Error("Password mismatch", logger.Error(err))
		c.JSON(http.StatusUnauthorized, response.Unauthorized("用户名或密码错误"))
		return
	}

	token, err := auth.GenerateToken(&auth.UserInfo{
		UserID:         user.ID,
		Username:       user.Username,
		RealName:       user.RealName,
		Role:           user.Role,
		OrganizationID: user.OrganizationID,
	})
	if err != nil {
		logger.Error("Generate token failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("生成Token失败"))
		return
	}

	cfg := config.GetConfig().JWT
	cacheKey := constants.RedisKeyPrefixToken + token
	cache.Set(ctx, cacheKey, user.ID, time.Duration(cfg.ExpireTime)*time.Second)

	var orgName string
	database.GetDB().Table("sys_organization").
		Select("org_name").
		Where("id = ?", user.OrganizationID).
		Scan(&orgName)

	roleName := constants.RoleMap[int(user.Role)]

	resp := &LoginResponse{
		Token:     token,
		ExpiresIn: cfg.ExpireTime,
		UserInfo: &UserInfoResp{
			UserID:         user.ID,
			Username:       user.Username,
			RealName:       user.RealName,
			Phone:          user.Phone,
			Avatar:         user.Avatar,
			Role:           user.Role,
			RoleName:       roleName,
			OrganizationID: user.OrganizationID,
			OrgName:        orgName,
			Position:       user.Position,
		},
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", user.ID).
		Update("last_login_at", time.Now())

	c.JSON(http.StatusOK, response.Success(resp))
}

func GetUserInfo(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	var user struct {
		ID             int64  `gorm:"column:id"`
		Username       string `gorm:"column:username"`
		RealName       string `gorm:"column:real_name"`
		Phone          string `gorm:"column:phone"`
		Avatar         string `gorm:"column:avatar"`
		Gender         int32  `gorm:"column:gender"`
		Email          string `gorm:"column:email"`
		Role           int32  `gorm:"column:role"`
		OrganizationID int64  `gorm:"column:organization_id"`
		Position       string `gorm:"column:position"`
		Specialty      string `gorm:"column:specialty"`
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", userInfo.UserID).
		First(&user)

	var orgName string
	database.GetDB().Table("sys_organization").
		Select("org_name").
		Where("id = ?", user.OrganizationID).
		Scan(&orgName)

	var permissions []string
	database.GetDB().Table("sys_role_permission").
		Select("permissions").
		Where("role_code = ?", getRoleCode(user.Role)).
		Scan(&permissions)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"userInfo": user,
		"orgName":  orgName,
		"roleName": constants.RoleMap[int(user.Role)],
		"permissions": permissions,
	}))
}

func ChangePassword(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var user struct {
		Password string `gorm:"column:password"`
	}
	database.GetDB().Table("sys_user").
		Select("password").
		Where("id = ?", userInfo.UserID).
		First(&user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("原密码错误"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("密码加密失败"))
		return
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", userInfo.UserID).
		Update("password", string(hashedPassword))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "密码修改成功"))
}

func KioskLogin(ctx context.Context, c *app.RequestContext) {
	var req struct {
		DeviceCode string `json:"deviceCode" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var device struct {
		ID             int64  `gorm:"column:id"`
		DeviceCode     string `gorm:"column:device_code"`
		OrganizationID int64  `gorm:"column:organization_id"`
		Status         int32  `gorm:"column:status"`
	}

	result := database.GetDB().Table("kiosk_device").
		Where("device_code = ?", req.DeviceCode).
		First(&device)

	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("终端不存在"))
		return
	}

	if device.Status != 1 {
		c.JSON(http.StatusForbidden, response.Forbidden("终端已离线或故障"))
		return
	}

	token, err := auth.GenerateToken(&auth.UserInfo{
		UserID:         device.ID,
		Username:       device.DeviceCode,
		RealName:       "自助终端",
		Role:           99,
		OrganizationID: device.OrganizationID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("生成Token失败"))
		return
	}

	cacheKey := constants.RedisKeyPrefixToken + token
	cache.Set(ctx, cacheKey, device.ID, time.Hour*24)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"token":     token,
		"expiresIn": 86400,
		"deviceId":  device.ID,
	}))
}

func MiniAppLogin(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var user struct {
		ID       int64  `gorm:"column:id"`
		RealName string `gorm:"column:real_name"`
		Phone    string `gorm:"column:phone"`
	}

	result := database.GetDB().Table("sys_user").
		Where("phone = ? AND deleted_at IS NULL", req.Phone).
		First(&user)

	var userId int64
	var userName string

	if result.Error != nil {
		userId = utils.GenerateID()
		userName = "小程序用户"
		database.GetDB().Table("sys_user").Create(map[string]interface{}{
			"id":       userId,
			"username": "mini_" + req.Phone,
			"password": "miniapp_user",
			"real_name": userName,
			"phone":     req.Phone,
			"role":      99,
			"organization_id": 1,
			"status":    1,
		})
	} else {
		userId = user.ID
		userName = user.RealName
	}

	token, err := auth.GenerateToken(&auth.UserInfo{
		UserID:         userId,
		Username:       req.Phone,
		RealName:       userName,
		Role:           99,
		OrganizationID: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("生成Token失败"))
		return
	}

	cacheKey := constants.RedisKeyPrefixToken + token
	cache.Set(ctx, cacheKey, userId, time.Hour*24*7)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"token":     token,
		"expiresIn": 604800,
		"userId":    userId,
		"userName":  userName,
	}))
}

func getRoleCode(role int32) string {
	switch role {
	case constants.RoleAdmin:
		return constants.RoleCodeAdmin
	case constants.RoleDirector:
		return constants.RoleCodeDirector
	case constants.RoleLeader:
		return constants.RoleCodeLeader
	case constants.RoleMediator:
		return constants.RoleCodeMediator
	default:
		return ""
	}
}
