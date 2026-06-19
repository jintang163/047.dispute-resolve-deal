package model

import "time"

type BlockchainCertificate struct {
	BaseModel
	CertNo         string     `gorm:"size:50;uniqueIndex;not null" json:"certNo"`
	EvidenceID     string     `gorm:"size:50;index;not null" json:"evidenceId"`
	EvidenceType   string     `gorm:"size:30;not null" json:"evidenceType"`
	EvidenceName   string     `gorm:"size:200;not null" json:"evidenceName"`
	EvidenceHash   string     `gorm:"size:64;not null" json:"evidenceHash"`
	CaseID         int64      `gorm:"index;not null" json:"caseId"`
	FlowID         string     `gorm:"size:50;index" json:"flowId"`
	TxID           string     `gorm:"size:100;not null" json:"txId"`
	BlockHeight    int64      `gorm:"not null" json:"blockHeight"`
	OnChainTime    *time.Time `json:"onChainTime"`
	CertURL        string     `gorm:"size:500" json:"certUrl"`
	QRCodeURL      string     `gorm:"size:500" json:"qrcodeUrl"`
	VerifyURL      string     `gorm:"size:500" json:"verifyUrl"`
	Status         int32      `gorm:"default:0;index" json:"status"`
	Metadata       string     `gorm:"type:text" json:"metadata"`
	CreatedBy      int64      `json:"createdBy"`
}

func (BlockchainCertificate) TableName() string {
	return "blockchain_certificate"
}
