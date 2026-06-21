package model

import "time"

type DisputeEvidence struct {
	BaseModel
	CaseID          int64      `gorm:"index;not null" json:"caseId"`
	CaseNo          string     `gorm:"size:32;index" json:"caseNo"`
	UploaderID      int64      `gorm:"index" json:"uploaderId"`
	UploaderName    string     `gorm:"size:50" json:"uploaderName"`
	OrganizationID  int64      `gorm:"index" json:"organizationId"`
	FileName        string     `gorm:"size:256;not null" json:"fileName"`
	FileType        int32      `gorm:"index" json:"fileType"`
	FileExt         string     `gorm:"size:20" json:"fileExt"`
	FileSize        int64      `json:"fileSize"`
	FileURL         string     `gorm:"size:512;not null" json:"fileUrl"`
	FilePath        string     `gorm:"size:512" json:"filePath"`
	MimeType        string     `gorm:"size:100" json:"mimeType"`
	ThumbnailURL    string     `gorm:"size:512" json:"thumbnailUrl"`
	OSSPath         string     `gorm:"size:512" json:"ossPath"`
	Description     string     `gorm:"size:500" json:"description"`
	IsVerified      int32      `gorm:"default:0;index" json:"isVerified"`
	VerifiedBy      int64      `json:"verifiedBy"`
	VerifiedAt      *time.Time `json:"verifiedAt"`
	SortOrder       int        `json:"sortOrder"`
	UploadFrom      int        `json:"uploadFrom"`
	Remark          string     `gorm:"size:500" json:"remark"`

	EvidenceCategory int       `gorm:"default:0;index" json:"evidenceCategory"`
	AICategory       int       `gorm:"default:0" json:"aiCategory"`
	AIConfidence     float64   `gorm:"type:decimal(5,4);default:0" json:"aiConfidence"`
	AIKeywords       string    `gorm:"size:256" json:"aiKeywords"`
	AIProcessed      int32     `gorm:"default:0;index" json:"aiProcessed"`
	AIProcessedAt    *time.Time `json:"aiProcessedAt"`
	ManualCategory   int       `gorm:"default:0" json:"manualCategory"`
	ManualUpdatedAt  *time.Time `json:"manualUpdatedAt"`
	ManualUpdatedBy  int64     `json:"manualUpdatedBy"`
}

func (DisputeEvidence) TableName() string {
	return "dispute_evidence"
}
