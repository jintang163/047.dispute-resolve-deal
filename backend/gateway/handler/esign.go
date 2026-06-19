package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

	"github.com/apache/rocketmq-client-go/v2/primitive"
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

	flowNo := fmt.Sprintf("ES%s", utils.GenerateIDStr())
	crossPageSeal := int32(0)
	if req.CrossPageSeal {
		crossPageSeal = 1
	}

	expireTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	flow := &model.EsignFlow{
		FlowNo:         flowNo,
		CaseID:         req.CaseID,
		CaseNo:         caseData.CaseNo,
		DocType:        req.DocType,
		DocTitle:       req.DocTitle,
		DocContent:     req.DocContent,
		DocURL:         req.DocURL,
		TemplateID:     req.TemplateID,
		TotalSignCount: len(req.SignerIDs),
		SignedCount:    0,
		Status:         constants.EsignStatusPending,
		CrossPageSeal:  crossPageSeal,
		ExpireTime:     &expireTime,
		CreatorID:      userInfo.UserID,
		CreatorName:    userInfo.RealName,
		OrganizationID: userInfo.OrganizationID,
	}

	tx := database.GetDB().Begin()

	if err := tx.Create(flow).Error; err != nil {
		tx.Rollback()
		logger.Error("Create esign flow failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建签署流程失败"))
		return
	}

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

		signer := &model.EsignSigner{
			FlowID:       flow.ID,
			UserID:       sid,
			UserName:     user.RealName,
			UserPhone:    user.Phone,
			IDCard:       user.IDCard,
			SignOrder:    i + 1,
			SignStatus:   constants.EsignSignerStatusPending,
			NotifyStatus: constants.EsignNotifyStatusNone,
		}
		if err := tx.Create(signer).Error; err != nil {
			logger.Error("Create esign signer failed", logger.Error(err))
		}

		signerNames = append(signerNames, user.RealName)

		fadadaSigners = append(fadadaSigners, fadada.FaDaDaSigner{
			CustomerID:   fmt.Sprintf("USR%d", sid),
			CustomerName: user.RealName,
			SignOrder:    i + 1,
			SignType:     "1",
			AutoSign:     false,
		})
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

			tx.Model(&model.EsignFlow{}).Where("id = ?", flow.ID).
				Update("fadada_flow_id", fadadaFlowID)

			for i, sid := range req.SignerIDs {
				signURL := fmt.Sprintf("%s&signer=%d", fadadaResult.ShortURL, i+1)
				tx.Model(&model.EsignSigner{}).
					Where("flow_id = ? AND user_id = ?", flow.ID, sid).
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
			cache.Set(ctx, fmt.Sprintf("esign:verify:%s:%d", flowNo, sid), verifyCode, 30*time.Minute)

			go func(phone, name, code string, uid int64) {
				msg := map[string]interface{}{
					"caseId":       req.CaseID,
					"caseNo":       caseData.CaseNo,
					"caseTitle":    caseData.Title,
					"flowId":       flowNo,
					"docTitle":     req.DocTitle,
					"verifyCode":   code,
					"expireHours":  expireHours,
					"signer":       name,
					"phone":        phone,
					"userId":       uid,
					"notifyType":   notifyType,
					"fadadaFlowId": fadadaFlowID,
				}
				callback := func(ctx context.Context, result *primitive.SendResult, err error) {
					if err != nil {
						logger.Error("Send esign notify message failed", logger.Error(err))
					}
				}
				mq.SendAsyncMessage(constants.MQTopicEsignNotify, msg, callback)

				now := time.Now()
				notifyStatus := constants.EsignNotifyStatusSMS
				if notifyType == "wechat" {
					notifyStatus = constants.EsignNotifyStatusWechat
				} else if notifyType == "all" {
					notifyStatus = constants.EsignNotifyStatusAll
				}
				database.GetDB().Model(&model.EsignSigner{}).
					Where("flow_id = ? AND user_id = ?", flow.ID, uid).
					Updates(map[string]interface{}{
						"notify_status":  notifyStatus,
						"notify_sent_at": now,
					})
			}(user.Phone, user.RealName, verifyCode, sid)
		}
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"flowId":        flowNo,
		"docTitle":      req.DocTitle,
		"signerCount":   len(req.SignerIDs),
		"expireTime":    expireTime.Format("2006-01-02 15:04:05"),
		"fadadaFlowId":  fadadaFlowID,
		"qrcodeUrl":     fadadaQRCodeURL,
		"crossPageSeal": crossPageSeal,
	}, "签署流程创建成功"))
}

