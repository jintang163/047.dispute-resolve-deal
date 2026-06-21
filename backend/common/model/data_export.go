package model

import "time"

type DataExportLog struct {
	BaseModel
	ExportNo           string    `gorm:"size:50;uniqueIndex" json:"exportNo"`
	ExportType         int       `gorm:"index" json:"exportType"`
	ExportName         string    `gorm:"size:200" json:"exportName"`
	FilterConditions   string    `gorm:"type:text" json:"filterConditions"`
	RecordCount        int       `json:"recordCount"`
	FileName           string    `gorm:"size:255" json:"fileName"`
	FilePath           string    `gorm:"size:500" json:"filePath"`
	FileSize           int64     `json:"fileSize"`
	EncryptionAlgorithm string   `gorm:"size:50" json:"encryptionAlgorithm"`
	PasswordSmsSent    int       `json:"passwordSmsSent"`
	PasswordSmsTime    *time.Time `json:"passwordSmsTime"`
	ExportStatus       int       `gorm:"index" json:"exportStatus"`
	ErrorMessage       string    `gorm:"type:text" json:"errorMessage"`
	OperatorID         int64     `gorm:"index" json:"operatorId"`
	OperatorName       string    `gorm:"size:50" json:"operatorName"`
	OperatorPhone      string    `gorm:"size:20" json:"operatorPhone"`
	OrgID              int64     `gorm:"index" json:"orgId"`
	IPAddress          string    `gorm:"size:50" json:"ipAddress"`
	UserAgent          string    `gorm:"size:500" json:"userAgent"`
	CompletedAt        *time.Time `json:"completedAt"`
	ExpiredAt          time.Time `gorm:"index" json:"expiredAt"`
}

func (DataExportLog) TableName() string {
	return "data_export_log"
}
