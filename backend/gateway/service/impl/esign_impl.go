package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
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

const (
	esignStatusPending   = 10
	esignStatusSigning   = 20
	esignStatusCompleted = 30
	esignStatusRevoked   = 40

	esignVerifyCodeTTL = 300
)

func (s *ESignServiceImpl) CreateEsignFlow(ctx context.Context, caseID int64, title string, documentIDs []int64, signerIDs []int64, creatorID int64) (map[string]interface{}, error) {
	var caseInfo map[string]interface{}
	database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseInfo)
	if caseInfo == nil {
		return nil, nil
	}

	flowNo := "ES" + utils.GenerateRandomString(12)

	flow := &model.ESignFlow{
		CaseID:       caseID,
		FlowNo:       flowNo,
		Title:        title,
		DocumentIDs:  documentIDs,
		SignerIDs:    signerIDs,
		Status:       esignStatusPending,
		CreatorID:    creatorID,
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
		"documentIDs":  documentIDs,
		"signerIDs":    signerIDs,
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

	for _, item := range list {
		if status, ok := item["status"].(int); ok {
			statusMap := map[int]string{
				10: "待签署", 20: "签署中", 30: "已完成", 40: "已撤销",
			}
			item["statusText"] = statusMap[status]
		}

		signerIDs, _ := item["signer_ids"].([]byte)
		if signerIDs != nil {
			idStr := string(signerIDs)
			signedCount := s.getSignedCount(item["id"].(int64))
			totalSigners := len(utils.ParseInt64Slice(idStr))
			item["signedCount"] = signedCount
			item["totalSigners"] = totalSigners
			item["progress"] = float64(signedCount) / float64(totalSigners) * 100
		}
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

	if status, ok := flow["status"].(int); ok {
		statusMap := map[int]string{
			10: "待签署", 20: "签署中", 30: "已完成", 40: "已撤销",
		}
		flow["statusText"] = statusMap[status]
	}

	signerIDs, _ := flow["signer_ids"].([]byte)
	if signerIDs != nil {
		var signers []map[string]interface{}
		idStr := string(signerIDs)
		database.GetDB().Table("esign_signer es").
			Select("es.*, u.real_name, u.avatar, u.mobile").
			Joins("LEFT JOIN user u ON es.signer_id = u.id").
			Where("es.flow_id = ?", esignID).
			Order("es.created_at ASC").
			Find(&signers)

		for _, signer := range signers {
			if status, ok := signer["sign_status"].(int); ok {
				signStatusMap := map[int]string{
					0: "待签署", 1: "已签署", 2: "已拒绝",
				}
				signer["signStatusText"] = signStatusMap[status]
			}
			if signer["signer_id"] == userID {
				flow["mySignStatus"] = signer["sign_status"]
				flow["mySignStatusText"] = signer["signStatusText"]
			}
		}

		flow["signers"] = signers
		flow["signedCount"] = s.getSignedCount(esignID)
		flow["totalSigners"] = len(signers)
	}

	documentIDs, _ := flow["document_ids"].([]byte)
	if documentIDs != nil {
		var documents []map[string]interface{}
		idStr := string(documentIDs)
		database.GetDB().Table("evidence").
			Select("id, file_name, file_url, file_size, file_type").
			Where("id IN (?)", idStr).
			Find(&documents)
		flow["documents"] = documents
	}

	canSign := s.checkCanSign(flow, userID)
	flow["canSign"] = canSign

	return flow, nil
}

func (s *ESignServiceImpl) SignDocument(ctx context.Context, esignID int64, userID int64, verifyCode string, signatureData string) error {
	var flow model.ESignFlow
	database.GetDB().Where("id = ? AND deleted_at IS NULL", esignID).First(&flow)
	if flow.ID == 0 {
		return nil
	}

	if flow.Status == esignStatusCompleted || flow.Status == esignStatusRevoked {
		return nil
	}

	if !s.isSigner(&flow, userID) {
		return nil
	}

	if !s.verifyCode(ctx, userID, verifyCode) {
		return nil
	}

	var signer model.ESignSigner
	database.GetDB().Where("flow_id = ? AND signer_id = ?", esignID, userID).First(&signer)
	if signer.ID == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"sign_status":   1,
		"sign_time":     time.Now(),
		"signature_data": signatureData,
		"sign_ip":       "",
	}

	err := database.GetDB().Model(&signer).Updates(updates).Error
	if err != nil {
		return err
	}

	if flow.Status == esignStatusPending {
		database.GetDB().Model(&flow).Update("status", esignStatusSigning)
	}

	signedCount := s.getSignedCount(esignID)
	totalSigners := len(flow.SignerIDs)
	if signedCount >= totalSigners {
		database.GetDB().Model(&flow).Update("status", esignStatusCompleted)
		s.sendEsignNotifications(flow.ID, flow.CaseID, flow.SignerIDs, flow.Title, "complete")
	}

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":      "esign_sign",
		"flowId":    esignID,
		"userId":    userID,
		"caseId":    flow.CaseID,
		"timestamp": time.Now(),
	})

	return nil
}

