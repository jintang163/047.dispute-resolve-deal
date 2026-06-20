package model

import "time"

type DisputeUrge struct {
	BaseModel
	CaseID           int64     `gorm:"index;not null" json:"caseId"`
	UrgeType         int       `gorm:"index;not null" json:"urgeType"`
	UrgeLevel        int       `gorm:"default:0" json:"urgeLevel"`
	EscalateTriggered int32    `gorm:"default:0" json:"escalateTriggered"`
	UrgeContent      string    `gorm:"size:1000;not null" json:"urgeContent"`
	OperatorID       int64     `json:"operatorId"`
	OperatorName     string    `gorm:"size:50" json:"operatorName"`
	CreatedAt        time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (DisputeUrge) TableName() string {
	return "dispute_urge"
}
