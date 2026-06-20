package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/crypto/bcrypt"
)

type UserCreateRequest struct {
	Username       string `json:"username" binding:"required"`
	RealName       string `json:"realName" binding:"required"`
	Password       string `json:"password" binding:"required"`
	Phone          string `json:"phone" binding:"required"`
	Email          string `json:"email"`
	IDCard         string `json:"idCard"`
	Role           int32  `json:"role" binding:"required"`
	OrganizationID int64  `json:"organizationId" binding:"required"`
	Specialty      string `json:"specialty"`
	Avatar         string `json:"avatar"`
}

type UserUpdateRequest struct {
	RealName       string `json:"realName"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	IDCard         string `json:"idCard"`
	Role           int32  `json:"role"`
	OrganizationID int64  `json:"organizationId"`
	Specialty    string `json:"specialty"`
	Status         int32  `json:"status"`
}

type OrganizationCreateRequest struct {
	OrgName    string `json:"orgName" binding:"required"`
	OrgCode    string `json:"orgCode" binding:"required"`
	OrgType    int32  `json:"orgType" binding:"required"`
	ParentID   int64  `json:"parentId"`
	Address    string `json:"address"`
	Contact    string `json:"contact"`
	Phone      string `json:"phone"`
	SortOrder  int32  `json:"sortOrder"`
}

type RoleCreateRequest struct {
	RoleName    string `json:"roleName" binding:"required"`
	RoleCode    string `json:"roleCode" binding:"required"`
	RoleLevel   int32  `json:"roleLevel" binding:"required"`
	Description string `json:"description"`
	Permissions []int64 `json:"permissions"`
}

func GetUserList(ctx context.Context, c *app.RequestContext) {
	var req struct {
		common.BaseQuery
		Role           int32  `form:"role"`
		Status         int32  `form:"status"`
		OrganizationID int64  `form:"organizationId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("sys_user su").
		Select("su.*, so.org_name, so.org_code").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.deleted_at IS NULL")

	if userInfo.Role == constants.RoleLeader {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
			Pluck("id", &childOrgs)
		db = db.Where("su.organization_id IN ?", childOrgs)
	} else if userInfo.Role == constants.RoleDirector {
		db = db.Where("su.role >= ?", constants.RoleMediator)
	}

	if req.Role > 0 {
		db = db.Where("su.role = ?", req.Role)
	}
	if req.Status > 0 {
		db = db.Where("su.status = ?", req.Status)
	}
	if req.OrganizationID > 0 {
		db = db.Where("su.organization_id = ?", req.OrganizationID)
	}
	if req.Keyword != "" {
		db = db.Where("su.real_name LIKE ? OR su.username LIKE ? OR su.phone LIKE ?", 
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("su.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if role, ok := item["role"].(int); ok {
			item["role_name"] = constants.RoleMap[role]
		}
		delete(item, "password")
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetUserDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var user map[string]interface{}
	result := database.GetDB().Table("sys_user su").
		Select("su.*, so.org_name, so.org_code").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.id = ? AND su.deleted_at IS NULL", id).
		Find(&user)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("用户不存在"))
		return
	}

	if role, ok := user["role"].(int); ok {
		user["role_name"] = constants.RoleMap[role]
	}
	delete(user, "password")

	c.JSON(http.StatusOK, response.Success(user))
}

