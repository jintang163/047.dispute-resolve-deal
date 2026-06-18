package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type JudicialConfirmationService interface {
	GetConfirmationList(ctx context.Context, userID int64, role int32, orgID int64, page, pageSize int, status int32, keyword string) ([]*model.JudicialConfirmation, int64, error)
	GetConfirmationDetail(ctx context.Context, confirmID int64, userID int64, role int32) (*model.JudicialConfirmation, error)
	GetConfirmationByNo(ctx context.Context, confirmNo string, idCard string) (*model.JudicialConfirmation, []*model.JudicialConfirmLog, error)
	CreateConfirmation(ctx context.Context, confirm *model.JudicialConfirmation, operatorID int64, operatorName string) (string, error)
	SubmitToCourt(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error
	UpdateStatusFromCourt(ctx context.Context, confirmID int64, status int32, reviewOpinion string, courtData map[string]interface{}) error
	QueryCourtStatus(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error
	GenerateConfirmationDocument(ctx context.Context, confirmID int64, operatorID int64, operatorName string) (string, error)
	SealDocument(ctx context.Context, confirmID int64, operatorID int64, operatorName string) error
	ArchiveDocumentToMinIO(ctx context.Context, confirmID int64, pdfData []byte, fileName string) (string, error)
	GetConfirmLogs(ctx context.Context, confirmID int64) ([]*model.JudicialConfirmLog, error)
	SendExpirationReminder(ctx context.Context, confirmID int64) error
	SendPerformanceReminder(ctx context.Context, confirmID int64) error
	CheckExpiredConfirmations(ctx context.Context) ([]*model.JudicialConfirmation, error)
	CheckPerformanceDeadline(ctx context.Context, daysBefore int) ([]*model.JudicialConfirmation, error)
	CourtConfigList(ctx context.Context, orgID int64, page, pageSize int) ([]*model.CourtConfig, int64, error)
	CourtConfigDetail(ctx context.Context, configID int64) (*model.CourtConfig, error)
	CreateCourtConfig(ctx context.Context, config *model.CourtConfig, operatorID int64) error
	UpdateCourtConfig(ctx context.Context, config *model.CourtConfig, operatorID int64) error
	DeleteCourtConfig(ctx context.Context, configID int64, operatorID int64) error
}
