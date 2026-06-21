package handler

import (
	"context"
	"fmt"
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

type MediationTemplateListRequest struct {
	model.BaseQuery
	Category string `form:"category"`
	Status   int32  `form:"status"`
	Keyword  string `form:"keyword"`
	IsSystem int32  `form:"isSystem"`
}

type MediationTemplateCreateRequest struct {
	TemplateName             string `json:"templateName" binding:"required"`
	TemplateCode             string `json:"templateCode" binding:"required"`
	Category                 string `json:"category" binding:"required"`
	DisputeTypeIDs           string `json:"disputeTypeIds"`
	RecordType               int32  `json:"recordType"`
	MediationPlace           string `json:"mediationPlace"`
	ProcessContentTemplate   string `json:"processContentTemplate"`
	DisputeFocusTemplate     string `json:"disputeFocusTemplate"`
	MediationOpinionTemplate string `json:"mediationOpinionTemplate"`
	AgreementContentTemplate string `json:"agreementContentTemplate"`
	NextStepTemplate         string `json:"nextStepTemplate"`
	DefaultDuration          int    `json:"defaultDuration"`
	ParticipantsTemplate     string `json:"participantsTemplate"`
	Tips                     string `json:"tips"`
	SortOrder                int    `json:"sortOrder"`
}

type MediationTemplateUpdateRequest struct {
	TemplateName             string `json:"templateName"`
	Category                 string `json:"category"`
	DisputeTypeIDs           string `json:"disputeTypeIds"`
	RecordType               int32  `json:"recordType"`
	MediationPlace           string `json:"mediationPlace"`
	ProcessContentTemplate   string `json:"processContentTemplate"`
	DisputeFocusTemplate     string `json:"disputeFocusTemplate"`
	MediationOpinionTemplate string `json:"mediationOpinionTemplate"`
	AgreementContentTemplate string `json:"agreementContentTemplate"`
	NextStepTemplate         string `json:"nextStepTemplate"`
	DefaultDuration          int    `json:"defaultDuration"`
	ParticipantsTemplate     string `json:"participantsTemplate"`
	Tips                     string `json:"tips"`
	SortOrder                int    `json:"sortOrder"`
	Status                   int32  `json:"status"`
}

type MediationTemplateApplyRequest struct {
	CaseID int64 `json:"caseId" binding:"required"`
}

func GetMediationTemplateList(ctx context.Context, c *app.RequestContext) {
	var req MediationTemplateListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("mediation_record_template mrt").
		Where("mrt.deleted_at IS NULL")

	if req.Category != "" {
		db = db.Where("mrt.category = ?", req.Category)
	}
	if req.Status > 0 {
		db = db.Where("mrt.status = ?", req.Status)
	} else {
		db = db.Where("mrt.status = 1")
	}
	if req.IsSystem >= 0 {
		if req.IsSystem == 1 {
			db = db.Where("mrt.is_system = 1")
		}
	}
	if req.Keyword != "" {
		db = db.Where("mrt.template_name LIKE ? OR mrt.template_code LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var templates []map[string]interface{}
	db.Select("mrt.*").
		Order("mrt.sort_order ASC, mrt.use_count DESC, mrt.id ASC").
		Offset(req.GetOffset()).Limit(req.GetLimit()).
		Find(&templates)

	for _, item := range templates {
		if category, ok := item["category"].(string); ok {
			if name, exists := constants.MediationTemplateCategoryMap[category]; exists {
				item["categoryName"] = name
			}
		}
	}

	c.JSON(http.StatusOK, response.SuccessPage(templates, total, req.Page, req.PageSize))
}

func GetMediationTemplateDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的模板ID"))
		return
	}

	var template model.MediationRecordTemplate
	if err := database.GetDB().Where("id = ?", id).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("模板不存在"))
		return
	}

	result := map[string]interface{}{
		"id":                       template.ID,
		"templateName":             template.TemplateName,
		"templateCode":             template.TemplateCode,
		"category":                 template.Category,
		"categoryName":             constants.MediationTemplateCategoryMap[template.Category],
		"disputeTypeIds":           template.DisputeTypeIDs,
		"recordType":               template.RecordType,
		"mediationPlace":           template.MediationPlace,
		"processContentTemplate":   template.ProcessContentTemplate,
		"disputeFocusTemplate":     template.DisputeFocusTemplate,
		"mediationOpinionTemplate": template.MediationOpinionTemplate,
		"agreementContentTemplate": template.AgreementContentTemplate,
		"nextStepTemplate":         template.NextStepTemplate,
		"defaultDuration":          template.DefaultDuration,
		"participantsTemplate":     template.ParticipantsTemplate,
		"tips":                     template.Tips,
		"isSystem":                 template.IsSystem,
		"useCount":                 template.UseCount,
		"sortOrder":                template.SortOrder,
		"status":                   template.Status,
		"creatorName":              template.CreatorName,
		"orgName":                  template.OrgName,
		"createdAt":                template.CreatedAt,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func CreateMediationTemplate(ctx context.Context, c *app.RequestContext) {
	var req MediationTemplateCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var count int64
	database.GetDB().Model(&model.MediationRecordTemplate{}).
		Where("template_code = ?", req.TemplateCode).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("模板编码已存在"))
		return
	}

	if req.DefaultDuration == 0 {
		req.DefaultDuration = 30
	}

	tpl := &model.MediationRecordTemplate{
		TemplateName:             req.TemplateName,
		TemplateCode:             req.TemplateCode,
		Category:                 req.Category,
		DisputeTypeIDs:           req.DisputeTypeIDs,
		RecordType:               req.RecordType,
		MediationPlace:           req.MediationPlace,
		ProcessContentTemplate:   req.ProcessContentTemplate,
		DisputeFocusTemplate:     req.DisputeFocusTemplate,
		MediationOpinionTemplate: req.MediationOpinionTemplate,
		AgreementContentTemplate: req.AgreementContentTemplate,
		NextStepTemplate:         req.NextStepTemplate,
		DefaultDuration:          req.DefaultDuration,
		ParticipantsTemplate:     req.ParticipantsTemplate,
		Tips:                     req.Tips,
		IsSystem:                 0,
		UseCount:                 0,
		SortOrder:                req.SortOrder,
		Status:                   1,
		CreatorID:                userInfo.UserID,
		CreatorName:              userInfo.RealName,
		OrgID:                    userInfo.OrganizationID,
		OrgName:                  userInfo.OrganizationName,
	}

	if err := database.GetDB().Create(tpl).Error; err != nil {
		logger.Error("创建调解模板失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(tpl))
}

func UpdateMediationTemplate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的模板ID"))
		return
	}

	var req MediationTemplateUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var existing model.MediationRecordTemplate
	if err := database.GetDB().Where("id = ?", id).First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("模板不存在"))
		return
	}

	if existing.IsSystem == 1 {
		updates := map[string]interface{}{}
		if req.Status > 0 {
			updates["status"] = req.Status
		}
		if req.SortOrder > 0 {
			updates["sort_order"] = req.SortOrder
		}
		if len(updates) > 0 {
			database.GetDB().Model(&model.MediationRecordTemplate{}).Where("id = ?", id).Updates(updates)
		}
		c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "系统内置模板仅允许修改状态和排序"))
		return
	}

	updates := map[string]interface{}{}
	if req.TemplateName != "" {
		updates["template_name"] = req.TemplateName
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	updates["dispute_type_ids"] = req.DisputeTypeIDs
	updates["record_type"] = req.RecordType
	updates["mediation_place"] = req.MediationPlace
	updates["process_content_template"] = req.ProcessContentTemplate
	updates["dispute_focus_template"] = req.DisputeFocusTemplate
	updates["mediation_opinion_template"] = req.MediationOpinionTemplate
	updates["agreement_content_template"] = req.AgreementContentTemplate
	updates["next_step_template"] = req.NextStepTemplate
	if req.DefaultDuration > 0 {
		updates["default_duration"] = req.DefaultDuration
	}
	updates["participants_template"] = req.ParticipantsTemplate
	updates["tips"] = req.Tips
	updates["sort_order"] = req.SortOrder
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	if err := database.GetDB().Model(&model.MediationRecordTemplate{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("更新调解模板失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("更新失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteMediationTemplate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的模板ID"))
		return
	}

	var existing model.MediationRecordTemplate
	if err := database.GetDB().Where("id = ?", id).First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("模板不存在"))
		return
	}

	if existing.IsSystem == 1 {
		c.JSON(http.StatusBadRequest, response.BadRequest("系统内置模板不允许删除"))
		return
	}

	if err := database.GetDB().Delete(&model.MediationRecordTemplate{}, id).Error; err != nil {
		logger.Error("删除调解模板失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("删除失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func ApplyMediationTemplate(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的模板ID"))
		return
	}

	var req MediationTemplateApplyRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var tpl model.MediationRecordTemplate
	if err := database.GetDB().Where("id = ? AND status = 1", id).First(&tpl).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("模板不存在或已禁用"))
		return
	}

	var caseData struct {
		CaseNo       string `gorm:"column:case_no"`
		Title        string `gorm:"column:title"`
		Status       int32  `gorm:"column:status"`
		MediatorID   int64  `gorm:"column:mediator_id"`
		MediatorName string `gorm:"column:mediator_name"`
	}
	database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).First(&caseData)

	if caseData.Status != constants.CaseStatusMediating {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有调解中状态的案件才能使用模板"))
		return
	}

	now := time.Now()
	processContent := tpl.ProcessContentTemplate
	disputeFocus := tpl.DisputeFocusTemplate
	mediationOpinion := tpl.MediationOpinionTemplate
	agreementContent := tpl.AgreementContentTemplate
	nextStep := tpl.NextStepTemplate

	recordType := tpl.RecordType
	if recordType == 0 {
		recordType = 1
	}

	mediationDuration := tpl.DefaultDuration
	if mediationDuration == 0 {
		mediationDuration = 30
	}

	recordID := utils.GenerateID()
	record := map[string]interface{}{
		"id":                recordID,
		"case_id":           req.CaseID,
		"case_no":           caseData.CaseNo,
		"record_type":       recordType,
		"mediator_id":       userInfo.UserID,
		"mediator_name":     userInfo.RealName,
		"participant_names": tpl.ParticipantsTemplate,
		"mediation_time":    now,
		"mediation_place":   tpl.MediationPlace,
		"mediation_duration": mediationDuration,
		"process_content":   processContent,
		"dispute_focus":     disputeFocus,
		"mediation_opinion": mediationOpinion,
		"agreement_content": agreementContent,
		"result":            0,
		"next_step":         nextStep,
		"is_key_record":     0,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_mediation_record").Create(record).Error; err != nil {
		tx.Rollback()
		logger.Error("套用模板创建调解记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.Error("创建失败"))
		return
	}

	tx.Model(&model.MediationRecordTemplate{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1"))

	useLog := &model.MediationRecordTemplateUseLog{
		TemplateID: id,
		CaseID:     req.CaseID,
		RecordID:   recordID,
		UserID:     userInfo.UserID,
		UserName:   userInfo.RealName,
	}
	tx.Create(useLog)

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "MEDIATION_TEMPLATE_APPLY",
		"operation_detail": fmt.Sprintf("使用模板「%s」创建调解记录", tpl.TemplateName),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"recordId":            recordID,
		"templateId":         id,
		"templateName":       tpl.TemplateName,
		"recordType":         recordType,
		"mediationPlace":     tpl.MediationPlace,
		"mediationDuration":  mediationDuration,
		"processContent":     processContent,
		"disputeFocus":       disputeFocus,
		"mediationOpinion":   mediationOpinion,
		"agreementContent":   agreementContent,
		"nextStep":           nextStep,
		"tips":               tpl.Tips,
		"isDraft":            true,
		"tip":                "模板已套用，请根据实际情况微调内容后保存",
	}, "模板套用成功"))
}

