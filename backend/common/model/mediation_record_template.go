package model

import "time"

type MediationRecordTemplate struct {
	BaseModel
	TemplateName            string `gorm:"size:128;not null" json:"templateName"`
	TemplateCode            string `gorm:"size:64;uniqueIndex;not null" json:"templateCode"`
	Category                string `gorm:"size:64;index" json:"category"`
	DisputeTypeIDs          string `gorm:"size:512" json:"disputeTypeIds"`
	RecordType              int32  `gorm:"default:1" json:"recordType"`
	MediationPlace          string `gorm:"size:256" json:"mediationPlace"`
	ProcessContentTemplate  string `gorm:"type:text" json:"processContentTemplate"`
	DisputeFocusTemplate    string `gorm:"type:text" json:"disputeFocusTemplate"`
	MediationOpinionTemplate string `gorm:"type:text" json:"mediationOpinionTemplate"`
	AgreementContentTemplate string `gorm:"type:text" json:"agreementContentTemplate"`
	NextStepTemplate        string `gorm:"size:512" json:"nextStepTemplate"`
	DefaultDuration         int    `gorm:"default:30" json:"defaultDuration"`
	ParticipantsTemplate    string `gorm:"size:512" json:"participantsTemplate"`
	Tips                    string `gorm:"type:text" json:"tips"`
	IsSystem                int32  `gorm:"index;default:0" json:"isSystem"`
	UseCount                int    `json:"useCount"`
	SortOrder               int    `json:"sortOrder"`
	Status                  int32  `gorm:"index;default:1" json:"status"`
	CreatorID               int64  `json:"creatorId"`
	CreatorName             string `gorm:"size:64" json:"creatorName"`
	OrgID                   int64  `gorm:"index" json:"orgId"`
	OrgName                 string `gorm:"size:128" json:"orgName"`
}

func (MediationRecordTemplate) TableName() string {
	return "mediation_record_template"
}

type MediationRecordTemplateUseLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID int64     `gorm:"index;not null" json:"templateId"`
	CaseID     int64     `gorm:"index" json:"caseId"`
	RecordID   int64     `gorm:"index" json:"recordId"`
	UserID     int64     `gorm:"index" json:"userId"`
	UserName   string    `gorm:"size:64" json:"userName"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (MediationRecordTemplateUseLog) TableName() string {
	return "mediation_record_template_use_log"
}
