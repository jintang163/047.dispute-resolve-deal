package model

import "time"

type CallbackRecord struct {
	BaseModel
	CaseID           int64      `gorm:"index;not null" json:"caseId"`
	CaseNo           string     `gorm:"size:50;index" json:"caseNo"`
	CaseTitle        string     `gorm:"size:200" json:"caseTitle"`
	ApplicantID      int64      `gorm:"index" json:"applicantId"`
	ApplicantName    string     `gorm:"size:50" json:"applicantName"`
	ApplicantPhone   string     `gorm:"size:20;index" json:"applicantPhone"`
	TaskID           string     `gorm:"size:100;index" json:"taskId"`
	CallID           string     `gorm:"size:100;index" json:"callId"`
	Status           int32      `gorm:"default:10;index" json:"status"`
	CallStatus       int32      `gorm:"index" json:"callStatus"`
	CallTime         *time.Time `json:"callTime"`
	CallDuration     int        `json:"callDuration"`
	RetryCount       int        `gorm:"default:0" json:"retryCount"`
	MaxRetryCount    int        `gorm:"default:3" json:"maxRetryCount"`
	NextRetryTime    *time.Time `json:"nextRetryTime"`
	ScheduledTime    *time.Time `json:"scheduledTime"`
	TranscriptText   string     `gorm:"type:text" json:"transcriptText"`
	SentimentResult  string     `gorm:"type:text" json:"sentimentResult"`
	SentimentScore   float64    `json:"sentimentScore"`
	Emotion          string     `gorm:"size:20" json:"emotion"`
	PerformanceScore int        `json:"performanceScore"`
	SatisfactionScore int       `json:"satisfactionScore"`
	RecordingURL     string     `gorm:"size:500" json:"recordingUrl"`
	RecordingSize    int64      `json:"recordingSize"`
	ExpireAt         *time.Time `json:"expireAt"`
	ResultData       string     `gorm:"type:json" json:"resultData"`
	Remark           string     `gorm:"size:500" json:"remark"`
}

func (CallbackRecord) TableName() string {
	return "callback_record"
}

const (
	CallbackStatusPending    = 10
	CallbackStatusCalling    = 20
	CallbackStatusSuccess    = 30
	CallbackStatusFailed     = 40
	CallbackStatusCancelled  = 99
)

const (
	CallStatusNotCalled      = 0
	CallStatusRinging        = 10
	CallStatusAnswered       = 20
	CallStatusNoAnswer       = 30
	CallStatusBusy           = 40
	CallStatusFailed         = 50
	CallStatusHangup         = 60
)

const (
	EmotionPositive = "positive"
	EmotionNeutral  = "neutral"
	EmotionNegative = "negative"
)

var CallbackStatusMap = map[int]string{
	CallbackStatusPending:   "待回访",
	CallbackStatusCalling:   "回访中",
	CallbackStatusSuccess:   "回访成功",
	CallbackStatusFailed:    "回访失败",
	CallbackStatusCancelled: "已取消",
}

var CallStatusMap = map[int]string{
	CallStatusNotCalled: "未呼叫",
	CallStatusRinging:   "振铃中",
	CallStatusAnswered:  "已接听",
	CallStatusNoAnswer:  "无人接听",
	CallStatusBusy:      "用户忙",
	CallStatusFailed:    "呼叫失败",
	CallStatusHangup:    "已挂断",
}

var EmotionMap = map[string]string{
	EmotionPositive: "正面",
	EmotionNeutral:  "中性",
	EmotionNegative: "负面",
}
