package service

import (
	"context"
)

type DispatchService interface {
	IntelligentDispatch(ctx context.Context, caseID int64, algorithm string, excludeIDs []int64, autoAssign bool, orgID int64) (map[string]interface{}, error)
	GetDispatchCandidates(ctx context.Context, caseID int64, orgID int64, topN int) ([]map[string]interface{}, error)
	GetDispatchConfig(ctx context.Context, orgID int64) (map[string]interface{}, error)
	UpdateDispatchConfig(ctx context.Context, orgID int64, config map[string]interface{}) error
	GetMediatorLoadStats(ctx context.Context, orgID int64) ([]map[string]interface{}, error)
	BatchIntelligentDispatch(ctx context.Context, caseIDs []int64, orgID int64, operatorID int64) (int, int, error)
}
