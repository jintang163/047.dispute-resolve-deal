package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type MediationRecordRequest struct {
	RecordType       int32    `json:"recordType"`
	MediationTime    string   `json:"mediationTime" binding:"required"`
	MediationPlace   string   `json:"mediationPlace"`
	MediationDuration int     `json:"mediationDuration"`
	ProcessContent   string   `json:"processContent" binding:"required"`
	DisputeFocus     string   `json:"disputeFocus"`
	MediationOpinion string   `json:"mediationOpinion"`
	AgreementContent string   `json:"agreementContent"`
	Result           int32    `json:"result"`
	NextStep         string   `json:"nextStep"`
	Participants     []string `json:"participants"`
	AssistMediators  []int64  `json:"assistMediators"`
	IsDraft          int32    `json:"isDraft"`
	TemplateID       int64    `json:"templateId"`
	TemplateName     string   `json:"templateName"`
}

func CreateMediationRecord(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var req MediationRecordRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
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
		Where("id = ?", caseID).
		First(&caseData)

	if caseData.Status != constants.CaseStatusMediating {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有调解中状态的案件才能记录调解信息"))
		return
	}

	if caseData.MediatorID != userInfo.UserID && userInfo.Role > constants.RoleMediator {
		c.JSON(http.StatusForbidden, response.Forbidden("只有案件调解员才能记录调解信息"))
		return
	}

	recordID := utils.GenerateID()
	participantNames := ""
	if len(req.Participants) > 0 {
		participantNames = strings.Join(req.Participants, ",")
	}

	assistMediators := ""
	if len(req.AssistMediators) > 0 {
		ids := make([]string, len(req.AssistMediators))
		for i, id := range req.AssistMediators {
			ids[i] = strconv.FormatInt(id, 10)
		}
		assistMediators = strings.Join(ids, ",")
	}

	isKeyRecord := int32(1)
	if req.IsDraft == 1 {
		isKeyRecord = 0
	}

	record := map[string]interface{}{
		"id":                  recordID,
		"case_id":             caseID,
		"case_no":             caseData.CaseNo,
		"record_type":         req.RecordType,
		"mediator_id":         userInfo.UserID,
		"mediator_name":       userInfo.RealName,
		"participant_names":   participantNames,
		"mediation_time":      req.MediationTime,
		"mediation_place":     req.MediationPlace,
		"mediation_duration":  req.MediationDuration,
		"process_content":     req.ProcessContent,
		"dispute_focus":       req.DisputeFocus,
		"mediation_opinion":   req.MediationOpinion,
		"agreement_content":   req.AgreementContent,
		"result":              req.Result,
		"next_step":           req.NextStep,
		"assist_mediators":    assistMediators,
		"is_key_record":       isKeyRecord,
		"is_draft":            req.IsDraft,
		"template_id":         req.TemplateID,
		"template_name":       req.TemplateName,
	}

	tx := database.GetDB().Begin()
	tx.Table("dispute_mediation_record").Create(record)

	progressUpdates := map[string]interface{}{
		"last_progress_time": time.Now(),
	}

	if req.IsDraft != 1 && req.Result > 0 {
		progressUpdates["mediation_result"] = req.Result
		if req.AgreementContent != "" {
			progressUpdates["agreement_content"] = req.AgreementContent
		}
		if req.Result == constants.MediationResultSuccess {
			progressUpdates["mediation_end_time"] = time.Now()
		}
	}
	tx.Table("dispute_case").Where("id = ?", caseID).Updates(progressUpdates)

	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "MEDIATION_RECORD",
		"operation_detail": fmt.Sprintf("记录调解信息，时长: %d分钟，结果: %d", req.MediationDuration, req.Result),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	if req.IsDraft == 1 {
		history["operation_type"] = "MEDIATION_RECORD_DRAFT"
		history["operation_detail"] = fmt.Sprintf("保存调解记录草稿，时长: %d分钟", req.MediationDuration)
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	if req.IsDraft != 1 {
		go func() {
			msg := map[string]interface{}{
				"caseId":    caseID,
				"caseNo":    caseData.CaseNo,
				"caseTitle": caseData.Title,
				"recordId":  recordID,
				"result":    req.Result,
				"duration":  req.MediationDuration,
				"mediator":  userInfo.RealName,
			}
			mq.SendMessage(constants.MQTopicAIProcess, msg)
		}()
	}

	responseData := map[string]interface{}{
		"id":      recordID,
		"isDraft": req.IsDraft == 1,
	}

	if req.IsDraft != 1 && req.Result == constants.MediationResultFail {
		var caseLocation struct {
			Longitude float64 `gorm:"column:longitude"`
			Latitude  float64 `gorm:"column:latitude"`
			TypeName  string  `gorm:"column:type_name"`
		}
		database.GetDB().Table("dispute_case").
			Select("longitude, latitude, type_name").
			Where("id = ?", caseID).
			First(&caseLocation)

		if caseLocation.Longitude != 0 && caseLocation.Latitude != 0 {
			distanceExpr := fmt.Sprintf(
				"(6371 * acos(cos(radians(%f)) * cos(radians(lao.latitude)) * cos(radians(lao.longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(lao.latitude))))",
				caseLocation.Latitude, caseLocation.Longitude, caseLocation.Latitude)

			var recommendOrgs []map[string]interface{}
			database.GetDB().Table("legal_aid_org lao").
				Select("lao.id, lao.org_code, lao.org_name, lao.org_type, lao.level, lao.address, "+
					"lao.contact_person, lao.contact_phone, lao.lawyer_count, lao.case_capacity, "+
					distanceExpr+" AS distance").
				Where("lao.deleted_at IS NULL AND lao.status = 1").
				Order("distance ASC, lao.case_capacity DESC").
				Limit(5).
				Find(&recommendOrgs)

			for _, org := range recommendOrgs {
				if d, ok := org["distance"]; ok {
					if dist, ok := d.(float64); ok {
						org["distance"] = math.Round(dist*100) / 100
					}
				}
			}

			responseData["legalAidRecommend"] = map[string]interface{}{
				"recommendOrgs": recommendOrgs,
				"tip":           "调解未达成协议，系统已为您推荐就近的法律援助机构，可一键转介申请法律援助。",
			}

			history := map[string]interface{}{
				"case_id":          caseID,
				"case_no":          caseData.CaseNo,
				"operation_type":   "LEGAL_AID_RECOMMEND",
				"operation_detail": fmt.Sprintf("调解失败，系统自动推荐%d家就近法律援助机构", len(recommendOrgs)),
				"operator_id":      0,
				"operator_name":    "系统自动",
				"old_status":       caseData.Status,
				"new_status":       caseData.Status,
			}
			database.GetDB().Table("dispute_case_history").Create(history)
		}
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(responseData, "调解记录创建成功"))
}

func GetMediationRecords(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var list []map[string]interface{}
	database.GetDB().Table("dispute_mediation_record").
		Where("case_id = ?", caseID).
		Order("mediation_time DESC, id DESC").
		Find(&list)

	for _, item := range list {
		if result, ok := item["result"].(int32); ok {
			resultMap := map[int]string{
				0: "进行中",
				1: "达成协议",
				2: "未达成",
				3: "部分达成",
			}
			item["result_name"] = resultMap[int(result)]
		}
		if recordType, ok := item["record_type"].(int32); ok {
			typeMap := map[int]string{
				1: "初次调解",
				2: "再次调解",
				3: "补充调解",
			}
			item["record_type_name"] = typeMap[int(recordType)]
		}
		if isDraft, ok := item["is_draft"]; ok {
			if draft, ok := isDraft.(int32); ok {
				item["is_draft"] = draft == 1
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func UpdateMediationRecord(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordID, _ := strconv.ParseInt(c.Param("recordId"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var req MediationRecordRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var record struct {
		MediatorID int64 `gorm:"column:mediator_id"`
		IsDraft    int32 `gorm:"column:is_draft"`
		CaseNo     string `gorm:"column:case_no"`
	}
	database.GetDB().Table("dispute_mediation_record").
		Select("mediator_id, is_draft, case_no").
		Where("id = ?", recordID).
		First(&record)

	if record.MediatorID != userInfo.UserID {
		c.JSON(http.StatusForbidden, response.Forbidden("只能修改自己创建的调解记录"))
		return
	}

	updates := map[string]interface{}{
		"mediation_time":     req.MediationTime,
		"mediation_place":    req.MediationPlace,
		"mediation_duration": req.MediationDuration,
		"process_content":    req.ProcessContent,
		"dispute_focus":      req.DisputeFocus,
		"mediation_opinion":  req.MediationOpinion,
		"agreement_content":  req.AgreementContent,
		"result":             req.Result,
		"next_step":          req.NextStep,
	}

	if len(req.Participants) > 0 {
		updates["participant_names"] = strings.Join(req.Participants, ",")
	}

	wasDraft := record.IsDraft == 1
	nowFormal := req.IsDraft != 1
	operationDetail := ""

	if wasDraft && nowFormal {
		updates["is_draft"] = 0
		updates["is_key_record"] = 1
		operationDetail = fmt.Sprintf("草稿转正式记录，时长: %d分钟，结果: %d", req.MediationDuration, req.Result)
	} else if wasDraft && req.IsDraft == 1 {
		operationDetail = fmt.Sprintf("更新草稿记录，时长: %d分钟", req.MediationDuration)
	} else {
		operationDetail = fmt.Sprintf("更新调解记录，时长: %d分钟，结果: %d", req.MediationDuration, req.Result)
	}

	tx := database.GetDB().Begin()
	tx.Table("dispute_mediation_record").
		Where("id = ?", recordID).
		Updates(updates)

	if wasDraft && nowFormal {
		progressUpdates := map[string]interface{}{
			"last_progress_time": time.Now(),
		}
		if req.Result > 0 {
			progressUpdates["mediation_result"] = req.Result
			if req.AgreementContent != "" {
				progressUpdates["agreement_content"] = req.AgreementContent
			}
			if req.Result == constants.MediationResultSuccess {
				progressUpdates["mediation_end_time"] = time.Now()
			}
		}
		tx.Table("dispute_case").Where("id = ?", caseID).Updates(progressUpdates)

		go func() {
			var caseData struct {
				CaseNo  string `gorm:"column:case_no"`
				Title   string `gorm:"column:title"`
			}
			database.GetDB().Table("dispute_case").
				Select("case_no, title").
				Where("id = ?", caseID).First(&caseData)

			msg := map[string]interface{}{
				"caseId":    caseID,
				"caseNo":    caseData.CaseNo,
				"caseTitle": caseData.Title,
				"recordId":  recordID,
				"result":    req.Result,
				"duration":  req.MediationDuration,
				"mediator":  userInfo.RealName,
			}
			mq.SendMessage(constants.MQTopicAIProcess, msg)
		}()
	}

	historyType := "MEDIATION_RECORD_UPDATE"
	if wasDraft && nowFormal {
		historyType = "MEDIATION_RECORD_DRAFT_TO_FORMAL"
	} else if wasDraft {
		historyType = "MEDIATION_RECORD_DRAFT_UPDATE"
	}
	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          record.CaseNo,
		"operation_type":   historyType,
		"operation_detail": operationDetail,
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"isDraft": req.IsDraft == 1,
		"wasDraft": wasDraft,
		"nowFormal": nowFormal,
	}, "调解记录更新成功"))
}

func GetAISummary(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	recordID, _ := strconv.ParseInt(c.Param("recordId"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var record struct {
		ProcessContent   string `gorm:"column:process_content"`
		DisputeFocus     string `gorm:"column:dispute_focus"`
		MediationOpinion string `gorm:"column:mediation_opinion"`
		AgreementContent string `gorm:"column:agreement_content"`
		AISummary        string `gorm:"column:ai_summary"`
	}
	database.GetDB().Table("dispute_mediation_record").
		Select("process_content, dispute_focus, mediation_opinion, agreement_content, ai_summary").
		Where("id = ?", recordID).
		First(&record)

	if record.AISummary != "" {
		c.JSON(http.StatusOK, response.Success(map[string]interface{}{
			"summary": record.AISummary,
			"cached":  true,
		}))
		return
	}

	content := fmt.Sprintf(`纠纷调解记录：
调解过程：%s
争议焦点：%s
调解意见：%s
协议内容：%s

请生成一份简明的调解摘要，包括：
1. 纠纷核心问题
2. 双方主要观点
3. 调解方案
4. 后续建议`, record.ProcessContent, record.DisputeFocus, record.MediationOpinion, record.AgreementContent)

	aiResp, err := ai.GenerateSummary(content, constants.SummaryTypeMediation)
	if err != nil {
		logger.Error("Generate AI summary failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("生成摘要失败，请稍后重试"))
		return
	}

	database.GetDB().Table("dispute_mediation_record").
		Where("id = ?", recordID).
		Update("ai_summary", aiResp.Summary)

	summaryRecord := map[string]interface{}{
		"case_id":         caseID,
		"record_id":       recordID,
		"summary_type":    constants.SummaryTypeMediation,
		"original_content": content,
		"ai_summary":      aiResp.Summary,
		"ai_suggestion":   aiResp.Suggestion,
		"risk_level":      aiResp.RiskLevel,
		"risk_points":     aiResp.RiskPoints,
		"tokens_used":     aiResp.TokensUsed,
		"cost_time":       aiResp.CostTime,
	}
	database.GetDB().Table("ai_mediation_summary").Create(summaryRecord)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"summary":    aiResp.Summary,
		"suggestion": aiResp.Suggestion,
		"riskLevel":  aiResp.RiskLevel,
		"riskPoints": aiResp.RiskPoints,
		"cached":     false,
	}))
}

func GetDisputeTypes(ctx context.Context, c *app.RequestContext) {
	cacheKey := "dispute_types_tree"
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	var allTypes []map[string]interface{}
	database.GetDB().Table("dispute_type").
		Where("status = 1").
		Order("level ASC, sort_order ASC, id ASC").
		Find(&allTypes)

	typeMap := make(map[int64]map[string]interface{})
	for _, t := range allTypes {
		id := t["id"].(int64)
		typeMap[id] = t
		t["children"] = []interface{}{}
	}

	root := make([]interface{}, 0)
	for _, t := range allTypes {
		parentID := t["parent_id"].(int64)
		if parentID == 0 {
			root = append(root, t)
		} else if parent, ok := typeMap[parentID]; ok {
			children := parent["children"].([]interface{})
			parent["children"] = append(children, t)
		}
	}

	jsonData, _ := json.Marshal(root)
	cache.Set(ctx, cacheKey, string(jsonData), time.Hour*24)

	c.JSON(http.StatusOK, response.Success(root))
}

type GenerateMediationProtocolRequest struct {
	CaseID           int64    `json:"caseId" binding:"required"`
	CaseNo           string   `json:"caseNo"`
	CaseTitle        string   `json:"caseTitle"`
	DisputeType      string   `json:"disputeType"`
	PartyAName       string   `json:"partyAName" binding:"required"`
	PartyAGender     string   `json:"partyAGender"`
	PartyAIDCard     string   `json:"partyAIDCard"`
	PartyAAddress    string   `json:"partyAAddress"`
	PartyAPhone      string   `json:"partyAPhone"`
	PartyBName       string   `json:"partyBName" binding:"required"`
	PartyBGender     string   `json:"partyBGender"`
	PartyBIDCard     string   `json:"partyBIDCard"`
	PartyBAddress    string   `json:"partyBAddress"`
	PartyBPhone      string   `json:"partyBPhone"`
	DisputeSummary   string   `json:"disputeSummary" binding:"required"`
	LiabilityParty   string   `json:"liabilityParty"`
	LiabilityRatioA  int      `json:"liabilityRatioA"`
	LiabilityRatioB  int      `json:"liabilityRatioB"`
	LiabilityReason  string   `json:"liabilityReason"`
	CompensationAmount float64 `json:"compensationAmount" binding:"required"`
	CompensationType string   `json:"compensationType"`
	PaymentMethod    string   `json:"paymentMethod"`
	PerformanceDate  string   `json:"performanceDate" binding:"required"`
	PaymentAccount   string   `json:"paymentAccount"`
	OtherTerms       []string `json:"otherTerms"`
	BreachClause     string   `json:"breachClause"`
	SignPlace        string   `json:"signPlace"`
	SignDate         string   `json:"signDate"`
	RegionPrefix     string   `json:"regionPrefix"`
	ProtocolYear     int      `json:"protocolYear"`
	ProtocolSeq      int      `json:"protocolSeq"`
}

func GenerateMediationProtocol(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var req GenerateMediationProtocolRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var caseData struct {
		CaseNo       string `gorm:"column:case_no"`
		Title        string `gorm:"column:title"`
		TypeName     string `gorm:"column:type_name"`
		Status       int32  `gorm:"column:status"`
		MediatorID   int64  `gorm:"column:mediator_id"`
		MediatorName string `gorm:"column:mediator_name"`
	}
	database.GetDB().Table("dispute_case").
		Where("id = ?", caseID).
		First(&caseData)

	if caseData.MediatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有案件调解员或领导可以生成调解协议"))
		return
	}

	if req.CaseNo == "" {
		req.CaseNo = caseData.CaseNo
	}
	if req.CaseTitle == "" {
		req.CaseTitle = caseData.Title
	}
	if req.DisputeType == "" {
		req.DisputeType = caseData.TypeName
	}

	if req.ProtocolYear == 0 {
		req.ProtocolYear = time.Now().Year()
	}
	if req.ProtocolSeq == 0 {
		var count int64
		database.GetDB().Table("mediation_protocol").
			Where("case_id = ?", caseID).
			Count(&count)
		req.ProtocolSeq = int(count) + 1
	}
	if req.SignDate == "" {
		req.SignDate = time.Now().Format("2006-01-02")
	}

	params := &ai.MediationProtocolParams{
		CaseID:             caseID,
		CaseNo:             req.CaseNo,
		CaseTitle:          req.CaseTitle,
		DisputeType:        req.DisputeType,
		PartyAName:         req.PartyAName,
		PartyAGender:       req.PartyAGender,
		PartyAIDCard:       req.PartyAIDCard,
		PartyAAddress:      req.PartyAAddress,
		PartyAPhone:        req.PartyAPhone,
		PartyBName:         req.PartyBName,
		PartyBGender:       req.PartyBGender,
		PartyBIDCard:       req.PartyBIDCard,
		PartyBAddress:      req.PartyBAddress,
		PartyBPhone:        req.PartyBPhone,
		DisputeSummary:     req.DisputeSummary,
		LiabilityParty:     req.LiabilityParty,
		LiabilityRatioA:    req.LiabilityRatioA,
		LiabilityRatioB:    req.LiabilityRatioB,
		LiabilityReason:    req.LiabilityReason,
		CompensationAmount: req.CompensationAmount,
		CompensationType:   req.CompensationType,
		PaymentMethod:      req.PaymentMethod,
		PerformanceDate:    req.PerformanceDate,
		PaymentAccount:     req.PaymentAccount,
		OtherTerms:         req.OtherTerms,
		BreachClause:       req.BreachClause,
		MediatorName:       caseData.MediatorName,
		SignPlace:          req.SignPlace,
		SignDate:           req.SignDate,
		RegionPrefix:       req.RegionPrefix,
		ProtocolYear:       req.ProtocolYear,
		ProtocolSeq:        req.ProtocolSeq,
	}

	result, err := ai.GenerateMediationProtocol(params)
	if err != nil {
		logger.Error("AI generate mediation protocol failed",
			logger.Int64("caseId", caseID),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.ServerError("生成调解协议失败，请稍后重试"))
		return
	}

	protocolID := utils.GenerateID()
	tx := database.GetDB().Begin()
	protocolData := map[string]interface{}{
		"id":               protocolID,
		"case_id":          caseID,
		"protocol_no":      result.ProtocolNo,
		"title":            result.Title,
		"content":          result.Content,
		"party_a_name":     result.PartyAName,
		"party_b_name":     result.PartyBName,
		"mediator_name":    result.MediatorName,
		"agreement_items":  result.AgreementItems,
		"breach_clause":    result.BreachClause,
		"is_signed":        0,
		"is_ai_generated":  1,
		"ai_generated_at":  time.Now(),
		"created_by":       userInfo.UserID,
	}
	if err := tx.Table("mediation_protocol").Create(protocolData).Error; err != nil {
		tx.Rollback()
		logger.Error("Save mediation protocol failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("保存协议失败"))
		return
	}

	legalBasisJSON, _ := json.Marshal(result.LegalBasis)
	aiLog := map[string]interface{}{
		"id":               utils.GenerateID(),
		"case_id":          caseID,
		"record_type":      "mediation_protocol",
		"ref_id":           protocolID,
		"original_params":  fmt.Sprintf("%+v", req),
		"ai_content":       result.Content,
		"legal_basis":      string(legalBasisJSON),
		"tokens_used":      0,
		"cost_time":        0,
		"created_by":       userInfo.UserID,
	}
	tx.Table("ai_generation_log").Create(aiLog)

	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "PROTOCOL_AI_GENERATE",
		"operation_detail": fmt.Sprintf("AI生成调解协议，协议编号: %s", result.ProtocolNo),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":             protocolID,
		"protocolId":     protocolID,
		"protocolNo":     result.ProtocolNo,
		"title":          result.Title,
		"content":        result.Content,
		"partyAName":     result.PartyAName,
		"partyBName":     result.PartyBName,
		"mediatorName":   result.MediatorName,
		"agreementItems": result.AgreementItems,
		"breachClause":   result.BreachClause,
		"legalBasis":     result.LegalBasis,
		"generatedAt":    result.GeneratedAt,
		"isAIGenerated":  1,
		"isSigned":       0,
		"isAdopted":      0,
	}))
}

func GetMediationProtocolList(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var rawList []map[string]interface{}
	database.GetDB().Table("mediation_protocol").
		Where("case_id = ?", caseID).
		Order("id DESC").
		Find(&rawList)

	result := make([]map[string]interface{}, 0, len(rawList))
	for _, item := range rawList {
		converted := make(map[string]interface{})
		for k, v := range item {
			switch k {
			case "protocol_no":
				converted["protocolNo"] = v
			case "party_a_name":
				converted["partyAName"] = v
			case "party_b_name":
				converted["partyBName"] = v
			case "mediator_name":
				converted["mediatorName"] = v
			case "agreement_items":
				converted["agreementItems"] = v
			case "breach_clause":
				converted["breachClause"] = v
			case "effective_date":
				converted["effectiveDate"] = v
			case "is_signed":
				converted["isSigned"] = v
			case "signed_at":
				converted["signedAt"] = v
			case "file_url":
				converted["fileUrl"] = v
			case "created_by":
				converted["createdBy"] = v
			case "is_ai_generated":
				converted["isAIGenerated"] = v
			case "ai_generated_at":
				converted["aiGeneratedAt"] = v
			case "is_adopted":
				converted["isAdopted"] = v
			case "adopted_by":
				converted["adoptedBy"] = v
			case "adopted_at":
				converted["adoptedAt"] = v
			case "created_at":
				converted["createdAt"] = v
			case "updated_at":
				converted["updatedAt"] = v
			case "deleted_at":
				converted["deletedAt"] = v
			default:
				converted[k] = v
			}
		}
		result = append(result, converted)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func AdoptMediationProtocol(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	protocolID, _ := strconv.ParseInt(c.Param("protocolId"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var protocol struct {
		IsSigned       int32  `gorm:"column:is_signed"`
		CaseID         int64  `gorm:"column:case_id"`
		Content        string `gorm:"column:content"`
		AgreementItems string `gorm:"column:agreement_items"`
	}
	database.GetDB().Table("mediation_protocol").
		Where("id = ?", protocolID).
		First(&protocol)

	if protocol.CaseID != caseID {
		c.JSON(http.StatusBadRequest, response.BadRequest("协议与案件不匹配"))
		return
	}

	updates := map[string]interface{}{
		"adopted_by":   userInfo.UserID,
		"adopted_at":   time.Now(),
		"is_adopted":   1,
	}

	tx := database.GetDB().Begin()
	tx.Table("mediation_protocol").Where("id = ?", protocolID).Updates(updates)

	tx.Table("dispute_case").
		Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"agreement_content": protocol.Content,
		})

	history := map[string]interface{}{
		"case_id":          caseID,
		"operation_type":   "PROTOCOL_ADOPT",
		"operation_detail": fmt.Sprintf("采用AI生成的调解协议，协议ID: %d", protocolID),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)
	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "协议已采用，已同步至案件协议内容"))
}
