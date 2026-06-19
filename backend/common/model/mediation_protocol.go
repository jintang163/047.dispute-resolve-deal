package model

import "time"

type MediationProtocol struct {
	BaseModel
	CaseID          int64      `gorm:"index;not null" json:"caseId"`
	ProtocolNo      string     `gorm:"size:50;uniqueIndex;not null" json:"protocolNo"`
	Title           string     `gorm:"size:200;not null" json:"title"`
	Content         string     `gorm:"type:text" json:"content"`
	PartyAName      string     `gorm:"size:50" json:"partyAName"`
	PartyBName      string     `gorm:"size:50" json:"partyBName"`
	MediatorName    string     `gorm:"size:50" json:"mediatorName"`
	AgreementItems  string     `gorm:"type:text" json:"agreementItems"`
	BreachClause    string     `gorm:"size:1000" json:"breachClause"`
	EffectiveDate   *time.Time `json:"effectiveDate"`
	IsSigned        int32      `gorm:"default:0;index" json:"isSigned"`
	SignedAt        *time.Time `json:"signedAt"`
	FileURL         string     `gorm:"size:500" json:"fileUrl"`
	CreatedBy       int64      `json:"createdBy"`
	IsAIGenerated   int32      `gorm:"default:0" json:"isAIGenerated"`
	AIGeneratedAt   *time.Time `json:"aiGeneratedAt"`
	IsAdopted       int32      `gorm:"default:0" json:"isAdopted"`
	AdoptedBy       int64      `json:"adoptedBy"`
	AdoptedAt       *time.Time `json:"adoptedAt"`
}

func (MediationProtocol) TableName() string {
	return "mediation_protocol"
}
