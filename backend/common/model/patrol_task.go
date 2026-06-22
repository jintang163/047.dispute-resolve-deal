package model

import "time"

type PatrolTask struct {
	BaseModel
	TaskNo              string     `gorm:"size:32;uniqueIndex" json:"taskNo"`
	Title               string     `gorm:"size:256;not null" json:"title"`
	Description         string     `gorm:"type:text" json:"description"`
	TaskType            int        `gorm:"default:1" json:"taskType"`
	Priority            int        `gorm:"default:3;index" json:"priority"`
	AssignerID          int64      `gorm:"index" json:"assignerId"`
	AssignerName        string     `gorm:"size:64" json:"assignerName"`
	AssigneeID          int64      `gorm:"index" json:"assigneeId"`
	AssigneeName        string     `gorm:"size:64" json:"assigneeName"`
	OrganizationID      int64      `gorm:"column:organization_id;index" json:"organizationId"`
	PlanStartTime       *time.Time `json:"planStartTime"`
	PlanEndTime         *time.Time `json:"planEndTime"`
	ActualStartTime     *time.Time `json:"actualStartTime"`
	ActualEndTime       *time.Time `json:"actualEndTime"`
	PointCount          int        `json:"pointCount"`
	CompletedPointCount int        `json:"completedPointCount"`
	Status              int32      `gorm:"default:10;index" json:"status"`
	PointsReward        int        `json:"pointsReward"`
	IsDeleted           int        `gorm:"default:0" json:"isDeleted"`

	Points []PatrolTaskPoint `gorm:"foreignKey:TaskID" json:"points,omitempty"`
}

func (PatrolTask) TableName() string {
	return "patrol_task"
}

type PatrolTaskQuery struct {
	BaseQuery
	Status         int    `form:"status" json:"status"`
	TaskType       int    `form:"taskType" json:"taskType"`
	Priority       int    `form:"priority" json:"priority"`
	OrgID          int64  `form:"orgId" json:"orgId"`
	AssigneeID     int64  `form:"assigneeId" json:"assigneeId"`
	AssignerID     int64  `form:"assignerId" json:"assignerId"`
	TaskNo         string `form:"taskNo" json:"taskNo"`
	DateRangeQuery
}

type CreatePatrolTaskRequest struct {
	Title         string              `json:"title" binding:"required"`
	Description   string              `json:"description"`
	TaskType      int                 `json:"taskType" binding:"required"`
	Priority      int                 `json:"priority"`
	AssigneeID    int64               `json:"assigneeId" binding:"required"`
	AssigneeName  string              `json:"assigneeName"`
	PlanStartTime string              `json:"planStartTime"`
	PlanEndTime   string              `json:"planEndTime"`
	PointsReward  int                 `json:"pointsReward"`
	Points        []TaskPointRequest  `json:"points" binding:"required,min=1"`
}

type TaskPointRequest struct {
	PointName     string  `json:"pointName" binding:"required"`
	PointType     string  `json:"pointType"`
	Address       string  `json:"address"`
	Longitude     float64 `json:"longitude" binding:"required"`
	Latitude      float64 `json:"latitude" binding:"required"`
	ContactPerson string  `json:"contactPerson"`
	ContactPhone  string  `json:"contactPhone"`
	CheckinRadius int     `json:"checkinRadius"`
	Remark        string  `json:"remark"`
}

type UpdatePatrolTaskRequest struct {
	ID              int64              `json:"id" binding:"required"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	TaskType        int                `json:"taskType"`
	Priority        int                `json:"priority"`
	AssigneeID      int64              `json:"assigneeId"`
	AssigneeName    string             `json:"assigneeName"`
	PlanStartTime   string             `json:"planStartTime"`
	PlanEndTime     string             `json:"planEndTime"`
	PointsReward    int                `json:"pointsReward"`
	Points          []TaskPointRequest `json:"points"`
}

type PatrolTaskPoint struct {
	BaseModel
	TaskID        int64      `gorm:"index" json:"taskId"`
	PointName     string     `gorm:"size:128;not null" json:"pointName"`
	PointType     string     `gorm:"size:64" json:"pointType"`
	Address       string     `gorm:"size:256" json:"address"`
	Longitude     float64    `json:"longitude"`
	Latitude      float64    `json:"latitude"`
	ContactPerson string     `gorm:"size:64" json:"contactPerson"`
	ContactPhone  string     `gorm:"size:20" json:"contactPhone"`
	CheckinRadius int        `gorm:"default:100" json:"checkinRadius"`
	SortOrder     int        `gorm:"default:0;index" json:"sortOrder"`
	IsChecked     int        `gorm:"default:0;index" json:"isChecked"`
	CheckinTime   *time.Time `json:"checkinTime"`
	Remark        string     `gorm:"size:512" json:"remark"`
}

func (PatrolTaskPoint) TableName() string {
	return "patrol_task_point"
}

type PatrolRoutePlanRequest struct {
	StartLongitude float64             `json:"startLongitude" binding:"required"`
	StartLatitude  float64             `json:"startLatitude" binding:"required"`
	Points         []RoutePointRequest `json:"points" binding:"required,min=1"`
	Strategy       int                 `json:"strategy"`
}

type RoutePointRequest struct {
	ID        int64   `json:"id"`
	Longitude float64 `json:"longitude" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
}

type RoutePlanResponse struct {
	TotalDistance float64       `json:"totalDistance"`
	TotalDuration int           `json:"totalDuration"`
	OrderedPoints []OrderedPoint `json:"orderedPoints"`
	RoutePolyline string        `json:"routePolyline"`
}

type OrderedPoint struct {
	ID        int64   `json:"id"`
	SortOrder int     `json:"sortOrder"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Distance  float64 `json:"distance"`
	Duration  int     `json:"duration"`
}
