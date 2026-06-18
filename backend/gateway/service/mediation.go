package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type MediationService interface {
	CreateMediationRecord(ctx context.Context, record *model.MediationRecord) error
	GetMediationRecords(ctx context.Context, caseID int64, page, pageSize int) ([]*model.MediationRecord, int64, error)
	UpdateMediationRecord(ctx context.Context, recordID int64, content string, result int32, userID int64) error
	GetAISummary(ctx context.Context, recordID int64, caseID int64) (string, error)
}
