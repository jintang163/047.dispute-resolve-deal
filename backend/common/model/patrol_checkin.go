package model

import "time"

type PatrolCheckin struct {
	BaseModel
	CheckinNo       string     `gorm:"size:32;uniqueIndex" json:"checkinNo"`
	TaskID          int64      `gorm:"index" json:"taskId"`
	PointID         int64      `gorm:"index" json:"pointId"`
	MemberID        int64      `gorm:"index;not null" json:"memberId"`
	MemberName      string     `gorm:"size:64" json:"memberName"`
	CheckinType     int        `gorm:"default:1" json:"checkinType"`
	CheckinTime     *time.Time `gorm:"default:CURRENT_TIMESTAMP;index" json:"checkinTime"`
	Longitude       float64    `json:"longitude"`
	Latitude        float64    `json:"latitude"`
	Address         string     `gorm:"size:256" json:"address"`
	LocationAccuracy float64   `json:"locationAccuracy"`
	PhotoURL        string     `gorm:"size:512" json:"photoUrl"`
	LivePhotoURL    string     `gorm:"size:512" json:"livePhotoUrl"`
	IsLiveVerified  int        `gorm:"default:0" json:"isLiveVerified"`
	LiveVerifyScore float64    `json:"liveVerifyScore"`
	CheckinDistance float64    `json:"checkinDistance"`
	IsValid         int        `gorm:"default:1;index" json:"isValid"`
	DeviceInfo      string     `gorm:"size:512" json:"deviceInfo"`
	IPAddress       string     `gorm:"size:64" json:"ipAddress"`
	Remark          string     `gorm:"size:256" json:"remark"`
}

func (PatrolCheckin) TableName() string {
	return "patrol_checkin"
}

type PatrolCheckinQuery struct {
	BaseQuery
	MemberID    int64  `form:"memberId" json:"memberId"`
	TaskID      int64  `form:"taskId" json:"taskId"`
	PointID     int64  `form:"pointId" json:"pointId"`
	CheckinType int    `form:"checkinType" json:"checkinType"`
	IsValid     int    `form:"isValid" json:"isValid"`
	CheckinNo   string `form:"checkinNo" json:"checkinNo"`
	DateRangeQuery
}

type PatrolCheckinRequest struct {
	TaskID          int64   `json:"taskId"`
	PointID         int64   `json:"pointId"`
	CheckinType     int     `json:"checkinType" binding:"required"`
	Longitude       float64 `json:"longitude" binding:"required"`
	Latitude        float64 `json:"latitude" binding:"required"`
	Address         string  `json:"address"`
	LocationAccuracy float64 `json:"locationAccuracy"`
	PhotoURL        string  `json:"photoUrl"`
	LivePhotoURL    string  `json:"livePhotoUrl"`
	IsLiveVerified  int     `json:"isLiveVerified"`
	LiveVerifyScore float64 `json:"liveVerifyScore"`
	DeviceInfo      string  `json:"deviceInfo"`
	Remark          string  `json:"remark"`
}

type LiveVerifyRequest struct {
	PhotoURL string `json:"photoUrl" binding:"required"`
	MemberID int64  `json:"memberId" binding:"required"`
}

type LiveVerifyResponse struct {
	IsVerified bool    `json:"isVerified"`
	Score      float64 `json:"score"`
	Message    string  `json:"message"`
}

type CheckinStatisticsResponse struct {
	TodayCheckinCount    int `json:"todayCheckinCount"`
	WeekCheckinCount     int `json:"weekCheckinCount"`
	MonthCheckinCount    int `json:"monthCheckinCount"`
	ContinuousCheckinDays int `json:"continuousCheckinDays"`
	TotalCheckinDays     int `json:"totalCheckinDays"`
}
