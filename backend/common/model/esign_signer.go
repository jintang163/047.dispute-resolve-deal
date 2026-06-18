package model

import "time"

type EsignSigner struct {
	BaseModel
	FlowID       int64      `gorm:"index;not null" json:"flowId"`
	UserID       int64      `gorm:"index;not null" json:"userId"`
	UserName     string     `gorm:"size:50;not null" json:"userName"`
	SignOrder    int        `gorm:"default:0" json:"signOrder"`
	Status       int32      `gorm:"default:0;index" json:"status"`
	SignedAt     *time.Time `json:"signedAt"`
	SignatureURL string     `gorm:"size:500" json:"signatureUrl"`
	VerifyCode   string     `gorm:"size:10" json:"verifyCode"`
	SignIP       string     `gorm:"size:50" json:"signIp"`
	Remark       string     `gorm:"size:500" json:"remark"`
}

func (EsignSigner) TableName() string {
	return "esign_signer"
}