func GetEsignList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	db := database.GetDB().Table("esign_flow ef").
		Select("ef.*, dc.title as case_title, dc.case_no").
		Joins("LEFT JOIN dispute_case dc ON ef.case_id = dc.id").
		Where("ef.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("ef.creator_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("ef.organization_id = ?", userInfo.OrganizationID)
	}

	if caseID > 0 {
		db = db.Where("ef.case_id = ?", caseID)
	}
	if status > 0 {
		db = db.Where("ef.status = ?", status)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("ef.created_at DESC").
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

	var flow model.EsignFlow
	result := database.GetDB().Where("flow_no = ?", flowID).First(&flow)
	if result.Error != nil || flow.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var caseInfo struct {
		Title string `gorm:"column:title"`
	}
	database.GetDB().Table("dispute_case").
		Select("title").
		Where("id = ?", flow.CaseID).
		Scan(&caseInfo)

	var signers []model.EsignSigner
	database.GetDB().Where("flow_id = ?", flow.ID).
		Order("sign_order ASC").
		Find(&signers)

	signStatusMap := map[int32]string{
		constants.EsignSignerStatusPending:  "待签署",
		constants.EsignSignerStatusSigned:   "已签署",
		constants.EsignSignerStatusRejected: "已拒绝",
	}

	notifyStatusMap := map[int32]string{
		constants.EsignNotifyStatusNone:   "未通知",
		constants.EsignNotifyStatusSMS:    "短信通知",
		constants.EsignNotifyStatusWechat: "微信通知",
		constants.EsignNotifyStatusAll:    "全部通知",
	}

	type signerVO struct {
		model.EsignSigner
		SignStatusName  string `json:"signStatusName"`
		NotifyStatusName string `json:"notifyStatusName"`
	}

	var signerVOs []signerVO
	for _, s := range signers {
		vo := signerVO{EsignSigner: s}
		vo.SignStatusName = signStatusMap[s.SignStatus]
		vo.NotifyStatusName = notifyStatusMap[s.NotifyStatus]
		signerVOs = append(signerVOs, vo)
	}

	statusMap := map[int32]string{
		constants.EsignStatusDraft:     "草稿",
		constants.EsignStatusPending:   "待签署",
		constants.EsignStatusSigning:   "签署中",
		constants.EsignStatusCompleted: "已完成",
		constants.EsignStatusExpired:   "已过期",
		constants.EsignStatusRevoked:   "已撤销",
	}

	record := map[string]interface{}{
		"id":                flow.ID,
		"flowId":            flow.FlowNo,
		"caseId":            flow.CaseID,
		"caseNo":            flow.CaseNo,
		"caseTitle":         caseInfo.Title,
		"docType":           flow.DocType,
		"docTitle":          flow.DocTitle,
		"docUrl":            flow.DocURL,
		"signedDocumentUrl": flow.SignedDocumentURL,
		"status":            flow.Status,
		"statusName":        statusMap[flow.Status],
		"signerCount":       flow.TotalSignCount,
		"signedCount":       flow.SignedCount,
		"crossPageSeal":     flow.CrossPageSeal,
		"expireTime":        flow.ExpireTime,
		"fadadaFlowId":      flow.FaDaDaFlowID,
		"bcCertNo":          flow.BCCertNo,
		"bcTxId":            flow.BCTxID,
		"bcOnChainTime":     flow.BCOnChainTime,
		"bcStatus":          flow.BCStatus,
		"creatorId":         flow.CreatorID,
		"creatorName":       flow.CreatorName,
		"createdAt":         flow.CreatedAt,
		"signers":           signerVOs,
	}

	var bcCert model.BlockchainCertificate
	database.GetDB().Where("flow_id = ? AND deleted_at IS NULL", flow.FlowNo).First(&bcCert)
	if bcCert.ID > 0 {
		record["blockchainCert"] = map[string]interface{}{
			"certNo":       bcCert.CertNo,
			"txId":         bcCert.TxID,
			"blockHeight":  bcCert.BlockHeight,
			"onChainTime":  bcCert.OnChainTime,
			"certUrl":      bcCert.CertURL,
			"qrcodeUrl":    bcCert.QRCodeURL,
			"verifyUrl":    bcCert.VerifyURL,
			"status":       bcCert.Status,
			"evidenceHash": bcCert.EvidenceHash,
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

	var flow model.EsignFlow
	result := database.GetDB().Where("id = ?", req.RecordID).First(&flow)

	if result.Error != nil || flow.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署记录不存在"))
		return
	}

	if flow.Status == constants.EsignStatusCompleted {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已完成"))
		return
	}
	if flow.Status == constants.EsignStatusExpired {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已过期"))
		return
	}
	if flow.Status == constants.EsignStatusRevoked {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已撤销"))
		return
	}

	if flow.ExpireTime != nil && time.Now().After(*flow.ExpireTime) {
		c.JSON(http.StatusBadRequest, response.BadRequest("该签署流程已过期"))
		return
	}

	var signer model.EsignSigner
	database.GetDB().Where("flow_id = ? AND user_id = ?", flow.ID, userInfo.UserID).First(&signer)

	if signer.ID == 0 {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该文件的签署人"))
		return
	}

	if signer.SignStatus == constants.EsignSignerStatusSigned {
		c.JSON(http.StatusBadRequest, response.BadRequest("您已签署该文件"))
		return
	}

	cacheKey := fmt.Sprintf("esign:verify:%s:%d", flow.FlowNo, userInfo.UserID)

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
	tx.Model(&signer).Updates(map[string]interface{}{
		"sign_status":   constants.EsignSignerStatusSigned,
		"sign_time":     now,
		"signature_url": signatureURL,
		"sign_ip":       c.ClientIP(),
	})

	newSignedCount := flow.SignedCount + 1
	status := int32(constants.EsignStatusSigning)
	allSigned := newSignedCount >= flow.TotalSignCount
	if allSigned {
		status = constants.EsignStatusCompleted
	}

	tx.Model(&flow).Updates(map[string]interface{}{
		"signed_count":   newSignedCount,
		"status":         status,
		"last_sign_time": now,
	})

	history := map[string]interface{}{
		"case_id":          flow.CaseID,
		"case_no":          flow.CaseNo,
		"operation_type":   "ESIGN_SIGN",
		"operation_detail": fmt.Sprintf("签署文件: %s，签署人: %s", flow.DocTitle, signer.UserName),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cache.Del(ctx, cacheKey)

	if allSigned && flow.FaDaDaFlowID != "" {
		go func() {
			fadadaClient := fadada.GetClient()
			if fadadaClient != nil {
				signedDocURL, err := fadadaClient.GetSignedDocument(flow.FaDaDaFlowID)
				if err != nil {
					logger.Error("Download signed document from FaDaDa failed", logger.Error(err))
				} else if signedDocURL != "" {
					database.GetDB().Model(&model.EsignFlow{}).
						Where("id = ?", flow.ID).
						Update("signed_document_url", signedDocURL)

					go storeSignedDocToBlockchain(flow.ID, flow.CaseID, flow.CaseNo, flow.DocTitle, flow.FlowNo, signedDocURL, userInfo.UserID)
				}
			}
		}()
	}

	go func() {
		notifyMsg := map[string]interface{}{
			"type":        "esign_sign_progress",
			"caseId":      flow.CaseID,
			"caseNo":      flow.CaseNo,
			"docTitle":    flow.DocTitle,
			"signer":      signer.UserName,
			"signedCount": newSignedCount,
			"totalCount":  flow.TotalSignCount,
			"allSigned":   allSigned,
			"signTime":    now.Format("2006-01-02 15:04:05"),
		}
		callback := func(ctx context.Context, result *primitive.SendResult, err error) {
			if err != nil {
				logger.Error("Send esign sign progress notify failed", logger.Error(err))
			}
		}
		mq.SendAsyncMessage(constants.MQTopicEsignNotify, notifyMsg, callback)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"signatureUrl": signatureURL,
		"signTime":     now.Format("2006-01-02 15:04:05"),
		"allSigned":    allSigned,
	}, "签署成功"))
}

