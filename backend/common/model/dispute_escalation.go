package model

import "time"

const (
	EscalationTypePendingTimeout   = 1
	EscalationTypeMediatingTimeout = 2

	EscalationStatusPending   = 10
	EscalationStatusProcessing = 20
	EscalationStatusDone      = 30
	EscalationStatusClosed    = 40
)

type DisputeEscalation struct {
	BaseModel
	CaseID         int64      `gorm:"index;not null" json:"caseId"`
	CaseNo         string     `gorm:"size:32;index" json:"caseNo"`
	EscalateType   int        `gorm:"index;not null;default:1" json:"escalateType"`
	FromLevel      int        `gorm:"default:0" json:"fromLevel"`
	ToLevel        int        `gorm:"index;not null" json:"toLevel"`
	FromUserID     int64      `gorm:"default:0" json:"fromUserId"`
	FromUserName   string     `gorm:"size:64" json:"fromUserName"`
	ToUserID       int64      `gorm:"default:0" json:"toUserId"`
	ToUserName     string     `gorm:"size:64" json:"toUserName"`
	ToOrgID        int64      `gorm:"default:0" json:"toOrgId"`
	ToOrgName      string     `gorm:"size:100" json:"toOrgName"`
	Reason         string     `gorm:"size:500;not null" json:"reason"`
	UrgeCount      int        `gorm:"default:0" json:"urgeCount"`
	FirstUrgeTime  *time.Time `json:"firstUrgeTime"`
	TimeoutHours   int        `gorm:"default:0" json:"timeoutHours"`
	OperatorID     int64      `gorm:"default:0" json:"operatorId"`
	OperatorName   string     `gorm:"size:64;default:'系统自动'" json:"operatorName"`
	Status         int32      `gorm:"index;default:10" json:"status"`
	Remark         string     `gorm:"size:500" json:"remark"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (DisputeEscalation) TableName() string {
	return "dispute_escalation"
}
