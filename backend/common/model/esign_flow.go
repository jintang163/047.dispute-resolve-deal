package model

import "time"

type EsignFlow struct {
	BaseModel
	FlowNo            string     `gorm:"size:50;uniqueIndex;not null" json:"flowNo"`
	CaseID            int64      `gorm:"index;not null" json:"caseId"`
	CaseNo            string     `gorm:"size:50" json:"caseNo"`
	DocType           int32      `gorm:"default:0" json:"docType"`
	DocTitle          string     `gorm:"size:200;not null" json:"docTitle"`
	DocContent        string     `gorm:"type:text" json:"docContent"`
	DocURL            string     `gorm:"size:500" json:"docUrl"`
	TemplateID        int64      `json:"templateId"`
	SignedDocumentURL string     `gorm:"size:500" json:"signedDocumentUrl"`
	Status            int32      `gorm:"default:0;index" json:"status"`
	CurrentSignIndex  int        `gorm:"default:0" json:"currentSignIndex"`
	TotalSignCount    int        `gorm:"default:0" json:"totalSignCount"`
	SignedCount       int        `gorm:"default:0" json:"signedCount"`
	ExpireTime        *time.Time `json:"expireTime"`
	FaDaDaFlowID      string     `gorm:"size:100" json:"fadadaFlowId"`
	CrossPageSeal     int32      `gorm:"default:0" json:"crossPageSeal"`
	BCCertNo          string     `gorm:"size:50;index" json:"bcCertNo"`
	BCTxID            string     `gorm:"size:100" json:"bcTxId"`
	BCOnChainTime     *time.Time `json:"bcOnChainTime"`
	BCStatus          int32      `gorm:"default:0" json:"bcStatus"`
	CreatorID         int64      `gorm:"index" json:"creatorId"`
	CreatorName       string     `gorm:"size:50" json:"creatorName"`
	OrganizationID    int64      `gorm:"index" json:"organizationId"`
	RevokeReason      string     `gorm:"size:500" json:"revokeReason"`
	RevokeTime        *time.Time `json:"revokeTime"`
	RevokeBy          int64      `json:"revokeBy"`
	LastSignTime      *time.Time `json:"lastSignTime"`
}

func (EsignFlow) TableName() string {
	return "esign_flow"
}
