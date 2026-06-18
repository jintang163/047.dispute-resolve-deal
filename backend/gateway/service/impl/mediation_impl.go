package impl

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/gateway/service"
)

type MediationServiceImpl struct{}

func NewMediationService() service.MediationService {
	return &MediationServiceImpl{}
}

func (s *MediationServiceImpl) CreateMediationRecord(ctx context.Context, record *model.MediationRecord) error {
	record.MediationTime = time.Now()
	return database.GetDB().Create(record).Error
}

func (s *MediationServiceImpl) GetMediationRecords(ctx context.Context, caseID int64, page, pageSize int) ([]*model.MediationRecord, int64, error) {
	var records []*model.MediationRecord
	var total int64

	db := database.GetDB().Model(&model.MediationRecord{}).Where("case_id = ? AND deleted_at IS NULL", caseID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records)

	return records, total, nil
}

func (s *MediationServiceImpl) UpdateMediationRecord(ctx context.Context, recordID int64, content string, result int32, userID int64) error {
	updates := map[string]interface{}{
		"content":     content,
		"result":      result,
		"updated_by":  userID,
		"updated_at":  time.Now(),
	}
	return database.GetDB().Model(&model.MediationRecord{}).Where("id = ?", recordID).Updates(updates).Error
}

func (s *MediationServiceImpl) GetAISummary(ctx context.Context, recordID int64, caseID int64) (string, error) {
	var record model.MediationRecord
	database.GetDB().Where("id = ?", recordID).First(&record)

	aiService := NewAIService()
	summary, err := aiService.GenerateMediationSummary(ctx, caseID, record.Content)
	if err != nil {
		return "", err
	}

	database.GetDB().Model(&record).Update("ai_summary", summary)
	return summary, nil
}
