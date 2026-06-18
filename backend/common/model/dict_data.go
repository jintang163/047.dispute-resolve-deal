package model

type DictData struct {
	BaseModel
	DictTypeID int64  `gorm:"index;not null" json:"dictTypeId"`
	DictLabel  string `gorm:"size:100;not null" json:"dictLabel"`
	DictValue  string `gorm:"size:100;not null" json:"dictValue"`
	DictSort   int    `gorm:"default:0" json:"dictSort"`
	Status     int32  `gorm:"default:1;index" json:"status"`
	Remark     string `gorm:"size:500" json:"remark"`
	CreatedBy  int64  `json:"createdBy"`
}

func (DictData) TableName() string {
	return "dict_data"
}
