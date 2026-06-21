package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type TransferTemplateListRequest struct {
	model.BaseQuery
	DeptType string `form:"deptType"`
	Status   int32  `form:"status"`
}

type TransferTemplateCreateRequest struct {
	TemplateName    string `json:"templateName" binding:"required"`
	DeptCode        string `json:"deptCode" binding:"required"`
	DeptName        string `json:"deptName" binding:"required"`
	DeptType        string `json:"deptType" binding:"required"`
	ContactPerson   string `json:"contactPerson"`
	ContactPhone    string `json:"contactPhone"`
	ContactEmail    string `json:"contactEmail"`
	Description     string `json:"description"`
	ApplicableTypes string `json:"applicableTypes"`
	SortOrder       int    `json:"sortOrder"`
}

type TransferTemplateUpdateRequest struct {
	TemplateName    string `json:"templateName"`
	DeptName        string `json:"deptName"`
	DeptType        string `json:"deptType"`
	ContactPerson   string `json:"contactPerson"`
	ContactPhone    string `json:"contactPhone"`
	ContactEmail    string `json:"contactEmail"`
	Description     string `json:"description"`
	ApplicableTypes string `json:"applicableTypes"`
	SortOrder       int    `json:"sortOrder"`
	Status          int32  `json:"status"`
}

type TransferCreateRequest struct {
	CaseID         int64   `json:"caseId" binding:"required"`
	TemplateID     int64   `json:"templateId"`
	ToDeptCode     string  `json:"toDeptCode" binding:"required"`
	TransferReason string  `json:"transferReason" binding:"required"`
	TransferRemark string  `json:"transferRemark"`
	AttachIDs      []int64 `json:"attachIds"`
	TimeoutHours   int     `json:"timeoutHours"`
}

type TransferListRequest struct {
	model.BaseQuery
	Status     int32  `form:"status"`
	ToDeptCode string `form:"toDeptCode"`
	FromDeptID int64  `form:"fromDeptId"`
	CaseID     int64  `form:"caseId"`
	IsTimeout  int32  `form:"isTimeout"`
	model.DateRangeQuery
}

type TransferReceiveRequest struct {
	ReceiveRemark string `json:"receiveRemark"`
}

type TransferRejectRequest struct {
	RejectReason string `json:"rejectReason" binding:"required"`
}

type TransferProcessRequest struct {
	ProcessResult string `json:"processResult"`
}

type TransferCompleteRequest struct {
	ProcessResult string `json:"processResult" binding:"required"`
}

type TransferUrgeRequest struct {
	UrgencyLevel int    `json:"urgencyLevel"`
	Content      string `json:"content" binding:"required"`
}

type TransferStatsRequest struct {
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	DeptCode  string `form:"deptCode"`
}

func GetTransferTemplateList(ctx context.Context, c *app.RequestContext) {
	var req TransferTemplateListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("dispute_transfer_template").
		Where("deleted_at IS NULL")

	if req.DeptType != "" {
		db = db.Where("dept_type = ?", req.DeptType)
	}
	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		db = db.Where("template_name LIKE ? OR dept_name LIKE ? OR dept_code LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("sort_order ASC, id DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetTransferTemplateDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var detail map[string]interface{}
	database.GetDB().Table("dispute_transfer_template").
		Where("id = ? AND deleted_at IS NULL", id).
		Find(&detail)

	if detail == nil {
		c.JSON(http.StatusNotFound, response.NotFound("转办模板不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func CreateTransferTemplate(ctx context.Context, c *app.RequestContext) {
	var req TransferTemplateCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var existing int64
	database.GetDB().Table("dispute_transfer_template").
		Where("dept_code = ? AND deleted_at IS NULL", req.DeptCode).
		Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("部门编码已存在"))
		return
	}

	id := utils.GenerateID()
	now := time.Now()

	data := map[string]interface{}{
		"id":               id,
		"template_name":    req.TemplateName,
		"dept_code":        req.DeptCode,
		"dept_name":        req.DeptName,
		"dept_type":        req.DeptType,
		"contact_person":   req.ContactPerson,
		"contact_phone":    req.ContactPhone,
		"contact_email":    req.ContactEmail,
		"description":      req.Description,
		"applicable_types": req.ApplicableTypes,
		"sort_order":       req.SortOrder,
		"status":           1,
		"created_at":       now,
		"updated_at":       now,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_transfer_template").Create(data).Error; err != nil {
		tx.Rollback()
		logger.Error("Create transfer template failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建转办模板失败"))
		return
	}
	tx.Commit()

	logger.Info("Transfer template created",
		logger.Int64("id", id),
		logger.String("deptCode", req.DeptCode),
		logger.String("deptName", req.DeptName),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": id,
	}, "转办模板创建成功"))
}

