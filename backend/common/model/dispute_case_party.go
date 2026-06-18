package model

type DisputeCaseParty struct {
	BaseModel
	CaseID             int64  `gorm:"index;not null" json:"caseId"`
	PartyType          int    `gorm:"index;not null" json:"partyType"`
	Name               string `gorm:"size:50;not null" json:"name"`
	Phone              string `gorm:"size:20;index" json:"phone"`
	IDCard             string `gorm:"size:18" json:"idcard"`
	Gender             int    `json:"gender"`
	Age                int    `json:"age"`
	Address            string `gorm:"size:255" json:"address"`
	Email              string `gorm:"size:100" json:"email"`
	Relationship       string `gorm:"size:50" json:"relationship"`
	IsLegalPerson      int32  `gorm:"default:0" json:"isLegalPerson"`
	LegalRepresentative string `gorm:"size:50" json:"legalRepresentative"`
	Remark             string `gorm:"size:500" json:"remark"`
}

func (DisputeCaseParty) TableName() string {
	return "dispute_case_party"
}
