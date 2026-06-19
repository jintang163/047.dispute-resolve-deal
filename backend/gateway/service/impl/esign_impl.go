package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/blockchain"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/fadada"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type ESignServiceImpl struct{}

func NewESignService() service.ESignService {
	return &ESignServiceImpl{}
}

type BlockchainServiceImpl struct{}

func NewBlockchainService() service.BlockchainService {
	return &BlockchainServiceImpl{}
}

const (
	esignVerifyCodeTTL = 300
)

func (s *ESignServiceImpl) CreateEsignFlow(ctx context.Context, caseID int64, title string, documentIDs []int64, signerIDs []int64, creatorID int64) (map[string]interface{}, error) {
	var caseInfo map[string]interface{}
	database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseInfo)
	if caseInfo == nil {
		return nil, fmt.Errorf("案件不存在")
	}

	flowNo := "ES" + utils.GenerateRandomString(12)

	var fadadaFlowID string
	fadadaClient := fadada.GetClient()
	if fadadaClient != nil {
		var fadadaSigners []fadada.FaDaDaSigner
		for i, sid := range signerIDs {
			var user struct {
				RealName string `gorm:"column:real_name"`
			}
			database.GetDB().Table("sys_user").Select("real_name").Where("id = ?", sid).First(&user)
			fadadaSigners = append(fadadaSigners, fadada.FaDaDaSigner{
				CustomerID:   fmt.Sprintf("USR%d", sid),
				CustomerName: user.RealName,
				SignOrder:    i + 1,
				SignType:     "1",
			})
		}

		docURL := ""
		if len(documentIDs) > 0 {
			database.GetDB().Table("evidence").
				Select("file_url").
				Where("id = ?", documentIDs[0]).
				Scan(&docURL)
		}

		if docURL != "" {
			result, err := fadadaClient.CreateSignFlow(&fadada.CreateSignFlowReq{
				DocTitle:    title,
				DocURL:      docURL,
				SignerIDs:   fadadaSigners,
				ExpireHours: 72,
			})
			if err != nil {
				logger.Error("FaDaDa create sign flow failed", logger.Error(err))
			} else {
				fadadaFlowID = result.FlowID
			}
		}
	}

	flow := &model.EsignFlow{
		CaseID:       caseID,
		FlowNo:       flowNo,
		DocumentName: title,
		Status:       constants.EsignStatusPending,
		TotalSignCount: len(signerIDs),
		FaDaDaFlowID: fadadaFlowID,
		CreatedBy:    creatorID,
	}

	if err := database.GetDB().Create(flow).Error; err != nil {
		return nil, err
	}

	s.createSignerRecords(flow.ID, signerIDs)
	s.sendEsignNotifications(flow.ID, caseID, signerIDs, title, "create")

	return map[string]interface{}{
		"flowId":       flow.ID,
		"flowNo":       flowNo,
		"title":        title,
		"status":       flow.Status,
		"fadadaFlowId": fadadaFlowID,
		"createdAt":    flow.CreatedAt,
	}, nil
}

func (s *ESignServiceImpl) GetEsignList(ctx context.Context, caseID int64, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("esign_flow ef").
		Select("ef.*, dc.case_no, dc.title as case_title, u.real_name as creator_name").
		Joins("LEFT JOIN dispute_case dc ON ef.case_id = dc.id").
		Joins("LEFT JOIN user u ON ef.creator_id = u.id").
		Where("ef.deleted_at IS NULL")

	if caseID > 0 {
		db = db.Where("ef.case_id = ?", caseID)
	}

	if userID > 0 && role != constants.RoleAdmin {
		db = db.Where("(ef.creator_id = ? OR FIND_IN_SET(?, ef.signer_ids))", userID, userID)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := (page - 1) * pageSize
	db.Order("ef.created_at DESC").
		Offset(offset).
		Limit(pageSize).
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
		if status, ok := item["status"].(int); ok {
			item["statusText"] = statusMap[status]
		}
		signedCount := s.getSignedCount(item["id"].(int64))
		item["signedCount"] = signedCount
	}

	return list, total, nil
}

