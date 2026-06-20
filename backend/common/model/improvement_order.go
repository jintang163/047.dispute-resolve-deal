package model

import "time"

const (
	ImprovementStatusPending    = 10
	ImprovementStatusProcessing = 20
	ImprovementStatusRectified  = 30
	ImprovementStatusReviewed   = 40
	ImprovementStatusClosed     = 99

	ImprovementPriorityHigh   = 1
	ImprovementPriorityMedium = 2
	ImprovementPriorityLow    = 3

	IssueTypeAttitude     = "attitude"
	IssueTypeEfficiency   = "efficiency"
	IssueTypeProfessional = "professional"
	IssueTypeResult       = "result"
	IssueTypeProcess      = "process"
	IssueTypeOther        = "other"
)

type ImprovementOrder struct {
	BaseModel
	OrderNo             string     `gorm:"size:32;uniqueIndex;not null" json:"orderNo"`
	CaseID              int64      `gorm:"index;not null" json:"caseId"`
	CaseNo              string     `gorm:"size:32;not null" json:"caseNo"`
	CaseTitle           string     `gorm:"size:256" json:"caseTitle"`
	ApplicantID         int64      `json:"applicantId"`
	ApplicantName       string     `gorm:"size:64" json:"applicantName"`
	MediatorID          int64      `gorm:"index;not null" json:"mediatorId"`
	MediatorName        string     `gorm:"size:64" json:"mediatorName"`
	OrgID               int64      `json:"orgId"`
	OrgName             string     `gorm:"size:100" json:"orgName"`
	SatisfactionScore   int        `json:"satisfactionScore"`
	SatisfactionComment string     `gorm:"type:text" json:"satisfactionComment"`
	SentimentEmotion    string     `gorm:"size:20;index" json:"sentimentEmotion"`
	SentimentScore      float64    `json:"sentimentScore"`
	SentimentSummary    string     `gorm:"size:500" json:"sentimentSummary"`
	IssueType           string     `gorm:"size:50" json:"issueType"`
	IssueDescription    string     `gorm:"type:text" json:"issueDescription"`
	ImprovementSuggestion string   `gorm:"type:text" json:"improvementSuggestion"`
	Status              int        `gorm:"default:10;index" json:"status"`
	Priority            int        `gorm:"default:2" json:"priority"`
	Deadline            *time.Time `json:"deadline"`
	AssignedAt          *time.Time `json:"assignedAt"`
	RectifyContent      string     `gorm:"type:text" json:"rectifyContent"`
	RectifyResult       string     `gorm:"type:text" json:"rectifyResult"`
	RectifiedAt         *time.Time `json:"rectifiedAt"`
	ReviewOpinion       string     `gorm:"size:500" json:"reviewOpinion"`
	ReviewedBy          int64      `json:"reviewedBy"`
	ReviewedByName      string     `gorm:"size:64" json:"reviewedByName"`
	ReviewedAt          *time.Time `json:"reviewedAt"`
	DeductionScore      float64    `json:"deductionScore"`
	DeductionReason     string     `gorm:"size:500" json:"deductionReason"`
	IsDeductionApplied  int        `gorm:"default:0" json:"isDeductionApplied"`
	DeductionAppliedAt  *time.Time `json:"deductionAppliedAt"`
	Remark              string     `gorm:"size:500" json:"remark"`
}

func (ImprovementOrder) TableName() string {
	return "improvement_order"
}
