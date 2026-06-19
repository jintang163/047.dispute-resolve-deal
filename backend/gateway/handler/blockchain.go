package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/blockchain"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

func StoreEvidenceToBlockchain(ctx context.Context, c *app.RequestContext) {
	var req struct {
		CaseID       int64  `json:"caseId" binding:"required"`
		EvidenceType string `json:"evidenceType" binding:"required"`
		EvidenceID   string `json:"evidenceId" binding:"required"`
		EvidenceName string `json:"evidenceName" binding:"required"`
		EvidenceHash string `json:"evidenceHash" binding:"required"`
		FlowID       string `json:"flowId"`
		Metadata     string `json:"metadata"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	bcClient := blockchain.GetClient()
	if bcClient == nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("区块链客户端未初始化"))
		return
	}

	var existingCert model.BlockchainCertificate
	database.GetDB().Where("evidence_id = ? AND evidence_type = ? AND status != ? AND deleted_at IS NULL",
		req.EvidenceID, req.EvidenceType, constants.BCStatusFailed).First(&existingCert)
	if existingCert.ID > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该证据已存证，证书编号: "+existingCert.CertNo))
		return
	}

	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   req.EvidenceID,
		EvidenceType: req.EvidenceType,
		EvidenceHash: req.EvidenceHash,
		EvidenceName: req.EvidenceName,
		Description:  fmt.Sprintf("司法存证链上链: %s", req.EvidenceName),
		Metadata:     req.Metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("区块链存证失败: "+err.Error()))
		return
	}

	now := time.Now()
	certNo := storeResult.CertNo

	var certResult *blockchain.CertificateResult
	certResult, _ = bcClient.GetCertificate(certNo)

	certURL := ""
	qrcodeURL := ""
	verifyURL := ""
	if certResult != nil {
		certURL = certResult.CertURL
		qrcodeURL = certResult.QRCodeURL
		verifyURL = certResult.VerifyURL
	}

	cert := &model.BlockchainCertificate{
		BaseModel: model.BaseModel{
			ID: utils.GenerateID(),
		},
		CertNo:       certNo,
		EvidenceID:   req.EvidenceID,
		EvidenceType: req.EvidenceType,
		EvidenceName: req.EvidenceName,
		EvidenceHash: req.EvidenceHash,
		CaseID:       req.CaseID,
		FlowID:       req.FlowID,
		TxID:         storeResult.TxID,
		BlockHeight:  storeResult.BlockHeight,
		OnChainTime:  &now,
		CertURL:      certURL,
		QRCodeURL:    qrcodeURL,
		VerifyURL:    verifyURL,
		Status:       constants.BCStatusOnChain,
		Metadata:     req.Metadata,
		CreatedBy:    userInfo.UserID,
	}

	if err := database.GetDB().Create(cert).Error; err != nil {
		logger.Error("Save blockchain certificate failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("存证证书保存失败"))
		return
	}

	if req.FlowID != "" {
		database.GetDB().Table("esign_record").
			Where("flow_id = ?", req.FlowID).
			Updates(map[string]interface{}{
				"bc_cert_no":      certNo,
				"bc_tx_id":        storeResult.TxID,
				"bc_on_chain_time": now,
				"bc_status":       constants.BCStatusOnChain,
			})
	}

	var caseNo string
	database.GetDB().Table("dispute_case").
		Select("case_no").
		Where("id = ?", req.CaseID).
		Scan(&caseNo)

	history := map[string]interface{}{
		"case_id":          req.CaseID,
		"case_no":          caseNo,
		"operation_type":   "BLOCKCHAIN_STORE",
		"operation_detail": fmt.Sprintf("区块链存证: %s，证书编号: %s", req.EvidenceName, certNo),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
	}
	database.GetDB().Table("dispute_case_history").Create(history)

	mq.SendAsync(constants.MQTopicBlockchainStore, map[string]interface{}{
		"type":        "blockchain_stored",
		"caseId":      req.CaseID,
		"caseNo":      caseNo,
		"evidenceId":  req.EvidenceID,
		"certNo":      certNo,
		"txId":        storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
		"evidenceName": req.EvidenceName,
	})

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"certNo":      certNo,
		"txId":        storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
		"onChainTime": now.Format("2006-01-02 15:04:05"),
		"certUrl":     certURL,
		"qrcodeUrl":   qrcodeURL,
		"verifyUrl":   verifyURL,
	}, "存证成功"))
}

func GetBlockchainCertList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	caseID, _ := strconv.ParseInt(c.Query("caseId"), 10, 64)
	evidenceType := c.Query("evidenceType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	db := database.GetDB().Table("blockchain_certificate bc").
		Select("bc.*, dc.case_no, dc.title as case_title").
		Joins("LEFT JOIN dispute_case dc ON bc.case_id = dc.id").
		Where("bc.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("bc.created_by = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("bc.case_id IN (SELECT id FROM dispute_case WHERE organization_id = ?)", userInfo.OrganizationID)
	}

	if caseID > 0 {
		db = db.Where("bc.case_id = ?", caseID)
	}
	if evidenceType != "" {
		db = db.Where("bc.evidence_type = ?", evidenceType)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := (page - 1) * pageSize
	db.Order("bc.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list)

	statusMap := map[int]string{
		constants.BCStatusPending:  "待存证",
		constants.BCStatusOnChain:  "已存证",
		constants.BCStatusFailed:   "存证失败",
		constants.BCStatusVerified: "已验证",
	}

	evidenceTypeMap := map[string]string{
		constants.BCTypeMediationProtocol: "调解协议",
		constants.BCTypeEsignDocument:     "签章文档",
		constants.BCTypeEvidence:          "证据材料",
	}

	for _, item := range list {
		if s, ok := item["status"].(int); ok {
			item["status_name"] = statusMap[s]
		}
		if et, ok := item["evidence_type"].(string); ok {
			item["evidence_type_name"] = evidenceTypeMap[et]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetBlockchainCertDetail(ctx context.Context, c *app.RequestContext) {
	certNo := c.Param("certNo")

	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("存证证书不存在"))
		return
	}

	var caseInfo struct {
		CaseNo string `gorm:"column:case_no"`
		Title  string `gorm:"column:title"`
	}
	database.GetDB().Table("dispute_case").
		Select("case_no, title").
		Where("id = ?", cert.CaseID).
		Scan(&caseInfo)

	statusMap := map[int]string{
		constants.BCStatusPending:  "待存证",
		constants.BCStatusOnChain:  "已存证",
		constants.BCStatusFailed:   "存证失败",
		constants.BCStatusVerified: "已验证",
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"certNo":        cert.CertNo,
		"evidenceId":    cert.EvidenceID,
		"evidenceType":  cert.EvidenceType,
		"evidenceName":  cert.EvidenceName,
		"evidenceHash":  cert.EvidenceHash,
		"caseId":        cert.CaseID,
		"caseNo":        caseInfo.CaseNo,
		"caseTitle":     caseInfo.Title,
		"flowId":        cert.FlowID,
		"txId":          cert.TxID,
		"blockHeight":   cert.BlockHeight,
		"onChainTime":   cert.OnChainTime,
		"certUrl":       cert.CertURL,
		"qrcodeUrl":     cert.QRCodeURL,
		"verifyUrl":     cert.VerifyURL,
		"status":        cert.Status,
		"statusName":    statusMap[cert.Status],
		"metadata":      cert.Metadata,
		"createdBy":     cert.CreatedBy,
		"createdAt":     cert.CreatedAt,
	}))
}

func VerifyBlockchainEvidence(ctx context.Context, c *app.RequestContext) {
	certNo := c.Query("certNo")
	if certNo == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("请提供存证证书编号"))
		return
	}

	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("存证证书不存在"))
		return
	}

	bcClient := blockchain.GetClient()
	if bcClient == nil {
		c.JSON(http.StatusOK, response.Success(map[string]interface{}{
			"valid":        true,
			"certNo":       cert.CertNo,
			"evidenceHash": cert.EvidenceHash,
			"txId":         cert.TxID,
			"blockHeight":  cert.BlockHeight,
			"onChainTime":  cert.OnChainTime,
			"evidenceName": cert.EvidenceName,
			"source":       "database",
		}))
		return
	}

	verifyResult, err := bcClient.VerifyEvidence(certNo)
	if err != nil {
		logger.Error("Verify blockchain evidence failed", logger.Error(err))
		c.JSON(http.StatusOK, response.Success(map[string]interface{}{
			"valid":        true,
			"certNo":       cert.CertNo,
			"evidenceHash": cert.EvidenceHash,
			"txId":         cert.TxID,
			"blockHeight":  cert.BlockHeight,
			"onChainTime":  cert.OnChainTime,
			"evidenceName": cert.EvidenceName,
			"source":       "database",
		}))
		return
	}

	if verifyResult.Valid {
		database.GetDB().Model(&cert).Update("status", constants.BCStatusVerified)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"valid":        verifyResult.Valid,
		"certNo":       cert.CertNo,
		"evidenceHash": verifyResult.EvidenceHash,
		"txId":         verifyResult.TxID,
		"blockHeight":  verifyResult.BlockHeight,
		"onChainTime":  verifyResult.Timestamp,
		"evidenceName": cert.EvidenceName,
		"source":       "blockchain",
	}))
}

func DownloadBlockchainCert(ctx context.Context, c *app.RequestContext) {
	certNo := c.Param("certNo")

	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("存证证书不存在"))
		return
	}

	if cert.CertURL == "" {
		bcClient := blockchain.GetClient()
		if bcClient != nil {
			certResult, err := bcClient.GetCertificate(certNo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.ServerError("获取证书失败"))
				return
			}
			cert.CertURL = certResult.CertURL
			cert.QRCodeURL = certResult.QRCodeURL
			cert.VerifyURL = certResult.VerifyURL

			database.GetDB().Model(&cert).Updates(map[string]interface{}{
				"cert_url":   cert.CertURL,
				"qrcode_url": cert.QRCodeURL,
				"verify_url": cert.VerifyURL,
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"certNo":      cert.CertNo,
		"certUrl":     cert.CertURL,
		"qrcodeUrl":   cert.QRCodeURL,
		"verifyUrl":   cert.VerifyURL,
		"txId":        cert.TxID,
		"blockHeight": cert.BlockHeight,
		"onChainTime": cert.OnChainTime,
		"evidenceHash": cert.EvidenceHash,
		"evidenceName": cert.EvidenceName,
	}))
}

func PublicVerifyEvidence(ctx context.Context, c *app.RequestContext) {
	certNo := c.Param("certNo")

	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("存证证书不存在或编号无效"))
		return
	}

	var caseInfo struct {
		CaseNo string `gorm:"column:case_no"`
		Title  string `gorm:"column:title"`
	}
	database.GetDB().Table("dispute_case").
		Select("case_no, title").
		Where("id = ?", cert.CaseID).
		Scan(&caseInfo)

	valid := false
	bcClient := blockchain.GetClient()
	if bcClient != nil {
		verifyResult, err := bcClient.VerifyEvidence(certNo)
		if err == nil {
			valid = verifyResult.Valid
		}
	} else {
		valid = cert.Status == constants.BCStatusOnChain || cert.Status == constants.BCStatusVerified
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"valid":        valid,
		"certNo":       cert.CertNo,
		"evidenceName": cert.EvidenceName,
		"evidenceHash": cert.EvidenceHash,
		"txId":         cert.TxID,
		"blockHeight":  cert.BlockHeight,
		"onChainTime":  cert.OnChainTime,
		"caseNo":       caseInfo.CaseNo,
		"caseTitle":    caseInfo.Title,
		"evidenceType": cert.EvidenceType,
	}))
}
