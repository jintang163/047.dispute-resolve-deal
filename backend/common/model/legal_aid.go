package model

import "time"

type LegalAidOrg struct {
	BaseModel
	OrgCode       string  `gorm:"size:64;uniqueIndex;not null" json:"orgCode"`
	OrgName       string  `gorm:"size:128;not null" json:"orgName"`
	OrgType       int     `gorm:"index;not null;default:1" json:"orgType"`
	Level         int     `gorm:"index;default:1" json:"level"`
	Address       string  `gorm:"size:256" json:"address"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	ContactPerson string  `gorm:"size:64" json:"contactPerson"`
	ContactPhone  string  `gorm:"size:20" json:"contactPhone"`
	ContactEmail  string  `gorm:"size:128" json:"contactEmail"`
	ServiceScope  string  `gorm:"type:text" json:"serviceScope"`
	WorkHours     string  `gorm:"size:256" json:"workHours"`
	Description   string  `gorm:"type:text" json:"description"`
	LawyerCount   int     `json:"lawyerCount"`
	CaseCapacity  int     `json:"caseCapacity"`
	AcceptCount   int     `json:"acceptCount"`
	SortOrder     int     `json:"sortOrder"`
	Status        int32   `gorm:"index;default:1" json:"status"`
}

func (LegalAidOrg) TableName() string {
	return "legal_aid_org"
}

type LegalAidLawyer struct {
	BaseModel
	OrgID            int64      `gorm:"index;not null" json:"orgId"`
	OrgName          string     `gorm:"size:128" json:"orgName"`
	LawyerName       string     `gorm:"size:64;not null" json:"lawyerName"`
	LicenseNo        string     `gorm:"size:64" json:"licenseNo"`
	Phone            string     `gorm:"size:20" json:"phone"`
	Email            string     `gorm:"size:128" json:"email"`
	Avatar           string     `gorm:"size:256" json:"avatar"`
	Gender           int        `gorm:"default:0" json:"gender"`
	Specialty        string     `gorm:"size:512" json:"specialty"`
	YearsOfExperience int       `json:"yearsOfExperience"`
	Title            string     `gorm:"size:64" json:"title"`
	Intro            string     `gorm:"type:text" json:"intro"`
	ConsultCount     int        `json:"consultCount"`
	ConsultRating    float64    `gorm:"type:decimal(3,2);default:0.00" json:"consultRating"`
	IsOnline         int32      `gorm:"index;default:0" json:"isOnline"`
	LastOnlineAt     *time.Time `json:"lastOnlineAt"`
	Status           int32      `gorm:"index;default:1" json:"status"`
}

func (LegalAidLawyer) TableName() string {
	return "legal_aid_lawyer"
}

type LegalAidApplication struct {
	BaseModel
	ApplyNo          string     `gorm:"size:64;uniqueIndex;not null" json:"applyNo"`
	CaseID           int64      `gorm:"index;not null" json:"caseId"`
	CaseNo           string     `gorm:"size:32;index" json:"caseNo"`
	ApplicantName    string     `gorm:"size:64;not null" json:"applicantName"`
	ApplicantPhone   string     `gorm:"size:20;index" json:"applicantPhone"`
	ApplicantIDCard  string     `gorm:"size:32" json:"applicantIdCard"`
	ApplicantAddress string     `gorm:"size:256" json:"applicantAddress"`
	IncomeLevel      int        `gorm:"default:3" json:"incomeLevel"`
	FamilySize       int        `json:"familySize"`
	MonthlyIncome    float64    `gorm:"type:decimal(10,2);default:0.00" json:"monthlyIncome"`
	AidReason        string     `gorm:"type:text" json:"aidReason"`
	DisputeType      string     `gorm:"size:128" json:"disputeType"`
	EvidenceSummary  string     `gorm:"type:text" json:"evidenceSummary"`
	MaterialURLs     string     `gorm:"type:text" json:"materialUrls"`
	Status           int32      `gorm:"index;default:10" json:"status"`
	AuditorID        int64      `json:"auditorId"`
	AuditorName      string     `gorm:"size:64" json:"auditorName"`
	AuditTime        *time.Time `json:"auditTime"`
	AuditOpinion     string     `gorm:"size:500" json:"auditOpinion"`
	RejectReason     string     `gorm:"size:500" json:"rejectReason"`
	SubmitterID      int64      `json:"submitterId"`
	SubmitterName    string     `gorm:"size:64" json:"submitterName"`
	SubmitTime       time.Time  `gorm:"autoCreateTime;index" json:"submitTime"`
	TransferID       int64      `gorm:"index" json:"transferId"`
	TransferNo       string     `gorm:"size:64" json:"transferNo"`
	Transferred      int32      `gorm:"default:0" json:"transferred"`
	TransferredAt    *time.Time `json:"transferredAt"`
}

func (LegalAidApplication) TableName() string {
	return "legal_aid_application"
}

type LegalAidTransfer struct {
	BaseModel
	TransferNo    string     `gorm:"size:64;uniqueIndex;not null" json:"transferNo"`
	CaseID        int64      `gorm:"index;not null" json:"caseId"`
	CaseNo        string     `gorm:"size:32;index" json:"caseNo"`
	CaseTitle     string     `gorm:"size:256" json:"caseTitle"`
	DisputeType   string     `gorm:"size:128" json:"disputeType"`
	FromOrgID     int64      `gorm:"index;default:0" json:"fromOrgId"`
	FromOrgName   string     `gorm:"size:128" json:"fromOrgName"`
	FromUserID    int64      `json:"fromUserId"`
	FromUserName  string     `gorm:"size:64" json:"fromUserName"`
	ToOrgID       int64      `gorm:"index;not null" json:"toOrgId"`
	ToOrgName     string     `gorm:"size:128" json:"toOrgName"`
	ToLawyerID    int64      `json:"toLawyerId"`
	ToLawyerName  string     `gorm:"size:64" json:"toLawyerName"`
	TransferReason string    `gorm:"type:text" json:"transferReason"`
	CaseSummary   string     `gorm:"type:text" json:"caseSummary"`
	AttachIDs     string     `gorm:"size:512" json:"attachIds"`
	AcceptStatus  int32      `gorm:"index;default:10" json:"acceptStatus"`
	AcceptTime    *time.Time `json:"acceptTime"`
	LegalCaseNo   string     `gorm:"size:64;index" json:"legalCaseNo"`
	RejectReason  string     `gorm:"size:500" json:"rejectReason"`
	CloseResult   string     `gorm:"type:text" json:"closeResult"`
	CloseTime     *time.Time `json:"closeTime"`
	TransferTime  time.Time  `gorm:"autoCreateTime;index" json:"transferTime"`
}

func (LegalAidTransfer) TableName() string {
	return "legal_aid_transfer"
}

type LegalAidConsult struct {
	BaseModel
	ConsultNo     string     `gorm:"size:64;uniqueIndex;not null" json:"consultNo"`
	TransferID    int64      `gorm:"index;default:0" json:"transferId"`
	CaseID        int64      `gorm:"index;default:0" json:"caseId"`
	CaseNo        string     `gorm:"size:32" json:"caseNo"`
	ApplicantID   int64      `gorm:"index;default:0" json:"applicantId"`
	ApplicantName string     `gorm:"size:64" json:"applicantName"`
	ApplicantPhone string    `gorm:"size:20" json:"applicantPhone"`
	LawyerID      int64      `gorm:"index;not null" json:"lawyerId"`
	LawyerName    string     `gorm:"size:64" json:"lawyerName"`
	OrgID         int64      `json:"orgId"`
	OrgName       string     `gorm:"size:128" json:"orgName"`
	ConsultType   int        `gorm:"default:1" json:"consultType"`
	ConsultStatus int32      `gorm:"index;default:10" json:"consultStatus"`
	QuestionTitle string     `gorm:"size:256" json:"questionTitle"`
	QuestionContent string   `gorm:"type:text" json:"questionContent"`
	AnswerContent string     `gorm:"type:text" json:"answerContent"`
	TotalDuration int        `json:"totalDuration"`
	FreeDuration  int        `gorm:"default:1800" json:"freeDuration"`
	UsedDuration  int        `json:"usedDuration"`
	IsFree        int32      `gorm:"default:1" json:"isFree"`
	Rating        int        `json:"rating"`
	Comment       string     `gorm:"size:500" json:"comment"`
	StartTime     *time.Time `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
}

