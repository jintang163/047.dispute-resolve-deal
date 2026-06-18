package impl

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/court"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JudicialConfirmationServiceImpl struct{}

func NewJudicialConfirmationService() service.JudicialConfirmationService {
	return &JudicialConfirmationServiceImpl{}
}

func (s *JudicialConfirmationServiceImpl) GetConfirmationList(ctx context.Context, userID int64, role int32, orgID int64, page, pageSize int, status int32, keyword string) ([]*model.JudicialConfirmation, int64, error) {
	var list []*model.JudicialConfirmation
	var total int64

	db := database.GetDB().Model(&model.JudicialConfirmation{}).Where("deleted_at IS NULL")

	if role == constants.RoleMediator {
		db = db.Where("submit_by = ?", userID)
	} else if role >= constants.RoleLeader && role <= constants.RoleDirector {
		db = db.Where("organization_id = ?", orgID)
	}

	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		db = db.Where("confirm_no LIKE ? OR case_no LIKE ? OR applicant_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list)

	return list, total, nil
}

func (s *JudicialConfirmationServiceImpl) GetConfirmationDetail(ctx context.Context, confirmID int64, userID int64, role int32) (*model.JudicialConfirmation, error) {
	cacheKey := constants.RedisKeyPrefixJudicial + "detail:" + utils.Int64ToString(confirmID)
	var confirm model.JudicialConfirmation

	if cache.Get(ctx, cacheKey, &confirm) == nil && confirm.ID > 0 {
		return &confirm, nil
	}

	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm)
	if result.Error != nil {
		return nil, result.Error
	}

	if role == constants.RoleMediator && confirm.SubmitBy != userID {
		return nil, fmt.Errorf("permission denied")
	}

	cache.Set(ctx, cacheKey, confirm, 10*time.Minute)
	return &confirm, nil
}

func (s *JudicialConfirmationServiceImpl) GetConfirmationByNo(ctx context.Context, confirmNo string, idCard string) (*model.JudicialConfirmation, []*model.JudicialConfirmLog, error) {
	var confirm model.JudicialConfirmation
	result := database.GetDB().Where("confirm_no = ? AND (applicant_id_card = ? OR respondent_id_card = ?) AND deleted_at IS NULL",
		confirmNo, idCard, idCard).First(&confirm)
	if result.Error != nil {
		return nil, nil, result.Error
	}

	logs, err := s.GetConfirmLogs(ctx, confirm.ID)
	if err != nil {
		logger.Warn("Get confirm logs failed", zap.Int64("confirmId", confirm.ID), logger.Error(err))
	}

	return &confirm, logs, nil
}

func (s *JudicialConfirmationServiceImpl) CreateConfirmation(ctx context.Context, confirm *model.JudicialConfirmation, operatorID int64, operatorName string) (string, error) {
	confirm.ConfirmNo = utils.GenerateConfirmNo(confirm.OrganizationID)
	confirm.Status = model.JudicialStatusSubmitted
	confirm.SealStatus = model.SealStatusPending
	confirm.SubmitBy = operatorID
	confirm.SubmitByName = operatorName
	now := time.Now()
	confirm.SubmitTime = &now

	tx := database.GetDB().Begin()
	if err := tx.Create(confirm).Error; err != nil {
		tx.Rollback()
		logger.Error("Create judicial confirmation failed", logger.Error(err))
		return "", err
	}

	log := &model.JudicialConfirmLog{
		ConfirmID:    confirm.ID,
		ConfirmNo:    confirm.ConfirmNo,
		ActionType:   model.ActionTypeSubmit,
		ActionName:   "提交司法确认申请",
		OperatorID:   operatorID,
		OperatorName: operatorName,
		OperatorType: model.OperatorTypeAdmin,
		CreatedAt:    &now,
	}
	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()

	mq.SendAsync(constants.MQTopicJudicialSubmit, map[string]interface{}{
		"confirmId": confirm.ID,
		"confirmNo": confirm.ConfirmNo,
		"caseId":    confirm.CaseID,
		"operatorId": operatorID,
	})

	logger.Info("Judicial confirmation created",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
	)

	return confirm.ConfirmNo, nil
}

