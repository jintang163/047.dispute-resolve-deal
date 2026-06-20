package service

import (
	"context"
	"sync"

	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/gateway/service/impl"
)

type TimeoutUrgeService interface {
	DetectAndUrgePendingCases(ctx context.Context) (int, error)
	DetectAndUrgeMediatingCases(ctx context.Context) (int, error)
	DetectAndEscalateUrgedCases(ctx context.Context) (int, error)

	UrgeCase(ctx context.Context, caseID int64, urgeType int, urgeContent string, operatorID int64, operatorName string) error
	EscalateCase(ctx context.Context, caseID int64, escalationType int, reason string) (*model.DisputeEscalation, error)

	GetEscalationList(ctx context.Context, orgID int64, toLevel int, status int32, page, pageSize int) ([]*model.DisputeEscalation, int64, error)
	GetEscalationDetail(ctx context.Context, escalationID int64) (*model.DisputeEscalation, error)
	GetCaseEscalationList(ctx context.Context, caseID int64) ([]*model.DisputeEscalation, error)
	GetCaseUrgeList(ctx context.Context, caseID int64) ([]*model.DisputeUrge, error)

	HandleEscalation(ctx context.Context, escalationID int64, operatorID int64, operatorName string, remark string) error
	CloseEscalation(ctx context.Context, escalationID int64, operatorID int64, operatorName string, remark string) error
}

var (
	timeoutUrgeServiceInstance TimeoutUrgeService
	timeoutUrgeServiceOnce     sync.Once
)

func InitTimeoutUrgeService() {
	timeoutUrgeServiceOnce.Do(func() {
		timeoutUrgeServiceInstance = impl.NewTimeoutUrgeService()
		logger.Info("TimeoutUrge service initialized")
	})
}

func TimeoutUrgeServiceInst() TimeoutUrgeService {
	if timeoutUrgeServiceInstance == nil {
		InitTimeoutUrgeService()
	}
	return timeoutUrgeServiceInstance
}
