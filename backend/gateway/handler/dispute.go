package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

type DisputeCreateRequest struct {
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	TypeID         int64   `json:"typeId" binding:"required"`
	CaseLevel      int32   `json:"caseLevel"`
	CaseSource     int32   `json:"caseSource"`
	ReporterName   string  `json:"reporterName"`
	ReporterPhone  string  `json:"reporterPhone"`
	ReporterIDCard string  `json:"reporterIdCard"`
	ReporterAddress string `json:"reporterAddress"`
	RespondentName string  `json:"respondentName"`
	RespondentPhone string `json:"respondentPhone"`
	RespondentAddress string `json:"respondentAddress"`
	OccurAddress   string  `json:"occurAddress"`
	OccurTime      string  `json:"occurTime"`
	Expectation    string  `json:"expectation"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	EvidenceIDs    []int64 `json:"evidenceIds"`
}

type DisputeListRequest struct {
	common.BaseQuery
	Status       int32  `form:"status"`
	CaseLevel    int32  `form:"caseLevel"`
	CaseSource   int32  `form:"caseSource"`
	TypeID       int64  `form:"typeId"`
	MediatorID   int64  `form:"mediatorId"`
	OrganizationID int64 `form:"organizationId"`
	common.DateRangeQuery
}

type DisputeAssignRequest struct {
	MediatorID int64  `json:"mediatorId" binding:"required"`
	Reason     string `json:"reason"`
}

func GetDisputeList(ctx context.Context, c *app.RequestContext) {
	var req DisputeListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	db := database.GetDB().Table("dispute_case dc").
		Select("dc.*, dt.type_name, dt.level_path as type_path, su.real_name as mediator_name, so.org_name").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Joins("LEFT JOIN sys_user su ON dc.mediator_id = su.id").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("dc.mediator_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("dc.organization_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)", 
			userInfo.OrganizationID, userInfo.OrganizationID)
	}

	if req.Status > 0 {
		db = db.Where("dc.status = ?", req.Status)
	}
	if req.CaseLevel > 0 {
		db = db.Where("dc.case_level = ?", req.CaseLevel)
	}
	if req.TypeID > 0 {
		db = db.Where("dc.type_id = ?", req.TypeID)
	}
	if req.MediatorID > 0 {
		db = db.Where("dc.mediator_id = ?", req.MediatorID)
	}
	if req.Keyword != "" {
		db = db.Where("dc.title LIKE ? OR dc.description LIKE ? OR dc.case_no LIKE ?", 
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.StartTime != "" {
		db = db.Where("dc.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dc.created_at <= ?", req.EndTime)
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
			item["status_name"] = constants.CaseStatusMap[int(status)]
		}
		if level, ok := item["case_level"].(int32); ok {
			item["case_level_name"] = constants.CaseLevelMap[int(level)]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetDisputeDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, id)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	var caseData map[string]interface{}
	database.GetDB().Table("dispute_case dc").
		Select("dc.*, dt.type_name, dt.level_path as type_path, su.real_name as mediator_name, so.org_name").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Joins("LEFT JOIN sys_user su ON dc.mediator_id = su.id").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.id = ? AND dc.deleted_at IS NULL", id).
		Find(&caseData)

	if caseData == nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	if status, ok := caseData["status"].(int32); ok {
		caseData["status_name"] = constants.CaseStatusMap[int(status)]
	}
	if level, ok := caseData["case_level"].(int32); ok {
		caseData["case_level_name"] = constants.CaseLevelMap[int(level)]
	}

	var evidence []map[string]interface{}
	database.GetDB().Table("dispute_evidence").
		Where("case_id = ?", id).
		Order("sort_order ASC, id DESC").
		Find(&evidence)
	caseData["evidence"] = evidence

	var history []map[string]interface{}
	database.GetDB().Table("dispute_case_history").
		Where("case_id = ?", id).
		Order("created_at DESC").
		Find(&history)
	caseData["history"] = history

	jsonData, _ := json.Marshal(caseData)
	cache.Set(ctx, cacheKey, string(jsonData), time.Duration(constants.RedisExpireCase)*time.Second)

	c.JSON(http.StatusOK, response.Success(caseData))
}

func CreateDispute(ctx context.Context, c *app.RequestContext) {
	var req DisputeCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := userInfo.OrganizationID
	if orgID == 0 {
		orgID = 1
	}

	var orgCode string
	database.GetDB().Table("sys_organization").
		Select("org_code").
		Where("id = ?", orgID).
		Scan(&orgCode)

	caseNo := utils.GenerateCaseNo(orgCode)
	caseID := utils.GenerateID()

	var typePath string
	database.GetDB().Table("dispute_type").
		Select("level_path, type_name").
		Where("id = ?", req.TypeID).
		Scan(&typePath)

	caseData := map[string]interface{}{
		"id":                caseID,
		"case_no":           caseNo,
		"title":             req.Title,
		"description":       req.Description,
		"type_id":           req.TypeID,
		"type_path":         typePath,
		"case_level":        req.CaseLevel,
		"case_source":       req.CaseSource,
		"status":            constants.CaseStatusPending,
		"reporter_name":     req.ReporterName,
		"reporter_phone":    req.ReporterPhone,
		"reporter_id_card":  req.ReporterIDCard,
		"reporter_address":  req.ReporterAddress,
		"respondent_name":   req.RespondentName,
		"respondent_phone":  req.RespondentPhone,
		"respondent_address": req.RespondentAddress,
		"occur_address":     req.OccurAddress,
		"expectation":       req.Expectation,
		"longitude":         req.Longitude,
		"latitude":          req.Latitude,
		"organization_id":   orgID,
		"created_by":        userInfo.UserID,
		"deadline":          time.Now().AddDate(0, 0, 15),
	}

	if req.OccurTime != "" {
		caseData["occur_time"] = req.OccurTime
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_case").Create(caseData).Error; err != nil {
		tx.Rollback()
		logger.Error("Create dispute case failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建案件失败"))
		return
	}

	if len(req.EvidenceIDs) > 0 {
		tx.Table("dispute_evidence").
			Where("id IN ?", req.EvidenceIDs).
			Updates(map[string]interface{}{
				"case_id": caseID,
				"case_no": caseNo,
			})
	}

	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          caseNo,
		"operation_type":   "CREATE",
		"operation_detail": "创建纠纷案件",
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"new_status":       constants.CaseStatusPending,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	go func() {
		msg := map[string]interface{}{
			"caseId":   caseID,
			"caseNo":   caseNo,
			"title":    req.Title,
			"source":   req.CaseSource,
			"orgId":    orgID,
			"createBy": userInfo.UserID,
		}
		mq.SendMessage(constants.MQTopicCaseCreate, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":     caseID,
		"caseNo": caseNo,
	}, "案件创建成功"))
}

func KioskCreateDispute(ctx context.Context, c *app.RequestContext) {
	var req DisputeCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	req.CaseSource = constants.CaseSourceKiosk

	deviceCode := c.GetHeader("X-Device-Code")
	req.Description += fmt.Sprintf("\n[自助终端登记，终端号: %s]", deviceCode)

	userInfo := &auth.UserInfo{
		UserID:         0,
		Username:       "kiosk",
		RealName:       "自助终端",
		Role:           99,
		OrganizationID: 1,
	}
	c.Set("userInfo", userInfo)

	ctx = context.WithValue(ctx, middleware.UserInfoKey, userInfo)

	CreateDispute(ctx, c)
}

func MiniAppCreateDispute(ctx context.Context, c *app.RequestContext) {
	var req DisputeCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	req.CaseSource = constants.CaseSourceMiniApp
	CreateDispute(ctx, c)
}

func AssignDispute(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req DisputeAssignRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		Status    int32  `gorm:"column:status"`
		CaseNo    string `gorm:"column:case_no"`
		Title     string `gorm:"column:title"`
		MediatorID int64 `gorm:"column:mediator_id"`
	}

	database.GetDB().Table("dispute_case").
		Select("status, case_no, title, mediator_id").
		Where("id = ?", id).
		First(&caseData)

	if caseData.Status != constants.CaseStatusPending {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有待分派状态的案件才能分派"))
		return
	}

	var mediator struct {
		RealName       string `gorm:"column:real_name"`
		Phone          string `gorm:"column:phone"`
		OrganizationID int64  `gorm:"column:organization_id"`
		Role           int32  `gorm:"column:role"`
		Status         int32  `gorm:"column:status"`
	}

	result := database.GetDB().Table("sys_user").
		Where("id = ? AND status = 1", req.MediatorID).
		First(&mediator)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("调解员不存在或已禁用"))
		return
	}

	if mediator.Role != constants.RoleMediator {
		c.JSON(http.StatusBadRequest, response.BadRequest("该用户不是调解员角色"))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"mediator_id":   req.MediatorID,
		"mediator_name": mediator.RealName,
		"mediator_time": now,
		"status":        constants.CaseStatusMediating,
		"mediation_start_time": now,
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_case").Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ServerError("分派失败"))
		return
	}

	history := map[string]interface{}{
		"case_id":          id,
		"case_no":          caseData.CaseNo,
		"operation_type":   "ASSIGN",
		"operation_detail": fmt.Sprintf("分派给调解员: %s，原因: %s", mediator.RealName, req.Reason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"operator_role":    userInfo.Role,
		"old_status":       constants.CaseStatusPending,
		"new_status":       constants.CaseStatusMediating,
		"remark":           req.Reason,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, id)
	cache.Del(ctx, cacheKey)

	go func() {
		msg := map[string]interface{}{
			"caseId":         id,
			"caseNo":         caseData.CaseNo,
			"title":          caseData.Title,
			"mediatorId":     req.MediatorID,
			"mediatorName":   mediator.RealName,
			"mediatorPhone":  mediator.Phone,
			"assignBy":       userInfo.UserID,
			"assignByName":   userInfo.RealName,
			"assignTime":     now.Format(time.RFC3339),
		}
		mq.SendMessage(constants.MQTopicCaseAssign, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "分派成功"))
}

func UrgeDispute(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		UrgencyLevel int32  `json:"urgencyLevel"`
		Content      string `json:"content"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo       string `gorm:"column:case_no"`
		Title        string `gorm:"column:title"`
		Status       int32  `gorm:"column:status"`
		MediatorID   int64  `gorm:"column:mediator_id"`
		MediatorName string `gorm:"column:mediator_name"`
		UrgencyCount int32  `gorm:"column:urgency_count"`
	}

	database.GetDB().Table("dispute_case").
		Where("id = ?", id).
		First(&caseData)

	if caseData.Status >= constants.CaseStatusClosed {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件已结案，无法催办"))
		return
	}

	if caseData.MediatorID == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件尚未分派调解员"))
		return
	}

	urgeType := constants.UrgeTypeUser
	if userInfo.Role == constants.RoleLeader || userInfo.Role == constants.RoleDirector {
		urgeType = constants.UrgeTypeLeader
	}

	tx := database.GetDB().Begin()

	urge := map[string]interface{}{
		"case_id":              id,
		"case_no":              caseData.CaseNo,
		"urge_type":            urgeType,
		"urge_source":          2,
		"operator_id":          userInfo.UserID,
		"operator_name":        userInfo.RealName,
		"current_handler_id":   caseData.MediatorID,
		"current_handler_name": caseData.MediatorName,
		"urgency_level":        req.UrgencyLevel,
		"urge_content":         req.Content,
		"notify_type":          "app,sms",
	}
	tx.Table("workflow_urge").Create(urge)

	tx.Table("dispute_case").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"urgency_time":  time.Now(),
			"urgency_count": caseData.UrgencyCount + 1,
		})

	history := map[string]interface{}{
		"case_id":          id,
		"case_no":          caseData.CaseNo,
		"operation_type":   "URGE",
		"operation_detail": fmt.Sprintf("催办案件，级别: %d，内容: %s", req.UrgencyLevel, req.Content),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, id)
	cache.Del(ctx, cacheKey)

	go func() {
		msg := map[string]interface{}{
			"caseId":         id,
			"caseNo":         caseData.CaseNo,
			"title":          caseData.Title,
			"urgencyLevel":   req.UrgencyLevel,
			"urgencyContent": req.Content,
			"handlerId":      caseData.MediatorID,
			"handlerName":    caseData.MediatorName,
			"urgeType":       urgeType,
			"urgeBy":         userInfo.RealName,
		}
		mq.SendMessage(constants.MQTopicCaseUrge, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "催办成功"))
}