func UpdateTransferTemplate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferTemplateUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var existing int64
	database.GetDB().Table("dispute_transfer_template").
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&existing)
	if existing == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办模板不存在"))
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.TemplateName != "" {
		updates["template_name"] = req.TemplateName
	}
	if req.DeptName != "" {
		updates["dept_name"] = req.DeptName
	}
	if req.DeptType != "" {
		updates["dept_type"] = req.DeptType
	}
	if req.ContactPerson != "" {
		updates["contact_person"] = req.ContactPerson
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}
	if req.ContactEmail != "" {
		updates["contact_email"] = req.ContactEmail
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ApplicableTypes != "" {
		updates["applicable_types"] = req.ApplicableTypes
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_transfer_template").Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Update transfer template failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("更新转办模板失败"))
		return
	}
	tx.Commit()

	logger.Info("Transfer template updated",
		logger.Int64("id", id),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "转办模板更新成功"))
}

func DeleteTransferTemplate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	userInfo := middleware.GetUserInfo(c)

	var existing int64
	database.GetDB().Table("dispute_transfer_template").
		Where("id = ? AND deleted_at IS NULL", id).
		Count(&existing)
	if existing == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办模板不存在"))
		return
	}

	now := time.Now()
	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_transfer_template").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error; err != nil {
		tx.Rollback()
		logger.Error("Delete transfer template failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("删除转办模板失败"))
		return
	}
	tx.Commit()

	logger.Info("Transfer template deleted",
		logger.Int64("id", id),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "转办模板删除成功"))
}

