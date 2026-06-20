package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

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

type CallbackListRequest struct {
	model.BaseQuery
	CaseID     int64  `form:"caseId"`
	Status     int32  `form:"status"`
	CallStatus int32  `form:"callStatus"`
	Keyword    string `form:"keyword"`
	model.DateRangeQuery
}

func GetCallbackList(ctx context.Context, c *app.RequestContext) {
	var req CallbackListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	db := database.GetDB().Table("callback_record cr").
		Select("cr.*, dc.case_no, dc.title as case_title").
		Joins("LEFT JOIN dispute_case dc ON cr.case_id = dc.id").
		Where("cr.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("dc.mediator_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("dc.org_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
			userInfo.OrganizationID, userInfo.OrganizationID)
	}

	if req.CaseID > 0 {
		db = db.Where("cr.case_id = ?", req.CaseID)
	}
	if req.Status > 0 {
		db = db.Where("cr.status = ?", req.Status)
	}
	if req.CallStatus > 0 {
		db = db.Where("cr.call_status = ?", req.CallStatus)
	}
	if req.Keyword != "" {
		db = db.Where("cr.case_no LIKE ? OR cr.applicant_name LIKE ? OR cr.applicant_phone LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.StartTime != "" {
		db = db.Where("cr.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("cr.created_at <= ?", req.EndTime)
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
			item["status_name"] = model.CallbackStatusMap[int(status)]
		}
		if callStatus, ok := item["call_status"].(int32); ok {
			item["call_status_name"] = model.CallStatusMap[int(callStatus)]
		}
		if emotion, ok := item["emotion"].(string); ok {
			item["emotion_name"] = model.EmotionMap[emotion]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetCallbackDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	record, err := callbackService.GetCallbackDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("回访记录不存在"))
		return
	}

	result := make(map[string]interface{})
	recordJSON, _ := json.Marshal(record)
	json.Unmarshal(recordJSON, &result)

	result["status_name"] = model.CallbackStatusMap[int(record.Status)]
	result["call_status_name"] = model.CallStatusMap[int(record.CallStatus)]
	if record.Emotion != "" {
		result["emotion_name"] = model.EmotionMap[record.Emotion]
	}

	if record.SentimentResult != "" {
		var sentimentResult map[string]interface{}
		if err := json.Unmarshal([]byte(record.SentimentResult), &sentimentResult); err == nil {
			result["sentiment_detail"] = sentimentResult
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func CreateCallback(ctx context.Context, c *app.RequestContext) {
	var req struct {
		CaseID int64 `json:"caseId" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	callbackService := service.CallbackServiceInst()
	record, err := callbackService.CreateCallbackRecord(ctx, req.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(record, "回访任务创建成功"))
}

func InitiateCallback(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	if err := callbackService.InitiateCall(ctx, id); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "回访电话已发起"))
}

func RetryCallback(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	if err := callbackService.RetryCallback(ctx, id); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "回访重试已发起"))
}

func CancelCallback(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	if err := callbackService.CancelCallback(ctx, id); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "回访已取消"))
}

func GetCallbacksByCase(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("caseId"), 10, 64)

	callbackService := service.CallbackServiceInst()
	records, err := callbackService.GetCallbackList(ctx, caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("查询失败"))
		return
	}

	result := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		item := make(map[string]interface{})
		recordJSON, _ := json.Marshal(record)
		json.Unmarshal(recordJSON, &item)
		item["status_name"] = model.CallbackStatusMap[int(record.Status)]
		item["call_status_name"] = model.CallStatusMap[int(record.CallStatus)]
		if record.Emotion != "" {
			item["emotion_name"] = model.EmotionMap[record.Emotion]
		}
		result = append(result, item)
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func AliyunVoiceCallback(ctx context.Context, c *app.RequestContext) {
	body, err := io.ReadAll(c.Request.Body())
	if err != nil {
		logger.Error("Read Aliyun voice callback body failed", logger.Error(err))
		c.JSON(http.StatusBadRequest, response.BadRequest("读取请求失败"))
		return
	}

	logger.Info("Received Aliyun voice callback",
		zap.String("body", string(body)),
	)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		logger.Warn("Parse Aliyun voice callback failed",
			zap.String("body", string(body)),
			logger.Error(err),
		)
		c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
		return
	}

	callbackService := service.CallbackServiceInst()
	if err := callbackService.HandleCallback(ctx, data); err != nil {
		logger.Warn("Handle Aliyun voice callback failed", logger.Error(err))
	}

	c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
}

func RefreshCallbackResult(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	record, err := callbackService.GetCallbackDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("回访记录不存在"))
		return
	}

	if record.CallID == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("该回访没有通话记录"))
		return
	}

	if err := callbackService.ProcessCallResult(ctx, record.CallID); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("刷新失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "回访结果已刷新"))
}

func DownloadAndArchiveRecording(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	callbackService := service.CallbackServiceInst()
	if err := callbackService.DownloadAndArchiveRecording(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("存档失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "录音已归档"))
}
