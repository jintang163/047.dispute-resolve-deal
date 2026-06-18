package model

type DictType struct {
	BaseModel
	DictName    string `gorm:"size:100;not null" json:"dictName"`
	DictCode    string `gorm:"size:50;uniqueIndex;not null" json:"dictCode"`
	Description string `gorm:"size:500" json:"description"`
	Status      int32  `gorm:"default:1;index" json:"status"`
	CreatedBy   int64  `json:"createdBy"`
}

func (DictType) TableName() string {
	return "dict_type"
}
