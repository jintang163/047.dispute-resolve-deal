package model

import "time"

type GridMember struct {
	BaseModel
	UserID         int64      `gorm:"column:user_id;index" json:"userId"`
	MemberNo       string     `gorm:"size:32;uniqueIndex" json:"memberNo"`
	RealName       string     `gorm:"size:64;not null" json:"realName"`
	Phone          string     `gorm:"size:20;not null" json:"phone"`
	IDCard         string     `gorm:"size:32;column:id_card" json:"idCard"`
	Gender         int        `json:"gender"`
	Avatar         string     `gorm:"size:256" json:"avatar"`
	OrganizationID int64      `gorm:"column:organization_id;index" json:"organizationId"`
	OrganizationName string   `gorm:"column:organization_name;size:128" json:"organizationName"`
	GridArea       string     `gorm:"size:256" json:"gridArea"`
	GridCode       string     `gorm:"size:64" json:"gridCode"`
	WorkAddress    string     `gorm:"size:256" json:"workAddress"`
	Longitude      float64    `json:"longitude"`
	Latitude       float64    `json:"latitude"`
	EntryDate      *time.Time `json:"entryDate"`
	Points         int        `json:"points"`
	TotalPoints    int        `json:"totalPoints"`
	TaskCount      int        `json:"taskCount"`
	VisitCount     int        `json:"visitCount"`
	DangerCount    int        `json:"dangerCount"`
	Status         int32      `gorm:"default:1;index" json:"status"`
}

func (GridMember) TableName() string {
	return "grid_member"
}

type GridMemberQuery struct {
	BaseQuery
	Status     int    `form:"status" json:"status"`
	OrgID      int64  `form:"orgId" json:"orgId"`
	MemberNo   string `form:"memberNo" json:"memberNo"`
	RealName   string `form:"realName" json:"realName"`
	Phone      string `form:"phone" json:"phone"`
}

type CreateGridMemberRequest struct {
	UserID         int64  `json:"userId" binding:"required"`
	MemberNo       string `json:"memberNo" binding:"required"`
	RealName       string `json:"realName" binding:"required"`
	Phone          string `json:"phone" binding:"required"`
	IDCard         string `json:"idCard"`
	Gender         int    `json:"gender"`
	OrganizationID int64  `json:"organizationId" binding:"required"`
	OrganizationName string `json:"organizationName"`
	GridArea       string `json:"gridArea"`
	GridCode       string `json:"gridCode"`
	WorkAddress    string `json:"workAddress"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	EntryDate      string `json:"entryDate"`
	Status         int32  `json:"status"`
}

type UpdateGridMemberRequest struct {
	ID             int64  `json:"id" binding:"required"`
	RealName       string `json:"realName"`
	Phone          string `json:"phone"`
	IDCard         string `json:"idCard"`
	Gender         int    `json:"gender"`
	Avatar         string `json:"avatar"`
	OrganizationID int64  `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	GridArea       string `json:"gridArea"`
	GridCode       string `json:"gridCode"`
	WorkAddress    string `json:"workAddress"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	EntryDate      string `json:"entryDate"`
	Status         int32  `json:"status"`
}
