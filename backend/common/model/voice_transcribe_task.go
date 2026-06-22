package model

import "time"

const (
	TranscribeStatusPending    = 0
	TranscribeStatusQueuing    = 1
	TranscribeStatusProcessing = 2
	TranscribeStatusCompleted  = 3
	TranscribeStatusFailed     = 4
	TranscribeStatusCanceled   = 5
)

var TranscribeStatusMap = map[int]string{
	0: "待提交",
	1: "排队中",
	2: "处理中",
	3: "已完成",
	4: "失败",
	5: "已取消",
}

type Sentence struct {
	Text      string `json:"text"`
	BeginTime int    `json:"beginTime"`
	EndTime   int    `json:"endTime"`
	SpeakerID int    `json:"speakerId"`
}

type VoiceTranscribeTask struct {
	BaseModel
	TaskID            string     `gorm:"size:128;uniqueIndex" json:"taskId"`
	TaskType          int        `gorm:"default:1" json:"taskType"`
	Status            int        `gorm:"default:0;index" json:"status"`
	FileName          string     `gorm:"size:255" json:"fileName"`
	FileURL           string     `gorm:"size:500;column:file_url" json:"fileUrl"`
	FileSize          int64      `json:"fileSize"`
	Format            string     `gorm:"size:20" json:"format"`
	Duration          int        `json:"duration"`
	TranscriptText    string     `gorm:"type:text" json:"transcriptText"`
	SpeakerCount      int        `gorm:"default:0" json:"speakerCount"`
	EnableDiarization bool       `gorm:"default:false" json:"enableDiarization"`
	Sentences         string     `gorm:"type:json" json:"sentences"`
	WordCount         int        `json:"wordCount"`
	ErrorMessage      string     `gorm:"size:500;column:error_msg" json:"errorMessage"`
	CaseID            int64      `gorm:"index" json:"caseId"`
	RecordID          int64      `gorm:"index" json:"recordId"`
	CreatedBy         int64      `gorm:"column:created_by" json:"createdBy"`
	RequestID         string     `gorm:"size:128" json:"requestId"`
	CompletedAt       *time.Time `json:"completedAt"`
	IsDeleted         int        `gorm:"default:0" json:"isDeleted"`
}

func (VoiceTranscribeTask) TableName() string {
	return "voice_transcribe_task"
}

type VoiceTranscribeTaskQuery struct {
	BaseQuery
	TaskID   string `form:"taskId" json:"taskId"`
	TaskType int    `form:"taskType" json:"taskType"`
	Status   int    `form:"status" json:"status"`
	CaseID   int64  `form:"caseId" json:"caseId"`
	RecordID int64  `form:"recordId" json:"recordId"`
	DateRangeQuery
}

type CreateVoiceTranscribeTaskRequest struct {
	TaskType  int    `json:"taskType" binding:"required"`
	FileName  string `json:"fileName"`
	FileURL   string `json:"fileUrl" binding:"required"`
	FileSize  int64  `json:"fileSize"`
	Format    string `json:"format"`
	Duration  int    `json:"duration"`
	CaseID    int64  `json:"caseId"`
	RecordID  int64  `json:"recordId"`
}

type UpdateVoiceTranscribeTaskRequest struct {
	ID             int64  `json:"id" binding:"required"`
	Status         int    `json:"status"`
	TranscriptText string `json:"transcriptText"`
	SpeakerCount   int    `json:"speakerCount"`
	Diarization    string `json:"diarization"`
	WordCount      int    `json:"wordCount"`
	ErrorMsg       string `json:"errorMsg"`
}
