package model

import "time"

type DisputeCase struct {
	BaseModel
	CaseNo           string     `gorm:"size:50;uniqueIndex;not null" json:"caseNo"`
	TypeID           int64      `gorm:"index" json:"typeId"`
	TypeName         string     `gorm:"size:100" json:"typeName"`
	Title            string     `gorm:"size:200;not null" json:"title"`
	Description      string     `gorm:"size:2000" json:"description"`
	ExpectedSolution string     `gorm:"size:1000" json:"expectedSolution"`
	Source           int        `gorm:"index" json:"source"`
	Level            int        `gorm:"index" json:"level"`
	UrgencyLevel     int        `gorm:"index" json:"urgencyLevel"`
	OrganizationID   int64      `gorm:"column:org_id;index" json:"organizationId"`
	OrganizationName string     `gorm:"column:org_name;size:100" json:"organizationName"`
	MediatorID       int64      `gorm:"index" json:"mediatorId"`
	MediatorName     string     `gorm:"size:50" json:"mediatorName"`
	ApplicantID      int64      `gorm:"index" json:"applicantId"`
	ApplicantName    string     `gorm:"size:50" json:"applicantName"`
	ApplicantPhone   string     `gorm:"size:20;index" json:"applicantPhone"`
	ApplicantIDCard  string     `gorm:"size:18" json:"applicantIdcard"`
	RespondentName   string     `gorm:"size:50" json:"respondentName"`
	RespondentPhone  string     `gorm:"size:20;index" json:"respondentPhone"`
	RespondentIDCard string     `gorm:"size:18" json:"respondentIdcard"`
	Status           int32      `gorm:"default:10;index" json:"status"`
	WorkflowID       int64      `gorm:"index" json:"workflowId"`
	IsMediated       int32      `gorm:"default:0" json:"isMediated"`
	MediationResult  string     `gorm:"size:500" json:"mediationResult"`
	ClosedAt         *time.Time `json:"closedAt"`
	SatisfactionScore int       `json:"satisfactionScore"`
	SatisfactionRemark string    `gorm:"size:500" json:"satisfactionRemark"`
	SentimentEmotion    string  `gorm:"size:20" json:"sentimentEmotion"`
	SentimentScoreVal   float64 `json:"sentimentScoreVal"`
	SentimentConfidence float64 `json:"sentimentConfidence"`
	SentimentKeywords   string  `gorm:"type:json" json:"sentimentKeywords"`
	SentimentSummary    string  `gorm:"size:500" json:"sentimentSummary"`
	SentimentAnalyzedAt *time.Time `json:"sentimentAnalyzedAt"`
	Longitude        float64    `json:"longitude"`
	Latitude         float64    `json:"latitude"`
	Address          string     `gorm:"size:255" json:"address"`
	CreatedBy        int64      `json:"createdBy"`
	CreatedFrom      int        `json:"createdFrom"`
	Remark           string     `gorm:"size:500" json:"remark"`

	MediatorTime       *time.Time `json:"mediatorTime"`
	MediationStartTime *time.Time `json:"mediationStartTime"`
	MediationEndTime   *time.Time `json:"mediationEndTime"`
	UrgencyTime        *time.Time `json:"urgencyTime"`
	UrgencyCount       int        `json:"urgencyCount"`
	EscalateLevel      int32      `gorm:"default:0" json:"escalateLevel"`
	EscalateTime       *time.Time `json:"escalateTime"`
	LastProgressTime   *time.Time `json:"lastProgressTime"`
	RiskFlag           int32      `gorm:"default:0" json:"riskFlag"`
	RiskReason         string     `gorm:"size:500" json:"riskReason"`
	RiskFlagTime       *time.Time `json:"riskFlagTime"`

	OrgID   int64  `gorm:"-" json:"orgId"`
	OrgName string `gorm:"-" json:"orgName"`
}

func (d *DisputeCase) AfterFind(tx interface{}) error {
	d.OrgID = d.OrganizationID
	d.OrgName = d.OrganizationName
	return nil
}

func (DisputeCase) TableName() string {
	return "dispute_case"
}
