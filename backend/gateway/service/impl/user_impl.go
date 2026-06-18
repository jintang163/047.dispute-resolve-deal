package impl

import (
	"context"
	"errors"
	"time"

	"github.com/dispute-resolve/common/auth"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserServiceImpl struct{}

func NewUserService() service.UserService {
	return &UserServiceImpl{}
}

func (s *UserServiceImpl) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	var user model.User
	result := database.GetDB().Where("username = ? AND deleted_at IS NULL", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户名或密码错误")
		}
		return nil, "", result.Error
	}

	if user.Status != 1 {
		return nil, "", errors.New("账号已被禁用")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		return nil, "", err
	}

	cacheKey := constants.RedisPrefixToken + utils.Int64ToString(user.ID)
	cache.Set(ctx, cacheKey, token, 24*time.Hour)

	return &user, token, nil
}

func (s *UserServiceImpl) KioskLogin(ctx context.Context, deviceNo string) (*model.User, string, error) {
	var kiosk model.Kiosk
	result := database.GetDB().Where("device_no = ? AND deleted_at IS NULL", deviceNo).First(&kiosk)
	if result.Error != nil {
		return nil, "", errors.New("终端设备不存在")
	}

	if kiosk.Status != 1 {
		return nil, "", errors.New("终端设备已被禁用")
	}

	var user model.User
	result = database.GetDB().Where("id = ? AND deleted_at IS NULL", kiosk.BindUserID).First(&user)
	if result.Error != nil {
		return nil, "", errors.New("终端绑定用户不存在")
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *UserServiceImpl) MiniAppLogin(ctx context.Context, openid string) (*model.User, string, error) {
	var user model.User
	result := database.GetDB().Where("openid = ? AND deleted_at IS NULL", openid).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = model.User{
				Username:       "wx_" + openid[:8],
				RealName:       "微信用户",
				Role:           constants.RoleUser,
				OpenID:         openid,
				Status:         1,
				OrganizationID: 1,
			}
			user.Password, _ = utils.HashPassword(utils.GenerateRandomString(16))
			database.GetDB().Create(&user)
		} else {
			return nil, "", result.Error
		}
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role, user.OrganizationID)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *UserServiceImpl) GetUserInfo(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	result := database.GetDB().Select("id, username, real_name, role, avatar, mobile, email, organization_id, created_at").
		Where("id = ? AND deleted_at IS NULL", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserServiceImpl) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	var user model.User
	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", userID).First(&user)
	if result.Error != nil {
		return result.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("原密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return database.GetDB().Model(&user).Update("password", string(hashedPassword)).Error
}

func (s *UserServiceImpl) GetUserList(ctx context.Context, page, pageSize int, keyword string, role int32, orgID int64) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	db := database.GetDB().Model(&model.User{}).Where("deleted_at IS NULL")

	if keyword != "" {
		db = db.Where("username LIKE ? OR real_name LIKE ? OR mobile LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role > 0 {
		db = db.Where("role = ?", role)
	}
	if orgID > 0 {
		db = db.Where("organization_id = ?", orgID)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

func (s *UserServiceImpl) GetUserDetail(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user *model.User) error {
	var count int64
	database.GetDB().Model(&model.User{}).Where("username = ? AND deleted_at IS NULL", user.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return database.GetDB().Create(user).Error
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, user *model.User) error {
	return database.GetDB().Model(user).Omit("password", "created_at").Updates(user).Error
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID int64) error {
	return database.GetDB().Model(&model.User{}).Where("id = ?", userID).Update("deleted_at", time.Now()).Error
}

func (s *UserServiceImpl) ResetPassword(ctx context.Context, userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return database.GetDB().Model(&model.User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error
}

func (s *UserServiceImpl) GetOrganizationTree(ctx context.Context, orgID int64) ([]*model.Organization, error) {
	var orgs []*model.Organization
	db := database.GetDB().Where("deleted_at IS NULL").Order("level, sort_order")
	if orgID > 0 {
		db = db.Where("parent_id = ? OR id = ?", orgID, orgID)
	}
	result := db.Find(&orgs)
	if result.Error != nil {
		return nil, result.Error
	}
	return orgs, nil
}

func (s *UserServiceImpl) CreateOrganization(ctx context.Context, org *model.Organization) error {
	return database.GetDB().Create(org).Error
}

func (s *UserServiceImpl) UpdateOrganization(ctx context.Context, org *model.Organization) error {
	return database.GetDB().Model(org).Omit("created_at").Updates(org).Error
}

func (s *UserServiceImpl) DeleteOrganization(ctx context.Context, orgID int64) error {
	return database.GetDB().Model(&model.Organization{}).Where("id = ?", orgID).Update("deleted_at", time.Now()).Error
}

func (s *UserServiceImpl) GetMediatorList(ctx context.Context, orgID int64, specialtyID int64) ([]*model.User, error) {
	var users []*model.User
	db := database.GetDB().Where("role = ? AND status = 1 AND deleted_at IS NULL", constants.RoleMediator)
	if orgID > 0 {
		db = db.Where("organization_id = ?", orgID)
	}
	if specialtyID > 0 {
		db = db.Joins("JOIN user_specialty us ON us.user_id = user.id").Where("us.specialty_id = ?", specialtyID)
	}
	result := db.Order("real_name").Find(&users)
	if result.Error != nil {
		logger.Error("GetMediatorList error", logger.Error(result.Error))
		return nil, result.Error
	}
	return users, nil
}
