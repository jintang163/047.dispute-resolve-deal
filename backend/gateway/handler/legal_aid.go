package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LegalAidOrgListRequest struct {
	model.BaseQuery
	OrgType  int     `form:"orgType"`
	Level    int     `form:"level"`
	Status   int32   `form:"status"`
	Keyword  string  `form:"keyword"`
	Longitude float64 `form:"longitude"`
	Latitude  float64 `form:"latitude"`
}

type LegalAidOrgCreateRequest struct {
	OrgCode       string  `json:"orgCode" binding:"required"`
	OrgName       string  `json:"orgName" binding:"required"`
	OrgType       int     `json:"orgType"`
	Level         int     `json:"level"`
	Address       string  `json:"address"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	ContactPerson string  `json:"contactPerson"`
	ContactPhone  string  `json:"contactPhone"`
	ContactEmail  string  `json:"contactEmail"`
	ServiceScope  string  `json:"serviceScope"`
	WorkHours     string  `json:"workHours"`
	Description   string  `json:"description"`
	LawyerCount   int     `json:"lawyerCount"`
	CaseCapacity  int     `json:"caseCapacity"`
	SortOrder     int     `json:"sortOrder"`
	Status        int32   `json:"status"`
}

type LegalAidOrgUpdateRequest struct {
	OrgName       string  `json:"orgName"`
	OrgType       int     `json:"orgType"`
	Level         int     `json:"level"`
	Address       string  `json:"address"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	ContactPerson string  `json:"contactPerson"`
	ContactPhone  string  `json:"contactPhone"`
	ContactEmail  string  `json:"contactEmail"`
	ServiceScope  string  `json:"serviceScope"`
	WorkHours     string  `json:"workHours"`
	Description   string  `json:"description"`
	LawyerCount   int     `json:"lawyerCount"`
	CaseCapacity  int     `json:"caseCapacity"`
	SortOrder     int     `json:"sortOrder"`
	Status        int32   `json:"status"`
}

type LegalAidApplyRequest struct {
	CaseID           int64    `json:"caseId" binding:"required"`
	ApplicantName    string   `json:"applicantName" binding:"required"`
	ApplicantPhone   string   `json:"applicantPhone"`
	ApplicantIDCard  string   `json:"applicantIdCard"`
	ApplicantAddress string   `json:"applicantAddress"`
	IncomeLevel      int      `json:"incomeLevel"`
	FamilySize       int      `json:"familySize"`
	MonthlyIncome    float64  `json:"monthlyIncome"`
	AidReason        string   `json:"aidReason"`
	EvidenceSummary  string   `json:"evidenceSummary"`
	MaterialURLs     []string `json:"materialUrls"`
}

type LegalAidApplyListRequest struct {
	model.BaseQuery
	Status     int32  `form:"status"`
	CaseID     int64  `form:"caseId"`
	ApplicantName string `form:"applicantName"`
	ApplicantPhone string `form:"applicantPhone"`
	model.DateRangeQuery
}

type LegalAidAuditRequest struct {
	Status      int32  `json:"status" binding:"required"`
	AuditOpinion string `json:"auditOpinion"`
	RejectReason string `json:"rejectReason"`
}

type LegalAidTransferRequest struct {
	CaseID         int64   `json:"caseId" binding:"required"`
	ToOrgID        int64   `json:"toOrgId" binding:"required"`
	ToLawyerID     int64   `json:"toLawyerId"`
	TransferReason string  `json:"transferReason"`
	CaseSummary    string  `json:"caseSummary"`
	AttachIDs      []int64 `json:"attachIds"`
}

type LegalAidTransferListRequest struct {
	model.BaseQuery
	AcceptStatus int32  `form:"acceptStatus"`
	CaseID       int64  `form:"caseId"`
	ToOrgID      int64  `form:"toOrgId"`
	FromOrgID    int64  `form:"fromOrgId"`
	LegalCaseNo  string `form:"legalCaseNo"`
	model.DateRangeQuery
}

