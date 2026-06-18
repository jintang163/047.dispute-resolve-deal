package model

import "time"

type VideoRoom struct {
	BaseModel
	CaseID       int64      `gorm:"index;not null" json:"caseId"`
	RoomNo       string     `gorm:"size:50;uniqueIndex;not null" json:"roomNo"`
	RoomName     string     `gorm:"size:100;not null" json:"roomName"`
	HostID       int64      `json:"hostId"`
	HostName     string     `gorm:"size:50" json:"hostName"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	Status       int32      `gorm:"default:0;index" json:"status"`
	Password     string     `gorm:"size:50" json:"password"`
	Token        string     `gorm:"size:500" json:"token"`
	RecordURL    string     `gorm:"size:500" json:"recordUrl"`
	Participants string     `gorm:"size:1000" json:"participants"`
	CreatedBy    int64      `json:"createdBy"`
}

func (VideoRoom) TableName() string {
	return "video_room"
}