func (s *JudicialConfirmationServiceImpl) addConfirmLog(tx *gorm.DB, confirmID int64, confirmNo string, actionType int, actionName string, operatorID int64, operatorName string, operatorType int32) error {
	now := time.Now()
	log := &model.JudicialConfirmLog{
		ConfirmID:    confirmID,
		ConfirmNo:    confirmNo,
		ActionType:   actionType,
		ActionName:   actionName,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		OperatorType: operatorType,
		CreatedAt:    &now,
	}
	return tx.Create(log).Error
}

func (s *JudicialConfirmationServiceImpl) SubmitToCourt(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	if confirm.CourtID == 0 {
		return fmt.Errorf("court not configured")
	}

	var courtConfig model.CourtConfig
	if err := database.GetDB().Where("id = ? AND status = 1", confirm.CourtID).First(&courtConfig).Error; err != nil {
		return fmt.Errorf("court config not found: %v", err)
	}

	client := court.NewMicroCourtClient(&courtConfig)

	evidenceList := make([]string, 0)
	var evidences []*model.DisputeEvidence
	database.GetDB().Where("case_id = ?", confirm.CaseID).Find(&evidences)
	for _, e := range evidences {
		evidenceList = append(evidenceList, e.FileURL)
	}

	deadlineStr := ""
	if confirm.PerformanceDeadline != nil {
		deadlineStr = confirm.PerformanceDeadline.Format("2006-01-02")
	}

	courtReq := &court.SubmitConfirmationRequest{
		ConfirmNo:          confirm.ConfirmNo,
		CaseNo:             confirm.CaseNo,
		CaseTitle:          confirm.CaseTitle,
		ApplicantName:      confirm.ApplicantName,
		ApplicantPhone:     confirm.ApplicantPhone,
		ApplicantIdCard:    confirm.ApplicantIDCard,
		ApplicantAddress:   confirm.ApplicantAddress,
		RespondentName:     confirm.RespondentName,
		RespondentPhone:    confirm.RespondentPhone,
		RespondentIdCard:   confirm.RespondentIDCard,
		RespondentAddress:  confirm.RespondentAddress,
		AgreementContent:   confirm.AgreementContent,
		ConfirmAmount:      fmt.Sprintf("%.2f", confirm.ConfirmAmount),
		PerformanceDeadline: deadlineStr,
		EvidenceList:       evidenceList,
	}

	resp, err := client.SubmitConfirmation(courtReq)
	if err != nil {
		logger.Error("Submit to court failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return err
	}

	now := time.Now()
	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"status":      model.JudicialStatusReviewing,
		"sub_status":  21,
		"review_time": &now,
	}
	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, model.ActionTypeCourtAccept, "法院已受理",
		operatorID, operatorName, model.OperatorTypeCourt)

	tx.Commit()

	cache.DelByPrefix(ctx, constants.RedisKeyPrefixJudicial+"detail:")

	mq.SendAsync(constants.MQTopicJudicialStatus, map[string]interface{}{
		"confirmId":    confirm.ID,
		"confirmNo":    confirm.ConfirmNo,
		"oldStatus":    model.JudicialStatusSubmitted,
		"newStatus":    model.JudicialStatusReviewing,
		"courtCaseNo":  resp.CourtCaseNo,
		"operatorId":   operatorID,
	})

	logger.Info("Confirmation submitted to court",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
		zap.String("courtCaseNo", resp.CourtCaseNo),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) UpdateStatusFromCourt(ctx context.Context, confirmID int64, status int32, reviewOpinion string, courtData map[string]interface{}) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	oldStatus := confirm.Status
	now := time.Now()

	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"status":         status,
		"review_opinion": reviewOpinion,
		"review_time":    &now,
	}

	if courtData != nil {
		if courtCaseNo, ok := courtData["courtCaseNo"].(string); ok {
			updates["court_case_no"] = courtCaseNo
		}
		if confirmDate, ok := courtData["confirmDate"].(string); ok {
			if t, err := time.Parse("2006-01-02", confirmDate); err == nil {
				updates["confirm_date"] = t
			}
		}
		if confirmCourt, ok := courtData["confirmCourt"].(string); ok {
			updates["confirm_court"] = confirmCourt
		}
		if confirmJudge, ok := courtData["confirmJudge"].(string); ok {
			updates["confirm_judge"] = confirmJudge
		}
		if confirmDocumentNo, ok := courtData["confirmDocumentNo"].(string); ok {
			updates["confirm_document_no"] = confirmDocumentNo
		}
		if documentUrl, ok := courtData["documentUrl"].(string); ok {
			updates["document_url"] = documentUrl
		}
		if sealStatus, ok := courtData["sealStatus"].(int); ok {
			updates["seal_status"] = sealStatus
			if sealStatus == 1 {
				updates["seal_time"] = &now
			}
		}
	}

	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	actionType := model.ActionTypePass
	actionName := "审核通过"
	if status == model.JudicialStatusRejected {
		actionType = model.ActionTypeReject
		actionName = "审核驳回"
	} else if status == model.JudicialStatusConfirmed {
		actionType = model.ActionTypePass
		actionName = "司法确认通过"
	} else if status == model.JudicialStatusExpired {
		actionType = model.ActionTypeExpired
		actionName = "已失效"
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, actionType, actionName,
		0, "法院系统", model.OperatorTypeCourt)

	tx.Commit()

	cache.DelByPrefix(ctx, constants.RedisKeyPrefixJudicial+"detail:")

	mq.SendAsync(constants.MQTopicJudicialStatus, map[string]interface{}{
		"confirmId":   confirm.ID,
		"confirmNo":   confirm.ConfirmNo,
		"oldStatus":   oldStatus,
		"newStatus":   status,
		"courtData":   courtData,
	})

	logger.Info("Confirmation status updated from court",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
		zap.Int32("oldStatus", oldStatus),
		zap.Int32("newStatus", status),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) QueryCourtStatus(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	if confirm.CourtID == 0 {
		return fmt.Errorf("court not configured")
	}

	var courtConfig model.CourtConfig
	if err := database.GetDB().Where("id = ? AND status = 1", confirm.CourtID).First(&courtConfig).Error; err != nil {
		return fmt.Errorf("court config not found: %v", err)
	}

	client := court.NewMicroCourtClient(&courtConfig)

	courtCaseNo := ""
	if v, ok := confirm.CourtCode; ok {
		courtCaseNo = v
	}

	resp, err := client.QueryConfirmationStatus(confirm.ConfirmNo, courtCaseNo)
	if err != nil {
		logger.Error("Query court status failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return err
	}

	localStatus := court.ConvertCourtStatusToLocal(resp.Status)

	courtData := map[string]interface{}{
		"courtCaseNo":       resp.CourtCaseNo,
		"confirmDate":       resp.ConfirmDate,
		"confirmCourt":      resp.ConfirmCourt,
		"confirmJudge":      resp.ConfirmJudge,
		"confirmDocumentNo": resp.ConfirmDocumentNo,
		"documentUrl":       resp.DocumentUrl,
		"sealStatus":        resp.SealStatus,
		"sealTime":          resp.SealTime,
	}

	return s.UpdateStatusFromCourt(ctx, confirmID, localStatus, resp.ReviewOpinion, courtData)
}

func (s *JudicialConfirmationServiceImpl) GenerateConfirmationDocument(ctx context.Context, confirmID int64, operatorID int64, operatorName string) (string, error) {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return "", err
	}

	if confirm.CourtID == 0 {
		return "", fmt.Errorf("court not configured")
	}

	var courtConfig model.CourtConfig
	if err := database.GetDB().Where("id = ? AND status = 1", confirm.CourtID).First(&courtConfig).Error; err != nil {
		return "", fmt.Errorf("court config not found: %v", err)
	}

	client := court.NewMicroCourtClient(&courtConfig)

	courtCaseNo := ""
	if confirm.CourtCode != "" {
		courtCaseNo = confirm.CourtCode
	}

	resp, err := client.GetConfirmationDocument(confirm.ConfirmNo, courtCaseNo, confirm.ConfirmDocumentNo)
	if err != nil {
		logger.Error("Get confirmation document from court failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return "", err
	}

	pdfData, err := base64.StdEncoding.DecodeString(resp.DocumentContent)
	if err != nil {
		logger.Error("Decode document content failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return "", fmt.Errorf("decode document failed: %v", err)
	}

	fileName := fmt.Sprintf("司法确认书_%s.pdf", confirm.ConfirmNo)
	documentURL, err := s.ArchiveDocumentToMinIO(ctx, confirmID, pdfData, fileName)
	if err != nil {
		logger.Error("Archive document to MinIO failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return "", err
	}

	now := time.Now()
	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"document_url":  documentURL,
		"seal_status":   resp.SealStatus,
	}
	if resp.SealStatus == 1 {
		updates["seal_time"] = &now
	}
	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, model.ActionTypeDeliver, "确认书已送达",
		operatorID, operatorName, model.OperatorTypeCourt)

	tx.Commit()

	cache.DelByPrefix(ctx, constants.RedisKeyPrefixJudicial+"detail:")

	mq.SendAsync(constants.MQTopicJudicialSeal, map[string]interface{}{
		"confirmId":    confirm.ID,
		"confirmNo":    confirm.ConfirmNo,
		"documentUrl":  documentURL,
		"sealStatus":   resp.SealStatus,
		"operatorId":   operatorID,
	})

	logger.Info("Confirmation document generated and archived",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
		zap.String("documentUrl", documentURL),
	)

	return documentURL, nil
}

func (s *JudicialConfirmationServiceImpl) SealDocument(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	if confirm.CourtID == 0 {
		return fmt.Errorf("court not configured")
	}

	var courtConfig model.CourtConfig
	if err := database.GetDB().Where("id = ? AND status = 1", confirm.CourtID).First(&courtConfig).Error; err != nil {
		return fmt.Errorf("court config not found: %v", err)
	}

	client := court.NewMicroCourtClient(&courtConfig)

	courtCaseNo := ""
	if confirm.CourtCode != "" {
		courtCaseNo = confirm.CourtCode
	}

	resp, err := client.SealDocument(confirm.ConfirmNo, courtCaseNo, confirm.ConfirmDocumentNo, courtConfig.SealCertNo)
	if err != nil {
		logger.Error("Seal document with court failed", zap.Int64("confirmId", confirmID), logger.Error(err))
		return err
	}

	now := time.Now()
	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"seal_status": model.SealStatusDone,
		"seal_time":   &now,
	}
	if resp.SealedDocumentUrl != "" {
		updates["document_url"] = resp.SealedDocumentUrl

		pdfData, err := s.downloadPDF(resp.SealedDocumentUrl)
		if err == nil {
			fileName := fmt.Sprintf("司法确认书_%s.pdf", confirm.ConfirmNo)
			minioURL, err := s.ArchiveDocumentToMinIO(ctx, confirmID, pdfData, fileName)
			if err == nil {
				updates["document_url"] = minioURL
			}
		}
	}

	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, model.ActionTypeSeal, "确认书已签章",
		operatorID, operatorName, model.OperatorTypeCourt)

	tx.Commit()

	cache.DelByPrefix(ctx, constants.RedisKeyPrefixJudicial+"detail:")

	mq.SendAsync(constants.MQTopicJudicialSeal, map[string]interface{}{
		"confirmId":    confirm.ID,
		"confirmNo":    confirm.ConfirmNo,
		"sealStatus":   model.SealStatusDone,
		"operatorId":   operatorID,
	})

	logger.Info("Confirmation document sealed",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) downloadPDF(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *JudicialConfirmationServiceImpl) ArchiveDocumentToMinIO(ctx context.Context, confirmID int64, pdfData []byte, fileName string) (string, error) {
	minioClient := database.GetMinioClient()
	if minioClient == nil {
		logger.Warn("MinIO client not available, skipping archive")
		return "", nil
	}

	bucketName := "dispute-resolution"
	objectName := fmt.Sprintf("%s/%d/%s", constants.MinIOPathJudicial, confirmID, fileName)

	reader := bytes.NewReader(pdfData)

	_, err := minioClient.PutObject(ctx, bucketName, objectName, reader, int64(len(pdfData)), "application/pdf")
	if err != nil {
		logger.Error("Upload PDF to MinIO failed",
			zap.String("objectName", objectName),
			logger.Error(err),
		)
		return "", err
	}

	url := fmt.Sprintf("/minio/%s/%s", bucketName, objectName)
	logger.Info("PDF archived to MinIO",
		zap.String("objectName", objectName),
		zap.Int("size", len(pdfData)),
	)

	return url, nil
}

