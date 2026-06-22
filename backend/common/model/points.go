package model

import "time"

type GridMemberPointsAccount struct {
	BaseModel
	MemberID         int64  `gorm:"uniqueIndex" json:"memberId"`
	MemberName       string `gorm:"size:64" json:"memberName"`
	MemberNo         string `gorm:"size:32" json:"memberNo"`
	OrganizationID   int64  `gorm:"column:organization_id;index" json:"organizationId"`
	Balance          int    `json:"balance"`
	TotalEarned      int    `json:"totalEarned"`
	TotalSpent       int    `json:"totalSpent"`
	TotalExpired     int    `json:"totalExpired"`
	Level            int    `gorm:"default:1;index" json:"level"`
	LevelName        string `gorm:"size:32;default:'初级网格员'" json:"levelName"`
	NextLevelPoints  int    `gorm:"default:1000" json:"nextLevelPoints"`
	CheckinDays      int    `json:"checkinDays"`
	TotalCheckinDays int    `json:"totalCheckinDays"`
	LastCheckinDate  *time.Time `json:"lastCheckinDate"`
}

func (GridMemberPointsAccount) TableName() string {
	return "grid_member_points_account"
}

type PointsRecord struct {
	BaseModel
	RecordNo      string     `gorm:"size:32;uniqueIndex" json:"recordNo"`
	MemberID      int64      `gorm:"index;not null" json:"memberId"`
	MemberName    string     `gorm:"size:64" json:"memberName"`
	OrganizationID int64     `gorm:"column:organization_id;index" json:"organizationId"`
	Type          int        `gorm:"index;not null" json:"type"`
	TypeName      string     `gorm:"size:32" json:"typeName"`
	BusinessType  string     `gorm:"size:64;index;not null" json:"businessType"`
	BusinessID    int64      `gorm:"index" json:"businessId"`
	BusinessNo    string     `gorm:"size:64" json:"businessNo"`
	Points        int        `json:"points"`
	BalanceBefore int        `json:"balanceBefore"`
	BalanceAfter  int        `json:"balanceAfter"`
	Description   string     `gorm:"size:256" json:"description"`
	OperatorID    int64      `json:"operatorId"`
	OperatorName  string     `gorm:"size:64" json:"operatorName"`
	ExpireDate    *time.Time `gorm:"index" json:"expireDate"`
	IsExpired     int        `gorm:"default:0" json:"isExpired"`
}

func (PointsRecord) TableName() string {
	return "points_record"
}

type PointsRule struct {
	BaseModel
	RuleCode         string `gorm:"size:64;uniqueIndex" json:"ruleCode"`
	RuleName         string `gorm:"size:128;not null" json:"ruleName"`
	RuleType         string `gorm:"size:64;index;not null" json:"ruleType"`
	Points           int    `gorm:"not null;default:0" json:"points"`
	MaxPointsPerDay  int    `gorm:"default:0" json:"maxPointsPerDay"`
	MaxPointsPerMonth int   `gorm:"default:0" json:"maxPointsPerMonth"`
	IsActive         int    `gorm:"default:1;index" json:"isActive"`
	Description      string `gorm:"size:512" json:"description"`
	ExpireDays       int    `gorm:"default:365" json:"expireDays"`
	SortOrder        int    `gorm:"default:0" json:"sortOrder"`
}

func (PointsRule) TableName() string {
	return "points_rule"
}

type PointsRecordQuery struct {
	BaseQuery
	MemberID     int64  `form:"memberId" json:"memberId"`
	OrgID        int64  `form:"orgId" json:"orgId"`
	Type         int    `form:"type" json:"type"`
	BusinessType string `form:"businessType" json:"businessType"`
	RecordNo     string `form:"recordNo" json:"recordNo"`
	IsExpired    int    `form:"isExpired" json:"isExpired"`
	DateRangeQuery
}

type PointsRuleQuery struct {
	BaseQuery
	RuleType string `form:"ruleType" json:"ruleType"`
	IsActive int    `form:"isActive" json:"isActive"`
	RuleCode string `form:"ruleCode" json:"ruleCode"`
	RuleName string `form:"ruleName" json:"ruleName"`
}

type CreatePointsRuleRequest struct {
	RuleCode         string `json:"ruleCode" binding:"required"`
	RuleName         string `json:"ruleName" binding:"required"`
	RuleType         string `json:"ruleType" binding:"required"`
	Points           int    `json:"points" binding:"required,min=0"`
	MaxPointsPerDay  int    `json:"maxPointsPerDay"`
	MaxPointsPerMonth int   `json:"maxPointsPerMonth"`
	IsActive         int    `json:"isActive"`
	Description      string `json:"description"`
	ExpireDays       int    `json:"expireDays"`
	SortOrder        int    `json:"sortOrder"`
}

type UpdatePointsRuleRequest struct {
	ID               int64  `json:"id" binding:"required"`
	RuleName         string `json:"ruleName"`
	RuleType         string `json:"ruleType"`
	Points           int    `json:"points"`
	MaxPointsPerDay  int    `json:"maxPointsPerDay"`
	MaxPointsPerMonth int   `json:"maxPointsPerMonth"`
	IsActive         int    `json:"isActive"`
	Description      string `json:"description"`
	ExpireDays       int    `json:"expireDays"`
	SortOrder        int    `json:"sortOrder"`
}

type PointsSummaryResponse struct {
	Balance         int `json:"balance"`
	TodayEarned     int `json:"todayEarned"`
	MonthEarned     int `json:"monthEarned"`
	TotalEarned     int `json:"totalEarned"`
	TotalSpent      int `json:"totalSpent"`
	Level           int `json:"level"`
	LevelName       string `json:"levelName"`
	NextLevelPoints int `json:"nextLevelPoints"`
	LevelProgress   float64 `json:"levelProgress"`
	CheckinDays     int `json:"checkinDays"`
}

type ExchangePointsRequest struct {
	GiftID   int64  `json:"giftId" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	ReceiverName    string `json:"receiverName" binding:"required"`
	ReceiverPhone   string `json:"receiverPhone" binding:"required"`
	ReceiverAddress string `json:"receiverAddress" binding:"required"`
	Remark          string `json:"remark"`
}
