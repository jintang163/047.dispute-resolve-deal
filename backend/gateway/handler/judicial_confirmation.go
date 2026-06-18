package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type JudicialConfirmationCreateRequest struct {
	CaseID             int64   `json:"caseId" binding:"required"`
	CaseNo             string  `json:"caseNo"`
	CaseTitle          string  `json:"caseTitle"`
	MediationRecordID  int64   `json:"mediationRecordId"`
	ProtocolID         int64   `json:"protocolId"`

	ApplicantName     string `json:"applicantName" binding:"required"`
	ApplicantPhone    string `json:"applicantPhone" binding:"required"`
	ApplicantIDCard   string `json:"applicantIdCard"`
	ApplicantAddress  string `json:"applicantAddress"`

	RespondentName    string `json:"respondentName" binding:"required"`
	RespondentPhone   string `json:"respondentPhone" binding:"required"`
	RespondentIDCard  string `json:"respondentIdCard"`
	RespondentAddress string `json:"respondentAddress"`

	CourtID           int64   `json:"courtId" binding:"required"`
	CourtName         string  `json:"courtName"`

	AgreementContent  string  `json:"agreementContent" binding:"required"`
	PerformanceDeadline string `json:"performanceDeadline"`
	ConfirmAmount     float64 `json:"confirmAmount"`

	Remark            string  `json:"remark"`
}

type JudicialConfirmationListRequest struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"pageSize,default=20"`
	Status    int32  `form:"status"`
	Keyword   string `form:"keyword"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
}

type JudicialConfirmQueryRequest struct {
	ConfirmNo string `form:"confirmNo" binding:"required"`
	IDCard    string `form:"idCard" binding:"required"`
}

type CourtConfigCreateRequest struct {
	CourtCode        string `json:"courtCode" binding:"required"`
	CourtName        string `json:"courtName" binding:"required"`
	CourtLevel       int32  `json:"courtLevel"`
	JurisdictionArea string `json:"jurisdictionArea"`
	Address          string `json:"address"`
	Contact          string `json:"contact"`
	Phone            string `json:"phone"`

	APIEndpoint   string `json:"apiEndpoint"`
	APIAppID      string `json:"apiAppId"`
	APISecret     string `json:"apiSecret"`
	APIPublicKey  string `json:"apiPublicKey"`

	SealCertNo    string `json:"sealCertNo"`
	SealImageURL  string `json:"sealImageUrl"`

	SortOrder       int    `json:"sortOrder"`
	Status          int32  `json:"status"`
}

