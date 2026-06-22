package model

import "time"

type HiddenDanger struct {
	BaseModel
	DangerNo        string     `gorm:"size:32;uniqueIndex" json:"dangerNo"`
	ReporterID      int64      `gorm:"index;not null" json:"reporterId"`
	ReporterName    string     `gorm:"size:64" json:"reporterName"`
	OrganizationID  int64      `gorm:"column:organization_id;index" json:"organizationId"`
	DangerType      int        `gorm:"default:1;index" json:"dangerType"`
	DangerLevel     int        `gorm:"default:3;index" json:"dangerLevel"`
	Title           string     `gorm:"size:256;not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Address         string     `gorm:"size:256" json:"address"`
	Longitude       float64    `json:"longitude"`
	Latitude        float64    `json:"latitude"`
	HappenTime      *time.Time `json:"happenTime"`
	InvolvedPerson  string     `gorm:"size:128" json:"involvedPerson"`
	InvolvedPhone   string     `gorm:"size:20" json:"involvedPhone"`
	PhotoURLs       string     `gorm:"type:text" json:"photoUrls"`
	VideoURL        string     `gorm:"size:512" json:"videoUrl"`
	IsDispute       int        `gorm:"default:0" json:"isDispute"`
	DisputeCaseID   int64      `json:"disputeCaseId"`
	HandleStatus    int32      `gorm:"default:10;index" json:"handleStatus"`
	HandlerID       int64      `json:"handlerId"`
	HandlerName     string     `gorm:"size:64" json:"handlerName"`
	HandleResult    string     `gorm:"type:text" json:"handleResult"`
	HandleTime      *time.Time `json:"handleTime"`
	PointsReward    int        `json:"pointsReward"`
}

func (HiddenDanger) TableName() string {
	return "hidden_danger"
}

type HiddenDangerQuery struct {
	BaseQuery
	ReporterID     int64  `form:"reporterId" json:"reporterId"`
	OrgID          int64  `form:"orgId" json:"orgId"`
	DangerType     int    `form:"dangerType" json:"dangerType"`
	DangerLevel    int    `form:"dangerLevel" json:"dangerLevel"`
	HandleStatus   int    `form:"handleStatus" json:"handleStatus"`
	DangerNo       string `form:"dangerNo" json:"dangerNo"`
	IsDispute      int    `form:"isDispute" json:"isDispute"`
	DateRangeQuery
}

type CreateHiddenDangerRequest struct {
	DangerType     int      `json:"dangerType" binding:"required"`
	DangerLevel    int      `json:"dangerLevel"`
	Title          string   `json:"title" binding:"required"`
	Description    string   `json:"description"`
	Address        string   `json:"address"`
	Longitude      float64  `json:"longitude"`
	Latitude       float64  `json:"latitude"`
	HappenTime     string   `json:"happenTime"`
	InvolvedPerson string   `json:"involvedPerson"`
	InvolvedPhone  string   `json:"involvedPhone"`
	PhotoURLs      []string `json:"photoUrls"`
	VideoURL       string   `json:"videoUrl"`
	IsDispute      int      `json:"isDispute"`
}

type HandleHiddenDangerRequest struct {
	ID           int64  `json:"id" binding:"required"`
	HandleStatus int32  `json:"handleStatus" binding:"required"`
	HandlerID    int64  `json:"handlerId"`
	HandlerName  string `json:"handlerName"`
	HandleResult string `json:"handleResult"`
}

type DangerStatisticsResponse struct {
	TotalCount      int `json:"totalCount"`
	PendingCount    int `json:"pendingCount"`
	ProcessingCount int `json:"processingCount"`
	CompletedCount  int `json:"completedCount"`
	ClosedCount     int `json:"closedCount"`
	DangerType1Count int `json:"dangerType1Count"`
	DangerType2Count int `json:"dangerType2Count"`
	DangerType3Count int `json:"dangerType3Count"`
	DangerType4Count int `json:"dangerType4Count"`
	DangerType5Count int `json:"dangerType5Count"`
}