func CreateUser(ctx context.Context, c *app.RequestContext) {
	var req UserCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > req.Role {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限创建该等级的用户"))
		return
	}

	var count int64
	database.GetDB().Table("sys_user").
		Where("username = ? OR phone = ?", req.Username, req.Phone).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("用户名或手机号已存在"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Hash password failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建用户失败"))
		return
	}

	userID := utils.GenerateID()
	user := map[string]interface{}{
		"id":              userID,
		"username":        req.Username,
		"real_name":       req.RealName,
		"password":        string(hashedPassword),
		"phone":           req.Phone,
		"email":           req.Email,
		"id_card":         req.IDCard,
		"role":            req.Role,
		"organization_id": req.OrganizationID,
		"specialty":       req.Specialty,
		"avatar":          req.Avatar,
		"status":          1,
		"created_by":      userInfo.UserID,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("sys_user").Create(user).Error; err != nil {
		tx.Rollback()
		logger.Error("Create user failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建用户失败"))
		return
	}

	rolePerm := map[string]interface{}{
		"user_id":      userID,
		"role_code":    getRoleCode(req.Role),
		"role_name":    constants.RoleMap[int(req.Role)],
		"granted_by":   userInfo.UserID,
		"granted_time": time.Now(),
	}
	tx.Table("sys_role_permission").Create(rolePerm)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": userID,
	}, "创建用户成功"))
}

func UpdateUser(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UserUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var existingUser struct {
		Role int32 `gorm:"column:role"`
	}
	database.GetDB().Table("sys_user").
		Select("role").
		Where("id = ?", id).
		First(&existingUser)

	if userInfo.Role > existingUser.Role && userInfo.UserID != id {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限修改该用户"))
		return
	}

	updates := map[string]interface{}{}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.IDCard != "" {
		updates["id_card"] = req.IDCard
	}
	if req.Role > 0 {
		updates["role"] = req.Role
	}
	if req.OrganizationID > 0 {
		updates["organization_id"] = req.OrganizationID
	}
	if req.Specialty != "" {
		updates["specialty"] = req.Specialty
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", id).
		Updates(updates)

	cache.Del(ctx, fmt.Sprintf("%s%d", constants.RedisKeyPrefixUser, id))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "更新用户成功"))
}

func DeleteUser(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	if userInfo.UserID == id {
		c.JSON(http.StatusBadRequest, response.BadRequest("不能删除自己"))
		return
	}

	var existingUser struct {
		Role int32 `gorm:"column:role"`
	}
	database.GetDB().Table("sys_user").
		Select("role").
		Where("id = ?", id).
		First(&existingUser)

	if userInfo.Role > existingUser.Role {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除该用户"))
		return
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", id).
		Update("deleted_at", time.Now())

	cache.Del(ctx, fmt.Sprintf("%s%d", constants.RedisKeyPrefixUser, id))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除用户成功"))
}

func ResetPassword(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var req struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var existingUser struct {
		Role int32 `gorm:"column:role"`
	}
	database.GetDB().Table("sys_user").
		Select("role").
		Where("id = ?", id).
		First(&existingUser)

	if userInfo.Role > existingUser.Role && userInfo.UserID != id {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限重置该用户密码"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("重置密码失败"))
		return
	}

	database.GetDB().Table("sys_user").
		Where("id = ?", id).
		Update("password", string(hashedPassword))

	cache.Del(ctx, fmt.Sprintf("%s%d", constants.RedisKeyPrefixUser, id))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "密码重置成功"))
}

func GetOrganizationTree(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var orgs []map[string]interface{}
	query := database.GetDB().Table("sys_organization").
		Select("id, parent_id, org_name, org_code, org_type, address, contact, phone, sort_order").
		Where("deleted_at IS NULL")

	if userInfo.Role >= constants.RoleLeader {
		query = query.Where("id = ? OR parent_id = ?", userInfo.OrganizationID, userInfo.OrganizationID)
	}

	query.Order("sort_order ASC, id ASC").Find(&orgs)

	orgTypeMap := map[int]string{
		constants.OrgTypeCenter:   "综治中心",
		constants.OrgTypeStreet:   "街道办",
		constants.OrgTypeCommunity: "社区",
		constants.OrgTypeVillage:  "村委会",
	}

	for _, item := range orgs {
		if ot, ok := item["org_type"].(int); ok {
			item["org_type_name"] = orgTypeMap[ot]
		}
	}

	tree := buildOrgTree(orgs, 0)

	c.JSON(http.StatusOK, response.Success(tree))
}