func GetJudicialConfirmationList(ctx context.Context, c *app.RequestContext) {
	var req JudicialConfirmationListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	list, total, err := service.JudicialConfirmationServiceInst().GetConfirmationList(
		ctx,
		userInfo.UserID,
		userInfo.Role,
		userInfo.OrganizationID,
		req.Page,
		req.PageSize,
		req.Status,
		req.Keyword,
	)
	if err != nil {
		logger.Error("Get judicial confirmation list failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		data := make(map[string]interface{})
		jsonBytes, _ := json.Marshal(item)
		json.Unmarshal(jsonBytes, &data)
		data["status_name"] = constants.JudicialStatusMap[int(item.Status)]
		if item.Status == model.JudicialStatusConfirmed && item.PerformanceDeadline != nil {
			daysLeft := int(time.Until(*item.PerformanceDeadline).Hours() / 24)
			data["days_left"] = daysLeft
		}
		result = append(result, data)
	}

	c.JSON(http.StatusOK, response.Page(result, total, req.Page, req.PageSize))
}

func GetJudicialConfirmationDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	confirm, err := service.JudicialConfirmationServiceInst().GetConfirmationDetail(
		ctx, id, userInfo.UserID, userInfo.Role)
	if err != nil {
		logger.Error("Get judicial confirmation detail failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}
	if confirm == nil {
		c.JSON(http.StatusNotFound, response.NotFound("记录不存在"))
		return
	}

	result := make(map[string]interface{})
	jsonBytes, _ := json.Marshal(confirm)
	json.Unmarshal(jsonBytes, &result)
	result["status_name"] = constants.JudicialStatusMap[int(confirm.Status)]

	logs, err := service.JudicialConfirmationServiceInst().GetConfirmLogs(ctx, id)
	if err == nil {
		result["logs"] = logs
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func QueryJudicialConfirmationByNo(ctx context.Context, c *app.RequestContext) {
	var req JudicialConfirmQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	confirm, logs, err := service.JudicialConfirmationServiceInst().GetConfirmationByNo(
		ctx, req.ConfirmNo, req.IDCard)
	if err != nil {
		logger.Error("Query judicial confirmation by no failed",
			zap.String("confirmNo", req.ConfirmNo),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}
	if confirm == nil {
		c.JSON(http.StatusNotFound, response.NotFound("记录不存在或身份证号不匹配"))
		return
	}

	result := make(map[string]interface{})
	jsonBytes, _ := json.Marshal(confirm)
	json.Unmarshal(jsonBytes, &result)
	result["status_name"] = constants.JudicialStatusMap[int(confirm.Status)]
	result["logs"] = logs

	c.JSON(http.StatusOK, response.Success(result))
}

func CreateJudicialConfirmation(ctx context.Context, c *app.RequestContext) {
	var req JudicialConfirmationCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var deadline *time.Time
	if req.PerformanceDeadline != "" {
		if t, err := time.Parse("2006-01-02", req.PerformanceDeadline); err == nil {
			deadline = &t
		}
	}

	confirm := &model.JudicialConfirmation{
		CaseID:            req.CaseID,
		CaseNo:            req.CaseNo,
		CaseTitle:         req.CaseTitle,
		MediationRecordID: req.MediationRecordID,
		ProtocolID:        req.ProtocolID,
		ApplicantName:     req.ApplicantName,
		ApplicantPhone:    req.ApplicantPhone,
		ApplicantIDCard:   req.ApplicantIDCard,
		ApplicantAddress:  req.ApplicantAddress,
		RespondentName:    req.RespondentName,
		RespondentPhone:   req.RespondentPhone,
		RespondentIDCard:  req.RespondentIDCard,
		RespondentAddress: req.RespondentAddress,
		CourtID:           req.CourtID,
		CourtName:         req.CourtName,
		AgreementContent:  req.AgreementContent,
		PerformanceDeadline: deadline,
		ConfirmAmount:     req.ConfirmAmount,
		OrganizationID:    userInfo.OrganizationID,
		OrganizationName:  userInfo.OrganizationName,
		Remark:            req.Remark,
	}

	confirmNo, err := service.JudicialConfirmationServiceInst().CreateConfirmation(
		ctx, confirm, userInfo.UserID, userInfo.RealName)
	if err != nil {
		logger.Error("Create judicial confirmation failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId": confirm.ID,
		"confirmNo": confirmNo,
	}))
}

func SubmitJudicialToCourt(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	err := service.JudicialConfirmationServiceInst().SubmitToCourt(
		ctx, id, userInfo.UserID, userInfo.RealName)
	if err != nil {
		logger.Error("Submit judicial to court failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("提交失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId": id,
		"status":    model.JudicialStatusReviewing,
	}))
}

func QueryCourtStatus(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	err := service.JudicialConfirmationServiceInst().QueryCourtStatus(
		ctx, id, userInfo.UserID, userInfo.RealName)
	if err != nil {
		logger.Error("Query court status failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败: "+err.Error()))
		return
	}

	confirm, _ := service.JudicialConfirmationServiceInst().GetConfirmationDetail(
		ctx, id, userInfo.UserID, userInfo.Role)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId": id,
		"status":    confirm.Status,
		"statusName": constants.JudicialStatusMap[int(confirm.Status)],
	}))
}

func GenerateConfirmationDocument(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	documentURL, err := service.JudicialConfirmationServiceInst().GenerateConfirmationDocument(
		ctx, id, userInfo.UserID, userInfo.RealName)
	if err != nil {
		logger.Error("Generate confirmation document failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("生成失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId":   id,
		"documentUrl": documentURL,
	}))
}

func SealConfirmationDocument(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	err := service.JudicialConfirmationServiceInst().SealDocument(
		ctx, id, userInfo.UserID, userInfo.RealName)
	if err != nil {
		logger.Error("Seal confirmation document failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("签章失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId":  id,
		"sealStatus": model.SealStatusDone,
	}))
}