func (s *ESignServiceImpl) GetEsignDetail(ctx context.Context, esignID int64, userID int64) (map[string]interface{}, error) {
	var flow map[string]interface{}
	result := database.GetDB().Table("esign_flow ef").
		Select("ef.*, dc.case_no, dc.title as case_title, u.real_name as creator_name").
		Joins("LEFT JOIN dispute_case dc ON ef.case_id = dc.id").
		Joins("LEFT JOIN user u ON ef.creator_id = u.id").
		Where("ef.id = ? AND ef.deleted_at IS NULL", esignID).
		First(&flow)

	if result.Error != nil {
		return nil, result.Error
	}

	statusMap := map[int]string{
		constants.EsignStatusDraft:     "草稿",
		constants.EsignStatusPending:   "待签署",
		constants.EsignStatusSigning:   "签署中",
		constants.EsignStatusCompleted: "已完成",
		constants.EsignStatusExpired:   "已过期",
		constants.EsignStatusRevoked:   "已撤销",
	}

	if status, ok := flow["status"].(int); ok {
		flow["statusText"] = statusMap[status]
	}

	var signers []map[string]interface{}
	database.GetDB().Table("esign_signer es").
		Select("es.*, u.real_name, u.avatar, u.phone").
		Joins("LEFT JOIN sys_user u ON es.user_id = u.id").
		Where("es.flow_id = ?", esignID).
		Order("es.sign_order ASC").
		Find(&signers)

	signStatusMap := map[int]string{
		constants.EsignSignerStatusPending:  "待签署",
		constants.EsignSignerStatusSigned:   "已签署",
		constants.EsignSignerStatusRejected: "已拒绝",
	}

	for _, signer := range signers {
		if status, ok := signer["sign_status"].(int); ok {
			signer["signStatusText"] = signStatusMap[status]
		}
	}

	flow["signers"] = signers
	flow["signedCount"] = s.getSignedCount(esignID)
	flow["totalSigners"] = len(signers)

	var bcCert model.BlockchainCertificate
	database.GetDB().Where("flow_id = ? AND deleted_at IS NULL", fmt.Sprintf("ES%d", esignID)).First(&bcCert)
	if bcCert.ID > 0 {
		flow["blockchainCert"] = map[string]interface{}{
			"certNo":      bcCert.CertNo,
			"txId":        bcCert.TxID,
			"blockHeight": bcCert.BlockHeight,
			"onChainTime": bcCert.OnChainTime,
			"certUrl":     bcCert.CertURL,
			"qrcodeUrl":   bcCert.QRCodeURL,
			"verifyUrl":   bcCert.VerifyURL,
			"status":      bcCert.Status,
		}
	}

	return flow, nil
}

