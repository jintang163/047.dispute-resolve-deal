package service

import "context"

type GiftService interface {
	GetGiftList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetGiftDetail(ctx context.Context, id int64) (map[string]interface{}, error)
	CreateGift(ctx context.Context, req map[string]interface{}) (int64, error)
	UpdateGift(ctx context.Context, id int64, req map[string]interface{}) error
	DeleteGift(ctx context.Context, id int64) error
	GetGiftCategories(ctx context.Context) ([]map[string]interface{}, error)
	CreateGiftCategory(ctx context.Context, req map[string]interface{}) (int64, error)
	UpdateGiftCategory(ctx context.Context, id int64, req map[string]interface{}) error
	DeleteGiftCategory(ctx context.Context, id int64) error
	GetExchangeList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetExchangeDetail(ctx context.Context, id int64) (map[string]interface{}, error)
	AuditExchange(ctx context.Context, id int64, status int32, remark string) error
	ShipExchange(ctx context.Context, id int64, expressCompany, expressNo string) error
	ReceiveExchange(ctx context.Context, id int64) error
	CancelExchange(ctx context.Context, id int64, reason string) error
	GetMemberExchanges(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetGiftStatistics(ctx context.Context) (map[string]interface{}, error)
}
