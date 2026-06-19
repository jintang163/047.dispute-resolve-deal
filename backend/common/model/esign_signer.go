package model

import "time"

type EsignSigner struct {
	BaseModel
	FlowID        int64      `gorm:"index;not null" json:"flowId"`
	UserID        int64      `gorm:"index;not null" json:"userId"`
	UserName      string     `gorm:"size:50;not null" json:"userName"`
	UserPhone     string     `gorm:"size:20" json:"userPhone"`
	IDCard        string     `gorm:"size:20" json:"idCard"`
	SignOrder     int        `gorm:"default:0" json:"signOrder"`
	SignStatus    int32      `gorm:"column:sign_status;default:0;index" json:"signStatus"`
	SignTime      *time.Time `gorm:"column:sign_time" json:"signTime"`
	SignatureURL  string     `gorm:"size:500" json:"signatureUrl"`
	VerifyCode    string     `gorm:"size:10" json:"verifyCode"`
	SignIP        string     `gorm:"size:50" json:"signIp"`
	Remark        string     `gorm:"size:500" json:"remark"`
	FaDaDaSignURL string     `gorm:"size:500" json:"fadadaSignUrl"`
	NotifyStatus  int32      `gorm:"default:0" json:"notifyStatus"`
	NotifySentAt  *time.Time `json:"notifySentAt"`
}

func (EsignSigner) TableName() string {
	return "esign_signer"
}
