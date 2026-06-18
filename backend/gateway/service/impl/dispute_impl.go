package impl

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type DisputeServiceImpl struct{}

func NewDisputeService() service.DisputeService {
	return &DisputeServiceImpl{}
}

func (s *DisputeServiceImpl) GetDisputeList(ctx context.Context, userID int64, role int32, orgID int64, page, pageSize int, status int32, typeID int64, keyword string) ([]*model.DisputeCase, int64, error) {
	cacheKey := constants.RedisPrefixDisputeList + utils.Int64ToString(userID) + ":" + utils.IntToString(page) + ":" + utils.IntToString(pageSize)
	var result struct {
		List  []*model.DisputeCase
		Total int64
	}

	if cache.Get(ctx, cacheKey, &result) == nil && result.List != nil {
		return result.List, result.Total, nil
	}

	var cases []*model.DisputeCase
	var total int64

	db := database.GetDB().Model(&model.DisputeCase{}).Where("deleted_at IS NULL")

	if role == constants.RoleMediator {
		db = db.Where("mediator_id = ?", userID)
	} else if role >= constants.RoleLeader && role <= constants.RoleDirector {
		db = db.Where("organization_id = ?", orgID)
	}

	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if typeID > 0 {
		db = db.Where("type_id = ?", typeID)
	}
	if keyword != "" {
		db = db.Where("title LIKE ? OR case_no LIKE ? OR applicant_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&cases)

	result.List = cases
	result.Total = total
	cache.Set(ctx, cacheKey, result, 5*time.Minute)

	return cases, total, nil
}

func (s *DisputeServiceImpl) GetDisputeDetail(ctx context.Context, caseID int64, userID int64, role int32) (*model.DisputeCase, error) {
	cacheKey := constants.RedisPrefixDisputeDetail + utils.Int64ToString(caseID)
	var caseData model.DisputeCase

	if cache.Get(ctx, cacheKey, &caseData) == nil && caseData.ID > 0 {
		return &caseData, nil
	}

	result := database.GetDB().Where("id = ? AND deleted_at IS NULL", caseID).First(&caseData)
	if result.Error != nil {
		return nil, result.Error
	}

	if role == constants.RoleMediator && caseData.MediatorID != userID {
		return nil, nil
	}

	cache.Set(ctx, cacheKey, caseData, 30*time.Minute)
	return &caseData, nil
}

func (s *DisputeServiceImpl) CreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence) error {
	dispute.CaseNo = utils.GenerateCaseNo(dispute.OrganizationID)
	dispute.Status = constants.CaseStatusPending

	tx := database.GetDB().Begin()
	if err := tx.Create(dispute).Error; err != nil {
		tx.Rollback()
		logger.Error("Create dispute error", logger.Error(err))
		return err
	}

	for _, e := range evidence {
		e.CaseID = dispute.ID
		if err := tx.Create(e).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	history := &model.DisputeHistory{
		CaseID:     dispute.ID,
		ActionType: constants.HistoryActionCreate,
		ActionName: "创建案件",
		OperatorID: dispute.CreatedBy,
	}
	tx.Create(history)

	tx.Commit()

	cache.DelByPrefix(ctx, constants.RedisPrefixDisputeList)

	mq.SendAsync(constants.MQTopicDisputeCreated, map[string]interface{}{
		"caseId":   dispute.ID,
		"caseNo":   dispute.CaseNo,
		"title":    dispute.Title,
		"orgId":    dispute.OrganizationID,
		"createAt": time.Now(),
	})

	return nil
}

func (s *DisputeServiceImpl) KioskCreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence, deviceID int64) (string, error) {
	dispute.Source = constants.SourceKiosk
	err := s.CreateDispute(ctx, dispute, evidence)
	if err != nil {
		return "", err
	}
	return dispute.CaseNo, nil
}

func (s *DisputeServiceImpl) MiniAppCreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence, userID int64) (string, error) {
	dispute.Source = constants.SourceMiniApp
	dispute.ApplicantID = userID
	dispute.CreatedBy = userID
	err := s.CreateDispute(ctx, dispute, evidence)
	if err != nil {
		return "", err
	}
	return dispute.CaseNo, nil
}

