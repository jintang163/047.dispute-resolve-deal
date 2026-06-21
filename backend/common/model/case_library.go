package model

import "time"

type CaseLibrary struct {
	BaseModel
	CaseNo          string     `gorm:"size:64;uniqueIndex;not null" json:"caseNo"`
	Title           string     `gorm:"size:256;not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	DisputeType     string     `gorm:"size:128" json:"disputeType"`
	TypeID          int64      `gorm:"index" json:"typeId"`
	MediationTactics string    `gorm:"type:text" json:"mediationTactics"`
	KeyPoints       string     `gorm:"type:text" json:"keyPoints"`
	ResultSummary   string     `gorm:"type:text" json:"resultSummary"`
	DifficultyLevel int        `gorm:"default:1" json:"difficultyLevel"`
	IsSuccess       int32      `gorm:"default:1" json:"isSuccess"`
	MediatorName    string     `gorm:"size:64" json:"mediatorName"`
	MediatorID      int64      `gorm:"index" json:"mediatorId"`
	OrgName         string     `gorm:"size:128" json:"orgName"`
	OrgID           int64      `gorm:"index" json:"orgId"`
	SourceCaseID    int64      `json:"sourceCaseId"`
	Keywords        string     `gorm:"size:500" json:"keywords"`
	Tags            string     `gorm:"size:500" json:"tags"`
	VectorID        string     `gorm:"size:100;index" json:"vectorId"`
	VectorStatus    int32      `gorm:"default:0" json:"vectorStatus"`
	ReferenceCount  int        `gorm:"default:0" json:"referenceCount"`
	AvgScore        float64    `gorm:"type:decimal(3,2);default:0.00" json:"avgScore"`
	ScoreCount      int        `gorm:"default:0" json:"scoreCount"`
	TotalScore      float64    `gorm:"type:decimal(10,2);default:0.00" json:"totalScore"`
	LastUsedAt      *time.Time `json:"lastUsedAt"`
	Status          int32      `gorm:"default:1;index" json:"status"`
	ArchivedAt      *time.Time `json:"archivedAt"`
	CreatedBy       int64      `json:"createdBy"`
}

func (CaseLibrary) TableName() string {
	return "case_library"
}

type CaseLibraryScore struct {
	BaseModel
	CaseID        int64  `gorm:"index;not null" json:"caseId"`
	CaseNo        string `gorm:"size:64" json:"caseNo"`
	UserID        int64  `gorm:"index;not null" json:"userId"`
	UserName      string `gorm:"size:64" json:"userName"`
	Score         int    `gorm:"not null" json:"score"`
	SourceCaseID  int64  `json:"sourceCaseId"`
	Comment       string `gorm:"size:500" json:"comment"`
}

func (CaseLibraryScore) TableName() string {
	return "case_library_score"
}

type CaseLibraryQuote struct {
	BaseModel
	SourceCaseID       int64  `gorm:"index;not null" json:"sourceCaseId"`
	LibraryCaseID      int64  `gorm:"index;not null" json:"libraryCaseId"`
	LibraryCaseNo      string `gorm:"size:64" json:"libraryCaseNo"`
	QuoteType          int32  `gorm:"default:1" json:"quoteType"`
	QuoteContent       string `gorm:"type:text" json:"quoteContent"`
	UserID             int64  `gorm:"index;not null" json:"userId"`
	UserName           string `gorm:"size:64" json:"userName"`
	MediationRecordID  int64  `json:"mediationRecordId"`
}

func (CaseLibraryQuote) TableName() string {
	return "case_library_quote"
}

type CaseLibraryArchive struct {
	BaseModel
	OriginalID    int64      `gorm:"index;not null" json:"originalId"`
	CaseNo        string     `gorm:"size:64;index" json:"caseNo"`
	Title         string     `gorm:"size:256" json:"title"`
	ArchiveReason int32      `gorm:"default:1" json:"archiveReason"`
	AvgScore      float64    `gorm:"type:decimal(3,2);default:0.00" json:"avgScore"`
	ReferenceCount int       `gorm:"default:0" json:"referenceCount"`
	LastUsedAt    *time.Time `json:"lastUsedAt"`
	ArchivedBy    int64      `json:"archivedBy"`
	CaseData      string     `gorm:"type:json" json:"caseData"`
}

func (CaseLibraryArchive) TableName() string {
	return "case_library_archive"
}
