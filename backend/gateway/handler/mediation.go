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
		"is_key_record":       1,
	}

	tx := database.GetDB().Begin()
	tx.Table("dispute_mediation_record").Create(record)

	if req.Result > 0 {
		updates := map[string]interface{}{
			"mediation_result": req.Result,
		}
		if req.AgreementContent != "" {
			updates["agreement_content"] = req.AgreementContent
		}
		if req.Result == constants.MediationResultSuccess {
			updates["mediation_end_time"] = time.Now()
		}
		tx.Table("dispute_case").Where("id = ?", caseID).Updates(updates)
	}

	history := map[string]interface{}{
		"case_id":       caseID,
		"case_no":       caseData.CaseNo,
		"operation_type": "MEDIATION_RECORD",
		"operation_detail": fmt.Sprintf("记录调解信息，时长: %d分钟，结果: %d", req.MediationDuration, req.Result),
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

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

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": recordID,
	}, "调解记录创建成功"))
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
	}
	database.GetDB().Table("dispute_mediation_record").
		Select("mediator_id").
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

	database.GetDB().Table("dispute_mediation_record").
		Where("id = ?", recordID).
		Updates(updates)

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "调解记录更新成功"))
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
