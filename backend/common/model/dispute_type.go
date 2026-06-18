package model

type DisputeType struct {
	BaseModel
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	ParentID    int64  `gorm:"index" json:"parentId"`
	ParentPath  string `gorm:"size:500" json:"parentPath"`
	Level       int    `gorm:"not null" json:"level"`
	SortOrder   int    `gorm:"default:0" json:"sortOrder"`
	Description string `gorm:"size:500" json:"description"`
	MediationDays int  `gorm:"default:30" json:"mediationDays"`
	WarningDays int    `gorm:"default:7" json:"warningDays"`
	Status      int32  `gorm:"default:1;index" json:"status"`
	CreatedBy   int64  `json:"createdBy"`
	UpdatedBy   int64  `json:"updatedBy"`
}

func (DisputeType) TableName() string {
	return "dispute_type"
}
