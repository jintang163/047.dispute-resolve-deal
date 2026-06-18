package model

import "time"

type AIConsultRecord struct {
	BaseModel
	UserID          int64     `gorm:"index;not null" json:"userId"`
	Question        string    `gorm:"type:text;not null" json:"question"`
	Answer          string    `gorm:"type:text;not null" json:"answer"`
	RelatedArticles string    `gorm:"type:text" json:"relatedArticles"`
	DurationMs      int64     `json:"durationMs"`
	TokensUsed      int64     `json:"tokensUsed"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (AIConsultRecord) TableName() string {
	return "ai_consult_record"
}
