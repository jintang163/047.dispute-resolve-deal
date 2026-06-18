package model

import "time"

type NotificationRecord struct {
	BaseModel
	ReceiverID   int64      `gorm:"index;not null" json:"receiverId"`
	ReceiverName string     `gorm:"size:50;not null" json:"receiverName"`
	TemplateID   int64      `gorm:"index" json:"templateId"`
	TemplateName string     `gorm:"size:100" json:"templateName"`
	TemplateType int        `gorm:"index" json:"templateType"`
	Title        string     `gorm:"size:200;not null" json:"title"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	Channel      int        `gorm:"index;not null" json:"channel"`
	Status       int32      `gorm:"default:0;index" json:"status"`
	ReadTime     *time.Time `json:"readTime"`
	Params       string     `gorm:"type:text" json:"params"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (NotificationRecord) TableName() string {
	return "notification_record"
}