func GetMediationTemplateCategories(ctx context.Context, c *app.RequestContext) {
	categories := make([]map[string]interface{}, 0, len(constants.MediationTemplateCategoryMap))
	for code, name := range constants.MediationTemplateCategoryMap {
		var count int64
		database.GetDB().Model(&model.MediationRecordTemplate{}).
			Where("category = ? AND status = 1 AND deleted_at IS NULL", code).
			Count(&count)
		categories = append(categories, map[string]interface{}{
			"code":  code,
			"name":  name,
			"count": count,
		})
	}

	c.JSON(http.StatusOK, response.Success(categories))
}

func RecommendMediationTemplates(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)

	var templates []model.MediationRecordTemplate
	db := database.GetDB().Where("status = 1 AND deleted_at IS NULL")

	if caseID > 0 {
		var caseData struct {
			TypeName string `gorm:"column:type_name"`
			TypeID   int64  `gorm:"column:type_id"`
		}
		database.GetDB().Table("dispute_case").
			Select("type_name, type_id").
			Where("id = ?", caseID).First(&caseData)

		if caseData.TypeName != "" {
			categoryMap := map[string][]string{
				"邻里":   {constants.MediationTemplateCategoryNeighborhood},
				"物业":   {constants.MediationTemplateCategoryProperty},
				"劳动":   {constants.MediationTemplateCategoryWage},
				"欠薪":   {constants.MediationTemplateCategoryWage},
				"合同":   {constants.MediationTemplateCategoryContract},
				"家庭":   {constants.MediationTemplateCategoryFamily},
				"婚姻":   {constants.MediationTemplateCategoryFamily},
				"交通":   {constants.MediationTemplateCategoryTraffic},
				"消费":   {constants.MediationTemplateCategoryConsumer},
			}

			matchedCategories := []string{}
			for keyword, cats := range categoryMap {
				if containsSubstring(caseData.TypeName, keyword) {
					matchedCategories = append(matchedCategories, cats...)
				}
			}

			if len(matchedCategories) > 0 {
				db = db.Where("category IN (?)", matchedCategories).
					Order(fmt.Sprintf("CASE WHEN category IN (%s) THEN 0 ELSE 1 END, sort_order ASC, use_count DESC",
						stringsJoin(matchedCategories, ",", "'")))
			}
		}
	}

	if caseID == 0 {
		db = db.Order("sort_order ASC, use_count DESC")
	}

	db.Limit(10).Find(&templates)

	result := make([]map[string]interface{}, 0, len(templates))
	for _, tpl := range templates {
		result = append(result, map[string]interface{}{
			"id":                tpl.ID,
			"templateName":      tpl.TemplateName,
			"templateCode":      tpl.TemplateCode,
			"category":          tpl.Category,
			"categoryName":      constants.MediationTemplateCategoryMap[tpl.Category],
			"recordType":        tpl.RecordType,
			"mediationPlace":    tpl.MediationPlace,
			"defaultDuration":   tpl.DefaultDuration,
			"participantsTemplate": tpl.ParticipantsTemplate,
			"tips":              tpl.Tips,
			"useCount":          tpl.UseCount,
			"isSystem":          tpl.IsSystem,
		})
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringsJoin(ss []string, sep string, wrap string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += wrap + s + wrap
	}
	return result
}
