package service

import "context"

type PointsService interface {
	GetPointsSummary(ctx context.Context, memberID int64) (map[string]interface{}, error)
	AddPoints(ctx context.Context, memberID int64, points int, businessType, businessNo string, description string) error
	DeductPoints(ctx context.Context, memberID int64, points int, businessType, businessNo string, description string) error
	GetPointsRecords(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetPointsRules(ctx context.Context, ruleType string) ([]map[string]interface{}, error)
	CreatePointsRule(ctx context.Context, req map[string]interface{}) error
	UpdatePointsRule(ctx context.Context, id int64, req map[string]interface{}) error
	DeletePointsRule(ctx context.Context, id int64) error
	ExchangeGift(ctx context.Context, memberID int64, giftID int64, quantity int, receiverName, receiverPhone, receiverAddress, remark string) (int64, error)
	ProcessExpiredPoints(ctx context.Context) error
}