func (s *ESignServiceImpl) SignDocument(ctx context.Context, esignID int64, userID int64, verifyCode string, signatureData string) error {
	var flow model.EsignFlow
	database.GetDB().Where("id = ? AND deleted_at IS NULL", esignID).First(&flow)
	if flow.ID == 0 {
		return fmt.Errorf("签署流程不存在")
	}

	if flow.Status == constants.EsignStatusCompleted || flow.Status == constants.EsignStatusRevoked {
		return fmt.Errorf("签署流程状态不允许签署")
	}

	var signer model.EsignSigner
	database.GetDB().Where("flow_id = ? AND user_id = ?", esignID, userID).First(&signer)
	if signer.ID == 0 {
		return fmt.Errorf("您不是签署人")
	}

	if signer.Status == constants.EsignSignerStatusSigned {
		return fmt.Errorf("已签署")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"sign_status":  constants.EsignSignerStatusSigned,
		"sign_time":    now,
		"signature_url": signatureData,
	}

	err := database.GetDB().Model(&signer).Updates(updates).Error
	if err != nil {
		return err
	}

	if flow.Status == constants.EsignStatusPending {
		database.GetDB().Model(&flow).Update("status", constants.EsignStatusSigning)
	}

	signedCount := s.getSignedCount(esignID)
	if signedCount >= flow.TotalSignCount {
		database.GetDB().Model(&flow).Update("status", constants.EsignStatusCompleted)
		s.sendEsignNotifications(flow.ID, flow.CaseID, nil, flow.DocumentName, "complete")

		if flow.FaDaDaFlowID != "" {
			go s.autoStoreToBlockchain(flow)
		}
	}

	mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
		"type":        "esign_sign",
		"flowId":      esignID,
		"userId":      userID,
		"caseId":      flow.CaseID,
		"signedCount": signedCount,
		"totalCount":  flow.TotalSignCount,
		"allSigned":   signedCount >= flow.TotalSignCount,
		"timestamp":   now,
	})

	return nil
}

func (s *ESignServiceImpl) RevokeEsignFlow(ctx context.Context, esignID int64, userID int64, reason string) error {
	var flow model.EsignFlow
	database.GetDB().Where("id = ? AND deleted_at IS NULL", esignID).First(&flow)
	if flow.ID == 0 {
		return fmt.Errorf("签署流程不存在")
	}

	if flow.FaDaDaFlowID != "" {
		fadadaClient := fadada.GetClient()
		if fadadaClient != nil {
			fadadaClient.RevokeSignFlow(flow.FaDaDaFlowID, reason)
		}
	}

	updates := map[string]interface{}{
		"status":       constants.EsignStatusRevoked,
		"revoke_reason": reason,
		"revoke_time":  time.Now(),
		"revoked_by":   userID,
	}

	return database.GetDB().Model(&flow).Updates(updates).Error
}

func (s *ESignServiceImpl) SendEsignVerifyCode(ctx context.Context, esignID int64, mobile string) error {
	var user model.User
	database.GetDB().Where("phone = ? AND status = 1", mobile).First(&user)
	if user.ID == 0 {
		return fmt.Errorf("用户不存在")
	}

	code := generateVerifyCode()

	key := fmt.Sprintf("esign:verify:%d:%d", user.ID, esignID)
	database.GetRedisClient().Set(ctx, key, code, time.Duration(esignVerifyCodeTTL)*time.Second)

	mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
		"type":      "sms_verify",
		"mobile":    mobile,
		"code":      code,
		"esignId":   esignID,
		"expire":    esignVerifyCodeTTL,
		"timestamp": time.Now(),
	})

	return nil
}

