package model

import "time"

type EsignFlow struct {
	BaseModel
	FlowNo            string     `gorm:"size:50;uniqueIndex;not null" json:"flowNo"`
	CaseID            int64      `gorm:"index;not null" json:"caseId"`
	TemplateID        int64      `json:"templateId"`
	DocumentName      string     `gorm:"size:200;not null" json:"documentName"`
	DocumentURL       string     `gorm:"size:500;not null" json:"documentUrl"`
	Status            int32      `gorm:"default:0;index" json:"status"`
	CurrentSignIndex  int        `gorm:"default:0" json:"currentSignIndex"`
	TotalSignCount    int        `gorm:"default:0" json:"totalSignCount"`
	SignedCount       int        `gorm:"default:0" json:"signedCount"`
	ExpireTime        *time.Time `json:"expireTime"`
	CreatedBy         int64      `json:"createdBy"`
}

func (EsignFlow) TableName() string {
	return "esign_flow"
}
