package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/blockchain"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/fadada"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type EsignCreateRequest struct {
	CaseID       int64   `json:"caseId" binding:"required"`
	DocType      int32   `json:"docType" binding:"required"`
	DocTitle     string  `json:"docTitle" binding:"required"`
	DocContent   string  `json:"docContent"`
	DocURL       string  `json:"docUrl"`
	TemplateID   int64   `json:"templateId"`
	SignerIDs    []int64 `json:"signerIds" binding:"required"`
	ExpireHours  int32   `json:"expireHours"`
	NeedNotify   bool    `json:"needNotify"`
	NotifyType   string  `json:"notifyType"`
	CrossPageSeal bool   `json:"crossPageSeal"`
}

type EsignSignRequest struct {
	RecordID     int64  `json:"recordId" binding:"required"`
	SignPassword string `json:"signPassword"`
	VerifyCode   string `json:"verifyCode"`
}

func CreateEsignFlow(ctx context.Context, c *app.RequestContext) {
	var req EsignCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo     string `gorm:"column:case_no"`
		Title      string `gorm:"column:title"`
		Status     int32  `gorm:"column:status"`
		MediatorID int64  `gorm:"column:mediator_id"`
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
	crossPageSeal := int32(0)
	if req.CrossPageSeal {
		crossPageSeal = 1
	}

	tx := database.GetDB().Begin()

	record := map[string]interface{}{
		"id":               utils.GenerateID(),
		"flow_id":          flowID,
		"case_id":          req.CaseID,
		"case_no":          caseData.CaseNo,
		"doc_type":         req.DocType,
		"doc_title":        req.DocTitle,
		"doc_content":      req.DocContent,
		"doc_url":          req.DocURL,
		"template_id":      req.TemplateID,
		"signer_count":     len(req.SignerIDs),
		"signed_count":     0,
		"status":           constants.EsignStatusPending,
		"cross_page_seal":  crossPageSeal,
		"expire_time":      time.Now().Add(time.Duration(expireHours) * time.Hour),
		"creator_id":       userInfo.UserID,
		"creator_name":     userInfo.RealName,
		"organization_id":  userInfo.OrganizationID,
	}

	if err := tx.Table("esign_record").Create(record).Error; err != nil {
		tx.Rollback()
		logger.Error("Create esign record failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建签署流程失败"))
		return
	}

	var signers []map[string]interface{}
	var signerNames []string
	var fadadaSigners []fadada.FaDaDaSigner

	for i, sid := range req.SignerIDs {
		var user struct {
			RealName string `gorm:"column:real_name"`
			Phone    string `gorm:"column:phone"`
			IDCard   string `gorm:"column:id_card"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name, phone, id_card").
			Where("id = ?", sid).
			First(&user)

		signers = append(signers, map[string]interface{}{
			"flow_id":      flowID,
			"user_id":      sid,
			"user_name":    user.RealName,
			"user_phone":   user.Phone,
			"id_card":      user.IDCard,
			"sign_order":   i + 1,
			"sign_status":  constants.EsignSignerStatusPending,
			"notify_status": constants.EsignNotifyStatusNone,
		})

		signerNames = append(signerNames, user.RealName)

		fadadaSigners = append(fadadaSigners, fadada.FaDaDaSigner{
			CustomerID:   fmt.Sprintf("USR%d", sid),
			CustomerName: user.RealName,
			SignOrder:    i + 1,
			SignType:     "1",
			AutoSign:     false,
		})
	}

	if len(signers) > 0 {
		tx.Table("esign_signer").Create(signers)
	}

	fadadaFlowID := ""
	fadadaQRCodeURL := ""
	fadadaClient := fadada.GetClient()
	if fadadaClient != nil && req.DocURL != "" {
		fadadaResult, err := fadadaClient.CreateSignFlow(&fadada.CreateSignFlowReq{
			DocTitle:      req.DocTitle,
			DocURL:        req.DocURL,
			SignerIDs:     fadadaSigners,
			ExpireHours:   int(expireHours),
			CrossPageSeal: req.CrossPageSeal,
		})
		if err != nil {
			logger.Error("FaDaDa create sign flow failed", logger.Error(err))
		} else {
			fadadaFlowID = fadadaResult.FlowID
			fadadaQRCodeURL = fadadaResult.QRCodeURL

			tx.Table("esign_record").
				Where("flow_id = ?", flowID).
				Updates(map[string]interface{}{
					"fadada_flow_id": fadadaFlowID,
				})

			for i, sid := range req.SignerIDs {
				signURL := fmt.Sprintf("%s&signer=%d", fadadaResult.ShortURL, i+1)
				tx.Table("esign_signer").
					Where("flow_id = ? AND user_id = ?", flowID, sid).
					Update("fadada_sign_url", signURL)
			}
		}
	}

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "ESIGN_CREATE",
		"operation_detail": fmt.Sprintf("创建电子签署: %s，签署人: %v", req.DocTitle, signerNames),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	if req.NeedNotify {
		notifyType := req.NotifyType
		if notifyType == "" {
			notifyType = "sms"
		}
		for _, sid := range req.SignerIDs {
			var user struct {
				RealName string `gorm:"column:real_name"`
				Phone    string `gorm:"column:phone"`
			}
			database.GetDB().Table("sys_user").
				Select("real_name, phone").
				Where("id = ?", sid).
				First(&user)

			verifyCode := utils.GenerateRandomNumber(6)
			cache.Set(ctx, fmt.Sprintf("esign:verify:%s:%d", flowID, sid), verifyCode, 30*time.Minute)

			go func(phone, name, code string, uid int64) {
				msg := map[string]interface{}{
					"caseId":      req.CaseID,
					"caseNo":      caseData.CaseNo,
					"caseTitle":   caseData.Title,
					"flowId":      flowID,
					"docTitle":    req.DocTitle,
					"verifyCode":  code,
					"expireHours": expireHours,
					"signer":      name,
					"phone":       phone,
					"userId":      uid,
					"notifyType":  notifyType,
					"fadadaFlowId": fadadaFlowID,
				}
				mq.SendMessage(constants.MQTopicEsignNotify, msg)

				now := time.Now()
				notifyStatus := constants.EsignNotifyStatusSMS
				if notifyType == "wechat" {
					notifyStatus = constants.EsignNotifyStatusWechat
				} else if notifyType == "all" {
					notifyStatus = constants.EsignNotifyStatusAll
				}
				database.GetDB().Table("esign_signer").
					Where("flow_id = ? AND user_id = ?", flowID, uid).
					Updates(map[string]interface{}{
						"notify_status": notifyStatus,
						"notify_sent_at": now,
					})
			}(user.Phone, user.RealName, verifyCode, sid)
		}
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"flowId":        flowID,
		"docTitle":      req.DocTitle,
		"signerCount":   len(req.SignerIDs),
		"expireTime":    time.Now().Add(time.Duration(expireHours) * time.Hour).Format("2006-01-02 15:04:05"),
		"fadadaFlowId":  fadadaFlowID,
		"qrcodeUrl":     fadadaQRCodeURL,
		"crossPageSeal": crossPageSeal,
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
		constants.EsignStatusDraft:     "草稿",
		constants.EsignStatusPending:   "待签署",
		constants.EsignStatusSigning:   "签署中",
		constants.EsignStatusCompleted: "已完成",
		constants.EsignStatusExpired:   "已过期",
		constants.EsignStatusRevoked:   "已撤销",
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
		constants.EsignSignerStatusPending:  "待签署",
		constants.EsignSignerStatusSigned:   "已签署",
		constants.EsignSignerStatusRejected: "已拒绝",
	}

	notifyStatusMap := map[int]string{
		constants.EsignNotifyStatusNone:   "未通知",
		constants.EsignNotifyStatusSMS:    "短信通知",
		constants.EsignNotifyStatusWechat: "微信通知",
		constants.EsignNotifyStatusAll:    "全部通知",
	}

	for _, item := range signers {
		if s, ok := item["sign_status"].(int); ok {
			item["sign_status_name"] = signStatusMap[s]
		}
		if ns, ok := item["notify_status"].(int); ok {
			item["notify_status_name"] = notifyStatusMap[ns]
		}
	}

	record["signers"] = signers

	var bcCert model.BlockchainCertificate
	database.GetDB().Where("flow_id = ? AND deleted_at IS NULL", flowID).First(&bcCert)
	if bcCert.ID > 0 {
		record["blockchain_cert"] = map[string]interface{}{
			"certNo":        bcCert.CertNo,
			"txId":          bcCert.TxID,
			"blockHeight":   bcCert.BlockHeight,
			"onChainTime":   bcCert.OnChainTime,
			"certUrl":       bcCert.CertURL,
			"qrcodeUrl":     bcCert.QRCodeURL,
			"verifyUrl":     bcCert.VerifyURL,
			"status":        bcCert.Status,
			"evidenceHash":  bcCert.EvidenceHash,
		}
	}

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
		ID              int64     `gorm:"column:id"`
		CaseID          int64     `gorm:"column:case_id"`
		CaseNo          string    `gorm:"column:case_no"`
		DocTitle        string    `gorm:"column:doc_title"`
		Status          int32     `gorm:"column:status"`
		SignerCount     int32     `gorm:"column:signer_count"`
		SignedCount     int32     `gorm:"column:signed_count"`
		ExpireTime      time.Time `gorm:"column:expire_time"`
		FaDaDaFlowID    string    `gorm:"column:fadada_flow_id"`
		DocURL          string    `gorm:"column:doc_url"`
	}

	result := database.GetDB().Table("esign_record").
		Where("id = ?", req.RecordID).
		First(&record)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("签署记录不存在"))
		return
	}

	if record.Status == constants.EsignStatusCompleted {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已完成"))
		return
	}
	if record.Status == constants.EsignStatusExpired {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已过期"))
		return
	}
	if record.Status == constants.EsignStatusRevoked {
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

	if signer.SignStatus == constants.EsignSignerStatusSigned {
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
			"sign_status":   constants.EsignSignerStatusSigned,
			"sign_time":     now,
			"signature_url": signatureURL,
			"sign_ip":       c.ClientIP(),
		})

	newSignedCount := record.SignedCount + 1
	status := int32(constants.EsignStatusSigning)
	allSigned := newSignedCount >= record.SignerCount
	if allSigned {
		status = constants.EsignStatusCompleted
	}

	tx.Table("esign_record").
		Where("id = ?", record.ID).
		Updates(map[string]interface{}{
			"signed_count":   newSignedCount,
			"status":         status,
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

	if allSigned && record.FaDaDaFlowID != "" {
		go func() {
			fadadaClient := fadada.GetClient()
			if fadadaClient != nil {
				signedDocURL, err := fadadaClient.GetSignedDocument(record.FaDaDaFlowID)
				if err != nil {
					logger.Error("Download signed document from FaDaDa failed", logger.Error(err))
				} else if signedDocURL != "" {
					database.GetDB().Table("esign_record").
						Where("id = ?", record.ID).
						Update("signed_document_url", signedDocURL)

					go storeSignedDocToBlockchain(record.ID, record.CaseID, record.CaseNo, record.DocTitle, signedDocURL, userInfo.UserID)
				}
			}
		}()
	}

	go func() {
		notifyMsg := map[string]interface{}{
			"type":       "esign_sign_progress",
			"caseId":     record.CaseID,
			"caseNo":     record.CaseNo,
			"docTitle":   record.DocTitle,
			"signer":     signer.UserName,
			"signedCount": newSignedCount,
			"totalCount": record.SignerCount,
			"allSigned":  allSigned,
			"signTime":   now.Format("2006-01-02 15:04:05"),
		}
		mq.SendMessage(constants.MQTopicEsignNotify, notifyMsg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"signatureUrl": signatureURL,
		"signTime":     now.Format("2006-01-02 15:04:05"),
		"allSigned":    allSigned,
	}, "签署成功"))
}

func storeSignedDocToBlockchain(recordID int64, caseID int64, caseNo, docTitle, docURL string, creatorID int64) {
	bcClient := blockchain.GetClient()
	if bcClient == nil {
		logger.Error("Blockchain client not initialized")
		return
	}

	pdfHash := ""
	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   fmt.Sprintf("ESIGN-%d", recordID),
		EvidenceType: constants.BCTypeEsignDocument,
		EvidenceHash: pdfHash,
		EvidenceName: docTitle,
		Description:  fmt.Sprintf("调解协议书电子签章文档上链: %s", docTitle),
		Metadata:     fmt.Sprintf(`{"caseId":%d,"caseNo":"%s","recordId":%d,"docUrl":"%s"}`, caseID, caseNo, recordID, docURL),
	})

	if err != nil {
		logger.Error("Store signed doc to blockchain failed", logger.Error(err))
		database.GetDB().Table("esign_record").
			Where("id = ?", recordID).
			Update("bc_status", constants.BCStatusFailed)
		return
	}

	now := time.Now()
	database.GetDB().Table("esign_record").
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"bc_cert_no":      storeResult.CertNo,
			"bc_tx_id":        storeResult.TxID,
			"bc_on_chain_time": now,
			"bc_status":       constants.BCStatusOnChain,
		})

	certNo := storeResult.CertNo
	certResult, err := bcClient.GetCertificate(certNo)
	if err != nil {
		logger.Error("Get blockchain certificate failed", logger.Error(err))
	} else {
		cert := &model.BlockchainCertificate{
			BaseModel: model.BaseModel{
				ID: utils.GenerateID(),
			},
			CertNo:       certNo,
			EvidenceID:   fmt.Sprintf("ESIGN-%d", recordID),
			EvidenceType: constants.BCTypeEsignDocument,
			EvidenceName: docTitle,
			EvidenceHash: pdfHash,
			CaseID:       caseID,
			FlowID:       fmt.Sprintf("ES%d", recordID),
			TxID:         storeResult.TxID,
			BlockHeight:  storeResult.BlockHeight,
			OnChainTime:  &now,
			CertURL:      certResult.CertURL,
			QRCodeURL:    certResult.QRCodeURL,
			VerifyURL:    certResult.VerifyURL,
			Status:       constants.BCStatusOnChain,
			Metadata:     fmt.Sprintf(`{"caseNo":"%s","docUrl":"%s"}`, caseNo, docURL),
			CreatedBy:    creatorID,
		}
		database.GetDB().Create(cert)
	}

	mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
		"type":       "esign_blockchain_stored",
		"caseId":     caseID,
		"caseNo":     caseNo,
		"docTitle":   docTitle,
		"certNo":     certNo,
		"txId":       storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
	})
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
		ID           int64  `gorm:"column:id"`
		CaseID       int64  `gorm:"column:case_id"`
		CaseNo       string `gorm:"column:case_no"`
		DocTitle     string `gorm:"column:doc_title"`
		Status       int32  `gorm:"column:status"`
		CreatorID    int64  `gorm:"column:creator_id"`
		FaDaDaFlowID string `gorm:"column:fadada_flow_id"`
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

	if record.Status == constants.EsignStatusCompleted {
		c.JSON(http.StatusBadRequest, response.BadRequest("已完成的流程不能撤销"))
		return
	}

	if record.FaDaDaFlowID != "" {
		fadadaClient := fadada.GetClient()
		if fadadaClient != nil {
			if err := fadadaClient.RevokeSignFlow(record.FaDaDaFlowID, req.Reason); err != nil {
				logger.Error("FaDaDa revoke sign flow failed", logger.Error(err))
			}
		}
	}

	tx := database.GetDB().Begin()

	tx.Table("esign_record").
		Where("id = ?", record.ID).
		Updates(map[string]interface{}{
			"status":        constants.EsignStatusRevoked,
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
				"notifyType":  "sms",
			}
			mq.SendMessage(constants.MQTopicEsignNotify, msg)
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
		mq.SendMessage(constants.MQTopicEsignNotify, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "验证码已发送"))
}

func FaDaDaCallback(ctx context.Context, c *app.RequestContext) {
	timestamp := c.Query("timestamp")
	flowID := c.Query("flowId")
	status := c.Query("status")
	sign := c.Query("sign")

	fadadaClient := fadada.GetClient()
	if fadadaClient == nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("法大大客户端未初始化"))
		return
	}

	if !fadadaClient.VerifyCallback(timestamp, flowID, status, sign) {
		c.JSON(http.StatusForbidden, response.Forbidden("签名验证失败"))
		return
	}

	var record struct {
		ID          int64  `gorm:"column:id"`
		CaseID      int64  `gorm:"column:case_id"`
		CaseNo      string `gorm:"column:case_no"`
		DocTitle    string `gorm:"column:doc_title"`
		SignerCount int32  `gorm:"column:signer_count"`
		Status      int32  `gorm:"column:status"`
	}

	database.GetDB().Table("esign_record").
		Where("fadada_flow_id = ?", flowID).
		First(&record)

	if record.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	if status == "2" {
		progress, err := fadadaClient.GetSignProgress(flowID)
		if err != nil {
			logger.Error("Get FaDaDa sign progress failed", logger.Error(err))
			c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "msg": "success"})
			return
		}

		signedCount := 0
		for _, s := range progress.Signers {
			if s.Status == "2" {
				signedCount++
				database.GetDB().Table("esign_signer").
					Where("flow_id = ? AND user_name = ?", record.ID, s.CustomerName).
					Updates(map[string]interface{}{
						"sign_status": constants.EsignSignerStatusSigned,
						"sign_time":   time.Now(),
					})
			}
		}

		newStatus := int32(constants.EsignStatusSigning)
		if signedCount >= int(record.SignerCount) {
			newStatus = constants.EsignStatusCompleted
		}

		database.GetDB().Table("esign_record").
			Where("id = ?", record.ID).
			Updates(map[string]interface{}{
				"status":        newStatus,
				"signed_count":  signedCount,
				"last_sign_time": time.Now(),
			})

		mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
			"type":        "esign_fadada_callback",
			"caseId":      record.CaseID,
			"caseNo":      record.CaseNo,
			"docTitle":    record.DocTitle,
			"signedCount": signedCount,
			"totalCount":  record.SignerCount,
			"allSigned":   signedCount >= int(record.SignerCount),
		})

		if newStatus == constants.EsignStatusCompleted {
			go func() {
				signedDocURL, _ := fadadaClient.GetSignedDocument(flowID)
				if signedDocURL != "" {
					database.GetDB().Table("esign_record").
						Where("id = ?", record.ID).
						Update("signed_document_url", signedDocURL)
				}
			}()
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "msg": "success"})
}

func GetEsignProgress(ctx context.Context, c *app.RequestContext) {
	flowID := c.Param("flowId")

	var record struct {
		ID          int64  `gorm:"column:id"`
		Status      int32  `gorm:"column:status"`
		SignerCount int32  `gorm:"column:signer_count"`
		SignedCount int32  `gorm:"column:signed_count"`
		FaDaDaFlowID string `gorm:"column:fadada_flow_id"`
	}

	database.GetDB().Table("esign_record").
		Where("flow_id = ?", flowID).
		First(&record)

	if record.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var signers []map[string]interface{}
	database.GetDB().Table("esign_signer").
		Where("flow_id = ?", flowID).
		Order("sign_order ASC").
		Find(&signers)

	signStatusMap := map[int]string{
		constants.EsignSignerStatusPending:  "待签署",
		constants.EsignSignerStatusSigned:   "已签署",
		constants.EsignSignerStatusRejected: "已拒绝",
	}

	for _, item := range signers {
		if s, ok := item["sign_status"].(int); ok {
			item["sign_status_name"] = signStatusMap[s]
		}
	}

	result := map[string]interface{}{
		"flowId":      flowID,
		"status":      record.Status,
		"signedCount": record.SignedCount,
		"totalCount":  record.SignerCount,
		"signers":     signers,
	}

	if record.FaDaDaFlowID != "" {
		fadadaClient := fadada.GetClient()
		if fadadaClient != nil {
			progress, err := fadadaClient.GetSignProgress(record.FaDaDaFlowID)
			if err == nil {
				result["fadadaProgress"] = progress
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}