func downloadPDFAndHash(docURL string) ([]byte, string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(docURL)
	if err != nil {
		return nil, "", fmt.Errorf("下载PDF失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载PDF失败, HTTP状态码: %d", resp.StatusCode)
	}

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取PDF数据失败: %v", err)
	}

	hash := sha256.Sum256(pdfData)
	pdfHash := hex.EncodeToString(hash[:])

	return pdfData, pdfHash, nil
}

func storeSignedDocToBlockchain(flowID int64, caseID int64, caseNo, docTitle, flowNo, docURL string, creatorID int64) {
	bcClient := blockchain.GetClient()
	if bcClient == nil {
		logger.Error("Blockchain client not initialized")
		return
	}

	_, pdfHash, err := downloadPDFAndHash(docURL)
	if err != nil {
		logger.Error("Download PDF and compute hash failed", logger.Error(err))
		pdfHash = ""
	} else {
		logger.Info("PDF hash computed", logger.String("pdfHash", pdfHash))
	}

	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   fmt.Sprintf("ESIGN-%d", flowID),
		EvidenceType: constants.BCTypeEsignDocument,
		EvidenceHash: pdfHash,
		EvidenceName: docTitle,
		Description:  fmt.Sprintf("调解协议书电子签章文档上链: %s", docTitle),
		Metadata:     fmt.Sprintf(`{"caseId":%d,"caseNo":"%s","flowNo":"%s","docUrl":"%s"}`, caseID, caseNo, flowNo, docURL),
	})

	if err != nil {
		logger.Error("Store signed doc to blockchain failed", logger.Error(err))
		database.GetDB().Model(&model.EsignFlow{}).
			Where("id = ?", flowID).
			Update("bc_status", constants.BCStatusFailed)
		return
	}

	now := time.Now()
	database.GetDB().Model(&model.EsignFlow{}).
		Where("id = ?", flowID).
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
			EvidenceID:   fmt.Sprintf("ESIGN-%d", flowID),
			EvidenceType: constants.BCTypeEsignDocument,
			EvidenceName: docTitle,
			EvidenceHash: pdfHash,
			CaseID:       caseID,
			FlowID:       flowNo,
			TxID:         storeResult.TxID,
			BlockHeight:  storeResult.BlockHeight,
			OnChainTime:  &now,
			CertURL:      certResult.CertURL,
			QRCodeURL:    certResult.QRCodeURL,
			VerifyURL:    certResult.VerifyURL,
			Status:       constants.BCStatusOnChain,
			Metadata:     fmt.Sprintf(`{"caseNo":"%s","flowNo":"%s","docUrl":"%s"}`, caseNo, flowNo, docURL),
			CreatedBy:    creatorID,
		}
		database.GetDB().Create(cert)
	}

	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send esign blockchain stored notify failed", logger.Error(err))
		}
	}
	mq.SendAsyncMessage(constants.MQTopicEsignNotify, map[string]interface{}{
		"type":        "esign_blockchain_stored",
		"caseId":      caseID,
		"caseNo":      caseNo,
		"docTitle":    docTitle,
		"certNo":      certNo,
		"txId":        storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
	}, callback)
}

