package model

import "time"

type MediationRecord struct {
	BaseModel
	CaseID            int64      `gorm:"index;not null" json:"caseId"`
	MediatorID        int64      `json:"mediatorId"`
	MediatorName      string     `gorm:"size:50" json:"mediatorName"`
	RecordType        int        `gorm:"index;not null" json:"recordType"`
	MediationDate     *time.Time `json:"mediationDate"`
	StartTime         string     `gorm:"size:10" json:"startTime"`
	EndTime           string     `gorm:"size:10" json:"endTime"`
	Duration          int        `json:"duration"`
	Location          string     `gorm:"size:255" json:"location"`
	Content           string     `gorm:"size:2000" json:"content"`
	Participant       string     `gorm:"size:500" json:"participant"`
	Result            string     `gorm:"size:500" json:"result"`
	NextMediationDate *time.Time `json:"nextMediationDate"`
	Remark            string     `gorm:"size:500" json:"remark"`
	CreatedBy         int64      `json:"createdBy"`
	AudioUrl          string     `gorm:"size:500" json:"audioUrl"`
	AudioDuration     int        `json:"audioDuration"`
	AudioFileSize     int64      `json:"audioFileSize"`
	TranscriptText    string     `gorm:"type:text" json:"transcriptText"`
	TranscribeStatus  int        `gorm:"default:0;index" json:"transcribeStatus"`
	TranscribeTaskId  string     `gorm:"size:128" json:"transcribeTaskId"`
	TranscribeAt      *time.Time `json:"transcribeAt"`
}

func (MediationRecord) TableName() string {
	return "mediation_record"
}