func GetConfirmationLogs(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	logs, err := service.JudicialConfirmationServiceInst().GetConfirmLogs(ctx, id)
	if err != nil {
		logger.Error("Get confirmation logs failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}

	result := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		data := make(map[string]interface{})
		jsonBytes, _ := json.Marshal(log)
		json.Unmarshal(jsonBytes, &data)
		data["action_type_name"] = constants.JudicialActionTypeMap[log.ActionType]
		result = append(result, data)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func SendPerformanceReminder(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := service.JudicialConfirmationServiceInst().SendPerformanceReminder(ctx, id)
	if err != nil {
		logger.Error("Send performance reminder failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("发送失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId": id,
		"sent":      true,
	}))
}

func SendExpirationReminder(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := service.JudicialConfirmationServiceInst().SendExpirationReminder(ctx, id)
	if err != nil {
		logger.Error("Send expiration reminder failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("发送失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"confirmId": id,
		"sent":      true,
	}))
}

func GetCourtConfigList(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	userInfo := middleware.GetUserInfo(c)

	list, total, err := service.JudicialConfirmationServiceInst().CourtConfigList(
		ctx, userInfo.OrganizationID, page, pageSize)
	if err != nil {
		logger.Error("Get court config list failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetCourtConfigDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	config, err := service.JudicialConfirmationServiceInst().CourtConfigDetail(ctx, id)
	if err != nil {
		logger.Error("Get court config detail failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("查询失败"))
		return
	}
	if config == nil {
		c.JSON(http.StatusNotFound, response.NotFound("配置不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(config))
}

func CreateCourtConfig(ctx context.Context, c *app.RequestContext) {
	var req CourtConfigCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	config := &model.CourtConfig{
		CourtCode:        req.CourtCode,
		CourtName:        req.CourtName,
		CourtLevel:       req.CourtLevel,
		JurisdictionArea: req.JurisdictionArea,
		Address:          req.Address,
		Contact:          req.Contact,
		Phone:            req.Phone,
		APIEndpoint:      req.APIEndpoint,
		APIAppID:         req.APIAppID,
		APISecret:        req.APISecret,
		APIPublicKey:     req.APIPublicKey,
		SealCertNo:       req.SealCertNo,
		SealImageURL:     req.SealImageURL,
		OrganizationID:   userInfo.OrganizationID,
		OrganizationName: userInfo.OrganizationName,
		SortOrder:        req.SortOrder,
		Status:           req.Status,
	}

	err := service.JudicialConfirmationServiceInst().CreateCourtConfig(ctx, config, userInfo.UserID)
	if err != nil {
		logger.Error("Create court config failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id": config.ID,
	}))
}

func UpdateCourtConfig(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req CourtConfigCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	config := &model.CourtConfig{
		BaseModel:        model.BaseModel{ID: id},
		CourtCode:        req.CourtCode,
		CourtName:        req.CourtName,
		CourtLevel:       req.CourtLevel,
		JurisdictionArea: req.JurisdictionArea,
		Address:          req.Address,
		Contact:          req.Contact,
		Phone:            req.Phone,
		APIEndpoint:      req.APIEndpoint,
		APIAppID:         req.APIAppID,
		APISecret:        req.APISecret,
		APIPublicKey:     req.APIPublicKey,
		SealCertNo:       req.SealCertNo,
		SealImageURL:     req.SealImageURL,
		SortOrder:        req.SortOrder,
		Status:           req.Status,
	}

	err := service.JudicialConfirmationServiceInst().UpdateCourtConfig(ctx, config, userInfo.UserID)
	if err != nil {
		logger.Error("Update court config failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("更新失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id": id,
	}))
}

func DeleteCourtConfig(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	err := service.JudicialConfirmationServiceInst().DeleteCourtConfig(ctx, id, userInfo.UserID)
	if err != nil {
		logger.Error("Delete court config failed", zap.Int64("id", id), logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("删除失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetCourtOptions(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var list []*model.CourtConfig
	db := database.GetDB().Model(&model.CourtConfig{}).
		Where("status = 1 AND deleted_at IS NULL")

	if userInfo.OrganizationID > 0 {
		db = db.Where("organization_id = ? OR organization_id = 0", userInfo.OrganizationID)
	}

	db.Order("sort_order ASC, id DESC").Find(&list)

	result := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		result = append(result, map[string]interface{}{
			"id":        item.ID,
			"courtCode": item.CourtCode,
			"courtName": item.CourtName,
			"courtLevel": item.CourtLevel,
			"address":   item.Address,
			"phone":     item.Phone,
		})
	}

	c.JSON(http.StatusOK, response.Success(result))
}