func (LegalAidConsult) TableName() string {
	return "legal_aid_consult"
}

type LegalAidConsultMessage struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConsultID   int64      `gorm:"index;not null" json:"consultId"`
	SenderID    int64      `gorm:"index;not null" json:"senderId"`
	SenderName  string     `gorm:"size:64" json:"senderName"`
	SenderType  int        `gorm:"not null" json:"senderType"`
	MessageType int        `gorm:"default:1" json:"messageType"`
	Content     string     `gorm:"type:text" json:"content"`
	FileURL     string     `gorm:"size:512" json:"fileUrl"`
	FileName    string     `gorm:"size:256" json:"fileName"`
	FileSize    int64      `json:"fileSize"`
	Duration    int        `json:"duration"`
	IsRead      int32      `gorm:"index;default:0" json:"isRead"`
	ReadTime    *time.Time `json:"readTime"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (LegalAidConsultMessage) TableName() string {
	return "legal_aid_consult_message"
}

type LegalAidMaterial struct {
	BaseModel
	CaseID         int64  `gorm:"index" json:"caseId"`
	ApplicationID  string `gorm:"size:64;index" json:"applicationId"`
	MaterialType   int    `gorm:"default:1" json:"materialType"`
	MaterialName   string `gorm:"size:64" json:"materialName"`
	FileName       string `gorm:"size:256" json:"fileName"`
	FilePath       string `gorm:"size:512" json:"filePath"`
	FileURL        string `gorm:"size:512" json:"fileUrl"`
	FileSize       int64  `json:"fileSize"`
	FileExt        string `gorm:"size:16" json:"fileExt"`
	MimeType       string `gorm:"size:128" json:"mimeType"`
	UploaderID     int64  `json:"uploaderId"`
	UploaderName   string `gorm:"size:64" json:"uploaderName"`
	OrganizationID int64  `json:"organizationId"`
}

func (LegalAidMaterial) TableName() string {
	return "legal_aid_material"
}