func (s *ESignServiceImpl) GetEsignProgress(ctx context.Context, flowID string) (map[string]interface{}, error) {
	var flow model.EsignFlow
	database.GetDB().Where("flow_no = ? AND deleted_at IS NULL", flowID).First(&flow)
	if flow.ID == 0 {
		return nil, fmt.Errorf("签署流程不存在")
	}

	var signers []map[string]interface{}
	database.GetDB().Table("esign_signer").
		Where("flow_id = ?", flow.ID).
		Order("sign_order ASC").
		Find(&signers)

	result := map[string]interface{}{
		"flowId":      flowID,
		"status":      flow.Status,
		"signedCount": flow.SignedCount,
		"totalCount":  flow.TotalSignCount,
		"signers":     signers,
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

	return result, nil
}

func (s *ESignServiceImpl) StoreSignedDocToBlockchain(ctx context.Context, esignID int64) (map[string]interface{}, error) {
	var flow model.EsignFlow
	database.GetDB().Where("id = ? AND deleted_at IS NULL", esignID).First(&flow)
	if flow.ID == 0 {
		return nil, fmt.Errorf("签署流程不存在")
	}

	if flow.Status != constants.EsignStatusCompleted {
		return nil, fmt.Errorf("只有已完成的签署流程才能存证")
	}

	bcClient := blockchain.GetClient()
	if bcClient == nil {
		return nil, fmt.Errorf("区块链客户端未初始化")
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"caseId":  flow.CaseID,
		"caseNo":  flow.CaseNo,
		"flowNo":  flow.FlowNo,
		"docName": flow.DocumentName,
	})

	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   flow.FlowNo,
		EvidenceType: constants.BCTypeEsignDocument,
		EvidenceHash: "",
		EvidenceName: flow.DocumentName,
		Description:  fmt.Sprintf("调解协议书电子签章文档上链: %s", flow.DocumentName),
		Metadata:     string(metadata),
	})
	if err != nil {
		database.GetDB().Model(&flow).Update("bc_status", constants.BCStatusFailed)
		return nil, err
	}

	now := time.Now()
	database.GetDB().Model(&flow).Updates(map[string]interface{}{
		"bc_cert_no":      storeResult.CertNo,
		"bc_tx_id":        storeResult.TxID,
		"bc_on_chain_time": now,
		"bc_status":       constants.BCStatusOnChain,
	})

	return map[string]interface{}{
		"certNo":      storeResult.CertNo,
		"txId":        storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
		"onChainTime": now.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *ESignServiceImpl) NotifySignerProgress(ctx context.Context, flowID string, signerID int64, notifyType string) error {
	mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
		"type":       "esign_progress",
		"flowId":     flowID,
		"signerId":   signerID,
		"notifyType": notifyType,
		"timestamp":  time.Now(),
	})

	now := time.Now()
	notifyStatus := constants.EsignNotifyStatusSMS
	if notifyType == "wechat" {
		notifyStatus = constants.EsignNotifyStatusWechat
	} else if notifyType == "all" {
		notifyStatus = constants.EsignNotifyStatusAll
	}

	database.GetDB().Table("esign_signer").
		Where("flow_id = (SELECT id FROM esign_flow WHERE flow_no = ?) AND user_id = ?", flowID, signerID).
		Updates(map[string]interface{}{
			"notify_status": notifyStatus,
			"notify_sent_at": now,
		})

	return nil
}

func (s *ESignServiceImpl) autoStoreToBlockchain(flow model.EsignFlow) {
	bcClient := blockchain.GetClient()
	if bcClient == nil {
		return
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"caseId": flow.CaseID,
		"caseNo": flow.CaseNo,
		"flowNo": flow.FlowNo,
	})

	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   flow.FlowNo,
		EvidenceType: constants.BCTypeEsignDocument,
		EvidenceHash: "",
		EvidenceName: flow.DocumentName,
		Description:  fmt.Sprintf("调解协议书签章完成自动上链: %s", flow.DocumentName),
		Metadata:     string(metadata),
	})
	if err != nil {
		logger.Error("Auto store to blockchain failed", logger.Error(err))
		database.GetDB().Model(&flow).Update("bc_status", constants.BCStatusFailed)
		return
	}

	now := time.Now()
	database.GetDB().Model(&flow).Updates(map[string]interface{}{
		"bc_cert_no":      storeResult.CertNo,
		"bc_tx_id":        storeResult.TxID,
		"bc_on_chain_time": now,
		"bc_status":       constants.BCStatusOnChain,
	})

	certResult, _ := bcClient.GetCertificate(storeResult.CertNo)
	certURL, qrcodeURL, verifyURL := "", "", ""
	if certResult != nil {
		certURL = certResult.CertURL
		qrcodeURL = certResult.QRCodeURL
		verifyURL = certResult.VerifyURL
	}

	cert := &model.BlockchainCertificate{
		BaseModel: model.BaseModel{
			ID: utils.GenerateID(),
		},
		CertNo:       storeResult.CertNo,
		EvidenceID:   flow.FlowNo,
		EvidenceType: constants.BCTypeEsignDocument,
		EvidenceName: flow.DocumentName,
		CaseID:       flow.CaseID,
		FlowID:       flow.FlowNo,
		TxID:         storeResult.TxID,
		BlockHeight:  storeResult.BlockHeight,
		OnChainTime:  &now,
		CertURL:      certURL,
		QRCodeURL:    qrcodeURL,
		VerifyURL:    verifyURL,
		Status:       constants.BCStatusOnChain,
		Metadata:     string(metadata),
		CreatedBy:    flow.CreatedBy,
	}
	database.GetDB().Create(cert)
}

