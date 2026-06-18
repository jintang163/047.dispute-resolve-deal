package model

type Organization struct {
	BaseModel
	Name         string  `gorm:"size:100;not null" json:"name"`
	Code         string  `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Type         int     `gorm:"not null;index" json:"type"`
	ParentID     int64   `gorm:"index" json:"parentId"`
	ParentPath   string  `gorm:"size:500" json:"parentPath"`
	SortOrder    int     `gorm:"default:0" json:"sortOrder"`
	LeaderID     int64   `json:"leaderId"`
	LeaderName   string  `gorm:"size:50" json:"leaderName"`
	Leader       string  `gorm:"size:50" json:"leader"`
	Contact      string  `gorm:"size:50" json:"contact"`
	Phone        string  `gorm:"size:20" json:"phone"`
	Address      string  `gorm:"size:255" json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	Description  string  `gorm:"size:500" json:"description"`
	Status       int32   `gorm:"default:1;index" json:"status"`
	CreatedBy    int64   `json:"createdBy"`
	UpdatedBy    int64   `json:"updatedBy"`
}

func (Organization) TableName() string {
	return "sys_organization"
}
