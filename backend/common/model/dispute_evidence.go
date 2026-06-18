package model

import "time"

type DisputeEvidence struct {
	BaseModel
	CaseID       int64      `gorm:"index;not null" json:"caseId"`
	UploaderID   int64      `json:"uploaderId"`
	UploaderName string     `gorm:"size:50" json:"uploaderName"`
	UploadedBy   int64      `json:"uploadedBy"`
	FileName     string     `gorm:"size:255;not null" json:"fileName"`
	FileType     string     `gorm:"size:50;index" json:"fileType"`
	FileSize     int64      `json:"fileSize"`
	FileURL      string     `gorm:"size:500;not null" json:"fileUrl"`
	ThumbnailURL string     `gorm:"size:500" json:"thumbnailUrl"`
	OSSPath      string     `gorm:"size:500" json:"ossPath"`
	Description  string     `gorm:"size:500" json:"description"`
	IsVerified   int32      `gorm:"default:0;index" json:"isVerified"`
	VerifiedBy   int64      `json:"verifiedBy"`
	VerifiedAt   *time.Time `json:"verifiedAt"`
	SortOrder    int        `json:"sortOrder"`
	UploadFrom   int        `json:"uploadFrom"`
	Remark       string     `gorm:"size:500" json:"remark"`
}

func (DisputeEvidence) TableName() string {
	return "dispute_evidence"
}