type LegalAidAcceptRequest struct {
	AcceptStatus int32  `json:"acceptStatus" binding:"required"`
	LegalCaseNo  string `json:"legalCaseNo"`
	RejectReason string `json:"rejectReason"`
	ToLawyerID   int64  `json:"toLawyerId"`
	ToLawyerName string `json:"toLawyerName"`
}

type LegalAidCloseRequest struct {
	CloseResult string `json:"closeResult" binding:"required"`
}

type LegalAidRecommendRequest struct {
	CaseID      int64   `json:"caseId" binding:"required"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	DisputeType string  `json:"disputeType"`
}

type LegalAidLawyerListRequest struct {
	model.BaseQuery
	OrgID    int64  `form:"orgId"`
	Status   int32  `form:"status"`
	IsOnline int32  `form:"isOnline"`
	Keyword  string `form:"keyword"`
	Specialty string `form:"specialty"`
}

type LegalAidConsultCreateRequest struct {
	TransferID      int64  `json:"transferId"`
	CaseID          int64  `json:"caseId"`
	LawyerID        int64  `json:"lawyerId" binding:"required"`
	ConsultType     int    `json:"consultType"`
	QuestionTitle   string `json:"questionTitle"`
	QuestionContent string `json:"questionContent"`
}

type LegalAidConsultListRequest struct {
	model.BaseQuery
	LawyerID      int64  `form:"lawyerId"`
	ApplicantID   int64  `form:"applicantId"`
	CaseID        int64  `form:"caseId"`
	ConsultStatus int32  `form:"consultStatus"`
	ConsultType   int    `form:"consultType"`
}

type LegalAidConsultMessageRequest struct {
	ConsultID   int64  `json:"consultId" binding:"required"`
	MessageType int    `json:"messageType"`
	Content     string `json:"content"`
	FileURL     string `json:"fileUrl"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	Duration    int    `json:"duration"`
}

func GetLegalAidOrgList(ctx context.Context, c *app.RequestContext) {
	var req LegalAidOrgListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("legal_aid_org lao").
		Where("lao.deleted_at IS NULL")

	if req.OrgType > 0 {
		db = db.Where("lao.org_type = ?", req.OrgType)
	}
	if req.Level > 0 {
		db = db.Where("lao.level = ?", req.Level)
	}
	if req.Status > 0 {
		db = db.Where("lao.status = ?", req.Status)
	}
	if req.Keyword != "" {
		db = db.Where("lao.org_name LIKE ? OR lao.contact_person LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	if req.Longitude != 0 && req.Latitude != 0 {
		db = db.Select("lao.*, "+
			fmt.Sprintf("(6371 * acos(cos(radians(%f)) * cos(radians(lao.latitude)) * cos(radians(lao.longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(lao.latitude)))) AS distance",
				req.Latitude, req.Longitude, req.Latitude))
		db = db.Order("distance ASC")
	} else {
		db = db.Order("lao.sort_order ASC, lao.id DESC")
	}

	var orgs []map[string]interface{}
	db.Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&orgs)

	c.JSON(http.StatusOK, response.SuccessPage(orgs, total, req.Page, req.PageSize))
}

func GetLegalAidOrgDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的机构ID"))
		return
	}

	var org model.LegalAidOrg
	if err := database.GetDB().Where("id = ?", id).First(&org).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("机构不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(org))
}