func RevokeEsignFlow(ctx context.Context, c *app.RequestContext) {
	flowNo := c.Param("flowId")
	userInfo := middleware.GetUserInfo(c)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var flow model.EsignFlow
	result := database.GetDB().Where("flow_no = ?", flowNo).First(&flow)

	if result.Error != nil || flow.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	if flow.CreatorID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("只有流程创建者或领导可以撤销"))
		return
	}

	if flow.Status == constants.EsignStatusCompleted {
		c.JSON(http.StatusBadRequest, response.BadRequest("已完成的流程不能撤销"))
		return
	}

	if flow.FaDaDaFlowID != "" {
		fadadaClient := fadada.GetClient()
		if fadadaClient != nil {
			if err := fadadaClient.RevokeSignFlow(flow.FaDaDaFlowID, req.Reason); err != nil {
				logger.Error("FaDaDa revoke sign flow failed", logger.Error(err))
			}
		}
	}

	tx := database.GetDB().Begin()

	now := time.Now()
	tx.Model(&flow).Updates(map[string]interface{}{
		"status":        constants.EsignStatusRevoked,
		"revoke_reason": req.Reason,
		"revoke_time":   now,
		"revoke_by":     userInfo.UserID,
	})

	var signers []model.EsignSigner
	database.GetDB().Where("flow_id = ?", flow.ID).Find(&signers)

	for _, s := range signers {
		go func(phone, name string) {
			msg := map[string]interface{}{
				"caseId":       flow.CaseID,
				"caseNo":       flow.CaseNo,
				"flowId":       flowNo,
				"docTitle":     flow.DocTitle,
				"revokeReason": req.Reason,
				"revokeBy":     userInfo.RealName,
				"signer":       name,
				"phone":        phone,
				"notifyType":   "sms",
			}
			callback := func(ctx context.Context, result *primitive.SendResult, err error) {
				if err != nil {
					logger.Error("Send esign revoke notify failed", logger.Error(err))
				}
			}
			mq.SendAsyncMessage(constants.MQTopicEsignNotify, msg, callback)
		}(s.UserPhone, s.UserName)
	}

	history := map[string]interface{}{
		"case_id":          flow.CaseID,
		"case_no":          flow.CaseNo,
		"operation_type":   "ESIGN_REVOKE",
		"operation_detail": fmt.Sprintf("撤销签署流程: %s，原因: %s", flow.DocTitle, req.Reason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "撤销成功"))
}