func (s *JudicialConfirmationServiceImpl) GetConfirmLogs(ctx context.Context, confirmID int64) ([]*model.JudicialConfirmLog, error) {
	var logs []*model.JudicialConfirmLog
	result := database.GetDB().Where("confirm_id = ?", confirmID).
		Order("created_at DESC").
		Find(&logs)
	return logs, result.Error
}

func (s *JudicialConfirmationServiceImpl) SendExpirationReminder(ctx context.Context, confirmID int64) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	if confirm.Status != model.JudicialStatusConfirmed {
		return fmt.Errorf("only confirmed confirmation can send expiration reminder")
	}

	if confirm.ExpirationRemindSent == 1 {
		return nil
	}

	reminderContent := fmt.Sprintf("【司法确认提醒】您申请的司法确认（编号：%s）已超过履行期限，请尽快履行。如未履行，对方可向法院申请强制执行。",
		confirm.ConfirmNo)

	if confirm.CourtID > 0 {
		var courtConfig model.CourtConfig
		if err := database.GetDB().Where("id = ? AND status = 1", confirm.CourtID).First(&courtConfig).Error; err == nil {
			client := court.NewMicroCourtClient(&courtConfig)
			_, err := client.SendReminder(confirm.ConfirmNo, 2, reminderContent, confirm.ApplicantPhone)
			if err != nil {
				logger.Warn("Send expiration reminder via court failed", zap.Int64("confirmId", confirmID), logger.Error(err))
			}
		}
	}

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":    "sms",
		"phone":   confirm.ApplicantPhone,
		"content": reminderContent,
		"confirmNo": confirm.ConfirmNo,
	})
	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":    "sms",
		"phone":   confirm.RespondentPhone,
		"content": reminderContent,
		"confirmNo": confirm.ConfirmNo,
	})

	now := time.Now()
	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"expiration_remind_sent": 1,
	}
	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, model.ActionTypeExpireRemind, "失效提醒已发送",
		0, "系统", model.OperatorTypeSystem)

	tx.Commit()

	mq.SendAsync(constants.MQTopicJudicialRemind, map[string]interface{}{
		"confirmId":    confirm.ID,
		"confirmNo":    confirm.ConfirmNo,
		"remindType":   "expiration",
		"targetPhones": []string{confirm.ApplicantPhone, confirm.RespondentPhone},
	})

	logger.Info("Expiration reminder sent",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) SendPerformanceReminder(ctx context.Context, confirmID int64) error {
	var confirm model.JudicialConfirmation
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", confirmID).First(&confirm).Error; err != nil {
		return err
	}

	if confirm.Status != model.JudicialStatusConfirmed {
		return fmt.Errorf("only confirmed confirmation can send performance reminder")
	}

	if confirm.PerformanceRemindSent == 1 {
		return nil
	}

	daysLeft := 0
	if confirm.PerformanceDeadline != nil {
		daysLeft = int(time.Until(*confirm.PerformanceDeadline).Hours() / 24)
	}

	reminderContent := fmt.Sprintf("【司法确认提醒】您申请的司法确认（编号：%s）履行期限即将到期，请尽快履行。剩余%d天。",
		confirm.ConfirmNo, daysLeft)

	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":    "sms",
		"phone":   confirm.ApplicantPhone,
		"content": reminderContent,
		"confirmNo": confirm.ConfirmNo,
	})
	mq.SendAsync(constants.MQTopicNotification, map[string]interface{}{
		"type":    "sms",
		"phone":   confirm.RespondentPhone,
		"content": reminderContent,
		"confirmNo": confirm.ConfirmNo,
	})

	now := time.Now()
	tx := database.GetDB().Begin()

	updates := map[string]interface{}{
		"performance_remind_sent": 1,
	}
	if err := tx.Model(&confirm).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.addConfirmLog(tx, confirm.ID, confirm.ConfirmNo, model.ActionTypePerformRemind, "履行提醒已发送",
		0, "系统", model.OperatorTypeSystem)

	tx.Commit()

	mq.SendAsync(constants.MQTopicJudicialRemind, map[string]interface{}{
		"confirmId":    confirm.ID,
		"confirmNo":    confirm.ConfirmNo,
		"remindType":   "performance",
		"targetPhones": []string{confirm.ApplicantPhone, confirm.RespondentPhone},
	})

	logger.Info("Performance reminder sent",
		zap.Int64("confirmId", confirm.ID),
		zap.String("confirmNo", confirm.ConfirmNo),
		zap.Int("daysLeft", daysLeft),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) CheckExpiredConfirmations(ctx context.Context) ([]*model.JudicialConfirmation, error) {
	now := time.Now()
	var list []*model.JudicialConfirmation

	result := database.GetDB().Model(&model.JudicialConfirmation{}).
		Where("status = ? AND performance_deadline < ? AND deleted_at IS NULL",
			model.JudicialStatusConfirmed, now).
		Find(&list)

	return list, result.Error
}