func CreateOrganization(ctx context.Context, c *app.RequestContext) {
	var req OrganizationCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限创建组织"))
		return
	}

	var count int64
	database.GetDB().Table("sys_organization").
		Where("org_code = ?", req.OrgCode).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("组织编码已存在"))
		return
	}

	orgID := utils.GenerateID()
	org := map[string]interface{}{
		"id":         orgID,
		"parent_id":  req.ParentID,
		"org_name":   req.OrgName,
		"org_code":   req.OrgCode,
		"org_type":   req.OrgType,
		"address":    req.Address,
		"contact":    req.Contact,
		"phone":      req.Phone,
		"sort_order": req.SortOrder,
		"created_by": userInfo.UserID,
	}

	database.GetDB().Table("sys_organization").Create(org)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": orgID,
	}, "创建组织成功"))
}

func UpdateOrganization(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req OrganizationCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限修改组织"))
		return
	}

	updates := map[string]interface{}{}
	if req.OrgName != "" {
		updates["org_name"] = req.OrgName
	}
	if req.OrgCode != "" {
		updates["org_code"] = req.OrgCode
	}
	if req.OrgType > 0 {
		updates["org_type"] = req.OrgType
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Contact != "" {
		updates["contact"] = req.Contact
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}

	database.GetDB().Table("sys_organization").
		Where("id = ?", id).
		Updates(updates)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "更新组织成功"))
}

func DeleteOrganization(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除组织"))
		return
	}

	var childCount int64
	database.GetDB().Table("sys_organization").
		Where("parent_id = ? AND deleted_at IS NULL", id).
		Count(&childCount)

	if childCount > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该组织下存在子组织，不能删除"))
		return
	}

	var userCount int64
	database.GetDB().Table("sys_user").
		Where("organization_id = ? AND deleted_at IS NULL", id).
		Count(&userCount)

	if userCount > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该组织下存在用户，不能删除"))
		return
	}

	database.GetDB().Table("sys_organization").
		Where("id = ?", id).
		Update("deleted_at", time.Now())

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除组织成功"))
}