func SendEsignVerifyCode(ctx context.Context, c *app.RequestContext) {
	flowNo := c.Query("flowId")
	userInfo := middleware.GetUserInfo(c)

	var flow model.EsignFlow
	result := database.GetDB().Where("flow_no = ?", flowNo).First(&flow)

	if result.Error != nil || flow.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var signer model.EsignSigner
	database.GetDB().Where("flow_id = ? AND user_id = ?", flow.ID, userInfo.UserID).First(&signer)

	if signer.UserPhone == "" {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是该文件的签署人"))
		return
	}

	verifyCode := utils.GenerateRandomNumber(6)
	cacheKey := fmt.Sprintf("esign:verify:%s:%d", flowNo, userInfo.UserID)
	cache.Set(ctx, cacheKey, verifyCode, 30*time.Minute)

	go func() {
		msg := map[string]interface{}{
			"caseId":     flow.CaseID,
			"caseNo":     flow.CaseNo,
			"flowId":     flowNo,
			"docTitle":   flow.DocTitle,
			"verifyCode": verifyCode,
			"signer":     signer.UserName,
			"phone":      signer.UserPhone,
			"notifyType": "sms",
		}
		callback := func(ctx context.Context, result *primitive.SendResult, err error) {
			if err != nil {
				logger.Error("Send esign verify code notify failed", logger.Error(err))
			}
		}
		mq.SendAsyncMessage(constants.MQTopicEsignNotify, msg, callback)
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

	var flow model.EsignFlow
	database.GetDB().Where("fadada_flow_id = ?", flowID).First(&flow)

	if flow.ID == 0 {
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
				database.GetDB().Model(&model.EsignSigner{}).
					Where("flow_id = ? AND user_name = ?", flow.ID, s.CustomerName).
					Updates(map[string]interface{}{
						"sign_status": constants.EsignSignerStatusSigned,
						"sign_time":   time.Now(),
					})
			}
		}

		newStatus := int32(constants.EsignStatusSigning)
		if signedCount >= flow.TotalSignCount {
			newStatus = constants.EsignStatusCompleted
		}

		database.GetDB().Model(&flow).Updates(map[string]interface{}{
			"status":         newStatus,
			"signed_count":   signedCount,
			"last_sign_time": time.Now(),
		})

		callback := func(ctx context.Context, result *primitive.SendResult, err error) {
			if err != nil {
				logger.Error("Send esign fadada callback notify failed", logger.Error(err))
			}
		}
		mq.SendAsyncMessage(constants.MQTopicEsignNotify, map[string]interface{}{
			"type":        "esign_fadada_callback",
			"caseId":      flow.CaseID,
			"caseNo":      flow.CaseNo,
			"docTitle":    flow.DocTitle,
			"signedCount": signedCount,
			"totalCount":  flow.TotalSignCount,
			"allSigned":   signedCount >= flow.TotalSignCount,
		}, callback)

		if newStatus == constants.EsignStatusCompleted {
			go func() {
				signedDocURL, _ := fadadaClient.GetSignedDocument(flowID)
				if signedDocURL != "" {
					database.GetDB().Model(&flow).Update("signed_document_url", signedDocURL)
					go storeSignedDocToBlockchain(flow.ID, flow.CaseID, flow.CaseNo, flow.DocTitle, flow.FlowNo, signedDocURL, flow.CreatorID)
				}
			}()
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "msg": "success"})
}

