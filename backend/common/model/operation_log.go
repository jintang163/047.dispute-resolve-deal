package model

import "time"

type OperationLog struct {
	BaseModel
	UserID      int64     `gorm:"index" json:"userId"`
	UserName    string    `gorm:"size:50" json:"userName"`
	OrgID       int64     `gorm:"index" json:"orgId"`
	Module      string    `gorm:"size:50;index" json:"module"`
	Action      string    `gorm:"size:50;index" json:"action"`
	Description string    `gorm:"size:500" json:"description"`
	Method      string    `gorm:"size:20" json:"method"`
	Params      string    `gorm:"type:text" json:"params"`
	Result      string    `gorm:"type:text" json:"result"`
	IPAddress   string    `gorm:"size:50" json:"ipAddress"`
	UserAgent   string    `gorm:"size:500" json:"userAgent"`
	DurationMs  int64     `json:"durationMs"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (OperationLog) TableName() string {
	return "operation_log"
}