func (s *JudicialConfirmationServiceImpl) CheckPerformanceDeadline(ctx context.Context, daysBefore int) ([]*model.JudicialConfirmation, error) {
	deadline := time.Now().AddDate(0, 0, daysBefore)
	var list []*model.JudicialConfirmation

	result := database.GetDB().Model(&model.JudicialConfirmation{}).
		Where("status = ? AND performance_deadline <= ? AND performance_remind_sent = 0 AND deleted_at IS NULL",
			model.JudicialStatusConfirmed, deadline).
		Find(&list)

	return list, result.Error
}

func (s *JudicialConfirmationServiceImpl) CourtConfigList(ctx context.Context, orgID int64, page, pageSize int) ([]*model.CourtConfig, int64, error) {
	var list []*model.CourtConfig
	var total int64

	db := database.GetDB().Model(&model.CourtConfig{}).Where("deleted_at IS NULL")

	if orgID > 0 {
		db = db.Where("organization_id = ?", orgID)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("sort_order ASC, id DESC").Find(&list)

	return list, total, nil
}

func (s *JudicialConfirmationServiceImpl) CourtConfigDetail(ctx context.Context, configID int64) (*model.CourtConfig, error) {
	var config model.CourtConfig
	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", configID).First(&config)
	if result.Error != nil {
		return nil, result.Error
	}
	return &config, nil
}

func (s *JudicialConfirmationServiceImpl) CreateCourtConfig(ctx context.Context, config *model.CourtConfig, operatorID int64) error {
	now := time.Now()
	config.CreatedAt = &now
	config.UpdatedAt = &now

	tx := database.GetDB().Begin()
	if err := tx.Create(config).Error; err != nil {
		tx.Rollback()
		logger.Error("Create court config failed", logger.Error(err))
		return err
	}
	tx.Commit()

	logger.Info("Court config created",
		zap.Int64("configId", config.ID),
		zap.String("courtName", config.CourtName),
		zap.Int64("operatorId", operatorID),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) UpdateCourtConfig(ctx context.Context, config *model.CourtConfig, operatorID int64) error {
	now := time.Now()
	config.UpdatedAt = &now

	tx := database.GetDB().Begin()
	if err := tx.Model(config).Updates(map[string]interface{}{
		"court_code":       config.CourtCode,
		"court_name":       config.CourtName,
		"court_level":      config.CourtLevel,
		"jurisdiction_area": config.JurisdictionArea,
		"address":          config.Address,
		"contact":          config.Contact,
		"phone":            config.Phone,
		"api_endpoint":     config.APIEndpoint,
		"api_app_id":       config.APIAppID,
		"api_secret":       config.APISecret,
		"api_public_key":   config.APIPublicKey,
		"seal_cert_no":     config.SealCertNo,
		"seal_image_url":   config.SealImageURL,
		"sort_order":       config.SortOrder,
		"status":           config.Status,
		"updated_at":       &now,
	}).Error; err != nil {
		tx.Rollback()
		logger.Error("Update court config failed", logger.Error(err))
		return err
	}
	tx.Commit()

	logger.Info("Court config updated",
		zap.Int64("configId", config.ID),
		zap.String("courtName", config.CourtName),
		zap.Int64("operatorId", operatorID),
	)

	return nil
}

func (s *JudicialConfirmationServiceImpl) DeleteCourtConfig(ctx context.Context, configID int64, operatorID int64) error {
	now := time.Now()
	tx := database.GetDB().Begin()

	if err := tx.Model(&model.CourtConfig{}).
		Where("id = ?", configID).
		Update("deleted_at", &now).Error; err != nil {
		tx.Rollback()
		logger.Error("Delete court config failed", logger.Error(err))
		return err
	}
	tx.Commit()

	logger.Info("Court config deleted",
		zap.Int64("configId", configID),
		zap.Int64("operatorId", operatorID),
	)

	return nil
}