func CreateLegalAidOrg(ctx context.Context, c *app.RequestContext) {
	var req LegalAidOrgCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var count int64
	database.GetDB().Model(&model.LegalAidOrg{}).Where("org_code = ?", req.OrgCode).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("机构编码已存在"))
		return
	}

	org := &model.LegalAidOrg{
		OrgCode:       req.OrgCode,
		OrgName:       req.OrgName,
		OrgType:       req.OrgType,
		Level:         req.Level,
		Address:       req.Address,
		Longitude:     req.Longitude,
		Latitude:      req.Latitude,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		ServiceScope:  req.ServiceScope,
		WorkHours:     req.WorkHours,
		Description:   req.Description,
		LawyerCount:   req.LawyerCount,
		CaseCapacity:  req.CaseCapacity,
		SortOrder:     req.SortOrder,
		Status:        req.Status,
	}

	if err := database.GetDB().Create(org).Error; err != nil {
		logger.Error("创建法援机构失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(org))
}

func UpdateLegalAidOrg(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的机构ID"))
		return
	}

	var req LegalAidOrgUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.OrgName != "" {
		updates["org_name"] = req.OrgName
	}
	if req.OrgType > 0 {
		updates["org_type"] = req.OrgType
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}
	updates["address"] = req.Address
	updates["longitude"] = req.Longitude
	updates["latitude"] = req.Latitude
	updates["contact_person"] = req.ContactPerson
	updates["contact_phone"] = req.ContactPhone
	updates["contact_email"] = req.ContactEmail
	updates["service_scope"] = req.ServiceScope
	updates["work_hours"] = req.WorkHours
	updates["description"] = req.Description
	updates["lawyer_count"] = req.LawyerCount
	updates["case_capacity"] = req.CaseCapacity
	updates["sort_order"] = req.SortOrder
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	if err := database.GetDB().Model(&model.LegalAidOrg{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("更新法援机构失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("更新失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteLegalAidOrg(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的机构ID"))
		return
	}

	if err := database.GetDB().Delete(&model.LegalAidOrg{}, id).Error; err != nil {
		logger.Error("删除法援机构失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("删除失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetLegalAidLawyerList(ctx context.Context, c *app.RequestContext) {
	var req LegalAidLawyerListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("legal_aid_lawyer lal").
		Joins("LEFT JOIN legal_aid_org lao ON lal.org_id = lao.id").
		Where("lal.deleted_at IS NULL")

	if req.OrgID > 0 {
		db = db.Where("lal.org_id = ?", req.OrgID)
	}
	if req.Status > 0 {
		db = db.Where("lal.status = ?", req.Status)
	}
	if req.IsOnline > 0 {
		db = db.Where("lal.is_online = ?", req.IsOnline)
	}
	if req.Keyword != "" {
		db = db.Where("lal.lawyer_name LIKE ? OR lal.license_no LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Specialty != "" {
		db = db.Where("lal.specialty LIKE ?", "%"+req.Specialty+"%")
	}

	var total int64
	db.Count(&total)

	var lawyers []map[string]interface{}
	db.Select("lal.*, lao.org_name as org_name").
		Order("lal.consult_rating DESC, lal.consult_count DESC").
		Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&lawyers)

	c.JSON(http.StatusOK, response.SuccessPage(lawyers, total, req.Page, req.PageSize))
}

func GetLegalAidLawyerDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的律师ID"))
		return
	}

	var lawyer model.LegalAidLawyer
	if err := database.GetDB().Where("id = ?", id).First(&lawyer).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("律师不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(lawyer))
}

func ApplyLegalAid(ctx context.Context, c *app.RequestContext) {
	var req LegalAidApplyRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var disputeCase model.DisputeCase
	if err := database.GetDB().Where("id = ?", req.CaseID).First(&disputeCase).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件不存在"))
		return
	}

	applyNo := fmt.Sprintf("LA%s%s", time.Now().Format("20060102"), utils.GenerateIDStr())

	materialURLs := ""
	if len(req.MaterialURLs) > 0 {
		materialJSON, _ := json.Marshal(req.MaterialURLs)
		materialURLs = string(materialJSON)
	}

	application := &model.LegalAidApplication{
		ApplyNo:          applyNo,
		CaseID:           req.CaseID,
		CaseNo:           disputeCase.CaseNo,
		ApplicantName:    req.ApplicantName,
		ApplicantPhone:   req.ApplicantPhone,
		ApplicantIDCard:  req.ApplicantIDCard,
		ApplicantAddress: req.ApplicantAddress,
		IncomeLevel:      req.IncomeLevel,
		FamilySize:       req.FamilySize,
		MonthlyIncome:    req.MonthlyIncome,
		AidReason:        req.AidReason,
		DisputeType:      disputeCase.TypeName,
		EvidenceSummary:  req.EvidenceSummary,
		MaterialURLs:     materialURLs,
		Status:           constants.LegalAidApplyStatusPending,
		SubmitterID:      userInfo.UserID,
		SubmitterName:    userInfo.RealName,
	}

	if err := database.GetDB().Create(application).Error; err != nil {
		logger.Error("提交法援申请失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("提交失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(application))
}

func GetLegalAidApplyList(ctx context.Context, c *app.RequestContext) {
	var req LegalAidApplyListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("legal_aid_application laa").
		Joins("LEFT JOIN dispute_case dc ON laa.case_id = dc.id").
		Where("laa.deleted_at IS NULL")

	if req.Status > 0 {
		db = db.Where("laa.status = ?", req.Status)
	}
	if req.CaseID > 0 {
		db = db.Where("laa.case_id = ?", req.CaseID)
	}
	if req.ApplicantName != "" {
		db = db.Where("laa.applicant_name LIKE ?", "%"+req.ApplicantName+"%")
	}
	if req.ApplicantPhone != "" {
		db = db.Where("laa.applicant_phone = ?", req.ApplicantPhone)
	}
	if req.StartTime != "" {
		db = db.Where("laa.submit_time >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("laa.submit_time <= ?", req.EndTime)
	}

	var total int64
	db.Count(&total)

	var applications []map[string]interface{}
	db.Select("laa.*, dc.title as case_title, dc.case_no as case_no").
		Order("laa.submit_time DESC").
		Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&applications)

	c.JSON(http.StatusOK, response.SuccessPage(applications, total, req.Page, req.PageSize))
}

func GetLegalAidApplyDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的申请ID"))
		return
	}

	var application model.LegalAidApplication
	if err := database.GetDB().Where("id = ?", id).First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("申请不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(application))
}

func AuditLegalAidApply(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的申请ID"))
		return
	}

	var req LegalAidAuditRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	now := time.Now()

	updates := map[string]interface{}{
		"status":        req.Status,
		"auditor_id":    userInfo.UserID,
		"auditor_name":  userInfo.RealName,
		"audit_time":    now,
		"audit_opinion": req.AuditOpinion,
	}

	if req.Status == constants.LegalAidApplyStatusRejected {
		updates["reject_reason"] = req.RejectReason
	}

	if err := database.GetDB().Model(&model.LegalAidApplication{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("审核法援申请失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("审核失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func RecommendLegalAidOrgs(ctx context.Context, c *app.RequestContext) {
	var req LegalAidRecommendRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var disputeCase model.DisputeCase
	longitude := req.Longitude
	latitude := req.Latitude

	if req.CaseID > 0 {
		if err := database.GetDB().Where("id = ?", req.CaseID).First(&disputeCase).Error; err == nil {
			if longitude == 0 && latitude == 0 {
				longitude = disputeCase.Longitude
				latitude = disputeCase.Latitude
			}
			if req.DisputeType == "" {
				req.DisputeType = disputeCase.TypeName
			}
		}
	}

	db := database.GetDB().Table("legal_aid_org lao").
		Where("lao.deleted_at IS NULL AND lao.status = 1")

	if longitude != 0 && latitude != 0 {
		distanceExpr := fmt.Sprintf(
			"(6371 * acos(cos(radians(%f)) * cos(radians(lao.latitude)) * cos(radians(lao.longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(lao.latitude))))",
			latitude, longitude, latitude)
		db = db.Select("lao.*, "+distanceExpr+" AS distance")
		db = db.Order("distance ASC, lao.case_capacity DESC")
	} else {
		db = db.Order("lao.sort_order ASC, lao.case_capacity DESC")
	}

	var orgs []map[string]interface{}
	db.Limit(5).Find(&orgs)

	for _, org := range orgs {
		if d, ok := org["distance"]; ok {
			if dist, ok := d.(float64); ok {
				org["distance"] = math.Round(dist*100) / 100
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(orgs))
}

func CreateLegalAidTransfer(ctx context.Context, c *app.RequestContext) {
	var req LegalAidTransferRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var disputeCase model.DisputeCase
	if err := database.GetDB().Where("id = ?", req.CaseID).First(&disputeCase).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件不存在"))
		return
	}

	var toOrg model.LegalAidOrg
	if err := database.GetDB().Where("id = ?", req.ToOrgID).First(&toOrg).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("法援机构不存在"))
		return
	}

	transferNo := fmt.Sprintf("LAT%s%s", time.Now().Format("20060102"), utils.GenerateIDStr())

	var lawyerName string
	if req.ToLawyerID > 0 {
		var lawyer model.LegalAidLawyer
		if err := database.GetDB().Where("id = ?", req.ToLawyerID).First(&lawyer).Error; err == nil {
			lawyerName = lawyer.LawyerName
		}
	}

	attachIDs := ""
	if len(req.AttachIDs) > 0 {
		for i, id := range req.AttachIDs {
			if i > 0 {
				attachIDs += ","
			}
			attachIDs += strconv.FormatInt(id, 10)
		}
	}

	transfer := &model.LegalAidTransfer{
		TransferNo:     transferNo,
		CaseID:         req.CaseID,
		CaseNo:         disputeCase.CaseNo,
		CaseTitle:      disputeCase.Title,
		DisputeType:    disputeCase.TypeName,
		FromOrgID:      userInfo.OrganizationID,
		FromOrgName:    userInfo.OrganizationName,
		FromUserID:     userInfo.UserID,
		FromUserName:   userInfo.RealName,
		ToOrgID:        req.ToOrgID,
		ToOrgName:      toOrg.OrgName,
		ToLawyerID:     req.ToLawyerID,
		ToLawyerName:   lawyerName,
		TransferReason: req.TransferReason,
		CaseSummary:    req.CaseSummary,
		AttachIDs:      attachIDs,
		AcceptStatus:   constants.LegalAidTransferStatusPending,
	}

	tx := database.GetDB().Begin()
	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		logger.Error("创建法援转介失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("转介失败"))
		return
	}

	tx.Model(&model.LegalAidOrg{}).Where("id = ?", req.ToOrgID).UpdateColumn("accept_count", gorm.Expr("accept_count + 1"))

	tx.Commit()

	c.JSON(http.StatusOK, response.Success(transfer))
}

func GetLegalAidTransferList(ctx context.Context, c *app.RequestContext) {
	var req LegalAidTransferListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("legal_aid_transfer lat").
		Where("lat.deleted_at IS NULL")

	if req.AcceptStatus > 0 {
		db = db.Where("lat.accept_status = ?", req.AcceptStatus)
	}
	if req.CaseID > 0 {
		db = db.Where("lat.case_id = ?", req.CaseID)
	}
	if req.ToOrgID > 0 {
		db = db.Where("lat.to_org_id = ?", req.ToOrgID)
	}
	if req.FromOrgID > 0 {
		db = db.Where("lat.from_org_id = ?", req.FromOrgID)
	}
	if req.LegalCaseNo != "" {
		db = db.Where("lat.legal_case_no LIKE ?", "%"+req.LegalCaseNo+"%")
	}
	if req.StartTime != "" {
		db = db.Where("lat.transfer_time >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("lat.transfer_time <= ?", req.EndTime)
	}

	var total int64
	db.Count(&total)

	var transfers []map[string]interface{}
	db.Order("lat.transfer_time DESC").
		Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&transfers)

	c.JSON(http.StatusOK, response.SuccessPage(transfers, total, req.Page, req.PageSize))
}

func GetLegalAidTransferDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的转介ID"))
		return
	}

	var transfer model.LegalAidTransfer
	if err := database.GetDB().Where("id = ?", id).First(&transfer).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("转介记录不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(transfer))
}

func AcceptLegalAidTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的转介ID"))
		return
	}

	var req LegalAidAcceptRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	now := time.Now()

	updates := map[string]interface{}{
		"accept_status": req.AcceptStatus,
	}

	if req.AcceptStatus == constants.LegalAidTransferStatusAccepted {
		updates["accept_time"] = now
		updates["legal_case_no"] = req.LegalCaseNo
		if req.ToLawyerID > 0 {
			updates["to_lawyer_id"] = req.ToLawyerID
			updates["to_lawyer_name"] = req.ToLawyerName
		}
	} else if req.AcceptStatus == constants.LegalAidTransferStatusRejected {
		updates["reject_reason"] = req.RejectReason
	}

	if err := database.GetDB().Model(&model.LegalAidTransfer{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("处理法援转介失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("处理失败"))
		return
	}

	_ = userInfo
	c.JSON(http.StatusOK, response.Success(nil))
}

func CloseLegalAidTransfer(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的转介ID"))
		return
	}

	var req LegalAidCloseRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	now := time.Now()

	updates := map[string]interface{}{
		"accept_status": constants.LegalAidTransferStatusClosed,
		"close_result":  req.CloseResult,
		"close_time":    now,
	}

	if err := database.GetDB().Model(&model.LegalAidTransfer{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("办结法援转介失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("办结失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func CreateLegalAidConsult(ctx context.Context, c *app.RequestContext) {
	var req LegalAidConsultCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var lawyer model.LegalAidLawyer
	if err := database.GetDB().Where("id = ?", req.LawyerID).First(&lawyer).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("律师不存在"))
		return
	}

	consultNo := fmt.Sprintf("LAC%s%s", time.Now().Format("20060102"), utils.GenerateIDStr())

	var caseNo string
	if req.CaseID > 0 {
		var disputeCase model.DisputeCase
		if err := database.GetDB().Where("id = ?", req.CaseID).First(&disputeCase).Error; err == nil {
			caseNo = disputeCase.CaseNo
		}
	}

	consult := &model.LegalAidConsult{
		ConsultNo:       consultNo,
		TransferID:      req.TransferID,
		CaseID:          req.CaseID,
		CaseNo:          caseNo,
		ApplicantID:     userInfo.UserID,
		ApplicantName:   userInfo.RealName,
		LawyerID:        req.LawyerID,
		LawyerName:      lawyer.LawyerName,
		OrgID:           lawyer.OrgID,
		OrgName:         lawyer.OrgName,
		ConsultType:     req.ConsultType,
		ConsultStatus:   constants.LegalAidConsultStatusPending,
		QuestionTitle:   req.QuestionTitle,
		QuestionContent: req.QuestionContent,
		FreeDuration:    constants.LegalAidFreeConsultDuration,
		IsFree:          1,
	}

	if err := database.GetDB().Create(consult).Error; err != nil {
		logger.Error("创建法援咨询失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(consult))
}

func GetLegalAidConsultList(ctx context.Context, c *app.RequestContext) {
	var req LegalAidConsultListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("legal_aid_consult lac").
		Where("lac.deleted_at IS NULL")

	if req.LawyerID > 0 {
		db = db.Where("lac.lawyer_id = ?", req.LawyerID)
	}
	if req.ApplicantID > 0 {
		db = db.Where("lac.applicant_id = ?", req.ApplicantID)
	}
	if req.CaseID > 0 {
		db = db.Where("lac.case_id = ?", req.CaseID)
	}
	if req.ConsultStatus > 0 {
		db = db.Where("lac.consult_status = ?", req.ConsultStatus)
	}
	if req.ConsultType > 0 {
		db = db.Where("lac.consult_type = ?", req.ConsultType)
	}

	var total int64
	db.Count(&total)

	var consults []map[string]interface{}
	db.Order("lac.created_at DESC").
		Offset(req.GetOffset()).Limit(req.GetLimit()).Find(&consults)

	c.JSON(http.StatusOK, response.SuccessPage(consults, total, req.Page, req.PageSize))
}

func GetLegalAidConsultDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的咨询ID"))
		return
	}

	var consult model.LegalAidConsult
	if err := database.GetDB().Where("id = ?", id).First(&consult).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("咨询不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(consult))
}

func StartLegalAidConsult(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的咨询ID"))
		return
	}

	now := time.Now()

	updates := map[string]interface{}{
		"consult_status": constants.LegalAidConsultStatusOngoing,
		"start_time":     now,
	}

	if err := database.GetDB().Model(&model.LegalAidConsult{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("开始咨询失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("开始失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func EndLegalAidConsult(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的咨询ID"))
		return
	}

	var consult model.LegalAidConsult
	if err := database.GetDB().Where("id = ?", id).First(&consult).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("咨询不存在"))
		return
	}

	now := time.Now()
	totalDuration := 0
	if consult.StartTime != nil {
		totalDuration = int(now.Sub(*consult.StartTime).Seconds())
	}

	usedDuration := totalDuration
	if usedDuration > consult.FreeDuration {
		usedDuration = consult.FreeDuration
	}

	updates := map[string]interface{}{
		"consult_status": constants.LegalAidConsultStatusCompleted,
		"end_time":       now,
		"total_duration": totalDuration,
		"used_duration":  usedDuration,
	}

	tx := database.GetDB().Begin()
	if err := tx.Model(&model.LegalAidConsult{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("结束咨询失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("结束失败"))
		return
	}

	tx.Model(&model.LegalAidLawyer{}).Where("id = ?", consult.LawyerID).
		UpdateColumn("consult_count", gorm.Expr("consult_count + 1"))

	tx.Commit()

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"totalDuration": totalDuration,
		"usedDuration":  usedDuration,
		"freeDuration":  consult.FreeDuration,
	}))
}

func SendLegalAidConsultMessage(ctx context.Context, c *app.RequestContext) {
	var req LegalAidConsultMessageRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var consult model.LegalAidConsult
	if err := database.GetDB().Where("id = ?", req.ConsultID).First(&consult).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("咨询不存在"))
		return
	}

	senderType := constants.LegalAidSenderTypeUser
	senderName := userInfo.RealName

	message := &model.LegalAidConsultMessage{
		ConsultID:   req.ConsultID,
		SenderID:    userInfo.UserID,
		SenderName:  senderName,
		SenderType:  senderType,
		MessageType: req.MessageType,
		Content:     req.Content,
		FileURL:     req.FileURL,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		Duration:    req.Duration,
	}

	if err := database.GetDB().Create(message).Error; err != nil {
		logger.Error("发送消息失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("发送失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(message))
}

func GetLegalAidConsultMessages(ctx context.Context, c *app.RequestContext) {
	consultID, _ := strconv.ParseInt(c.Param("consultId"), 10, 64)
	if consultID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的咨询ID"))
		return
	}

	var messages []model.LegalAidConsultMessage
	database.GetDB().Where("consult_id = ?", consultID).
		Order("created_at ASC").Find(&messages)

	c.JSON(http.StatusOK, response.Success(messages))
}

func RateLegalAidConsult(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的咨询ID"))
		return
	}

	type RateRequest struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	var req RateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var consult model.LegalAidConsult
	if err := database.GetDB().Where("id = ?", id).First(&consult).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("咨询不存在"))
		return
	}

	tx := database.GetDB().Begin()

	tx.Model(&model.LegalAidConsult{}).Where("id = ?", id).Updates(map[string]interface{}{
		"rating":  req.Rating,
		"comment": req.Comment,
	})

	var lawyer model.LegalAidLawyer
	if err := tx.Where("id = ?", consult.LawyerID).First(&lawyer).Error; err == nil {
		newCount := lawyer.ConsultCount + 1
		newTotalScore := lawyer.ConsultRating * float64(lawyer.ConsultCount)
		newTotalScore += float64(req.Rating)
		newRating := newTotalScore / float64(newCount)

		tx.Model(&model.LegalAidLawyer{}).Where("id = ?", consult.LawyerID).Updates(map[string]interface{}{
			"consult_count":  newCount,
			"consult_rating": fmt.Sprintf("%.2f", newRating),
		})
	}

	tx.Commit()

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetCaseLegalAidRecords(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("caseId"), 10, 64)
	if caseID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的案件ID"))
		return
	}

	var applications []model.LegalAidApplication
	database.GetDB().Where("case_id = ?", caseID).
		Order("submit_time DESC").Find(&applications)

	var transfers []model.LegalAidTransfer
	database.GetDB().Where("case_id = ?", caseID).
		Order("transfer_time DESC").Find(&transfers)

	var consults []model.LegalAidConsult
	database.GetDB().Where("case_id = ?", caseID).
		Order("created_at DESC").Find(&consults)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"applications": applications,
		"transfers":    transfers,
		"consults":     consults,
	}))
}
