package model

import "time"

type CallbackTemplate struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	Code         string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Type         int        `gorm:"default:1" json:"type"`
	WelcomeText  string     `gorm:"size:500" json:"welcomeText"`
	QuestionFlow string     `gorm:"type:json" json:"questionFlow"`
	EndText      string     `gorm:"size:500" json:"endText"`
	VoiceType    string     `gorm:"size:50;default:xiaoyun" json:"voiceType"`
	Speed        int        `gorm:"default:0" json:"speed"`
	Volume       int        `gorm:"default:0" json:"volume"`
	PauseTime    int        `gorm:"default:800" json:"pauseTime"`
	IsDefault    int        `gorm:"default:0" json:"isDefault"`
	Status       int        `gorm:"default:1" json:"status"`
	CreatedBy    int64      `json:"createdBy"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}

func (CallbackTemplate) TableName() string {
	return "callback_template"
}
