package model

type NotificationTemplate struct {
	BaseModel
	TemplateName    string `gorm:"size:100;not null" json:"templateName"`
	TemplateCode    string `gorm:"size:50;uniqueIndex;not null" json:"templateCode"`
	TemplateType    int    `gorm:"index;not null" json:"templateType"`
	TitleTemplate   string `gorm:"size:200" json:"titleTemplate"`
	ContentTemplate string `gorm:"type:text;not null" json:"contentTemplate"`
	Channel         int    `gorm:"index;not null" json:"channel"`
	Status          int32  `gorm:"default:1;index" json:"status"`
	CreatedBy       int64  `json:"createdBy"`
}

func (NotificationTemplate) TableName() string {
	return "notification_template"
}