func (s *ESignServiceImpl) createSignerRecords(flowID int64, signerIDs []int64) {
	for _, signerID := range signerIDs {
		signer := &model.EsignSigner{
			FlowID: flowID,
			UserID: signerID,
			Status: constants.EsignSignerStatusPending,
		}
		database.GetDB().Create(signer)
	}
}

func (s *ESignServiceImpl) sendEsignNotifications(flowID, caseID int64, signerIDs []int64, title, action string) {
	for _, sid := range signerIDs {
		mq.SendAsync(constants.MQTopicEsignNotify, map[string]interface{}{
			"type":      "esign_" + action,
			"flowId":    flowID,
			"userId":    sid,
			"caseId":    caseID,
			"title":     title,
			"timestamp": time.Now(),
		})
	}
}

func (s *ESignServiceImpl) getSignedCount(flowID int64) int {
	var count int64
	database.GetDB().Table("esign_signer").
		Where("flow_id = ? AND sign_status = ?", flowID, constants.EsignSignerStatusSigned).
		Count(&count)
	return int(count)
}

func (bs *BlockchainServiceImpl) StoreEvidence(ctx context.Context, caseID int64, evidenceID string, evidenceType string, evidenceName string, evidenceHash string, flowID string, metadata string, creatorID int64) (map[string]interface{}, error) {
	bcClient := blockchain.GetClient()
	if bcClient == nil {
		return nil, fmt.Errorf("区块链客户端未初始化")
	}

	storeResult, err := bcClient.StoreEvidence(&blockchain.StoreEvidenceReq{
		EvidenceID:   evidenceID,
		EvidenceType: evidenceType,
		EvidenceHash: evidenceHash,
		EvidenceName: evidenceName,
		Description:  fmt.Sprintf("司法存证链上链: %s", evidenceName),
		Metadata:     metadata,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	certResult, _ := bcClient.GetCertificate(storeResult.CertNo)
	certURL, qrcodeURL, verifyURL := "", "", ""
	if certResult != nil {
		certURL = certResult.CertURL
		qrcodeURL = certResult.QRCodeURL
		verifyURL = certResult.VerifyURL
	}

	cert := &model.BlockchainCertificate{
		BaseModel: model.BaseModel{
			ID: utils.GenerateID(),
		},
		CertNo:       storeResult.CertNo,
		EvidenceID:   evidenceID,
		EvidenceType: evidenceType,
		EvidenceName: evidenceName,
		EvidenceHash: evidenceHash,
		CaseID:       caseID,
		FlowID:       flowID,
		TxID:         storeResult.TxID,
		BlockHeight:  storeResult.BlockHeight,
		OnChainTime:  &now,
		CertURL:      certURL,
		QRCodeURL:    qrcodeURL,
		VerifyURL:    verifyURL,
		Status:       constants.BCStatusOnChain,
		Metadata:     metadata,
		CreatedBy:    creatorID,
	}

	if err := database.GetDB().Create(cert).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"certNo":      storeResult.CertNo,
		"txId":        storeResult.TxID,
		"blockHeight": storeResult.BlockHeight,
		"onChainTime": now.Format("2006-01-02 15:04:05"),
		"certUrl":     certURL,
		"qrcodeUrl":   qrcodeURL,
		"verifyUrl":   verifyURL,
	}, nil
}

func (bs *BlockchainServiceImpl) GetCertList(ctx context.Context, caseID int64, evidenceType string, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("blockchain_certificate bc").
		Select("bc.*, dc.case_no, dc.title as case_title").
		Joins("LEFT JOIN dispute_case dc ON bc.case_id = dc.id").
		Where("bc.deleted_at IS NULL")

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

	for _, item := range list {
		if s, ok := item["status"].(int); ok {
			item["statusText"] = statusMap[s]
		}
	}

	return list, total, nil
}

func (bs *BlockchainServiceImpl) GetCertDetail(ctx context.Context, certNo string) (map[string]interface{}, error) {
	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		return nil, result.Error
	}

	return map[string]interface{}{
		"certNo":       cert.CertNo,
		"evidenceId":   cert.EvidenceID,
		"evidenceType": cert.EvidenceType,
		"evidenceName": cert.EvidenceName,
		"evidenceHash": cert.EvidenceHash,
		"caseId":       cert.CaseID,
		"flowId":       cert.FlowID,
		"txId":         cert.TxID,
		"blockHeight":  cert.BlockHeight,
		"onChainTime":  cert.OnChainTime,
		"certUrl":      cert.CertURL,
		"qrcodeUrl":    cert.QRCodeURL,
		"verifyUrl":    cert.VerifyURL,
		"status":       cert.Status,
		"metadata":     cert.Metadata,
	}, nil
}

