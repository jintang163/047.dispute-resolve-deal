package model

import "time"

const (
	JudicialStatusSubmitted   = 10
	JudicialStatusReviewing   = 20
	JudicialStatusConfirmed   = 30
	JudicialStatusRejected    = 40
	JudicialStatusExpired     = 50

	SealStatusPending = 0
	SealStatusDone    = 1
)

type JudicialConfirmation struct {
	BaseModel
	ConfirmNo           string     `gorm:"size:32;uniqueIndex;not null" json:"confirmNo"`
	CaseID              int64      `gorm:"index" json:"caseId"`
	CaseNo              string     `gorm:"size:32;index" json:"caseNo"`
	CaseTitle           string     `gorm:"size:256" json:"caseTitle"`
	MediationRecordID   int64      `json:"mediationRecordId"`
	ProtocolID          int64      `json:"protocolId"`

	ApplicantName      string `gorm:"size:64;not null" json:"applicantName"`
	ApplicantPhone     string `gorm:"size:20;not null" json:"applicantPhone"`
	ApplicantIDCard    string `gorm:"size:32" json:"applicantIdCard"`
	ApplicantAddress   string `gorm:"size:256" json:"applicantAddress"`

	RespondentName     string `gorm:"size:64;not null" json:"respondentName"`
	RespondentPhone    string `gorm:"size:20;not null" json:"respondentPhone"`
	RespondentIDCard   string `gorm:"size:32" json:"respondentIdCard"`
	RespondentAddress  string `gorm:"size:256" json:"respondentAddress"`

	CourtID            int64      `gorm:"index" json:"courtId"`
	CourtName          string     `gorm:"size:128" json:"courtName"`
	CourtCode          string     `gorm:"size:64" json:"courtCode"`

	AgreementContent   string     `gorm:"type:text" json:"agreementContent"`
	PerformanceDeadline *time.Time `gorm:"index" json:"performanceDeadline"`
	ConfirmAmount      float64    `gorm:"type:decimal(15,2)" json:"confirmAmount"`

	Status             int32      `gorm:"default:10;index" json:"status"`
	SubStatus          int32      `json:"subStatus"`
	ReviewOpinion      string     `gorm:"size:512" json:"reviewOpinion"`
	ConfirmDate        *time.Time `json:"confirmDate"`
	ConfirmCourt       string     `gorm:"size:128" json:"confirmCourt"`
	ConfirmJudge       string     `gorm:"size:64" json:"confirmJudge"`
	ConfirmDocumentNo  string     `gorm:"size:64" json:"confirmDocumentNo"`

	DocumentURL        string     `gorm:"size:512" json:"documentUrl"`
	SealStatus         int32      `gorm:"default:0" json:"sealStatus"`
	SealTime           *time.Time `json:"sealTime"`

	SubmitTime         *time.Time `json:"submitTime"`
	SubmitBy           int64      `json:"submitBy"`
	SubmitByName       string     `gorm:"size:64" json:"submitByName"`

	ReviewTime         *time.Time `json:"reviewTime"`
	ReviewBy           int64      `json:"reviewBy"`
	ReviewByName       string     `gorm:"size:64" json:"reviewByName"`

	ExpirationRemindSent  int32 `gorm:"default:0" json:"expirationRemindSent"`
	PerformanceRemindSent int32 `gorm:"default:0" json:"performanceRemindSent"`

	OrganizationID   int64  `gorm:"column:org_id;index;not null" json:"organizationId"`
	OrganizationName string `gorm:"column:org_name;size:100" json:"organizationName"`
	Remark           string `gorm:"size:512" json:"remark"`

	OrgID   int64  `gorm:"-" json:"orgId"`
	OrgName string `gorm:"-" json:"orgName"`
}

func (j *JudicialConfirmation) AfterFind(tx interface{}) error {
	j.OrgID = j.OrganizationID
	j.OrgName = j.OrganizationName
	return nil
}

func (JudicialConfirmation) TableName() string {
	return "judicial_confirmation"
}