func GetTransferList(ctx context.Context, c *app.RequestContext) {
	var req TransferListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := userInfo.OrganizationID

	db := database.GetDB().Table("dispute_transfer dt").
		Select("dt.*, so.org_name as from_dept_org_name").
		Joins("LEFT JOIN sys_organization so ON dt.from_dept_id = so.id").
		Where("dt.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("dt.from_user_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("dt.from_dept_id = ? OR dt.from_dept_id IN (SELECT id FROM sys_organization WHERE parent_id = ?)",
			orgID, orgID)
	}

	if req.Status > 0 {
		db = db.Where("dt.status = ?", req.Status)
	}
	if req.ToDeptCode != "" {
		db = db.Where("dt.to_dept_code = ?", req.ToDeptCode)
	}
	if req.FromDeptID > 0 {
		db = db.Where("dt.from_dept_id = ?", req.FromDeptID)
	}
	if req.CaseID > 0 {
		db = db.Where("dt.case_id = ?", req.CaseID)
	}
	if req.IsTimeout > 0 {
		db = db.Where("dt.is_timeout = ?", req.IsTimeout)
	}
	if req.Keyword != "" {
		db = db.Where("dt.transfer_no LIKE ? OR dt.case_no LIKE ? OR dt.case_title LIKE ? OR dt.to_dept_name LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.StartTime != "" {
		db = db.Where("dt.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dt.created_at <= ?", req.EndTime)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order(req.GetSort()).
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.TransferStatusMap[int(status)]
		} else if status, ok := item["status"].(int64); ok {
			item["status_name"] = constants.TransferStatusMap[int(status)]
		}
		if deptType, ok := item["to_dept_type"].(string); ok {
			item["to_dept_type_name"] = constants.TransferDeptTypeMap[deptType]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetTransferDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var detail map[string]interface{}
	database.GetDB().Table("dispute_transfer dt").
		Select("dt.*, so.org_name as from_dept_org_name").
		Joins("LEFT JOIN sys_organization so ON dt.from_dept_id = so.id").
		Where("dt.id = ? AND dt.deleted_at IS NULL", id).
		Find(&detail)

	if detail == nil {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if status, ok := detail["status"].(int32); ok {
		detail["status_name"] = constants.TransferStatusMap[int(status)]
	} else if status, ok := detail["status"].(int64); ok {
		detail["status_name"] = constants.TransferStatusMap[int(status)]
	}
	if deptType, ok := detail["to_dept_type"].(string); ok {
		detail["to_dept_type_name"] = constants.TransferDeptTypeMap[deptType]
	}

	var urgeList []map[string]interface{}
	database.GetDB().Table("dispute_transfer_urge").
		Where("transfer_id = ?", id).
		Order("created_at DESC").
		Find(&urgeList)
	detail["urge_records"] = urgeList

	c.JSON(http.StatusOK, response.Success(detail))
}

func CreateTransfer(ctx context.Context, c *app.RequestContext) {
	var req TransferCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := userInfo.OrganizationID
	if orgID == 0 {
		orgID = 1
	}

	var caseData struct {
		CaseNo  string `gorm:"column:case_no"`
		Title   string `gorm:"column:title"`
		Status  int32  `gorm:"column:status"`
		OrgID   int64  `gorm:"column:organization_id"`
		OrgName string `gorm:"column:org_name"`
	}
	database.GetDB().Table("dispute_case").
		Select("case_no, title, status, organization_id, org_name").
		Where("id = ? AND deleted_at IS NULL", req.CaseID).
		First(&caseData)

	if caseData.CaseNo == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件不存在"))
		return
	}

	if caseData.Status == constants.CaseStatusClosed || caseData.Status == constants.CaseStatusCancelled {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件已结案或已取消，无法转办"))
		return
	}

	var templateData struct {
		ID            int64  `gorm:"column:id"`
		DeptCode      string `gorm:"column:dept_code"`
		DeptName      string `gorm:"column:dept_name"`
		DeptType      string `gorm:"column:dept_type"`
		ContactPerson string `gorm:"column:contact_person"`
		ContactPhone  string `gorm:"column:contact_phone"`
	}
	if req.TemplateID > 0 {
		database.GetDB().Table("dispute_transfer_template").
			Select("id, dept_code, dept_name, dept_type, contact_person, contact_phone").
			Where("id = ? AND status = 1 AND deleted_at IS NULL", req.TemplateID).
			First(&templateData)
		if templateData.DeptCode == "" {
			c.JSON(http.StatusBadRequest, response.BadRequest("转办模板不存在或已禁用"))
			return
		}
	} else {
		database.GetDB().Table("dispute_transfer_template").
			Select("id, dept_code, dept_name, dept_type, contact_person, contact_phone").
			Where("dept_code = ? AND status = 1 AND deleted_at IS NULL", req.ToDeptCode).
			First(&templateData)
		if templateData.DeptCode == "" {
			c.JSON(http.StatusBadRequest, response.BadRequest("目标部门不存在或已禁用"))
			return
		}
	}

	transferID := utils.GenerateID()
	transferNo := utils.GenerateTransferNo()
	now := time.Now()

	timeoutHours := req.TimeoutHours
	if timeoutHours <= 0 {
		timeoutHours = 72
	}

	var attachIDsStr string
	if len(req.AttachIDs) > 0 {
		idStrs := make([]string, len(req.AttachIDs))
		for i, aid := range req.AttachIDs {
			idStrs[i] = strconv.FormatInt(aid, 10)
		}
		attachIDsStr = strings.Join(idStrs, ",")
	}

	transferData := map[string]interface{}{
		"id":                transferID,
		"transfer_no":       transferNo,
		"case_id":           req.CaseID,
		"case_no":           caseData.CaseNo,
		"case_title":        caseData.Title,
		"template_id":       templateData.ID,
		"from_dept_id":      orgID,
		"from_dept_name":    caseData.OrgName,
		"from_user_id":      userInfo.UserID,
		"from_user_name":    userInfo.RealName,
		"to_dept_code":      templateData.DeptCode,
		"to_dept_name":      templateData.DeptName,
		"to_dept_type":      templateData.DeptType,
		"to_contact_person": templateData.ContactPerson,
		"to_contact_phone":  templateData.ContactPhone,
		"transfer_reason":   req.TransferReason,
		"transfer_remark":   req.TransferRemark,
		"attach_ids":        attachIDsStr,
		"status":            constants.TransferStatusPending,
		"timeout_hours":     timeoutHours,
		"is_timeout":        0,
		"urge_count":        0,
		"created_at":        now,
		"updated_at":        now,
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_transfer").Create(transferData).Error; err != nil {
		tx.Rollback()
		logger.Error("Create transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建转办失败"))
		return
	}

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "TRANSFER",
		"operation_detail": fmt.Sprintf("转办至%s，原因: %s", templateData.DeptName, req.TransferReason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"operator_role":    userInfo.Role,
		"old_status":       caseData.Status,
		"new_status":       caseData.Status,
		"remark":           req.TransferReason,
		"created_at":       now,
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		logger.Error("Create case history failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建案件历史失败"))
		return
	}

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, req.CaseID)
	cache.Del(ctx, cacheKey)

	go func() {
		msg := map[string]interface{}{
			"transferId":     transferID,
			"transferNo":     transferNo,
			"caseId":         req.CaseID,
			"caseNo":         caseData.CaseNo,
			"caseTitle":      caseData.Title,
			"fromDeptId":     orgID,
			"fromDeptName":   caseData.OrgName,
			"fromUserId":     userInfo.UserID,
			"fromUserName":   userInfo.RealName,
			"toDeptCode":     templateData.DeptCode,
			"toDeptName":     templateData.DeptName,
			"toDeptType":     templateData.DeptType,
			"toContactPhone": templateData.ContactPhone,
			"transferReason": req.TransferReason,
			"timeoutHours":   timeoutHours,
			"createTime":     now.Format(time.RFC3339),
		}
		mq.SendMessage(constants.MQTopicTransferCreate, msg)
	}()

	logger.Info("Transfer created",
		logger.Int64("transferId", transferID),
		logger.String("transferNo", transferNo),
		logger.Int64("caseId", req.CaseID),
		logger.String("toDept", templateData.DeptName),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":         transferID,
		"transferNo": transferNo,
	}, "转办创建成功"))
}

func ReceiveTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferReceiveRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var transferData struct {
		ID           int64  `gorm:"column:id"`
		TransferNo   string `gorm:"column:transfer_no"`
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		CaseTitle    string `gorm:"column:case_title"`
		Status       int32  `gorm:"column:status"`
		ToDeptName   string `gorm:"column:to_dept_name"`
		FromUserName string `gorm:"column:from_user_name"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, case_title, status, to_dept_name, from_user_name").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status != constants.TransferStatusPending {
		c.JSON(http.StatusBadRequest, response.BadRequest("当前状态不允许接收"))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":            constants.TransferStatusReceived,
		"receive_time":      now,
		"receive_user_id":   userInfo.UserID,
		"receive_user_name": userInfo.RealName,
		"receive_remark":    req.ReceiveRemark,
		"process_start_time": now,
		"updated_at":        now,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Receive transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("接收转办失败"))
		return
	}
	tx.Commit()

	go func() {
		msg := map[string]interface{}{
			"transferId":   id,
			"transferNo":   transferData.TransferNo,
			"caseId":       transferData.CaseID,
			"caseNo":       transferData.CaseNo,
			"caseTitle":    transferData.CaseTitle,
			"toDeptName":   transferData.ToDeptName,
			"receiveUserId": userInfo.UserID,
			"receiveUserName": userInfo.RealName,
			"receiveTime":  now.Format(time.RFC3339),
			"receiveRemark": req.ReceiveRemark,
			"fromUserName": transferData.FromUserName,
		}
		mq.SendMessage(constants.MQTopicTransferReceive, msg)
	}()

	logger.Info("Transfer received",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "接收成功"))
}

func RejectTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferRejectRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var transferData struct {
		ID           int64  `gorm:"column:id"`
		TransferNo   string `gorm:"column:transfer_no"`
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		Status       int32  `gorm:"column:status"`
		ToDeptName   string `gorm:"column:to_dept_name"`
		FromDeptID   int64  `gorm:"column:from_dept_id"`
		FromUserName string `gorm:"column:from_user_name"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, status, to_dept_name, from_dept_id, from_user_name").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status != constants.TransferStatusPending {
		c.JSON(http.StatusBadRequest, response.BadRequest("当前状态不允许驳回"))
		return
	}

	now := time.Now()
	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(map[string]interface{}{
		"status":        constants.TransferStatusRejected,
		"reject_reason": req.RejectReason,
		"reject_time":   now,
		"closed_at":     now,
		"updated_at":    now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Reject transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("驳回转办失败"))
		return
	}

	history := map[string]interface{}{
		"case_id":          transferData.CaseID,
		"case_no":          transferData.CaseNo,
		"operation_type":   "TRANSFER_REJECT",
		"operation_detail": fmt.Sprintf("%s驳回转办，原因: %s", transferData.ToDeptName, req.RejectReason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"remark":           req.RejectReason,
		"created_at":       now,
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		logger.Error("Create case history failed", logger.Error(err))
	}

	tx.Commit()

	logger.Info("Transfer rejected",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "驳回成功"))
}

func StartProcessTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferProcessRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var transferData struct {
		ID         int64  `gorm:"column:id"`
		TransferNo string `gorm:"column:transfer_no"`
		CaseID     int64  `gorm:"column:case_id"`
		CaseNo     string `gorm:"column:case_no"`
		Status     int32  `gorm:"column:status"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, status").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status != constants.TransferStatusReceived {
		c.JSON(http.StatusBadRequest, response.BadRequest("当前状态不允许开始处理"))
		return
	}

	now := time.Now()
	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(map[string]interface{}{
		"status":             constants.TransferStatusProcessing,
		"process_start_time": now,
		"process_result":     req.ProcessResult,
		"updated_at":         now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Start process transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("开始处理失败"))
		return
	}
	tx.Commit()

	logger.Info("Transfer processing started",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已开始处理"))
}

func CompleteTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferCompleteRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var transferData struct {
		ID               int64      `gorm:"column:id"`
		TransferNo       string     `gorm:"column:transfer_no"`
		CaseID           int64      `gorm:"column:case_id"`
		CaseNo           string     `gorm:"column:case_no"`
		CaseTitle        string     `gorm:"column:case_title"`
		Status           int32      `gorm:"column:status"`
		ToDeptCode       string     `gorm:"column:to_dept_code"`
		ToDeptName       string     `gorm:"column:to_dept_name"`
		ToDeptType       string     `gorm:"column:to_dept_type"`
		ProcessStartTime *time.Time `gorm:"column:process_start_time"`
		FromDeptID       int64      `gorm:"column:from_dept_id"`
		FromDeptName     string     `gorm:"column:from_dept_name"`
		FromUserID       int64      `gorm:"column:from_user_id"`
		FromUserName     string     `gorm:"column:from_user_name"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, case_title, status, to_dept_code, to_dept_name, to_dept_type, process_start_time, from_dept_id, from_dept_name, from_user_id, from_user_name").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status != constants.TransferStatusReceived && transferData.Status != constants.TransferStatusProcessing {
		c.JSON(http.StatusBadRequest, response.BadRequest("当前状态不允许办结"))
		return
	}

	now := time.Now()
	var processDuration int
	if transferData.ProcessStartTime != nil {
		processDuration = int(now.Sub(*transferData.ProcessStartTime).Hours())
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(map[string]interface{}{
		"status":           constants.TransferStatusCompleted,
		"process_end_time": now,
		"process_result":   req.ProcessResult,
		"process_duration": processDuration,
		"closed_at":        now,
		"updated_at":       now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Complete transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("办结转办失败"))
		return
	}

	history := map[string]interface{}{
		"case_id":          transferData.CaseID,
		"case_no":          transferData.CaseNo,
		"operation_type":   "TRANSFER_COMPLETE",
		"operation_detail": fmt.Sprintf("%s办结转办，结果: %s", transferData.ToDeptName, req.ProcessResult),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"remark":           req.ProcessResult,
		"created_at":       now,
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		logger.Error("Create case history failed", logger.Error(err))
	}

	tx.Commit()

	go func() {
		msg := map[string]interface{}{
			"transferId":     id,
			"transferNo":     transferData.TransferNo,
			"caseId":         transferData.CaseID,
			"caseNo":         transferData.CaseNo,
			"caseTitle":      transferData.CaseTitle,
			"toDeptCode":     transferData.ToDeptCode,
			"toDeptName":     transferData.ToDeptName,
			"toDeptType":     transferData.ToDeptType,
			"fromDeptId":     transferData.FromDeptID,
			"fromDeptName":   transferData.FromDeptName,
			"fromUserId":     transferData.FromUserID,
			"fromUserName":   transferData.FromUserName,
			"processResult":  req.ProcessResult,
			"processDuration": processDuration,
			"completeTime":   now.Format(time.RFC3339),
		}
		mq.SendMessage(constants.MQTopicTransferComplete, msg)
	}()

	logger.Info("Transfer completed",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int("duration", processDuration),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "办结成功"))
}

func UrgeTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req TransferUrgeRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var transferData struct {
		ID            int64      `gorm:"column:id"`
		TransferNo    string     `gorm:"column:transfer_no"`
		CaseID        int64      `gorm:"column:case_id"`
		CaseNo        string     `gorm:"column:case_no"`
		CaseTitle     string     `gorm:"column:case_title"`
		Status        int32      `gorm:"column:status"`
		ToDeptCode    string     `gorm:"column:to_dept_code"`
		ToDeptName    string     `gorm:"column:to_dept_name"`
		UrgeCount     int        `gorm:"column:urge_count"`
		FirstUrgeTime *time.Time `gorm:"column:first_urge_time"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, case_title, status, to_dept_code, to_dept_name, urge_count, first_urge_time").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status == constants.TransferStatusCompleted ||
		transferData.Status == constants.TransferStatusRejected ||
		transferData.Status == constants.TransferStatusCancelled {
		c.JSON(http.StatusBadRequest, response.BadRequest("当前状态不允许催办"))
		return
	}

	urgeType := constants.TransferUrgeTypeUser
	if userInfo.Role == constants.RoleLeader || userInfo.Role == constants.RoleDirector {
		urgeType = constants.TransferUrgeTypeLeader
	}
	if req.UrgencyLevel <= 0 {
		req.UrgencyLevel = 2
	}

	now := time.Now()
	firstUrgeTime := transferData.FirstUrgeTime
	if firstUrgeTime == nil {
		firstUrgeTime = &now
	}

	tx := database.GetDB().Begin()

	urgeRecord := map[string]interface{}{
		"transfer_id":   id,
		"transfer_no":   transferData.TransferNo,
		"urge_type":     urgeType,
		"urge_source":   constants.TransferUrgeSourceManual,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.RealName,
		"urgency_level": req.UrgencyLevel,
		"urge_content":  req.Content,
		"notify_type":   "app,sms",
		"created_at":    now,
	}
	if err := tx.Table("dispute_transfer_urge").Create(urgeRecord).Error; err != nil {
		tx.Rollback()
		logger.Error("Create transfer urge record failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建催办记录失败"))
		return
	}

	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(map[string]interface{}{
		"urge_count":      transferData.UrgeCount + 1,
		"last_urge_time":  now,
		"first_urge_time": firstUrgeTime,
		"updated_at":      now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Update transfer urge info failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("更新催办信息失败"))
		return
	}

	tx.Commit()

	go func() {
		msg := map[string]interface{}{
			"transferId":   id,
			"transferNo":   transferData.TransferNo,
			"caseId":       transferData.CaseID,
			"caseNo":       transferData.CaseNo,
			"caseTitle":    transferData.CaseTitle,
			"toDeptCode":   transferData.ToDeptCode,
			"toDeptName":   transferData.ToDeptName,
			"urgeType":     urgeType,
			"urgencyLevel": req.UrgencyLevel,
			"urgeContent":  req.Content,
			"urgeCount":    transferData.UrgeCount + 1,
			"urgeBy":       userInfo.RealName,
			"urgeTime":     now.Format(time.RFC3339),
		}
		mq.SendMessage(constants.MQTopicTransferUrge, msg)
	}()

	logger.Info("Transfer urged",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int("urgeCount", transferData.UrgeCount+1),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "催办成功"))
}

func CancelTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	userInfo := middleware.GetUserInfo(c)

	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var transferData struct {
		ID           int64  `gorm:"column:id"`
		TransferNo   string `gorm:"column:transfer_no"`
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		Status       int32  `gorm:"column:status"`
		FromUserID   int64  `gorm:"column:from_user_id"`
		ToDeptName   string `gorm:"column:to_dept_name"`
	}
	database.GetDB().Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, status, from_user_id, to_dept_name").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&transferData)

	if transferData.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("转办记录不存在"))
		return
	}

	if transferData.Status != constants.TransferStatusPending {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有待接收状态可以取消"))
		return
	}

	if transferData.FromUserID != userInfo.UserID &&
		userInfo.Role != constants.RoleLeader &&
		userInfo.Role != constants.RoleDirector &&
		userInfo.Role != constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限取消此转办"))
		return
	}

	now := time.Now()
	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_transfer").Where("id = ?", id).Updates(map[string]interface{}{
		"status":     constants.TransferStatusCancelled,
		"closed_at":  now,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Cancel transfer failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("取消转办失败"))
		return
	}

	history := map[string]interface{}{
		"case_id":          transferData.CaseID,
		"case_no":          transferData.CaseNo,
		"operation_type":   "TRANSFER_CANCEL",
		"operation_detail": fmt.Sprintf("取消转办至%s，备注: %s", transferData.ToDeptName, req.Remark),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"remark":           req.Remark,
		"created_at":       now,
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		logger.Error("Create case history failed", logger.Error(err))
	}

	tx.Commit()

	logger.Info("Transfer cancelled",
		logger.Int64("transferId", id),
		logger.String("transferNo", transferData.TransferNo),
		logger.Int64("userId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "取消成功"))
}

func GetTransferUrgeList(ctx context.Context, c *app.RequestContext) {
	transferID, _ := strconv.ParseInt(c.Param("transferId"), 10, 64)

	var list []map[string]interface{}
	database.GetDB().Table("dispute_transfer_urge").
		Where("transfer_id = ?", transferID).
		Order("created_at DESC").
		Find(&list)

	c.JSON(http.StatusOK, response.Success(list))
}

func GetTransferDeptStats(ctx context.Context, c *app.RequestContext) {
	var req TransferStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	_ = userInfo

	db := database.GetDB().Table("dispute_transfer dt").
		Select("dt.to_dept_code, dt.to_dept_name, dt.to_dept_type, "+
			"COUNT(*) as total_count, "+
			"SUM(CASE WHEN dt.status = 10 THEN 1 ELSE 0 END) as pending_count, "+
			"SUM(CASE WHEN dt.status = 20 THEN 1 ELSE 0 END) as received_count, "+
			"SUM(CASE WHEN dt.status = 30 THEN 1 ELSE 0 END) as processing_count, "+
			"SUM(CASE WHEN dt.status = 40 THEN 1 ELSE 0 END) as completed_count, "+
			"SUM(CASE WHEN dt.status = 50 THEN 1 ELSE 0 END) as rejected_count, "+
			"SUM(CASE WHEN dt.is_timeout = 1 THEN 1 ELSE 0 END) as timeout_count, "+
			"AVG(CASE WHEN dt.status = 40 AND dt.process_duration > 0 THEN dt.process_duration ELSE NULL END) as avg_duration, "+
			"SUM(dt.urge_count) as total_urge_count").
		Where("dt.deleted_at IS NULL")

	if req.StartTime != "" {
		db = db.Where("dt.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dt.created_at <= ?", req.EndTime)
	}
	if req.DeptCode != "" {
		db = db.Where("dt.to_dept_code = ?", req.DeptCode)
	}

	var list []map[string]interface{}
	db.Group("dt.to_dept_code, dt.to_dept_name, dt.to_dept_type").
		Order("total_count DESC").
		Find(&list)

	for _, item := range list {
		if deptType, ok := item["to_dept_type"].(string); ok {
			item["to_dept_type_name"] = constants.TransferDeptTypeMap[deptType]
		}
		totalCount, _ := item["total_count"].(int64)
		completedCount, _ := item["completed_count"].(int64)
		timeoutCount, _ := item["timeout_count"].(int64)
		if totalCount > 0 {
			item["complete_rate"] = fmt.Sprintf("%.1f%%", float64(completedCount)/float64(totalCount)*100)
			item["timeout_rate"] = fmt.Sprintf("%.1f%%", float64(timeoutCount)/float64(totalCount)*100)
		} else {
			item["complete_rate"] = "0%"
			item["timeout_rate"] = "0%"
		}
		if avgDur, ok := item["avg_duration"]; ok && avgDur != nil {
			item["avg_duration"] = fmt.Sprintf("%.1f", avgDur)
		} else {
			item["avg_duration"] = "0"
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetTransferDurationRanking(ctx context.Context, c *app.RequestContext) {
	var req TransferStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	db := database.GetDB().Table("dispute_transfer dt").
		Select("dt.id, dt.transfer_no, dt.case_id, dt.case_no, dt.case_title, "+
			"dt.to_dept_code, dt.to_dept_name, dt.to_dept_type, "+
			"dt.process_duration, dt.created_at, dt.closed_at, dt.urge_count").
		Where("dt.status = ? AND dt.process_duration > 0 AND dt.deleted_at IS NULL", constants.TransferStatusCompleted)

	if req.StartTime != "" {
		db = db.Where("dt.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dt.created_at <= ?", req.EndTime)
	}

	sortType := c.DefaultQuery("sort", "duration")
	sortOrder := "process_duration DESC"
	if sortType == "urge" {
		sortOrder = "urge_count DESC"
	} else if sortType == "fastest" {
		sortOrder = "process_duration ASC"
	}

	var list []map[string]interface{}
	db.Order(sortOrder).
		Limit(limit).
		Find(&list)

	for i, item := range list {
		item["rank"] = i + 1
		if deptType, ok := item["to_dept_type"].(string); ok {
			item["to_dept_type_name"] = constants.TransferDeptTypeMap[deptType]
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":  list,
		"total": len(list),
	}))
}

func GetTransferTrendStats(ctx context.Context, c *app.RequestContext) {
	var req TransferStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -29)
	if req.StartTime != "" {
		if t, err := time.Parse("2006-01-02", req.StartTime); err == nil {
			startTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse("2006-01-02", req.EndTime); err == nil {
			endTime = t.AddDate(0, 0, 1)
		}
	}

	days := int(endTime.Sub(startTime).Hours() / 24)
	if days <= 0 || days > 365 {
		days = 30
	}

	trendData := make([]map[string]interface{}, 0, days)

	for i := 0; i < days; i++ {
		dayStart := startTime.AddDate(0, 0, i).Format("2006-01-02")
		dayEnd := startTime.AddDate(0, 0, i+1).Format("2006-01-02")

		var dayStats struct {
			NewCount       int64 `gorm:"column:new_count"`
			CompletedCount int64 `gorm:"column:completed_count"`
			TimeoutCount   int64 `gorm:"column:timeout_count"`
		}

		database.GetDB().Table("dispute_transfer").
			Select(`COUNT(*) as new_count,
				SUM(CASE WHEN status = 40 AND closed_at >= ? AND closed_at < ? THEN 1 ELSE 0 END) as completed_count,
				SUM(CASE WHEN is_timeout = 1 AND created_at >= ? AND created_at < ? THEN 1 ELSE 0 END) as timeout_count`,
				dayStart, dayEnd, dayStart, dayEnd).
			Where("created_at >= ? AND created_at < ? AND deleted_at IS NULL", dayStart, dayEnd).
			Scan(&dayStats)

		trendData = append(trendData, map[string]interface{}{
			"date":            dayStart,
			"newCount":        dayStats.NewCount,
			"completedCount":  dayStats.CompletedCount,
			"timeoutCount":    dayStats.TimeoutCount,
		})
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"trendData": trendData,
		"days":      days,
	}))
}

func ProcessTransferTimeoutCheck() int {
	ctx := context.Background()
	now := time.Now()

	db := database.GetDB()

	type TimeoutTransfer struct {
		ID             int64  `gorm:"column:id"`
		TransferNo     string `gorm:"column:transfer_no"`
		CaseID         int64  `gorm:"column:case_id"`
		CaseNo         string `gorm:"column:case_no"`
		CaseTitle      string `gorm:"column:case_title"`
		ToDeptCode     string `gorm:"column:to_dept_code"`
		ToDeptName     string `gorm:"column:to_dept_name"`
		ToContactPhone string `gorm:"column:to_contact_phone"`
		FromDeptID     int64  `gorm:"column:from_dept_id"`
		FromDeptName   string `gorm:"column:from_dept_name"`
		FromUserID     int64  `gorm:"column:from_user_id"`
		FromUserName   string `gorm:"column:from_user_name"`
		TimeoutHours   int    `gorm:"column:timeout_hours"`
		UrgeCount      int    `gorm:"column:urge_count"`
		IsTimeout      int32  `gorm:"column:is_timeout"`
		CreatedAt      time.Time `gorm:"column:created_at"`
		Status         int32  `gorm:"column:status"`
	}

	var transfers []TimeoutTransfer
	err := db.Table("dispute_transfer").
		Select("id, transfer_no, case_id, case_no, case_title, to_dept_code, to_dept_name, to_contact_phone, from_dept_id, from_dept_name, from_user_id, from_user_name, timeout_hours, urge_count, is_timeout, created_at, status").
		Where("status IN (?, ?) AND deleted_at IS NULL",
			constants.TransferStatusPending, constants.TransferStatusReceived).
		Scan(&transfers).Error
	if err != nil {
		logger.Error("Query timeout transfers failed", logger.Error(err))
		return 0
	}

	processedCount := 0

	for _, t := range transfers {
		elapsed := now.Sub(t.CreatedAt)
		timeoutDuration := time.Duration(t.TimeoutHours) * time.Hour

		if elapsed <= timeoutDuration {
			continue
		}

		overdueHours := int(elapsed.Hours())

		if t.IsTimeout == 0 {
			updates := map[string]interface{}{
				"is_timeout": 1,
				"updated_at": now,
			}
			if t.UrgeCount == 0 {
				updates["first_urge_time"] = now
			}
			updates["urge_count"] = t.UrgeCount + 1
			updates["last_urge_time"] = now

			tx := db.Begin()
			if err := tx.Table("dispute_transfer").Where("id = ?", t.ID).Updates(updates).Error; err != nil {
				tx.Rollback()
				logger.Warn("Update transfer timeout flag failed",
					logger.Int64("transferId", t.ID),
					logger.Error(err))
				continue
			}

			urgeRecord := map[string]interface{}{
				"transfer_id":   t.ID,
				"transfer_no":   t.TransferNo,
				"urge_type":     constants.TransferUrgeTypeSystem,
				"urge_source":   constants.TransferUrgeSourceSystem,
				"operator_id":   0,
				"operator_name": "系统自动",
				"urgency_level": 2,
				"urge_content":  fmt.Sprintf("转办超时提醒：已超时%d小时，请尽快处理", overdueHours),
				"notify_type":   "app,sms",
				"created_at":    now,
			}
			if err := tx.Table("dispute_transfer_urge").Create(urgeRecord).Error; err != nil {
				tx.Rollback()
				logger.Warn("Create timeout urge record failed",
					logger.Int64("transferId", t.ID),
					logger.Error(err))
				continue
			}
			tx.Commit()

			go func(t TimeoutTransfer, overdueHours int) {
				msg := map[string]interface{}{
					"transferId":     t.ID,
					"transferNo":     t.TransferNo,
					"caseId":         t.CaseID,
					"caseNo":         t.CaseNo,
					"caseTitle":      t.CaseTitle,
					"toDeptCode":     t.ToDeptCode,
					"toDeptName":     t.ToDeptName,
					"toContactPhone": t.ToContactPhone,
					"fromDeptId":     t.FromDeptID,
					"fromDeptName":   t.FromDeptName,
					"fromUserId":     t.FromUserID,
					"fromUserName":   t.FromUserName,
					"overdueHours":   overdueHours,
					"timeoutHours":   t.TimeoutHours,
					"urgeCount":      t.UrgeCount + 1,
					"urgeTime":       now.Format(time.RFC3339),
				}
				mq.SendMessage(constants.MQTopicTransferTimeout, msg)
			}(t, overdueHours)

			processedCount++
			logger.Info("Transfer timeout detected and urged",
				logger.Int64("transferId", t.ID),
				logger.String("transferNo", t.TransferNo),
				logger.Int("overdueHours", overdueHours))
		}
	}

	return processedCount
}

func GetCaseTransferList(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var list []map[string]interface{}
	database.GetDB().Table("dispute_transfer").
		Where("case_id = ? AND deleted_at IS NULL", caseID).
		Order("created_at DESC").
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.TransferStatusMap[int(status)]
		} else if status, ok := item["status"].(int64); ok {
			item["status_name"] = constants.TransferStatusMap[int(status)]
		}
		if deptType, ok := item["to_dept_type"].(string); ok {
			item["to_dept_type_name"] = constants.TransferDeptTypeMap[deptType]
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}