func GetMediatorList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	specialty := c.Query("specialty")

	var orgs []int64
	database.GetDB().Table("sys_organization").
		Select("id").
		Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
		Pluck("id", &orgs)

	db := database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, su.phone, su.avatar, su.specialty, so.org_name").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.role = ?", constants.RoleMediator).
		Where("su.status = 1").
		Where("su.deleted_at IS NULL").
		Where("su.organization_id IN ?", orgs)

	if specialty != "" {
		db = db.Where("su.specialty LIKE ?", "%"+specialty+"%")
	}

	var list []map[string]interface{}
	db.Order("su.id ASC").Find(&list)

	mediatorIDs := make([]int64, 0, len(list))
	mediatorIDMap := make(map[int64]map[string]interface{}, len(list))
	for _, item := range list {
		var id int64
		switch v := item["id"].(type) {
		case int64:
			id = v
		case int:
			id = int64(v)
		case float64:
			id = int64(v)
		}
		if id > 0 {
			mediatorIDs = append(mediatorIDs, id)
			mediatorIDMap[id] = item
		}
	}

	if len(mediatorIDs) > 0 {
		var loadResults []struct {
			MediatorID  int64 `gorm:"column:mediator_id"`
			PendingCount int64 `gorm:"column:pending_count"`
		}
		database.GetDB().Table("dispute_case").
			Select("mediator_id, COUNT(*) as pending_count").
			Where("mediator_id IN ?", mediatorIDs).
			Where("status < ? AND deleted_at IS NULL", constants.CaseStatusClosed).
			Group("mediator_id").
			Find(&loadResults)

		for _, r := range loadResults {
			if item, ok := mediatorIDMap[r.MediatorID]; ok {
				item["pending_case_count"] = r.PendingCount
				if r.PendingCount >= 10 {
					item["is_high_load"] = true
				} else {
					item["is_high_load"] = false
				}
			}
		}
		for _, item := range list {
			if _, exists := item["pending_case_count"]; !exists {
				item["pending_case_count"] = int64(0)
				item["is_high_load"] = false
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetMediatorLoad(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid mediator id"))
		return
	}

	var mediator struct {
		ID             int64  `gorm:"column:id"`
		RealName       string `gorm:"column:real_name"`
		Phone          string `gorm:"column:phone"`
		OrganizationID int64  `gorm:"column:organization_id"`
		Role           int32  `gorm:"column:role"`
		Status         int32  `gorm:"column:status"`
	}
	result := database.GetDB().Table("sys_user").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&mediator)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("调解员不存在"))
		return
	}
	if mediator.Status != 1 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该调解员已被禁用"))
		return
	}
	if mediator.Role != constants.RoleMediator {
		c.JSON(http.StatusBadRequest, response.BadRequest("该用户不是调解员角色"))
		return
	}

	var pendingCount int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status < ? AND deleted_at IS NULL", id, constants.CaseStatusClosed).
		Count(&pendingCount)

	var mediatingCount int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status = ? AND deleted_at IS NULL", id, constants.CaseStatusMediating).
		Count(&mediatingCount)

	var pendingAssignCount int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status = ? AND deleted_at IS NULL", id, constants.CaseStatusPending).
		Count(&pendingAssignCount)

	isHighLoad := pendingCount >= 10
	suggestion := ""
	if isHighLoad {
		suggestion = "该调解员负载较高，建议选择其他人"
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"mediatorId":          mediator.ID,
		"mediatorName":        mediator.RealName,
		"pendingCaseCount":    pendingCount,
		"mediatingCaseCount":  mediatingCount,
		"pendingAssignCount":  pendingAssignCount,
		"isHighLoad":          isHighLoad,
		"loadThreshold":       10,
		"suggestion":          suggestion,
	}))
}

func GetOperationLogList(ctx context.Context, c *app.RequestContext) {
	var req struct {
		common.BaseQuery
		OperationType string `form:"operationType"`
		OperatorID    int64  `form:"operatorId"`
		common.DateRangeQuery
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("sys_operation_log sol").
		Select("sol.*, su.real_name as operator_name").
		Joins("LEFT JOIN sys_user su ON sol.operator_id = su.id").
		Where("1=1")

	if userInfo.Role >= constants.RoleLeader {
		db = db.Where("sol.organization_id = ?", userInfo.OrganizationID)
	}

	if req.OperationType != "" {
		db = db.Where("sol.operation_type = ?", req.OperationType)
	}
	if req.OperatorID > 0 {
		db = db.Where("sol.operator_id = ?", req.OperatorID)
	}
	if req.StartTime != "" {
		db = db.Where("sol.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("sol.created_at <= ?", req.EndTime)
	}
	if req.Keyword != "" {
		db = db.Where("sol.operation_detail LIKE ?", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("sol.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func buildOrgTree(orgs []map[string]interface{}, parentID int64) []map[string]interface{} {
	var tree []map[string]interface{}
	for _, org := range orgs {
		pid := org["parent_id"].(int64)
		if pid == parentID {
			children := buildOrgTree(orgs, org["id"].(int64))
			if len(children) > 0 {
				org["children"] = children
			}
			tree = append(tree, org)
		}
	}
	return tree
}

func getRoleCode(role int32) string {
	switch role {
	case constants.RoleDirector:
		return constants.RoleCodeDirector
	case constants.RoleLeader:
		return constants.RoleCodeLeader
	case constants.RoleMediator:
		return constants.RoleCodeMediator
	case constants.RoleAdmin:
		return constants.RoleCodeAdmin
	default:
		return ""
	}
}