func GetDisputeProgress(ctx context.Context, c *app.RequestContext) {
	caseNo := c.Query("caseNo")
	phone := c.Query("phone")

	if caseNo == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("请输入案件编号"))
		return
	}

	var caseData map[string]interface{}
	database.GetDB().Table("dispute_case dc").
		Select("dc.case_no, dc.title, dc.status, dt.type_name, dc.created_at, dc.mediator_name, dc.satisfaction_score").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Where("dc.case_no = ?", caseNo).
		Find(&caseData)

	if caseData == nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	if phone != "" {
		var reporterPhone string
		database.GetDB().Table("dispute_case").
			Select("reporter_phone").
			Where("case_no = ?", caseNo).
			Scan(&reporterPhone)
		if reporterPhone != phone {
			c.JSON(http.StatusForbidden, response.Forbidden("手机号不匹配"))
			return
		}
	}

	if status, ok := caseData["status"].(int32); ok {
		caseData["status_name"] = constants.CaseStatusMap[int(status)]
	}

	var history []map[string]interface{}
	database.GetDB().Table("dispute_case_history").
		Select("operation_type, operation_detail, operator_name, created_at").
		Where("case_no = ?", caseNo).
		Order("created_at ASC").
		Find(&history)
	caseData["progress"] = history

	c.JSON(http.StatusOK, response.Success(caseData))
}

func GetDisputeHistory(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var list []map[string]interface{}
	database.GetDB().Table("dispute_case_history").
		Where("case_id = ?", id).
		Order("created_at DESC").
		Find(&list)

	c.JSON(http.StatusOK, response.Success(list))
}

func UpdateDisputeStatus(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Status int32  `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var oldStatus int32
	var caseNo string
	database.GetDB().Table("dispute_case").
		Select("status, case_no").
		Where("id = ?", id).
		Row().Scan(&oldStatus, &caseNo)

	database.GetDB().Table("dispute_case").
		Where("id = ?", id).
		Update("status", req.Status)

	history := map[string]interface{}{
		"case_id":       id,
		"case_no":       caseNo,
		"operation_type": "STATUS_CHANGE",
		"operation_detail": req.Remark,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.RealName,
		"old_status":    oldStatus,
		"new_status":    req.Status,
		"remark":        req.Remark,
	}
	database.GetDB().Table("dispute_case_history").Create(history)

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, id)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "状态更新成功"))
}