func (bs *BlockchainServiceImpl) VerifyEvidence(ctx context.Context, certNo string) (map[string]interface{}, error) {
	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		return nil, result.Error
	}

	bcClient := blockchain.GetClient()
	if bcClient == nil {
		return map[string]interface{}{
			"valid":  true,
			"certNo": cert.CertNo,
			"txId":   cert.TxID,
			"source": "database",
		}, nil
	}

	verifyResult, err := bcClient.VerifyEvidence(certNo)
	if err != nil {
		return map[string]interface{}{
			"valid":  true,
			"certNo": cert.CertNo,
			"txId":   cert.TxID,
			"source": "database",
		}, nil
	}

	if verifyResult.Valid {
		database.GetDB().Model(&cert).Update("status", constants.BCStatusVerified)
	}

	return map[string]interface{}{
		"valid":        verifyResult.Valid,
		"certNo":       cert.CertNo,
		"txId":         verifyResult.TxID,
		"blockHeight":  verifyResult.BlockHeight,
		"evidenceHash": verifyResult.EvidenceHash,
		"source":       "blockchain",
	}, nil
}

func (bs *BlockchainServiceImpl) DownloadCert(ctx context.Context, certNo string) (map[string]interface{}, error) {
	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		return nil, result.Error
	}

	return map[string]interface{}{
		"certNo":      cert.CertNo,
		"certUrl":     cert.CertURL,
		"qrcodeUrl":   cert.QRCodeURL,
		"verifyUrl":   cert.VerifyURL,
		"txId":        cert.TxID,
		"blockHeight": cert.BlockHeight,
		"onChainTime": cert.OnChainTime,
	}, nil
}

func (bs *BlockchainServiceImpl) PublicVerify(ctx context.Context, certNo string) (map[string]interface{}, error) {
	var cert model.BlockchainCertificate
	result := database.GetDB().Where("cert_no = ? AND deleted_at IS NULL", certNo).First(&cert)
	if result.Error != nil {
		return nil, result.Error
	}

	return map[string]interface{}{
		"valid":        cert.Status == constants.BCStatusOnChain || cert.Status == constants.BCStatusVerified,
		"certNo":       cert.CertNo,
		"evidenceName": cert.EvidenceName,
		"evidenceHash": cert.EvidenceHash,
		"txId":         cert.TxID,
		"blockHeight":  cert.BlockHeight,
		"onChainTime":  cert.OnChainTime,
	}, nil
}

func generateVerifyCode() string {
	return utils.GenerateRandomNumber(6)
}
