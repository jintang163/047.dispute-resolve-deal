package model

type WorkflowDefinition struct {
	BaseModel
	Name           string `gorm:"size:100;not null" json:"name"`
	Code           string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	DisputeTypeID  int64  `gorm:"index" json:"disputeTypeId"`
	Description    string `gorm:"size:500" json:"description"`
	FlowConfig     string `gorm:"type:text" json:"flowConfig"`
	Version        int    `gorm:"default:1" json:"version"`
	Status         int32  `gorm:"default:1;index" json:"status"`
	CreatedBy      int64  `json:"createdBy"`
}

func (WorkflowDefinition) TableName() string {
	return "workflow_definition"
}