func GetEsignProgress(ctx context.Context, c *app.RequestContext) {
	flowNo := c.Param("flowId")

	var flow model.EsignFlow
	database.GetDB().Where("flow_no = ?", flowNo).First(&flow)

	if flow.ID == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("签署流程不存在"))
		return
	}

	var signers []model.EsignSigner
	database.GetDB().Where("flow_id = ?", flow.ID).
		Order("sign_order ASC").
		Find(&signers)

	signStatusMap := map[int32]string{
		constants.EsignSignerStatusPending:  "待签署",
		constants.EsignSignerStatusSigned:   "已签署",
		constants.EsignSignerStatusRejected: "已拒绝",
	}

	type signerVO struct {
		model.EsignSigner
		SignStatusName string `json:"signStatusName"`
	}

	var signerVOs []signerVO
	for _, s := range signers {
		vo := signerVO{EsignSigner: s}
		vo.SignStatusName = signStatusMap[s.SignStatus]
		signerVOs = append(signerVOs, vo)
	}

	result := map[string]interface{}{
		"flowId":      flowNo,
		"status":      flow.Status,
		"signedCount": flow.SignedCount,
		"totalCount":  flow.TotalSignCount,
		"signers":     signerVOs,
	}

	if flow.FaDaDaFlowID != "" {
		fadadaClient := fadada.GetClient()
		if fadadaClient != nil {
			progress, err := fadadaClient.GetSignProgress(flow.FaDaDaFlowID)
			if err == nil {
				result["fadadaProgress"] = progress
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}