func (s *DisputeServiceImpl) AssignDispute(ctx context.Context, caseID int64, mediatorID int64, assignorID int64) error {
	var mediator model.User
	database.GetDB().Select("real_name").Where("id = ?", mediatorID).First(&mediator)

	updates := map[string]interface{}{
		"mediator_id":   mediatorID,
		"mediator_name": mediator.RealName,
		"status":        constants.CaseStatusMediating,
		"assigned_at":   time.Now(),
	}

	result := database.GetDB().Model(&model.DisputeCase{}).Where("id = ?", caseID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	history := &model.DisputeHistory{
		CaseID:     caseID,
		ActionType: constants.HistoryActionAssign,
		ActionName: "分派案件",
		Remark:     "分派给调解员：" + mediator.RealName,
		OperatorID: assignorID,
	}
	database.GetDB().Create(history)

	cache.DelByPrefix(ctx, constants.RedisPrefixDisputeList)
	cache.Del(ctx, constants.RedisPrefixDisputeDetail+utils.Int64ToString(caseID))

	mq.SendAsync(constants.MQTopicDisputeAssigned, map[string]interface{}{
		"caseId":     caseID,
		"mediatorId": mediatorID,
		"assignorId": assignorID,
	})

	return nil
}

func (s *DisputeServiceImpl) UrgeDispute(ctx context.Context, caseID int64, userID int64, urgeType int32, remark string) error {
	history := &model.DisputeHistory{
		CaseID:     caseID,
		ActionType: constants.HistoryActionUrge,
		ActionName: "催办案件",
		Remark:     remark,
		OperatorID: userID,
	}
	database.GetDB().Create(history)

	mq.SendAsync(constants.MQTopicDisputeUrged, map[string]interface{}{
		"caseId":   caseID,
		"userId":   userID,
		"urgeType": urgeType,
		"remark":   remark,
	})

	return nil
}

func (s *DisputeServiceImpl) UpdateDisputeStatus(ctx context.Context, caseID int64, status int32, userID int64, remark string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == constants.CaseStatusClosed {
		updates["closed_at"] = time.Now()
	}

	result := database.GetDB().Model(&model.DisputeCase{}).Where("id = ?", caseID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	history := &model.DisputeHistory{
		CaseID:     caseID,
		ActionType: constants.HistoryActionStatus,
		ActionName: "更新状态",
		Remark:     remark,
		OperatorID: userID,
	}
	database.GetDB().Create(history)

	cache.DelByPrefix(ctx, constants.RedisPrefixDisputeList)
	cache.Del(ctx, constants.RedisPrefixDisputeDetail+utils.Int64ToString(caseID))

	return nil
}

func (s *DisputeServiceImpl) GetDisputeHistory(ctx context.Context, caseID int64) ([]*model.DisputeHistory, error) {
	var history []*model.DisputeHistory
	result := database.GetDB().Where("case_id = ?", caseID).Order("created_at ASC").Find(&history)
	if result.Error != nil {
		return nil, result.Error
	}
	return history, nil
}

func (s *DisputeServiceImpl) GetDisputeProgress(ctx context.Context, caseNo string, idCard string) (*model.DisputeCase, []*model.DisputeHistory, error) {
	var caseData model.DisputeCase
	result := database.GetDB().Where("case_no = ? AND applicant_idcard = ? AND deleted_at IS NULL",
		caseNo, idCard).First(&caseData)
	if result.Error != nil {
		return nil, nil, result.Error
	}

	history, err := s.GetDisputeHistory(ctx, caseData.ID)
	if err != nil {
		return nil, nil, err
	}

	return &caseData, history, nil
}

func (s *DisputeServiceImpl) GetDisputeTypes(ctx context.Context) ([]*model.DisputeType, error) {
	cacheKey := constants.RedisPrefixDisputeTypes
	var types []*model.DisputeType

	if cache.Get(ctx, cacheKey, &types) == nil && len(types) > 0 {
		return types, nil
	}

	result := database.GetDB().Where("status = 1 AND deleted_at IS NULL").Order("level, sort_order").Find(&types)
	if result.Error != nil {
		return nil, result.Error
	}

	cache.Set(ctx, cacheKey, types, 24*time.Hour)
	return types, nil
}

func (s *DisputeServiceImpl) UploadEvidence(ctx context.Context, caseID int64, fileType int32, fileName, fileURL string, fileSize int64, remark string, uploadFrom int32, userID int64) (*model.Evidence, error) {
	evidence := &model.Evidence{
		CaseID:     caseID,
		FileType:   fileType,
		FileName:   fileName,
		FileURL:    fileURL,
		FileSize:   fileSize,
		Remark:     remark,
		SortOrder:  0,
		UploadFrom: uploadFrom,
		UploadedBy: userID,
	}

	result := database.GetDB().Create(evidence)
	if result.Error != nil {
		return nil, result.Error
	}

	return evidence, nil
}

func (s *DisputeServiceImpl) GetEvidenceList(ctx context.Context, caseID int64, page, pageSize int) ([]*model.Evidence, int64, error) {
	var evidence []*model.Evidence
	var total int64

	db := database.GetDB().Model(&model.Evidence{}).Where("case_id = ? AND deleted_at IS NULL", caseID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("sort_order, created_at DESC").Find(&evidence)

	return evidence, total, nil
}

func (s *DisputeServiceImpl) DeleteEvidence(ctx context.Context, evidenceID int64, userID int64) error {
	return database.GetDB().Model(&model.Evidence{}).Where("id = ?", evidenceID).Update("deleted_at", time.Now()).Error
}

func (s *DisputeServiceImpl) BatchDeleteEvidence(ctx context.Context, evidenceIDs []int64, userID int64) error {
	return database.GetDB().Model(&model.Evidence{}).Where("id IN ?", evidenceIDs).Update("deleted_at", time.Now()).Error
}

func (s *DisputeServiceImpl) UpdateEvidenceRemark(ctx context.Context, evidenceID int64, remark string, userID int64) error {
	return database.GetDB().Model(&model.Evidence{}).Where("id = ?", evidenceID).Update("remark", remark).Error
}
