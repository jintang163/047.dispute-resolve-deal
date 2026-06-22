package model

import "time"

type Gift struct {
	BaseModel
	GiftNo         string     `gorm:"size:32;uniqueIndex" json:"giftNo"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	CategoryID     int64      `gorm:"index" json:"categoryId"`
	CategoryName   string     `gorm:"size:64" json:"categoryName"`
	Description    string     `gorm:"type:text" json:"description"`
	ImageURL       string     `gorm:"size:512" json:"imageUrl"`
	Images         string     `gorm:"type:text" json:"images"`
	PointsRequired int        `gorm:"not null;default:0;index" json:"pointsRequired"`
	MarketPrice    float64    `json:"marketPrice"`
	Stock          int        `gorm:"default:0" json:"stock"`
	SoldCount      int        `gorm:"default:0" json:"soldCount"`
	SortOrder      int        `gorm:"default:0;index" json:"sortOrder"`
	IsHot          int        `gorm:"default:0;index" json:"isHot"`
	IsNew          int        `gorm:"default:0" json:"isNew"`
	Status         int        `gorm:"default:1;index" json:"status"`
	ExchangeLimit  int        `gorm:"default:0" json:"exchangeLimit"`
	ValidStartDate *time.Time `json:"validStartDate"`
	ValidEndDate   *time.Time `json:"validEndDate"`
}

func (Gift) TableName() string {
	return "gift"
}

type GiftCategory struct {
	BaseModel
	CategoryCode string `gorm:"size:64;uniqueIndex" json:"categoryCode"`
	CategoryName string `gorm:"size:64;not null" json:"categoryName"`
	Icon         string `gorm:"size:256" json:"icon"`
	SortOrder    int    `gorm:"default:0;index" json:"sortOrder"`
	Status       int    `gorm:"default:1;index" json:"status"`
}

func (GiftCategory) TableName() string {
	return "gift_category"
}

type GiftExchange struct {
	BaseModel
	ExchangeNo      string     `gorm:"size:32;uniqueIndex" json:"exchangeNo"`
	MemberID        int64      `gorm:"index;not null" json:"memberId"`
	MemberName      string     `gorm:"size:64" json:"memberName"`
	MemberPhone     string     `gorm:"size:20" json:"memberPhone"`
	OrganizationID  int64      `gorm:"column:organization_id;index" json:"organizationId"`
	GiftID          int64      `gorm:"index;not null" json:"giftId"`
	GiftName        string     `gorm:"size:128" json:"giftName"`
	GiftImage       string     `gorm:"size:512" json:"giftImage"`
	GiftPoints      int        `json:"giftPoints"`
	Quantity        int        `gorm:"default:1" json:"quantity"`
	TotalPoints     int        `json:"totalPoints"`
	ReceiverName    string     `gorm:"size:64" json:"receiverName"`
	ReceiverPhone   string     `gorm:"size:20" json:"receiverPhone"`
	ReceiverAddress string     `gorm:"size:512" json:"receiverAddress"`
	ExpressCompany  string     `gorm:"size:64" json:"expressCompany"`
	ExpressNo       string     `gorm:"size:64" json:"expressNo"`
	Status          int32      `gorm:"default:10;index" json:"status"`
	AuditID         int64      `json:"auditId"`
	AuditTime       *time.Time `json:"auditTime"`
	AuditRemark     string     `gorm:"size:512" json:"auditRemark"`
	ShipTime        *time.Time `json:"shipTime"`
	ReceiveTime     *time.Time `json:"receiveTime"`
	CancelReason    string     `gorm:"size:512" json:"cancelReason"`
	Remark          string     `gorm:"size:512" json:"remark"`
}

func (GiftExchange) TableName() string {
	return "gift_exchange"
}

type GiftQuery struct {
	BaseQuery
	CategoryID int64  `form:"categoryId" json:"categoryId"`
	Status     int    `form:"status" json:"status"`
	IsHot      int    `form:"isHot" json:"isHot"`
	IsNew      int    `form:"isNew" json:"isNew"`
	GiftNo     string `form:"giftNo" json:"giftNo"`
	Name       string `form:"name" json:"name"`
	MinPoints  int    `form:"minPoints" json:"minPoints"`
	MaxPoints  int    `form:"maxPoints" json:"maxPoints"`
}

type GiftExchangeQuery struct {
	BaseQuery
	MemberID int64  `form:"memberId" json:"memberId"`
	OrgID    int64  `form:"orgId" json:"orgId"`
	GiftID   int64  `form:"giftId" json:"giftId"`
	Status   int    `form:"status" json:"status"`
	ExchangeNo string `form:"exchangeNo" json:"exchangeNo"`
	DateRangeQuery
}

type CreateGiftRequest struct {
	GiftNo         string   `json:"giftNo" binding:"required"`
	Name           string   `json:"name" binding:"required"`
	CategoryID     int64    `json:"categoryId" binding:"required"`
	CategoryName   string   `json:"categoryName"`
	Description    string   `json:"description"`
	ImageURL       string   `json:"imageUrl"`
	Images         []string `json:"images"`
	PointsRequired int      `json:"pointsRequired" binding:"required,min=0"`
	MarketPrice    float64  `json:"marketPrice"`
	Stock          int      `json:"stock"`
	SortOrder      int      `json:"sortOrder"`
	IsHot          int      `json:"isHot"`
	IsNew          int      `json:"isNew"`
	Status         int      `json:"status"`
	ExchangeLimit  int      `json:"exchangeLimit"`
	ValidStartDate string   `json:"validStartDate"`
	ValidEndDate   string   `json:"validEndDate"`
}

type UpdateGiftRequest struct {
	ID             int64    `json:"id" binding:"required"`
	Name           string   `json:"name"`
	CategoryID     int64    `json:"categoryId"`
	CategoryName   string   `json:"categoryName"`
	Description    string   `json:"description"`
	ImageURL       string   `json:"imageUrl"`
	Images         []string `json:"images"`
	PointsRequired int      `json:"pointsRequired"`
	MarketPrice    float64  `json:"marketPrice"`
	Stock          int      `json:"stock"`
	SortOrder      int      `json:"sortOrder"`
	IsHot          int      `json:"isHot"`
	IsNew          int      `json:"isNew"`
	Status         int      `json:"status"`
	ExchangeLimit  int      `json:"exchangeLimit"`
	ValidStartDate string   `json:"validStartDate"`
	ValidEndDate   string   `json:"validEndDate"`
}

type CreateGiftCategoryRequest struct {
	CategoryCode string `json:"categoryCode" binding:"required"`
	CategoryName string `json:"categoryName" binding:"required"`
	Icon         string `json:"icon"`
	SortOrder    int    `json:"sortOrder"`
	Status       int    `json:"status"`
}

type UpdateGiftCategoryRequest struct {
	ID           int64  `json:"id" binding:"required"`
	CategoryName string `json:"categoryName"`
	Icon         string `json:"icon"`
	SortOrder    int    `json:"sortOrder"`
	Status       int    `json:"status"`
}

type AuditGiftExchangeRequest struct {
	ID          int64  `json:"id" binding:"required"`
	Status      int32  `json:"status" binding:"required"`
	AuditRemark string `json:"auditRemark"`
}

type ShipGiftExchangeRequest struct {
	ID             int64  `json:"id" binding:"required"`
	ExpressCompany string `json:"expressCompany" binding:"required"`
	ExpressNo      string `json:"expressNo" binding:"required"`
}

type GiftStatisticsResponse struct {
	TotalCount      int `json:"totalCount"`
	OnSaleCount     int `json:"onSaleCount"`
	OffSaleCount    int `json:"offSaleCount"`
	LowStockCount   int `json:"lowStockCount"`
	TotalStock      int `json:"totalStock"`
	TotalSoldCount  int `json:"totalSoldCount"`
	ExchangeCount   int `json:"exchangeCount"`
	PendingAuditCount int `json:"pendingAuditCount"`
	PendingShipCount int `json:"pendingShipCount"`
	ShippedCount    int `json:"shippedCount"`
	CompletedCount  int `json:"completedCount"`
	TotalPointsUsed int `json:"totalPointsUsed"`
}