func (s *ESignServiceImpl) RevokeEsignFlow(ctx context.Context, esignID int64, userID int64, reason string) error {
	var flow model.ESignFlow
	database.GetDB().Where("id = ? AND deleted_at IS NULL", esignID).First(&flow)
	if flow.ID == 0 {
		return nil
	}

	if flow.CreatorID != userID {
		return nil
	}

	updates := map[string]interface{}{
		"status":       esignStatusRevoked,
		"revoke_reason": reason,
		"revoke_time":  time.Now(),
		"revoked_by":   userID,
	}

	err := database.GetDB().Model(&flow).Updates(updates).Error
	if err != nil {
		return err
	}

	s.sendEsignNotifications(flow.ID, flow.CaseID, flow.SignerIDs, flow.Title, "revoke")

	return nil
}

func (s *ESignServiceImpl) SendEsignVerifyCode(ctx context.Context, esignID int64, mobile string) error {
	var user model.User
	database.GetDB().Where("mobile = ? AND status = 1", mobile).First(&user)
	if user.ID == 0 {
		return nil
	}

	code := generateVerifyCode()

	key := fmt.Sprintf("esign:verify:%d:%d", user.ID, esignID)
	database.GetRedisClient().Set(ctx, key, code, esignVerifyCodeTTL*time.Second)

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":      "sms_verify",
		"mobile":    mobile,
		"code":      code,
		"esignId":   esignID,
		"expire":    esignVerifyCodeTTL,
		"timestamp": time.Now(),
	})

	logger.Info("Esign verify code sent", logger.String("mobile", mobile), logger.String("code", code))

	return nil
}

func (s *ESignServiceImpl) createSignerRecords(flowID int64, signerIDs []int64) {
	for _, signerID := range signerIDs {
		signer := &model.ESignSigner{
			FlowID:     flowID,
			SignerID:   signerID,
			SignStatus: 0,
		}
		database.GetDB().Create(signer)
	}
}

func (s *ESignServiceImpl) sendEsignNotifications(flowID, caseID int64, signerIDs []int64, title, action string) {
	for _, sid := range signerIDs {
		mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
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
		Where("flow_id = ? AND sign_status = 1", flowID).
		Count(&count)
	return int(count)
}

func (s *ESignServiceImpl) isSigner(flow *model.ESignFlow, userID int64) bool {
	for _, sid := range flow.SignerIDs {
		if sid == userID {
			return true
		}
	}
	return false
}

func (s *ESignServiceImpl) checkCanSign(flow map[string]interface{}, userID int64) bool {
	status, _ := flow["status"].(int)
	if status == esignStatusCompleted || status == esignStatusRevoked {
		return false
	}

	signerIDs, _ := flow["signer_ids"].([]byte)
	if signerIDs != nil {
		idStr := string(signerIDs)
		return utils.Int64InSlice(userID, utils.ParseInt64Slice(idStr))
	}

	return false
}

func (s *ESignServiceImpl) verifyCode(ctx context.Context, userID int64, code string) bool {
	key := fmt.Sprintf("esign:verify:%d:*", userID)
	keys, _ := database.GetRedisClient().Keys(ctx, key).Result()
	if len(keys) == 0 {
		return false
	}

	for _, k := range keys {
		savedCode, _ := database.GetRedisClient().Get(ctx, k).Result()
		if savedCode == code {
			database.GetRedisClient().Del(ctx, k)
			return true
		}
	}

	return false
}
