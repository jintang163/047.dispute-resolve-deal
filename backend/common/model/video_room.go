package model

import "time"

type VideoRoom struct {
	BaseModel
	CaseID           int64      `gorm:"index;not null" json:"caseId"`
	RoomNo           string     `gorm:"size:50;uniqueIndex;not null" json:"roomNo"`
	RoomName         string     `gorm:"size:100;not null" json:"roomName"`
	HostID           int64      `json:"hostId"`
	HostName         string     `gorm:"size:50" json:"hostName"`
	StartTime        *time.Time `json:"startTime"`
	EndTime          *time.Time `json:"endTime"`
	Status           int32      `gorm:"default:0;index" json:"status"`
	Password         string     `gorm:"size:50" json:"password"`
	Token            string     `gorm:"size:500" json:"token"`
	RecordURL        string     `gorm:"size:500" json:"recordUrl"`
	RecordTaskID     string     `gorm:"size:128" json:"recordTaskId"`
	RecordStatus     int32      `gorm:"default:0" json:"recordStatus"`
	HasMeetingMinutes int32     `gorm:"default:0" json:"hasMeetingMinutes"`
	TRTCRoomID       int32      `json:"trtcRoomId"`
	ScreenShareUserID int64     `json:"screenShareUserId"`
	VirtualBGEnabled int32      `gorm:"default:0" json:"virtualBgEnabled"`
	BeautyEnabled    int32      `gorm:"default:0" json:"beautyEnabled"`
	Participants     string     `gorm:"size:1000" json:"participants"`
	CreatedBy        int64      `json:"createdBy"`
}

func (VideoRoom) TableName() string {
	return "video_room"
}

type VideoRecordSegment struct {
	BaseModel
	RoomID       int64     `gorm:"index;not null" json:"roomId"`
	CaseID       int64     `json:"caseId"`
	CaseNo       string    `gorm:"size:32" json:"caseNo"`
	TaskID       string    `gorm:"size:128;index" json:"taskId"`
	SegmentIndex int       `json:"segmentIndex"`
	SegmentSec   int       `gorm:"default:600" json:"segmentSec"`
	Status       int32     `gorm:"default:1;index" json:"status"`
	FileURL      string    `gorm:"size:512" json:"fileUrl"`
	FileSize     int64     `json:"fileSize"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	StartTimeMs  int64     `json:"startTimeMs"`
	EndTimeMs    int64     `json:"endTimeMs"`
	DurationSec  int       `json:"durationSec"`
	StoragePath  string    `gorm:"size:512" json:"storagePath"`
}

func (VideoRecordSegment) TableName() string {
	return "video_record_segment"
}

type VideoQueue struct {
	BaseModel
	CaseID         int64     `gorm:"index;not null" json:"caseId"`
	CaseNo         string    `gorm:"size:32" json:"caseNo"`
	MediatorID     int64     `gorm:"index" json:"mediatorId"`
	MediatorName   string    `gorm:"size:64" json:"mediatorName"`
	PartyName      string    `gorm:"size:64" json:"partyName"`
	PartyPhone     string    `gorm:"size:20" json:"partyPhone"`
	PartyUserID    int64     `json:"partyUserId"`
	Priority       int       `gorm:"default:3" json:"priority"`
	Status         int32     `gorm:"default:1;index" json:"status"`
	EnqueueTime    *time.Time `json:"enqueueTime"`
	DequeueTime    *time.Time `json:"dequeueTime"`
	NotifyCount    int       `gorm:"default:0" json:"notifyCount"`
	LastNotifyTime *time.Time `json:"lastNotifyTime"`
	Remark         string    `gorm:"size:512" json:"remark"`
}

func (VideoQueue) TableName() string {
	return "video_queue"
}

type VideoMeetingMinutes struct {
	BaseModel
	RoomID           int64     `gorm:"index;not null" json:"roomId"`
	CaseID           int64     `gorm:"index;not null" json:"caseId"`
	CaseNo           string    `gorm:"size:32" json:"caseNo"`
	MeetingTitle     string    `gorm:"size:256" json:"meetingTitle"`
	MeetingTime      *time.Time `json:"meetingTime"`
	Duration         string    `gorm:"size:64" json:"duration"`
	Participants     string    `gorm:"type:json" json:"participants"`
	Summary          string    `gorm:"type:text" json:"summary"`
	KeyPoints        string    `gorm:"type:json" json:"keyPoints"`
	DisputeFocus     string    `gorm:"type:json" json:"disputeFocus"`
	MediationProcess string    `gorm:"type:text" json:"mediationProcess"`
	EvidenceDiscussed string   `gorm:"type:json" json:"evidenceDiscussed"`
	Agreement        string    `gorm:"type:text" json:"agreement"`
	NextSteps        string    `gorm:"type:json" json:"nextSteps"`
	RiskPoints       string    `gorm:"type:json" json:"riskPoints"`
	EmotionalState   string    `gorm:"type:text" json:"emotionalState"`
	MediatorAdvice   string    `gorm:"type:text" json:"mediatorAdvice"`
	Transcript       string    `gorm:"type:text" json:"transcript"`
	AIModel          string    `gorm:"size:64;default:'deepseek'" json:"aiModel"`
	TokensUsed       int       `json:"tokensUsed"`
	CostTime         int       `json:"costTime"`
	IsApproved       int32     `gorm:"default:0" json:"isApproved"`
	ApprovedBy       int64     `json:"approvedBy"`
	ApprovedAt       *time.Time `json:"approvedAt"`
	Status           int32     `gorm:"default:1;index" json:"status"`
}

func (VideoMeetingMinutes) TableName() string {
	return "video_meeting_minutes"
}
