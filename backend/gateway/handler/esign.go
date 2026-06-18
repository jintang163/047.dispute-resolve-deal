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
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type EsignCreateRequest struct {
	CaseID          int64   `json:"caseId" binding:"required"`
	DocType         int32   `json:"docType" binding:"required"`
	DocTitle        string  `json:"docTitle" binding:"required"`
	DocContent      string  `json:"docContent"`
	TemplateID      int64   `json:"templateId"`
	SignerIDs       []int64 `json:"signerIds" binding:"required"`
	ExpireHours     int32   `json:"expireHours"`
	NeedNotify      bool    `json:"needNotify"`
}

type EsignSignRequest struct {
	RecordID    int64  `json:"recordId" binding:"required"`
	SignPassword string `json:"signPassword"`
	VerifyCode  string `json:"verifyCode"`
}

func CreateEsignFlow(ctx context.Context, c *app.RequestContext) {
	var req EsignCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo   string `gorm:"column:case_no"`
		Title    string `gorm:"column:title"`
		Status   int32  `gorm:"column:status"`
		MediatorID int64 `gorm:"column:mediator_id"`
	}

	result := database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).
		First(&caseData)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	if caseData.Status >= constants.CaseStatusClosed {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件已结案，无法创建签署流程"))
		return
	}

	if caseData.MediatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有案件调解员或领导可以创建签署流程"))
		return
	}

	expireHours := req.ExpireHours
	if expireHours == 0 {
		expireHours = 72
	}

	flowID := fmt.Sprintf("ES%s", utils.GenerateIDStr())

	tx := database.GetDB().Begin()

	record := map[string]interface{}{
		"id":              utils.GenerateID(),
		"flow_id":         flowID,
		"case_id":         req.CaseID,
		"case_no":         caseData.CaseNo,
		"doc_type":        req.DocType,
		"doc_title":       req.DocTitle,
		"doc_content":     req.DocContent,
		"template_id":     req.TemplateID,
		"signer_count":    len(req.SignerIDs),
		"signed_count":    0,
		"status":          10,
		"expire_time":     time.Now().Add(time.Duration(expireHours) * time.Hour),
		"creator_id":      userInfo.UserID,
		"creator_name":    userInfo.RealName,
		"organization_id": userInfo.OrganizationID,
	}

	if err := tx.Table("esign_record").Create(record).Error; err != nil {
		tx.Rollback()
		logger.Error("Create esign record failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建签署流程失败"))
		return
	}

	var signers []map[string]interface{}
	var signerNames []string

	for i, sid := range req.SignerIDs {
		var user struct {
			RealName  string `gorm:"column:real_name"`
			Phone     string `gorm:"column:phone"`
			IDCard    string `gorm:"column:id_card"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name, phone, id_card").
			Where("id = ?", sid).
			First(&user)

		signers = append(signers, map[string]interface{}{
			"flow_id":    flowID,
			"user_id":    sid,
			"user_name":  user.RealName,
			"user_phone": user.Phone,
			"id_card":    user.IDCard,
			"sign_order": i + 1,
			"sign_status": 10,
		})

		signerNames = append(signerNames, user.RealName)

		if req.NeedNotify {
			verifyCode := utils.GenerateRandomNumber(6)
			cache.Set(ctx, fmt.Sprintf("esign:verify:%s:%d", flowID, sid), verifyCode, 30*time.Minute)

			go func(phone, name, code string) {
				msg := map[string]interface{}{
					"caseId":     req.CaseID,
					"caseNo":     caseData.CaseNo,
					"caseTitle":  caseData.Title,
					"flowId":     flowID,
					"docTitle":   req.DocTitle,
					"verifyCode": code,
					"expireHours": expireHours,
					"signer":     name,
					"phone":      phone,
					"notifyType": "sms",
				}
				mq.SendMessage(constants.MQTopicNotification, msg)
			}(user.Phone, user.RealName, verifyCode)
		}
	}

	if len(signers) > 0 {
		tx.Table("esign_signer").Create(signers)
	}

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "ESIGN_CREATE",
		"operation_detail": fmt.Sprintf("创建电子签署: %s，签署人: %s", 
			req.DocTitle, fmt.Sprintf("%v", signerNames)),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"flowId":      flowID,
		"docTitle":    req.DocTitle,
		"signerCount": len(req.SignerIDs),
		"expireTime":  time.Now().Add(time.Duration(expireHours) * time.Hour).Format("2006-01-02 15:04:05"),
	}, "签署流程创建成功"))
}

func GetEsignList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	db := database.GetDB().Table("esign_record er").
		Select("er.*, dc.title as case_title, dc.case_no").
		Joins("LEFT JOIN dispute_case dc ON er.case_id = dc.id").
		Where("er.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("er.creator_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("er.organization_id = ?", userInfo.OrganizationID)
	}

	if caseID > 0 {
		db = db.Where("er.case_id = ?", caseID)
	}
	if status > 0 {
		db = db.Where("er.status = ?", status)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("er.created_at DESC").
		Limit(50).
		Find(&list)

	statusMap := map[int]string{
		10: "待签署",
		20: "签署中",
		30: "已完成",
		40: "已过期",
		50: "已撤销",
	}

	for _, item := range list {
		if s, ok := item["status"].(int); ok {
			item["status_name"] = statusMap[s]
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetEsignDetail(ctx context.Context, c *app.RequestContext) {
	flowID := c.Param("flowId")
	userInfo := middleware.GetUserInfo(c)

	var record map[string]interface{}
	result := database.GetDB().Table("esign_record er").
		Select("er.*, dc.title as case_title, dc.case_no").
		Joins("LEFT JOIN dispute_case dc ON er.case_id = dc.id").
		Where("er.flow_id = ?", flowID).
		Find(&record)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var signers []map[string]interface{}
	database.GetDB().Table("esign_signer").
		Where("flow_id = ?", flowID).
		Order("sign_order ASC").
		Find(&signers)

	signStatusMap := map[int]string{
		10: "待签署",
		20: "已签署",
		30: "已拒绝",
	}

	for _, item := range signers {
		if s, ok := item["sign_status"].(int); ok {
			item["sign_status_name"] = signStatusMap[s]
		}
	}

	record["signers"] = signers

	c.JSON(http.StatusOK, response.Success(record))
}

func SignDocument(ctx context.Context, c *app.RequestContext) {
	var req EsignSignRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var record struct {
		ID            int64  `gorm:"column:id"`
		CaseID        int64  `gorm:"column:case_id"`
		CaseNo        string `gorm:"column:case_no"`
		DocTitle      string `gorm:"column:doc_title"`
		Status        int32  `gorm:"column:status"`
		SignerCount   int32  `gorm:"column:signer_count"`
		SignedCount   int32  `gorm:"column:signed_count"`
		ExpireTime    time.Time `gorm:"column:expire_time"`
	}

	result := database.GetDB().Table("esign_record").
		Where("id = ?", req.RecordID).
		First(&record)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("签署记录不存在"))
		return
	}

	if record.Status == 30 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已完成"))
		return
	}
	if record.Status == 40 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已过期"))
		return
	}
	if record.Status == 50 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已撤销"))
		return
	}

	if time.Now().After(record.ExpireTime) {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已过期"))
		return
	}

	var signer struct {
		ID         int64  `gorm:"column:id"`
		SignStatus int32  `gorm:"column:sign_status"`
		SignOrder  int32  `gorm:"column:sign_order"`
		UserName   string `gorm:"column:user_name"`
	}

	database.GetDB().Table("esign_signer").
		Where("flow_id = ? AND user_id = ?", 
			database.GetDB().Table("esign_record").Select("flow_id").Where("id = ?", req.RecordID),
			userInfo.UserID).
		First(&signer)

	if signer.ID == 0 {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该文件的签署人"))
		return
	}

	if signer.SignStatus == 20 {
		c.JSON(http.StatusBadRequest, response.BadRequest("您已签署该文件"))
		return
	}

	cacheKey := fmt.Sprintf("esign:verify:%s:%d", 
		database.GetDB().Table("esign_record").Select("flow_id").Where("id = ?", req.RecordID),
		userInfo.UserID)
	
	if req.VerifyCode != "" {
		cachedCode, err := cache.Get(ctx, cacheKey)
		if err != nil || cachedCode != req.VerifyCode {
			c.JSON(http.StatusBadRequest, response.BadRequest("验证码错误或已过期"))
			return
		}
	}

	signatureURL := fmt.Sprintf("/api/v1/public/signature/%s.png", utils.GenerateUUID())

	tx := database.GetDB().Begin()

	now := time.Now()
	tx.Table("esign_signer").
		Where("id = ?", signer.ID).
		Updates(map[string]interface{}{
			"sign_status":   20,
			"sign_time":     now,
			"signature_url": signatureURL,
			"sign_ip":       c.ClientIP(),
		})

	newSignedCount := record.SignedCount + 1
	status := int32(20)
	if newSignedCount >= record.SignerCount {
		status = 30
	}

	tx.Table("esign_record").
		Where("id = ?", record.ID).
		Updates(map[string]interface{}{
			"signed_count": newSignedCount,
			"status":       status,
			"last_sign_time": now,
		})

	history := map[string]interface{}{
		"case_id":          record.CaseID,
		"case_no":          record.CaseNo,
		"operation_type":   "ESIGN_SIGN",
		"operation_detail": fmt.Sprintf("签署文件: %s，签署人: %s", record.DocTitle, signer.UserName),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"signatureUrl": signatureURL,
		"signTime":     now.Format("2006-01-02 15:04:05"),
		"allSigned":    newSignedCount >= record.SignerCount,
	}, "签署成功"))
}

func RevokeEsignFlow(ctx context.Context, c *app.RequestContext) {
	flowID := c.Param("flowId")
	userInfo := middleware.GetUserInfo(c)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var record struct {
		ID         int64  `gorm:"column:id"`
		CaseID     int64  `gorm:"column:case_id"`
		CaseNo     string `gorm:"column:case_no"`
		DocTitle   string `gorm:"column:doc_title"`
		Status     int32  `gorm:"column:status"`
		CreatorID  int64  `gorm:"column:creator_id"`
	}

	result := database.GetDB().Table("esign_record").
		Where("flow_id = ?", flowID).
		First(&record)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	if record.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有流程创建者或领导可以撤销"))
		return
	}

	if record.Status == 30 {
		c.JSON(http.StatusBadRequest, response.BadRequest("已完成的流程不能撤销"))
	}

	tx := database.GetDB().Begin()

	tx.Table("esign_record").
		Where("id = ?", record.ID).
		Updates(map[string]interface{}{
			"status":       50,
			"revoke_reason": req.Reason,
			"revoke_time":   time.Now(),
			"revoke_by":     userInfo.UserID,
		})

	var signers []map[string]interface{}
	database.GetDB().Table("esign_signer").
		Select("user_id, user_name, user_phone").
		Where("flow_id = ?", flowID).
		Find(&signers)

	for _, s := range signers {
		go func(phone, name string) {
			msg := map[string]interface{}{
				"caseId":      record.CaseID,
				"caseNo":      record.CaseNo,
				"flowId":      flowID,
				"docTitle":    record.DocTitle,
				"revokeReason": req.Reason,
				"revokeBy":    userInfo.RealName,
				"signer":      name,
				"phone":       phone,
			}
			mq.SendMessage(constants.MQTopicNotification, msg)
		}(s["user_phone"].(string), s["user_name"].(string))
	}

	history := map[string]interface{}{
		"case_id":          record.CaseID,
		"case_no":          record.CaseNo,
		"operation_type":   "ESIGN_REVOKE",
		"operation_detail": fmt.Sprintf("撤销签署流程: %s，原因: %s", record.DocTitle, req.Reason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "撤销成功"))
}

func SendEsignVerifyCode(ctx context.Context, c *app.RequestContext) {
	flowID := c.Query("flowId")
	userInfo := middleware.GetUserInfo(c)

	var record struct {
		CaseID   int64  `gorm:"column:case_id"`
		CaseNo   string `gorm:"column:case_no"`
		DocTitle string `gorm:"column:doc_title"`
	}

	result := database.GetDB().Table("esign_record").
		Where("flow_id = ?", flowID).
		First(&record)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var signer struct {
		UserPhone string `gorm:"column:user_phone"`
		UserName  string `gorm:"column:user_name"`
	}

	database.GetDB().Table("esign_signer").
		Where("flow_id = ? AND user_id = ?", flowID, userInfo.UserID).
		First(&signer)

	if signer.UserPhone == "" {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该文件的签署人"))
		return
	}

	verifyCode := utils.GenerateRandomNumber(6)
	cacheKey := fmt.Sprintf("esign:verify:%s:%d", flowID, userInfo.UserID)
	cache.Set(ctx, cacheKey, verifyCode, 30*time.Minute)

	go func() {
		msg := map[string]interface{}{
			"caseId":     record.CaseID,
			"caseNo":     record.CaseNo,
			"flowId":     flowID,
			"docTitle":   record.DocTitle,
			"verifyCode": verifyCode,
			"signer":     signer.UserName,
			"phone":      signer.UserPhone,
			"notifyType": "sms",
		}
		mq.SendMessage(constants.MQTopicNotification, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "验证码已发送"))
}
