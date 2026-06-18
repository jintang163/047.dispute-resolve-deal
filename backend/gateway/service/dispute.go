package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type DisputeService interface {
	GetDisputeList(ctx context.Context, userID int64, role int32, orgID int64, page, pageSize int, status int32, typeID int64, keyword string) ([]*model.DisputeCase, int64, error)
	GetDisputeDetail(ctx context.Context, caseID int64, userID int64, role int32) (*model.DisputeCase, error)
	CreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence) error
	KioskCreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence, deviceID int64) (string, error)
	MiniAppCreateDispute(ctx context.Context, dispute *model.DisputeCase, evidence []*model.Evidence, userID int64) (string, error)
	AssignDispute(ctx context.Context, caseID int64, mediatorID int64, assignorID int64) error
	UrgeDispute(ctx context.Context, caseID int64, userID int64, urgeType int32, remark string) error
	UpdateDisputeStatus(ctx context.Context, caseID int64, status int32, userID int64, remark string) error
	GetDisputeHistory(ctx context.Context, caseID int64) ([]*model.DisputeHistory, error)
	GetDisputeProgress(ctx context.Context, caseNo string, idCard string) (*model.DisputeCase, []*model.DisputeHistory, error)
	GetDisputeTypes(ctx context.Context) ([]*model.DisputeType, error)
	UploadEvidence(ctx context.Context, caseID int64, fileType int32, fileName, fileURL string, fileSize int64, remark string, uploadFrom int32, userID int64) (*model.Evidence, error)
	GetEvidenceList(ctx context.Context, caseID int64, page, pageSize int) ([]*model.Evidence, int64, error)
	DeleteEvidence(ctx context.Context, evidenceID int64, userID int64) error
	BatchDeleteEvidence(ctx context.Context, evidenceIDs []int64, userID int64) error
	UpdateEvidenceRemark(ctx context.Context, evidenceID int64, remark string, userID int64) error
}
